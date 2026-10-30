package integration

import (
	"context"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Never restores a dump onto an existing database. Create an additional, uniquely
// named test database inside the verified test cluster; leave it for inspection.
func TestCatalogRolesMigrationAndHTTP(t *testing.T) {
	dsn := os.Getenv("INTEGRATION_DATABASE_URL")
	if dsn == "" {
		t.Skip("Requiere PostgreSQL aislado")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_test" {
		t.Fatal("Solo cluster de pruebas")
	}
	ctx := context.Background()
	parent, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer parent.Close()
	name := "jac_catalog_migration_" + strings.ToLower(rand.Text()[:10])
	if _, err = parent.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()); err != nil {
		t.Fatal(err)
	}
	cfg = cfg.Copy()
	cfg.ConnConfig.Database = name
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	h := workshopHarness{pool: pool, ctx: ctx, password: rand.Text(), suffix: rand.Text()[:8]}
	schema, err := os.ReadFile("../../database/postgresql/esquema_actual.sql")
	if err != nil {
		t.Fatal(err)
	}
	lines := []string{}
	for _, line := range strings.Split(string(schema), "\n") {
		if !strings.HasPrefix(line, "\\") {
			lines = append(lines, line)
		}
	}
	h.exec(t, strings.Join(lines, "\n"))
	h.exec(t, `SET search_path TO public`)
	for _, m := range []string{postgres.FunctionalMigration, postgres.BrandsMigration, postgres.AttentionMigration, postgres.DiagnosisWorkflowMigration, postgres.AuditMigration, postgres.PaymentsMigration, postgres.EmployeesMigration} {
		h.exec(t, m)
	}
	key := []byte(rand.Text() + rand.Text())
	handler, err := httpapi.New(&postgres.Store{Pool: pool}, httpapi.Options{Origin: "http://localhost:8080", Key: key})
	if err != nil {
		t.Fatal(err)
	}
	h.api = httptest.NewServer(handler)
	defer h.api.Close()
	all := []string{}
	for m, actions := range domain.PermissionPolicy().Modules {
		for _, a := range actions {
			all = append(all, m+"."+a)
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(h.password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	legacyUser := func(label, role string, codes []string) string {
		t.Helper()
		rid := h.scalar(t, `INSERT INTO roles(nombre_rol) VALUES($1) RETURNING id_rol::text`, role)
		id := h.scalar(t, `INSERT INTO usuarios(nombre,usuario,password_hash) VALUES($1,$2,$3) RETURNING id_usuario::text`, label, label+"_"+h.suffix, string(hash))
		h.exec(t, `INSERT INTO usuario_roles(id_usuario,id_rol) VALUES($1,$2)`, id, rid)
		for _, code := range codes {
			h.exec(t, `INSERT INTO permisos(codigo) VALUES($1) ON CONFLICT DO NOTHING`, code)
			h.exec(t, `INSERT INTO rol_permisos(id_rol,id_permiso) SELECT $1,id_permiso FROM permisos WHERE codigo=$2`, rid, code)
		}
		return id
	}
	owner := legacyUser("owner", "Propietario", all)
	oldAdmin := legacyUser("legacy", "Administrador", all)
	oldAux := legacyUser("historical", "Auxiliar", []string{"auxilio.ver"})
	unknown := h.scalar(t, `INSERT INTO clientes(nombre,tipo_documento,numero_documento,telefono,direccion) VALUES('Histórico','Licencia','abc-001','987654321','Lima') RETURNING id_cliente::text`)
	// Historical auxiliary participation, created before the retirement migration.
	vehicle := h.scalar(t, `INSERT INTO vehiculos(id_cliente,placa,marca,modelo) VALUES($1,'MIG-001','JAC','Prueba') RETURNING id_vehiculo::text`, unknown)
	reception := h.scalar(t, `INSERT INTO recepciones(id_vehiculo,id_cliente_atencion,id_recepcionista,km_ingreso,sintomas,tipo_atencion) VALUES($1,$2,$3,0,'Histórico','Directo') RETURNING id_recepcion::text`, vehicle, unknown, oldAdmin)
	quote := h.scalar(t, `INSERT INTO cotizaciones(id_recepcion,estado,total) VALUES($1,'aprobada',0) RETURNING id_cotizacion::text`, reception)
	order := h.scalar(t, `INSERT INTO ordenes_trabajo(id_cotizacion,codigo_ot,id_usuario_creacion) VALUES($1,'OT-MIG',$2) RETURNING id_orden::text`, quote, oldAdmin)
	h.exec(t, `INSERT INTO asignaciones_ot(id_orden,id_usuario,rol_en_trabajo) VALUES($1,$2,'AUXILIAR')`, order, oldAux)
	auxBefore := h.scalar(t, `SELECT md5(jsonb_agg(to_jsonb(a) ORDER BY id_asignacion)::text) FROM asignaciones_ot a`)
	h.exec(t, `INSERT INTO clientes(nombre,tipo_documento,numero_documento,telefono,direccion) VALUES('Alias','dni','12345678','987654321','Lima'),('Duplicado','DNI','12345678','987654321','Lima')`)
	t.Run("colisión aborta sin pérdida", func(t *testing.T) {
		conn, e := pool.Acquire(ctx)
		if e != nil {
			t.Fatal(e)
		}
		defer conn.Release()
		if _, e = conn.Exec(ctx, postgres.DocumentTypesMigration); e == nil {
			t.Fatal("Debió rechazar colisión")
		}
		if _, e = conn.Exec(ctx, `ROLLBACK`); e != nil {
			t.Fatal(e)
		}
		if h.scalar(t, `SELECT (to_regclass('public.tipos_documento') IS NULL)::text`) != "true" {
			t.Fatal("DDL parcial")
		}
		if h.scalar(t, `SELECT numero_documento FROM clientes WHERE id_cliente=$1`, unknown) != "abc-001" {
			t.Fatal("Cambió identidad")
		}
	})
	// Resolve only our deliberately duplicated synthetic fixture.
	h.exec(t, `UPDATE clientes SET numero_documento='87654321' WHERE nombre='Duplicado'`)
	fingerprint := `SELECT md5(coalesce((SELECT jsonb_agg(to_jsonb(u) ORDER BY id_usuario)::text FROM usuarios u),'')||coalesce((SELECT jsonb_agg(to_jsonb(ur) ORDER BY id_usuario,id_rol)::text FROM usuario_roles ur),'')||coalesce((SELECT jsonb_agg(to_jsonb(rp) ORDER BY id_rol,id_permiso)::text FROM rol_permisos rp WHERE rp.id_rol IN (SELECT id_rol FROM usuario_roles)),''))`
	before := h.scalar(t, fingerprint)
	h.exec(t, postgres.DocumentTypesMigration)
	h.exec(t, postgres.ThreeRolesMigration)
	if h.scalar(t, fingerprint) != before {
		t.Fatal("Cambió usuarios, asignaciones o permisos anteriores")
	}
	if h.scalar(t, `SELECT nombre_rol FROM roles r JOIN usuario_roles USING(id_rol) WHERE id_usuario=$1`, owner) != "Gerente" {
		t.Fatal("Propietario no conservó ID")
	}
	if h.scalar(t, `SELECT nombre_rol FROM roles r JOIN usuario_roles USING(id_rol) WHERE id_usuario=$1`, oldAdmin) != "Administrador (legado)" {
		t.Fatal("Administrador anterior no conservado")
	}
	if h.scalar(t, `SELECT count(*)::text FROM roles WHERE activo`) != "3" {
		t.Fatal("Deben existir tres roles ofertados")
	}
	login := func(label string) workshopSession {
		t.Helper()
		r := h.send(workshopSession{}, "POST", "/auth/login", map[string]string{"username": label + "_" + h.suffix, "password": h.password})
		if r.status != 200 {
			t.Fatal("login", r.status)
		}
		var v struct {
			CSRF string `json:"csrfToken"`
		}
		decodeTest(t, r.raw, &v)
		return workshopSession{r.cookies[0], v.CSRF}
	}
	gerente := login("owner")
	legacy := login("legacy")
	adminID, admin := h.user(t, "operations", "Administrador", nil)
	_, mechanic := h.user(t, "mechanic", "Mecánico", nil)
	h.call(t, gerente, "GET", "/usuarios", nil, 200)
	h.call(t, legacy, "GET", "/usuarios", nil, 200)
	h.call(t, gerente, "PATCH", "/usuarios/"+oldAux, map[string]string{"name": "Histórico conservado"}, 200)
	h.call(t, admin, "GET", "/usuarios", nil, 403)
	h.call(t, mechanic, "GET", "/usuarios", nil, 403)
	h.call(t, mechanic, "GET", "/ordenes/"+order+"/pagos", nil, 403)
	t.Run("auxiliar histórico conservado y nuevas asignaciones rechazadas", func(t *testing.T) {
		if h.scalar(t, `SELECT md5(jsonb_agg(to_jsonb(a) ORDER BY id_asignacion)::text) FROM asignaciones_ot a`) != auxBefore {
			t.Fatal("Cambió asignación histórica")
		}
		saved := dataObject(t, h.call(t, gerente, "GET", "/ordenes/"+order, nil, 200))
		if saved["assignments"].([]any)[0].(map[string]any)["role"] != "AUXILIAR" {
			t.Fatal("Historial ausente")
		}
		if _, e := pool.Exec(ctx, `INSERT INTO asignaciones_ot(id_orden,id_usuario,rol_en_trabajo) VALUES($1,$2,'AUXILIAR')`, order, oldAux); e == nil {
			t.Fatal("SQL permitió nuevo auxiliar")
		}
		h.call(t, gerente, "PATCH", "/ordenes/"+order, map[string]string{"state": "En reparación", "note": "No alterar auxiliar", "assistantId": oldAux}, 422)
	})
	gerenteRole := h.scalar(t, `SELECT id_rol::text FROM roles WHERE nombre_rol='Gerente'`)
	h.call(t, admin, "PATCH", "/usuarios/"+adminID, map[string]string{"roleId": gerenteRole}, 403)
	h.call(t, gerente, "PATCH", "/usuarios/"+owner, map[string]string{"roleId": gerenteRole}, 403)
	h.call(t, legacy, "PATCH", "/usuarios/"+owner, map[string]bool{"active": false}, 422)
	h.call(t, legacy, "PATCH", "/usuarios/"+owner, map[string]string{"removeRoleId": gerenteRole}, 422)
	h.call(t, gerente, "PUT", "/roles/"+gerenteRole+"/permisos", map[string]any{"permissions": map[string][]string{}}, 403)
	auxRole := h.scalar(t, `SELECT id_rol::text FROM roles WHERE nombre_rol='Auxiliar'`)
	h.call(t, gerente, "PATCH", "/usuarios/"+adminID, map[string]string{"roleId": auxRole}, 422)
	t.Run("catálogo clientes contactos validación y búsqueda", func(t *testing.T) {
		list := h.call(t, admin, "GET", "/tipos-documento", nil, 200)["data"].([]any)
		if len(list) != 5 {
			t.Fatal("Cuatro actuales y un histórico", len(list))
		}
		h.call(t, mechanic, "GET", "/tipos-documento", nil, 403)
		h.call(t, workshopSession{}, "GET", "/tipos-documento", nil, 401)
		body := map[string]string{"name": "Cliente catálogo", "documentType": "PASAPORTE", "documentNumber": "DOC001", "phone": "987654321", "address": "Lima"}
		client := responseID(t, h.call(t, admin, "POST", "/clientes", body, 201))
		if h.scalar(t, `SELECT d.codigo FROM clientes c JOIN tipos_documento d USING(id_tipo_documento) WHERE c.id_cliente=$1`, client) != "PASAPORTE" {
			t.Fatal("FK no persistida")
		}
		h.call(t, admin, "GET", "/clientes?q=DOC001", nil, 200)
		h.call(t, admin, "POST", "/clientes", body, 409)
		body["documentType"] = "Desconocido"
		body["documentNumber"] = "DOC002"
		h.call(t, admin, "POST", "/clientes", body, 422)
		h.call(t, admin, "PATCH", "/clientes/"+client, map[string]string{"documentType": "CE", "documentNumber": "DOC002"}, 200)
		h.call(t, admin, "POST", "/clientes/"+client+"/contactos", map[string]string{"name": "Contacto autorizado", "documentType": "CE", "documentNumber": "CONT001", "phone": "987654321", "relation": "Conductor"}, 201)
		h.call(t, admin, "PATCH", "/clientes/"+unknown, map[string]string{"name": "Editado sin cambiar identidad"}, 200)
		if h.scalar(t, `SELECT tipo_documento||':'||numero_documento FROM clientes WHERE id_cliente=$1`, unknown) != "Licencia:abc-001" {
			t.Fatal("Histórico alterado")
		}
		h.call(t, admin, "PATCH", "/clientes/"+unknown, map[string]string{"documentNumber": "DOC003"}, 422)
		h.exec(t, `UPDATE tipos_documento SET activo=false WHERE codigo='CE'`)
		body["documentType"] = "CE"
		h.call(t, admin, "POST", "/clientes", body, 422)
		h.call(t, admin, "PATCH", "/clientes/"+client, map[string]string{"name": "Nombre actualizado"}, 200)
	})
	t.Run("JWT activo revocación y sesión tras cambiar roles", func(t *testing.T) {
		// Separate HTTP server/window for login-heavy cases; production rate
		// limiting stays unchanged. Same signing key and real persisted users.
		handler, e := httpapi.New(&postgres.Store{Pool: pool}, httpapi.Options{Origin: "http://localhost:8080", Key: key})
		if e != nil {
			t.Fatal(e)
		}
		h.api = httptest.NewServer(handler)
		defer h.api.Close()
		gerente = login("owner")
		admin = login("operations")
		mechanic = login("mechanic")

		bad := mechanic
		bad.cookie = &http.Cookie{Name: mechanic.cookie.Name, Value: "invalid.jwt"}
		h.call(t, bad, "GET", "/auth/me", nil, 401)
		h.call(t, gerente, "PATCH", "/usuarios/"+adminID, map[string]bool{"active": false}, 200)
		h.call(t, admin, "GET", "/auth/me", nil, 401)
		h.call(t, gerente, "PATCH", "/usuarios/"+adminID, map[string]bool{"active": true}, 200)
		admin = login("operations")
		role := h.scalar(t, `SELECT id_rol::text FROM roles WHERE nombre_rol='Administrador'`)
		h.call(t, gerente, "PUT", "/roles/"+role+"/permisos", map[string]any{"permissions": map[string][]string{"clientes": {"ver"}}}, 200)
		h.call(t, admin, "GET", "/clientes", nil, 200)
		h.call(t, admin, "POST", "/clientes", map[string]string{}, 403)
		h.call(t, gerente, "PATCH", "/usuarios/"+adminID, map[string]string{"roleId": gerenteRole}, 200)
		h.call(t, admin, "GET", "/auth/me", nil, 401)
		admin = login("operations")
		h.call(t, admin, "GET", "/usuarios", nil, 200)
		// Reapplying migrations must never replenish a deliberately reduced matrix.
		before := h.scalar(t, `SELECT count(*)::text FROM rol_permisos WHERE id_rol=$1`, role)
		h.exec(t, postgres.DocumentTypesMigration)
		h.exec(t, postgres.ThreeRolesMigration)
		if h.scalar(t, `SELECT count(*)::text FROM rol_permisos WHERE id_rol=$1`, role) != before {
			t.Fatal("Reaplicación amplió privilegios")
		}
	})
	t.Log("Migración en base nueva", name, "sin eliminarla; colisiones, historial, FK, tres roles, último Gerente y HTTP verificados")
}
