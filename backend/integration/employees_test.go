package integration

import (
	"context"
	"crypto/rand"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestEmployeesPostgreSQL(t *testing.T) {
	dsn := os.Getenv("INTEGRATION_DATABASE_URL")
	if dsn == "" {
		t.Skip("Requiere PostgreSQL aislado elbravodeljac_test")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_test" {
		t.Fatal("Solo base de pruebas")
	}
	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	handler, err := httpapi.New(&postgres.Store{Pool: pool}, httpapi.Options{Origin: "http://localhost:8080", Key: []byte(rand.Text() + rand.Text())})
	if err != nil {
		t.Fatal(err)
	}
	api := httptest.NewServer(handler)
	defer api.Close()
	h := workshopHarness{pool: pool, api: api, ctx: ctx, password: rand.Text(), suffix: rand.Text()[:8]}
	perms := []string{}
	for module, actions := range domain.PermissionPolicy().Modules {
		for _, action := range actions {
			perms = append(perms, module+"."+action)
		}
	}
	adminID, admin := h.user(t, "employer", "EmployeeAdmin_"+h.suffix, perms)
	// This fixture must not become an extra administrator for later tests of
	// last-admin protection. Disable only this test's account; retain its records.
	defer h.exec(t, `UPDATE usuarios SET activo=false WHERE id_usuario=$1`, adminID)
	workerID, worker := h.user(t, "employeeuser", "AccesoTecnico_"+h.suffix, []string{"ordenes.ver"})
	fingerprint := `SELECT md5(coalesce((SELECT jsonb_agg(to_jsonb(u) ORDER BY id_usuario)::text FROM usuarios u),'')||coalesce((SELECT jsonb_agg(to_jsonb(ur) ORDER BY id_usuario,id_rol)::text FROM usuario_roles ur),'')||coalesce((SELECT jsonb_agg(to_jsonb(r) ORDER BY id_rol)::text FROM roles r),'')||coalesce((SELECT jsonb_agg(to_jsonb(rp) ORDER BY id_rol,id_permiso)::text FROM rol_permisos rp),'')||coalesce((SELECT jsonb_agg(to_jsonb(p) ORDER BY id_permiso)::text FROM permisos p),''))`
	before := h.scalar(t, fingerprint)
	h.exec(t, postgres.EmployeesMigration)
	if h.scalar(t, fingerprint) != before {
		t.Fatal("Migración alteró cuentas/RBAC existentes")
	}
	if h.scalar(t, `SELECT (id_empleado IS NULL)::text FROM usuarios WHERE id_usuario=$1`, workerID) != "true" {
		t.Fatal("No inferir empleado/cargo")
	}
	h.call(t, worker, "GET", "/auth/me", nil, 200)
	h.call(t, worker, "GET", "/empleados", nil, 403)
	h.call(t, workshopSession{}, "GET", "/empleados", nil, 401)
	h.call(t, admin, "POST", "/empleados", map[string]any{"name": strings.Repeat("ñ", 151), "employeeType": "Mecánico"}, 422)
	h.call(t, admin, "POST", "/empleados", map[string]any{"name": "Válido", "employeeType": strings.Repeat("a", 41)}, 422)
	h.call(t, admin, "POST", "/empleados", map[string]any{"name": "Válido", "employeeType": "   "}, 422)
	employee := dataObject(t, h.call(t, admin, "POST", "/empleados", map[string]any{"name": "Persona laboral " + h.suffix, "employeeType": "Mecánico"}, 201))
	eid := employee["id"].(string)
	if employee["userId"] != nil {
		t.Fatal("Empleado sin cuenta debe ser posible")
	}
	h.call(t, admin, "GET", "/empleados?q="+h.suffix, nil, 200)
	h.call(t, admin, "GET", "/empleados/"+eid, nil, 200)
	h.call(t, worker, "PATCH", "/empleados/"+eid, map[string]string{"employeeType": "Administrador"}, 403)
	linked := dataObject(t, h.call(t, admin, "PATCH", "/usuarios/"+workerID, map[string]string{"employeeId": eid}, 200))
	if linked["name"] != employee["name"] || linked["employeeType"] != "Mecánico" || len(linked["roles"].([]any)) != 1 {
		t.Fatal("Vínculo incorrecto o cambió roles")
	}
	role := h.scalar(t, `INSERT INTO roles(nombre_rol) VALUES($1) RETURNING id_rol::text`, "SupervisorAcceso_"+h.suffix)
	h.exec(t, `INSERT INTO rol_permisos(id_rol,id_permiso) SELECT $1,id_permiso FROM permisos WHERE codigo='ordenes.ver'`, role)
	linked = dataObject(t, h.call(t, admin, "PATCH", "/usuarios/"+workerID, map[string]string{"roleId": role}, 200))
	if len(linked["roles"].([]any)) != 2 || linked["employeeType"] != "Mecánico" {
		t.Fatal("Cargo y roles deben ser independientes")
	}
	// Changing access roles intentionally revokes the old JWT.
	h.call(t, worker, "GET", "/auth/me", nil, 401)
	login := h.send(workshopSession{}, "POST", "/auth/login", map[string]string{"username": "employeeuser_" + h.suffix, "password": h.password})
	if login.err != nil || login.status != 200 {
		t.Fatal("No se pudo renovar sesión tras cambiar roles")
	}
	var session struct {
		CSRF string `json:"csrfToken"`
	}
	decodeTest(t, login.raw, &session)
	worker = workshopSession{login.cookies[0], session.CSRF}
	h.call(t, admin, "PATCH", "/empleados/"+eid, map[string]string{"name": "Nombre laboral actualizado", "employeeType": "Administrador"}, 200)
	if h.scalar(t, `SELECT nombre FROM usuarios WHERE id_usuario=$1`, workerID) != "Nombre laboral actualizado" {
		t.Fatal("Nombre principal no sincronizado")
	}
	h.call(t, worker, "GET", "/empleados", nil, 403)
	h.call(t, admin, "PATCH", "/usuarios/"+workerID, map[string]string{"name": "Nombre contradictorio"}, 422)
	otherID, _ := h.user(t, "otheremployee", "AccesoOtro_"+h.suffix, []string{"ordenes.ver"})
	oldName := h.scalar(t, `SELECT nombre FROM usuarios WHERE id_usuario=$1`, otherID)
	h.call(t, admin, "PATCH", "/usuarios/"+otherID, map[string]string{"employeeId": eid, "name": "No persistir", "roleId": role}, 409)
	if h.scalar(t, `SELECT nombre FROM usuarios WHERE id_usuario=$1`, otherID) != oldName {
		t.Fatal("Fallo de transacción modificó usuario")
	}
	body := map[string]any{"name": "Cuenta nueva", "username": "atomic_" + h.suffix, "password": h.password, "roleId": role, "employeeId": eid}
	h.call(t, admin, "POST", "/usuarios", body, 409)
	if h.scalar(t, `SELECT count(*)::text FROM usuarios WHERE usuario=$1`, body["username"]) != "0" {
		t.Fatal("Creación parcial tras fallar vínculo")
	}
	body["employeeId"] = "9223372036854775807"
	h.call(t, admin, "POST", "/usuarios", body, 404)
	h.call(t, admin, "PATCH", "/empleados/"+eid, map[string]bool{"active": false}, 200)
	h.call(t, worker, "GET", "/auth/me", nil, 200)
	if h.scalar(t, `SELECT activo::text FROM usuarios WHERE id_usuario=$1`, workerID) != "true" {
		t.Fatal("Estado laboral cambió habilitación de cuenta")
	}
	h.call(t, admin, "PATCH", "/empleados/"+eid, map[string]any{"active": true, "employeeType": "Mecánico"}, 200)
	t.Log("Migración repetible, conservación de cuentas/RBAC, empleado sin cuenta, vínculo único atómico, nombres, permisos y cargo independiente de dos roles: APROBADOS")
}
