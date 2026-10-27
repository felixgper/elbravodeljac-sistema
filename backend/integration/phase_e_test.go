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
	"github.com/jackc/pgx/v5/pgxpool"
	"net"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPhaseEFrontendBrowserPostgreSQLMongoDB(t *testing.T) {
	if os.Getenv("FRONTEND_E2E") != "1" {
		t.Skip("Navegador E: requiere FRONTEND_E2E=1, Node/Playwright y bases aisladas")
	}
	dsn, uri, name := os.Getenv("INTEGRATION_DATABASE_URL"), os.Getenv("INTEGRATION_MONGODB_URI"), os.Getenv("INTEGRATION_MONGODB_DATABASE")
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_test" || uri == "" || name != "elbravodeljac_tecnico_test" {
		t.Fatal("E2E requiere exclusivamente las dos bases de prueba")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
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
	for m, actions := range domain.PermissionPolicy().Modules {
		for _, a := range actions {
			all = append(all, m+"."+a)
		}
	}
	adminID, admin := h.user(t, "emanager", "EAdmin_"+h.suffix, all)
	mechanicID, _ := h.user(t, "emechanic", "Mecánico", []string{"ordenes.ver", "ordenes.editar"})
	limitedID, _ := h.user(t, "elimited", "ELimited_"+h.suffix, []string{"ordenes.ver", "ordenes.editar"})
	serviceID := responseID(t, h.call(t, admin, "POST", "/servicios", map[string]string{"name": "Servicio E " + h.suffix, "referencePrice": "100.00"}, 201))
	partID := responseID(t, h.call(t, admin, "POST", "/repuestos", map[string]string{"name": "Repuesto E " + h.suffix, "referencePrice": "20.00", "stock": "10"}, 201))
	unitID := responseID(t, h.call(t, admin, "POST", "/unidades-auxilio", map[string]string{"code": "E-" + h.suffix, "description": "Unidad de pruebas"}, 201))
	// No form needs permission on an unrelated directory. Auxiliary projections
	// remain authenticated, module-bound and respect assignment scope.
	h.call(t, admin, "GET", "/frontend/opciones?module=pagos&resource=usuarios", nil, 403)
	h.call(t, admin, "GET", "/frontend/opciones?module=ordenes&resource=clientes", nil, 200)
	for _, resource := range []string{"clientes", "vehiculos", "recepciones", "diagnosticos", "cotizaciones", "mecanicos"} {
		h.call(t, admin, "GET", "/frontend/opciones?module=ordenes&resource="+resource, nil, 200)
	}
	h.call(t, admin, "GET", "/frontend/opciones?module=mecanicos&resource=ordenes", nil, 200)
	fixture := map[string]string{"url": origin, "username": "emanager_" + h.suffix, "password": h.password, "suffix": h.suffix, "adminId": adminID, "mechanicId": mechanicID, "limitedId": limitedID, "limitedUsername": "elimited_" + h.suffix, "serviceId": serviceID, "partId": partID, "unitId": unitID}
	input, _ := json.Marshal(fixture)
	cmd := exec.CommandContext(ctx, "node", filepath.Join("..", "..", "checks", "frontend-browser.mjs"))
	cmd.Stdin = bytes.NewReader(input)
	output, err := cmd.CombinedOutput()
	_ = os.WriteFile(filepath.Join("..", ".cache", "phase-e-browser-last.log"), output, 0600)
	if err != nil {
		t.Fatalf("Navegador E: %v\n%s", err, output)
	}
	t.Log(strings.TrimSpace(string(output)))
}
