package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"net"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// Visual QA uses only a new test cluster. No production code or policy changes.
func TestCarbonDesignBrowser(t *testing.T) {
	if os.Getenv("CARBON_DESIGN_QA") != "1" {
		t.Skip("Requiere Chrome y PostgreSQL exclusivo de diseño")
	}
	cfg, err := pgxpool.ParseConfig(os.Getenv("INTEGRATION_DATABASE_URL"))
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_test" || cfg.ConnConfig.Host != "127.0.0.1" || cfg.ConnConfig.Port != 25440 {
		t.Fatal("Solo cluster QA nuevo del puerto 25440")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	origin := "http://" + ln.Addr().String()
	handler, err := httpapi.New(&postgres.Store{Pool: pool}, httpapi.Options{Origin: origin, Key: []byte(rand.Text() + rand.Text()), FrontendDir: filepath.Join("..", "..", "dist")})
	if err != nil {
		t.Fatal(err)
	}
	api := httptest.NewUnstartedServer(handler)
	api.Listener = ln
	api.Start()
	defer api.Close()
	h := workshopHarness{ctx: ctx, pool: pool, api: api, origin: origin, password: rand.Text(), suffix: rand.Text()[:8]}
	hash, err := bcrypt.GenerateFromPassword([]byte(h.password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	fixture := map[string]string{"url": origin, "password": h.password, "suffix": h.suffix}
	for _, a := range []struct{ key, role, job string }{{"admin", "Administrador", "Administración"}, {"manager", "Gerente", "Gerencia"}, {"mechanic", "Mecánico", "Mecánico"}} {
		name := a.role + " QA " + h.suffix
		username := a.key + "_" + h.suffix
		email := username + "@example.invalid"
		employee := h.scalar(t, `INSERT INTO empleados(nombre,tipo_empleado) VALUES($1,$2) RETURNING id_empleado::text`, name, a.job)
		user := h.scalar(t, `INSERT INTO usuarios(nombre,usuario,correo,password_hash,id_empleado) VALUES($1,$2,$3,$4,$5) RETURNING id_usuario::text`, name, username, email, string(hash), employee)
		h.exec(t, `INSERT INTO usuario_roles(id_usuario,id_rol) SELECT $1,id_rol FROM roles WHERE nombre_rol=$2 AND activo`, user, a.role)
		fixture[a.key+"Username"] = username
		fixture[a.key+"Email"] = email
		fixture[a.key+"Id"] = user
	}
	raw, _ := json.Marshal(fixture)
	cmd := exec.CommandContext(ctx, "node", filepath.Join("..", "..", "checks", "carbon-design-browser.mjs"))
	cmd.Stdin = bytes.NewReader(raw)
	output, err := cmd.CombinedOutput()
	dir := filepath.Join("..", ".cache", "diseno-login-dashboard")
	_ = os.MkdirAll(dir, 0700)
	_ = os.WriteFile(filepath.Join(dir, "browser.log"), output, 0600)
	if err != nil {
		t.Fatalf("Diseño en Chrome: %v\n%s", err, output)
	}
	t.Log(string(output))
}
