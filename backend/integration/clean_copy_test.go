package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"net"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	mgstore "github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/mongodb"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Explicit opt-in. Only the newly restored QA copy, never development/production.
func TestCleanRestoredCopy(t *testing.T) {
	dsn := os.Getenv("INTEGRATION_CLEAN_DATABASE_URL")
	if dsn == "" {
		t.Skip("Requiere copia restaurada aislada para comprobar estado vacío")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil || !regexp.MustCompile(`^jac_restore_\d{14}$`).MatchString(cfg.ConnConfig.Database) || cfg.ConnConfig.Host != "127.0.0.1" || cfg.ConnConfig.Port != 25435 {
		t.Fatal("Solo copia QA explícita en puerto 25435")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	tables := []string{"clientes", "contactos_cliente", "vehiculos", "historial_propietarios", "recepciones", "diagnosticos", "cotizaciones", "detalle_cotizacion", "ordenes_trabajo", "asignaciones_ot", "avances_ot", "pruebas_ot", "historial_estados_ot", "pagos", "entregas", "auxilios"}
	for _, table := range tables {
		var n int
		if err = pool.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&n); err != nil || n != 0 {
			t.Fatal("Copia no vacía:", table, err)
		}
	}
	name := "elbravodeljac_restore_" + strings.TrimPrefix(cfg.ConnConfig.Database, "jac_restore_")
	mg, err := mgstore.Open("mongodb://127.0.0.1:27032/?directConnection=true&retryWrites=false", "elbravodeljac_tecnico_test")
	if err != nil {
		t.Fatal(err)
	}
	defer mg.Close(context.Background())
	// Test adapter targets only the corresponding newly restored QA database.
	mg.DB = mg.Client.Database(name)
	for _, col := range []string{"resultados_escaner", "evidencias_tecnicas", "archivos_tecnicos.files", "archivos_tecnicos.chunks"} {
		n, e := mg.DB.Collection(col).CountDocuments(ctx, map[string]any{})
		if e != nil || n != 0 {
			t.Fatal("Mongo copia no vacío", col, e)
		}
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	origin := "http://" + ln.Addr().String()
	handler, err := httpapi.New(&postgres.Store{Pool: pool}, httpapi.Options{Origin: origin, Key: []byte(rand.Text() + rand.Text()), Technical: mg, FrontendDir: filepath.Join("..", "..", "dist")})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(handler)
	server.Listener = ln
	server.Start()
	defer server.Close()
	// Existing private credential is never logged or committed.
	access, err := os.ReadFile(filepath.Join("..", ".cache", "acceso-desarrollo.txt"))
	if err != nil {
		t.Fatal("Falta acceso privado para comprobar cuenta existente")
	}
	username, password := "", ""
	for _, line := range strings.Split(string(access), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Usuario:") {
			username = strings.TrimSpace(strings.TrimPrefix(line, "Usuario:"))
		}
		if strings.HasPrefix(line, "Contraseña:") {
			password = strings.TrimSpace(strings.TrimPrefix(line, "Contraseña:"))
		}
	}
	if username == "" || password == "" {
		t.Fatal("Formato de acceso privado inválido")
	}
	fixture := map[string]string{"url": origin, "username": username, "password": password, "newPassword": rand.Text(), "suffix": rand.Text()[:8], "database": cfg.ConnConfig.Database}
	raw, _ := json.Marshal(fixture)
	cmd := exec.CommandContext(ctx, "node", filepath.Join("..", "..", "checks", "clean-copy-browser.mjs"))
	cmd.Stdin = bytes.NewReader(raw)
	out, err := cmd.CombinedOutput()
	_ = os.WriteFile(filepath.Join("..", ".cache", "auditoria-20261103", "clean-browser.log"), out, 0600)
	if err != nil {
		t.Fatalf("Prueba funcional de copia limpia: %v\n%s", err, out)
	}
	var clients, vehicles int
	if err = pool.QueryRow(ctx, "SELECT (SELECT count(*) FROM clientes),(SELECT count(*) FROM vehiculos)").Scan(&clients, &vehicles); err != nil || clients != 1 || vehicles != 1 {
		t.Fatal("Altas de copia no persistieron", err)
	}
	t.Log("Copia limpia: cuenta Gerente existente, Administrador y mecánicos nuevos por API, estados vacíos y altas cliente/vehículo persistidas; desarrollo no utilizado")
}
