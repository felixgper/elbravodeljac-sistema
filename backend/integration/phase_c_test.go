package integration

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestPhaseCRelationalHTTP(t *testing.T) {
	dsn := os.Getenv("INTEGRATION_DATABASE_URL")
	if dsn == "" {
		t.Skip("Integración PostgreSQL no ejecutada: falta INTEGRATION_DATABASE_URL")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_test" {
		t.Fatal("Solo se permite elbravodeljac_test")
	}
	cfg.ConnConfig.RuntimeParams["timezone"] = "UTC"
	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	key := make([]byte, 32)
	if _, err = rand.Read(key); err != nil {
		t.Fatal(err)
	}
	handler, err := httpapi.New(&postgres.Store{Pool: pool}, httpapi.Options{Origin: "http://localhost:8080", Key: key})
	if err != nil {
		t.Fatal(err)
	}
	api := httptest.NewServer(handler)
	defer api.Close()
	h := workshopHarness{pool: pool, api: api, ctx: ctx, password: rand.Text(), suffix: rand.Text()[:10]}
	all := []string{}
	for module, actions := range domain.PermissionPolicy().Modules {
		for _, a := range actions {
			all = append(all, module+"."+a)
		}
	}
	adminID, admin := h.user(t, "cmanager", "CAdmin_"+h.suffix, all)
	secondID, second := h.user(t, "csecond", "CSecond_"+h.suffix, all)
	_, limited := h.user(t, "climited", "CLimited_"+h.suffix, []string{"usuarios.ver", "usuarios.roles"})
	adminRole := h.scalar(t, `SELECT id_rol::text FROM usuario_roles WHERE id_usuario=$1`, adminID)
	lowRole := h.scalar(t, `INSERT INTO roles(nombre_rol) VALUES($1) RETURNING id_rol::text`, "CWorker_"+h.suffix)
	extraRole := h.scalar(t, `INSERT INTO roles(nombre_rol) VALUES($1) RETURNING id_rol::text`, "CExtra_"+h.suffix)
	auxRole := h.scalar(t, `INSERT INTO roles(nombre_rol) VALUES('Auxiliar') ON CONFLICT(nombre_rol) DO UPDATE SET nombre_rol=excluded.nombre_rol RETURNING id_rol::text`)
	h.exec(t, `INSERT INTO roles(nombre_rol) VALUES('Mecánico') ON CONFLICT DO NOTHING`)
	hash, err := bcrypt.GenerateFromPassword([]byte(h.password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	tech := h.scalar(t, `INSERT INTO usuarios(nombre,usuario,password_hash) VALUES('Técnico fixture',$1,$2) RETURNING id_usuario::text`, "ctech_"+h.suffix, string(hash))
	h.exec(t, `INSERT INTO usuario_roles(id_usuario,id_rol) SELECT $1,id_rol FROM roles WHERE nombre_rol='Mecánico'`, tech)
	var worker string
	var workerSession workshopSession
	t.Run("usuarios matriz y asignación N sin pérdidas", func(t *testing.T) {
		for _, path := range []string{"/usuarios", "/roles", "/permisos"} {
			h.call(t, admin, "GET", path, nil, 200)
		}
		worker = responseID(t, h.call(t, admin, "POST", "/usuarios", map[string]any{"name": "Trabajador fixture", "username": "cworker_" + h.suffix, "password": h.password, "roleId": lowRole}, 201))
		h.call(t, admin, "PATCH", "/usuarios/"+worker, map[string]string{"roleId": extraRole}, 200)
		if count := h.scalar(t, `SELECT count(*)::text FROM usuario_roles WHERE id_usuario=$1`, worker); count != "2" {
			t.Fatal("N assignments overwritten", count)
		}
		h.call(t, admin, "PATCH", "/usuarios/"+worker, map[string]string{"removeRoleId": extraRole}, 200)
		if count := h.scalar(t, `SELECT count(*)::text FROM usuario_roles WHERE id_usuario=$1`, worker); count != "1" {
			t.Fatal(count)
		}
		matrix := map[string]any{"permissions": map[string][]string{"clientes": {"ver"}}}
		h.call(t, admin, "PUT", "/roles/"+lowRole+"/permisos", matrix, 200)
		h.call(t, admin, "GET", "/roles/"+lowRole+"/permisos", nil, 200)
		h.call(t, admin, "PUT", "/roles/"+lowRole+"/permisos", map[string]any{"permissions": map[string][]string{"clientes": {"crear"}}}, 422)
		if count := h.scalar(t, `SELECT count(*)::text FROM rol_permisos WHERE id_rol=$1`, lowRole); count != "1" {
			t.Fatal("matrix rollback", count)
		}
		h.call(t, admin, "PUT", "/roles/"+adminRole+"/permisos", matrix, 403)
		h.call(t, admin, "PATCH", "/usuarios/"+adminID, map[string]string{"roleId": lowRole}, 403)
		h.call(t, limited, "PATCH", "/usuarios/"+worker, map[string]string{"roleId": adminRole}, 403)
		h.call(t, admin, "DELETE", "/usuarios/"+adminID, nil, 403)
		h.call(t, admin, "POST", "/roles/"+auxRole+"/permisos/restablecer", nil, 422)
		res := h.send(workshopSession{}, "POST", "/auth/login", map[string]string{"username": "cworker_" + h.suffix, "password": h.password})
		if res.err != nil || res.status != 200 {
			t.Fatalf("worker login %d %v", res.status, res.err)
		}
		var parsed struct {
			CSRF string `json:"csrfToken"`
		}
		decodeTest(t, res.raw, &parsed)
		workerSession = workshopSession{res.cookies[0], parsed.CSRF}
		h.call(t, workerSession, "PUT", "/roles/"+lowRole+"/permisos", matrix, 403)
		h.call(t, admin, "PATCH", "/usuarios/"+worker, map[string]string{"name": strings.Repeat("ñ", 151)}, 422)
		h.call(t, admin, "GET", "/usuarios/"+worker, nil, 200)
		newPassword := rand.Text()
		h.call(t, admin, "PATCH", "/usuarios/"+worker, map[string]string{"password": newPassword}, 200)
		h.call(t, workerSession, "GET", "/auth/me", nil, 401)
		disposable := responseID(t, h.call(t, admin, "POST", "/usuarios", map[string]string{"name": "Eliminar fixture", "username": "delete_" + h.suffix, "password": h.password, "roleId": lowRole}, 201))
		h.call(t, admin, "DELETE", "/usuarios/"+disposable, nil, 204)
	})
	t.Run("catálogos proveedores referencias y especialidad", func(t *testing.T) {
		for _, body := range []map[string]string{{"name": "Precio inválido", "referencePrice": "-1"}, {"name": "Decimales", "referencePrice": "1.001"}} {
			h.call(t, admin, "POST", "/servicios", body, 422)
		}
		service := responseID(t, h.call(t, admin, "POST", "/servicios", map[string]string{"name": "CServicio_" + h.suffix, "referencePrice": "120.00"}, 201))
		if service == "" {
			t.Fatal("service")
		}
		h.call(t, admin, "POST", "/servicios", map[string]string{"name": "CServicio_" + h.suffix, "referencePrice": "120.00"}, 409)
		h.call(t, admin, "POST", "/repuestos", map[string]string{"name": "Stock inválido", "referencePrice": "1", "stock": "-1"}, 422)
		part := responseID(t, h.call(t, admin, "POST", "/repuestos", map[string]string{"name": "CRepuesto_" + h.suffix, "referencePrice": "35.00", "stock": "10.25"}, 201))
		h.call(t, admin, "POST", "/proveedores", map[string]string{"name": "No perder texto", "ruc": "12345678901", "parts": "Especialidad libre"}, 422)
		provider := responseID(t, h.call(t, admin, "POST", "/proveedores", map[string]string{"name": "Proveedor fixture", "document": "P" + h.suffix, "contact": "Contacto explícito"}, 201))
		h.call(t, admin, "GET", "/proveedores", nil, 200)
		h.call(t, admin, "GET", "/proveedores/"+provider, nil, 200)
		h.call(t, admin, "PATCH", "/proveedores/"+provider, map[string]string{"phone": "+51987654321"}, 200)
		h.call(t, admin, "PATCH", "/proveedores/"+provider, map[string]string{"parts": "No guardar en contacto"}, 422)
		if value := h.scalar(t, `SELECT contacto FROM proveedores WHERE id_proveedor=$1`, provider); value != "Contacto explícito" {
			t.Fatal("information lost", value)
		}
		link := "/proveedores/" + provider + "/repuestos/" + part
		h.call(t, admin, "PUT", link, map[string]string{"referenceCost": "15.50"}, 200)
		h.call(t, admin, "PUT", link, map[string]string{"referenceCost": "16.25"}, 200)
		h.call(t, admin, "GET", "/proveedores/"+provider+"/repuestos", nil, 200)
		h.call(t, admin, "DELETE", "/proveedores/"+provider, nil, 409)
		h.call(t, admin, "PUT", link, map[string]string{"referenceCost": "-1"}, 422)
		h.call(t, admin, "PUT", "/proveedores/"+provider+"/repuestos/9223372036854775807", map[string]string{"referenceCost": "1"}, 409)
		h.call(t, admin, "DELETE", link, nil, 204)
		h.call(t, admin, "DELETE", "/proveedores/"+provider, nil, 204)
	})
	t.Run("auxilios asignación y unidades", func(t *testing.T) {
		client := responseID(t, h.call(t, admin, "POST", "/clientes", map[string]string{"name": "Auxilio fixture", "documentType": "Pasaporte", "documentNumber": "C" + strings.ToUpper(h.suffix), "phone": "+51987654321", "address": "Lima"}, 201))
		vehicle := responseID(t, h.call(t, admin, "POST", "/vehiculos", map[string]any{"clientId": client, "plate": "C-" + strings.ToUpper(h.suffix[:6]), "brand": "JAC", "model": "T8", "year": 2024, "mileage": "10"}, 201))
		unit := responseID(t, h.call(t, admin, "POST", "/unidades-auxilio", map[string]string{"code": "CUnit_" + h.suffix, "description": "Unidad fixture"}, 201))
		h.call(t, admin, "POST", "/unidades-auxilio", map[string]string{"code": "CUnit_" + h.suffix}, 409)
		body := map[string]string{"clientId": client, "vehicleId": vehicle, "technicianId": tech, "unitId": unit, "location": "Lima", "incident": "Falla mecánica", "description": "Auxilio fixture"}
		aux := responseID(t, h.call(t, admin, "POST", "/auxilios", body, 201))
		h.call(t, admin, "GET", "/auxilios", nil, 200)
		h.call(t, admin, "GET", "/auxilios/"+aux, nil, 200)
		h.call(t, admin, "GET", "/unidades-auxilio", nil, 200)
		h.call(t, admin, "PATCH", "/auxilios/"+aux, map[string]string{"technicianId": worker}, 422)
		h.call(t, admin, "PATCH", "/auxilios/"+aux, map[string]string{"unitId": unit, "technicianId": tech, "state": "en_atencion"}, 200)
		h.call(t, admin, "PATCH", "/auxilios/"+aux, map[string]string{"unitId": "9223372036854775807"}, 404)
		if state := h.scalar(t, `SELECT estado FROM auxilios WHERE id_auxilio=$1`, aux); state != "en_atencion" {
			t.Fatal("rollback", state)
		}
		h.call(t, admin, "DELETE", "/usuarios/"+tech, nil, 409)
		if n := h.scalar(t, `SELECT count(*)::text FROM usuario_roles WHERE id_usuario=$1`, tech); n != "1" {
			t.Fatal("failed user delete lost roles", n)
		}
		h.call(t, admin, "PATCH", "/auxilios/"+aux, map[string]string{"state": "finalizado"}, 200)
		h.call(t, admin, "PATCH", "/auxilios/"+aux, map[string]string{"state": "pendiente"}, 409)
	})
	t.Run("dashboard reportes sin multiplicar pagos", func(t *testing.T) {
		// Build our own multi-detail order, independent of suite/file execution order.
		customer := responseID(t, h.call(t, admin, "POST", "/clientes", map[string]string{"name": "Estadística fixture", "documentType": "Pasaporte", "documentNumber": "S" + strings.ToUpper(h.suffix), "phone": "+51987654321", "address": "Lima"}, 201))
		car := responseID(t, h.call(t, admin, "POST", "/vehiculos", map[string]any{"clientId": customer, "plate": "S-" + strings.ToUpper(h.suffix[:6]), "brand": "JAC", "model": "T8", "year": 2024, "mileage": "10"}, 201))
		reception := responseID(t, h.call(t, admin, "POST", "/recepciones", map[string]string{"clientId": customer, "vehicleId": car, "mileage": "10", "symptoms": "Estadística"}, 201))
		h.call(t, admin, "PATCH", "/recepciones/"+reception+"/mecanico", map[string]any{"technicianId": tech, "version": 0}, 200)
		diagnosis := responseID(t, h.call(t, admin, "POST", "/diagnosticos", map[string]string{"status": "Finalizado", "evaluation": "Evaluación técnica de prueba", "recommendations": "Servicios recomendados de prueba", "receptionId": reception, "findings": "Fixture de agregación"}, 201))
		service := responseID(t, h.call(t, admin, "POST", "/servicios", map[string]string{"name": "Stats_" + h.suffix, "referencePrice": "100"}, 201))
		quote := responseID(t, h.call(t, admin, "POST", "/cotizaciones", map[string]any{"diagnosisId": diagnosis, "status": "Aprobada", "items": []map[string]string{{"serviceId": service, "quantity": "1", "unitPrice": "100"}, {"serviceId": service, "quantity": "1", "unitPrice": "100"}}}, 201))
		order := responseID(t, h.call(t, admin, "POST", "/ordenes", map[string]string{"quoteId": quote, "mechanicId": tech}, 201))
		h.confirmExecution(t, admin, order)
		h.call(t, admin, "PATCH", "/ordenes/"+order, map[string]string{"state": "Finalizada", "note": "Comprobado", "test": "Ruta", "result": "Satisfactoria"}, 200)
		for _, amount := range []string{"50.00", "150.00"} {
			h.call(t, admin, "POST", "/ordenes/"+order+"/pagos", map[string]string{"amount": amount, "method": "Efectivo"}, 201)
		}
		report := dataObject(t, h.call(t, admin, "GET", "/reportes?period=Últimos%206%20meses", nil, 200))
		expected := h.scalar(t, `SELECT coalesce(sum(monto),0)::numeric(18,2)::text FROM jac_movimientos_pagos WHERE fecha >= (date_trunc('month',statement_timestamp() AT TIME ZONE 'America/Lima')-interval '5 months') AT TIME ZONE 'America/Lima' AND fecha < (date_trunc('month',statement_timestamp() AT TIME ZONE 'America/Lima')+interval '1 month') AT TIME ZONE 'America/Lima'`)
		if report["income"] != expected {
			t.Fatal("duplicated income", report["income"], expected)
		}
		dashboard := dataObject(t, h.call(t, admin, "GET", "/dashboard", nil, 200))
		monthExpected := h.scalar(t, `SELECT coalesce(sum(monto),0)::numeric(18,2)::text FROM jac_movimientos_pagos WHERE fecha >= date_trunc('month',statement_timestamp() AT TIME ZONE 'America/Lima') AT TIME ZONE 'America/Lima' AND fecha < (date_trunc('month',statement_timestamp() AT TIME ZONE 'America/Lima')+interval '1 month') AT TIME ZONE 'America/Lima'`)
		if dashboard["monthlyIncome"] != monthExpected {
			t.Fatal("dashboard multiplied income", dashboard["monthlyIncome"], monthExpected)
		}
		h.call(t, admin, "GET", "/reportes", nil, 200)
		h.call(t, admin, "GET", "/reportes?desde=2026-01-01T00:00:00Z&hasta=2026-02-01T00:00:00Z", nil, 200)
		h.call(t, admin, "GET", "/reportes?desde=incorrecto&hasta=incorrecto", nil, 422)
		h.call(t, limited, "GET", "/dashboard", nil, 403)
	})
	t.Run("último acceso administrativo transacción y concurrencia", func(t *testing.T) {
		// Remove the second admin first: a non-admin permission manager cannot remove the last.
		h.call(t, admin, "PATCH", "/usuarios/"+secondID, map[string]bool{"active": false}, 200)
		h.call(t, limited, "PUT", "/roles/"+adminRole+"/permisos", map[string]any{"permissions": map[string][]string{}}, 200)
		// Other test managers exist in this persistent cluster; clearing this
		// custom role is legitimate. Last Gerente is tested in an isolated database.
		h.call(t, limited, "PUT", "/roles/"+adminRole+"/permisos", map[string]any{"permissions": domain.PermissionPolicy().Defaults["Gerente"]}, 403)
		h.exec(t, `INSERT INTO rol_permisos(id_rol,id_permiso) SELECT $1,id_permiso FROM permisos ON CONFLICT DO NOTHING`, adminRole)
		h.call(t, admin, "PATCH", "/usuarios/"+secondID, map[string]bool{"active": true}, 200)
		// Re-authenticate second: deactivation invalidated the former JWT.
		res := h.send(workshopSession{}, "POST", "/auth/login", map[string]string{"username": "csecond_" + h.suffix, "password": h.password})
		if res.status != 200 {
			t.Fatalf("second login %d %s", res.status, res.raw)
		}
		var p struct {
			CSRF string `json:"csrfToken"`
		}
		decodeTest(t, res.raw, &p)
		second = workshopSession{res.cookies[0], p.CSRF}
		var wg sync.WaitGroup
		results := make(chan workshopResponse, 2)
		wg.Add(2)
		go func() {
			defer wg.Done()
			results <- h.send(admin, "PATCH", "/usuarios/"+secondID, map[string]bool{"active": false})
		}()
		go func() {
			defer wg.Done()
			results <- h.send(second, "PATCH", "/usuarios/"+adminID, map[string]bool{"active": false})
		}()
		wg.Wait()
		close(results)
		for result := range results {
			if result.err != nil {
				t.Fatal(result.err)
			}
			if result.status != 200 && result.status != 401 && result.status != 403 {
				t.Fatalf("concurrent admins %d %s", result.status, result.raw)
			}
		}
		if active := h.scalar(t, `SELECT count(*)::text FROM usuarios WHERE id_usuario IN ($1,$2) AND activo`, adminID, secondID); active != "1" {
			t.Fatal("administrative race", active)
		}
	})
	if tables := h.scalar(t, `SELECT count(*)::text FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'`); tables != "28" {
		t.Fatal(tables)
	}
	if fks := h.scalar(t, `SELECT count(*)::text FROM pg_constraint WHERE contype='f' AND connamespace='public'::regnamespace`); fks != "50" {
		t.Fatal(fks)
	}
}

func decodeTest(t *testing.T, b []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatal(err)
	}
}
