package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	mongoadapter "github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/mongodb"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestAuditBrowserIsolated(t *testing.T) {
	if os.Getenv("FRONTEND_AUDIT") != "1" {
		t.Skip("Navegador E: requiere FRONTEND_AUDIT=1, Node/Playwright y bases aisladas")
	}
	dsn, uri, name := os.Getenv("INTEGRATION_DATABASE_URL"), os.Getenv("INTEGRATION_MONGODB_URI"), os.Getenv("INTEGRATION_MONGODB_DATABASE")
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_test" || uri == "" || name != "elbravodeljac_tecnico_test" {
		t.Fatal("E2E requiere exclusivamente las dos bases de prueba")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	trace := &auditNavigationTrace{}
	cfg.ConnConfig.Tracer = trace
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	mg, err := mongoadapter.Open(uri, name)
	if err != nil {
		t.Fatal(err)
	}
	defer mg.Close(context.Background())
	if err = mg.Init(ctx); err != nil {
		t.Fatal(err)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	origin := "http://" + ln.Addr().String()
	key := make([]byte, 32)
	if _, err = rand.Read(key); err != nil {
		t.Fatal(err)
	}
	handler, err := httpapi.New(&postgres.Store{Pool: pool}, httpapi.Options{Origin: origin, Key: key, Technical: mg, FrontendDir: filepath.Join("..", "..", "dist")})
	if err != nil {
		t.Fatal(err)
	}
	api := httptest.NewUnstartedServer(handler)
	api.Listener = ln
	api.Start()
	defer api.Close()
	h := workshopHarness{pool: pool, api: api, ctx: ctx, password: rand.Text(), suffix: rand.Text()[:10], origin: origin}
	all := []string{}
	for module, actions := range domain.PermissionPolicy().Modules {
		for _, action := range actions {
			all = append(all, module+"."+action)
		}
	}
	_, _ = h.user(t, "auditbrowser", "AuditBrowser_"+h.suffix, all)
	// Explicit access-role fixture: a clean isolated database need not already
	// contain the Supervisor role used to prove independence from job titles.
	h.exec(t, `INSERT INTO roles(nombre_rol) VALUES('Administrador') ON CONFLICT DO NOTHING`)
	h.exec(t, `INSERT INTO rol_permisos(id_rol,id_permiso) SELECT r.id_rol,p.id_permiso FROM roles r CROSS JOIN permisos p WHERE r.nombre_rol='Administrador' AND p.codigo='ordenes.ver' ON CONFLICT DO NOTHING`)
	mechanic, _ := h.user(t, "auditworker", "Mecánico", []string{"inicio.ver", "recepciones.ver", "diagnosticos.ver", "diagnosticos.crear", "ordenes.ver", "ordenes.editar", "vehiculos.ver"})
	fixture := map[string]string{"url": origin, "username": "auditbrowser_" + h.suffix, "password": h.password, "workerUsername": "auditworker_" + h.suffix, "mechanicId": mechanic, "suffix": h.suffix}
	input, _ := json.Marshal(fixture)
	cmd := exec.CommandContext(ctx, "node", filepath.Join("..", "..", "checks", "audit-browser.mjs"))
	cmd.Stdin = bytes.NewReader(input)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Auditoría Chrome: %v %s", err, out)
	}
	if trace.unexpected.Load() > 0 {
		t.Fatal("Errores SQL inesperados durante el navegador:", trace.unexpected.Load())
	}
	t.Log("Consultas canceladas por navegación:", trace.canceled.Load())
	t.Log(string(out))
}

type auditNavigationTrace struct {
	canceled   atomic.Int64
	unexpected atomic.Int64
}

func (t *auditNavigationTrace) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	return ctx
}
func (t *auditNavigationTrace) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	if data.Err != nil {
		if ctx.Err() != nil {
			t.canceled.Add(1)
		} else {
			t.unexpected.Add(1)
		}
	}
}
