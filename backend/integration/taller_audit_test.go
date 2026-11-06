package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"net"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	mgstore "github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/mongodb"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Real browser, three independent sessions, canonical roles and employee links.
func TestTallerFourLinkedAccounts(t *testing.T) {
	if os.Getenv("FRONTEND_AUDIT") != "1" {
		t.Skip("Requiere navegador y cluster QA")
	}
	cfg, err := pgxpool.ParseConfig(os.Getenv("INTEGRATION_DATABASE_URL"))
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_test" {
		t.Fatal("Solo PostgreSQL QA")
	}
	if os.Getenv("INTEGRATION_MONGODB_DATABASE") != "elbravodeljac_tecnico_test" {
		t.Fatal("Solo MongoDB QA")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	trace := &tallerSQLTrace{}
	cfg.ConnConfig.Tracer = trace
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	token := requireFreshBrowserRun(t, pool)
	mg, err := mgstore.Open(os.Getenv("INTEGRATION_MONGODB_URI"), "elbravodeljac_tecnico_test")
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
	handler, err := httpapi.New(&postgres.Store{Pool: pool}, httpapi.Options{Origin: origin, Key: []byte(rand.Text() + rand.Text()), Technical: mg, FrontendDir: filepath.Join("..", "..", "dist")})
	if err != nil {
		t.Fatal(err)
	}
	api := httptest.NewUnstartedServer(browserIsolationHandler(handler, token))
	api.Listener = ln
	api.Start()
	defer api.Close()
	h := workshopHarness{ctx: ctx, pool: pool, api: api, origin: origin, password: rand.Text(), suffix: rand.Text()[:8]}
	hash, err := bcrypt.GenerateFromPassword([]byte(h.password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	seed := func(label, role, job string, active bool) (string, string) {
		t.Helper()
		employee := h.scalar(t, `INSERT INTO empleados(nombre,tipo_empleado,activo) VALUES($1,$2,$3) RETURNING id_empleado::text`, label+" "+h.suffix, job, active)
		user := h.scalar(t, `INSERT INTO usuarios(nombre,usuario,password_hash,id_empleado) VALUES($1,$2,$3,$4) RETURNING id_usuario::text`, label+" "+h.suffix, label+"_"+h.suffix, string(hash), employee)
		h.exec(t, `INSERT INTO usuario_roles(id_usuario,id_rol) SELECT $1,id_rol FROM roles WHERE nombre_rol=$2`, user, role)
		return user, employee
	}
	manager, _ := seed("taller_gerente", "Gerente", "Gerencia", true)
	h.exec(t, `INSERT INTO clientes(nombre,tipo_documento,numero_documento,telefono,direccion) SELECT 'Carga QA '||$1||' '||g,'Pasaporte','LOAD'||$1||g,'+51987654321','QA aislado' FROM generate_series(1,1200) g`, h.suffix)
	admin, adminEmployee := seed("taller_admin", "Administrador", "Administración", true)
	first, firstEmployee := seed("taller_m1", "Mecánico", "Mecánico", true)
	second, secondEmployee := seed("taller_m2", "Mecánico", "Mecánico", true)
	inactive, inactiveEmployee := seed("taller_inactivo", "Mecánico", "Mecánico", false)
	fixture := map[string]string{"isolationToken": token, "managerUsername": "taller_gerente_" + h.suffix, "managerId": manager, "url": origin, "suffix": h.suffix, "password": h.password, "username": "taller_admin_" + h.suffix, "workerUsername": "taller_m1_" + h.suffix, "otherUsername": "taller_m2_" + h.suffix, "adminId": admin, "adminEmployee": adminEmployee, "mechanicId": first, "mechanicEmployee": firstEmployee, "otherId": second, "otherEmployee": secondEmployee, "inactiveId": inactive, "inactiveEmployee": inactiveEmployee}
	input, _ := json.Marshal(fixture)
	cmd := exec.CommandContext(ctx, "node", filepath.Join("..", "..", "checks", "taller-browser.mjs"))
	cmd.Stdin = bytes.NewReader(input)
	out, err := cmd.CombinedOutput()
	_ = os.MkdirAll(filepath.Join("..", ".cache", "taller-audit"), 0700)
	_ = os.WriteFile(filepath.Join("..", ".cache", "taller-audit", "browser.log"), out, 0600)
	if err != nil {
		t.Fatalf("Taller con cuentas distintas: %v\n%s", err, out)
	}
	raw, err := os.ReadFile(filepath.Join("..", ".cache", "taller-audit", "resultado.json"))
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		ClientID    string   `json:"clientId"`
		VehicleID   string   `json:"vehicleId"`
		ReceptionID string   `json:"receptionId"`
		DiagnosisID string   `json:"diagnosisId"`
		OrderID     string   `json:"orderId"`
		AuxIDs      []string `json:"auxIds"`
	}
	if err = json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	if h.scalar(t, `SELECT (r.id_recepcionista=$2::bigint AND r.id_mecanico_diagnostico=$3::bigint AND d.id_tecnico=$3::bigint AND r.id_cliente_atencion=$4::bigint AND r.id_vehiculo=$5::bigint)::text FROM recepciones r JOIN diagnosticos d USING(id_recepcion) WHERE r.id_recepcion=$1`, result.ReceptionID, admin, first, result.ClientID, result.VehicleID) != "true" {
		t.Fatal("Identidades/actores incorrectos")
	}
	if h.scalar(t, `SELECT count(*)::text FROM diagnosticos d WHERE d.id_diagnostico=$1 AND d.estado='Finalizado' AND d.historial @> jsonb_build_array(jsonb_build_object('by',$2::text))`, result.DiagnosisID, first) != "1" {
		t.Fatal("Falta autoría de diagnóstico")
	}
	if h.scalar(t, `SELECT count(*)::text FROM auxilios WHERE id_auxilio::text=ANY($1) AND id_tecnico=$2 AND id_cliente_solicitante=$3`, result.AuxIDs, first, result.ClientID) != "2" {
		t.Fatal("No persisten auxilios independientes")
	}
	// Controlled calendar fixture: both creations are real HTTP operations;
	// only this test row's date is moved to yesterday to exercise historical listing.
	h.exec(t, `UPDATE auxilios SET fecha=fecha-interval '1 day' WHERE id_auxilio=$1`, result.AuxIDs[0])
	if h.scalar(t, `SELECT (count(DISTINCT fecha::date)=2)::text FROM auxilios WHERE id_auxilio::text=ANY($1)`, result.AuxIDs) != "true" {
		t.Fatal("Auxilios no pertenecen a fechas diferentes")
	}
	if h.scalar(t, `SELECT (e.ficha_pdf IS NOT NULL AND e.documento_receptor_snap=c.numero_documento)::text FROM entregas e JOIN ordenes_trabajo o USING(id_orden) JOIN cotizaciones q USING(id_cotizacion) JOIN recepciones r USING(id_recepcion) JOIN clientes c ON c.id_cliente=r.id_cliente_atencion WHERE e.id_orden=$1`, result.OrderID) != "true" {
		t.Fatal("Entrega/PDF no persistidos")
	}
	t.Log(string(out))
	if trace.unexpected.Load() != 0 {
		t.Fatal("Errores SQL inesperados:", trace.unexpected.Load())
	}
	t.Logf("SQL: %d cancelaciones por navegación/F5, %d restricciones esperadas y 0 errores inesperados", trace.canceled.Load(), trace.expected.Load())
	t.Log("PostgreSQL confirma actores de Administración/Mecánico, vínculos laborales, borrador/final, dos auxilios fechados y entrega archivada")
}

type tallerSQLTrace struct{ canceled, expected, unexpected atomic.Int64 }

func (t *tallerSQLTrace) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	return ctx
}
func (t *tallerSQLTrace) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, d pgx.TraceQueryEndData) {
	if d.Err == nil {
		return
	}
	if ctx.Err() != nil {
		t.canceled.Add(1)
		return
	}
	var constraint *pgconn.PgError
	if errors.As(d.Err, &constraint) && (constraint.Code == "23505" || constraint.Code == "23514" || constraint.Code == "23503") {
		t.expected.Add(1)
		return
	}
	t.unexpected.Add(1)
}
