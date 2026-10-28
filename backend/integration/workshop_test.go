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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
)

type workshopSession struct {
	cookie *http.Cookie
	csrf   string
}
type workshopHarness struct {
	origin   string
	pool     *pgxpool.Pool
	api      *httptest.Server
	ctx      context.Context
	password string
	suffix   string
}
type workshopResponse struct {
	status  int
	raw     []byte
	cookies []*http.Cookie
	err     error
}

func (h *workshopHarness) send(session workshopSession, method, path string, payload any) workshopResponse {
	var b []byte
	var err error
	if payload != nil {
		// Existing fixtures now explicitly supply the newer business metadata.
		// Audit-specific negative tests pass empty/null fields to exercise rejection.
		fixture, _ := json.Marshal(payload)
		var fields map[string]any
		json.Unmarshal(fixture, &fields)
		if method == "POST" && strings.HasSuffix(path, "/pagos") {
			if _, ok := fields["requestId"]; !ok {
				fields["requestId"] = rand.Text()
			}
		}
		if method == "POST" && strings.HasSuffix(path, "/entrega") {
			if _, ok := fields["conformity"]; !ok {
				fields["conformity"] = true
			}
			if _, ok := fields["recipientKind"]; !ok {
				if fields["kind"] == "Tercero" {
					fields["recipientKind"] = "Persona autorizada"
				} else {
					fields["recipientKind"] = "Propietario o titular"
				}
			}
		}
		if (method == "POST" || method == "PUT") && (path == "/cotizaciones" || strings.HasPrefix(path, "/cotizaciones/")) {
			if fields["status"] == "Aprobada" {
				if _, ok := fields["approval"]; !ok {
					fields["approval"] = map[string]string{"name": "Cliente de la prueba aislada", "method": "Presencial"}
				}
			}
			if method == "PUT" {
				if _, ok := fields["version"]; !ok {
					var version int64
					_ = h.pool.QueryRow(h.ctx, `SELECT version FROM cotizaciones WHERE id_cotizacion=$1`, strings.TrimPrefix(path, "/cotizaciones/")).Scan(&version)
					fields["version"] = version
				}
			}
		}
		payload = fields
		b, err = json.Marshal(payload)
		if err != nil {
			return workshopResponse{err: err}
		}
	}
	r, err := http.NewRequest(method, h.api.URL+"/api/v1"+path, bytes.NewReader(b))
	if err != nil {
		return workshopResponse{err: err}
	}
	r.Header.Set("Content-Type", "application/json")
	origin := h.origin
	if origin == "" {
		origin = "http://localhost:8080"
	}
	r.Header.Set("Origin", origin)
	r.Header.Set("X-CSRF-Token", session.csrf)
	if session.cookie != nil {
		r.AddCookie(session.cookie)
	}
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return workshopResponse{err: err}
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	return workshopResponse{res.StatusCode, raw, res.Cookies(), err}
}
func (h *workshopHarness) call(t *testing.T, s workshopSession, method, path string, payload any, status int) map[string]any {
	t.Helper()
	res := h.send(s, method, path, payload)
	if res.err != nil {
		t.Fatal(res.err)
	}
	if res.status != status {
		t.Fatalf("%s %s: HTTP %d, esperado %d: %s", method, path, res.status, status, res.raw)
	}
	if status == 204 {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(res.raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}
func dataObject(t *testing.T, response map[string]any) map[string]any {
	t.Helper()
	out, ok := response["data"].(map[string]any)
	if !ok {
		t.Fatal("response missing data object", response)
	}
	return out
}
func responseID(t *testing.T, response map[string]any) string {
	t.Helper()
	id, ok := dataObject(t, response)["id"].(string)
	if !ok || id == "" {
		t.Fatal("missing string ID", response)
	}
	return id
}
func (h *workshopHarness) exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	if _, err := h.pool.Exec(h.ctx, sql, args...); err != nil {
		t.Fatal(err)
	}
}
func (h *workshopHarness) scalar(t *testing.T, sql string, args ...any) string {
	t.Helper()
	var v string
	if err := h.pool.QueryRow(h.ctx, sql, args...).Scan(&v); err != nil {
		t.Fatal(err)
	}
	return v
}
func (h *workshopHarness) user(t *testing.T, label, role string, permissions []string) (string, workshopSession) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(h.password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	username := label + "_" + h.suffix
	id := h.scalar(t, `INSERT INTO usuarios(nombre,usuario,password_hash) VALUES($1,$2,$3) RETURNING id_usuario::text`, label, username, string(hash))
	h.exec(t, `INSERT INTO roles(nombre_rol) VALUES($1) ON CONFLICT(nombre_rol) DO NOTHING`, role)
	roleID := h.scalar(t, `SELECT id_rol::text FROM roles WHERE nombre_rol=$1`, role)
	h.exec(t, `INSERT INTO usuario_roles(id_usuario,id_rol) VALUES($1,$2)`, id, roleID)
	for _, permission := range permissions {
		h.exec(t, `INSERT INTO permisos(codigo) VALUES($1) ON CONFLICT(codigo) DO NOTHING`, permission)
		h.exec(t, `INSERT INTO rol_permisos(id_rol,id_permiso) SELECT $1,id_permiso FROM permisos WHERE codigo=$2 ON CONFLICT DO NOTHING`, roleID, permission)
	}
	res := h.send(workshopSession{}, "POST", "/auth/login", map[string]string{"username": username, "password": h.password})
	if res.err != nil || res.status != 200 {
		t.Fatalf("fixture login: %d %s %v", res.status, res.raw, res.err)
	}
	var parsed struct {
		CSRF string `json:"csrfToken"`
	}
	if err = json.Unmarshal(res.raw, &parsed); err != nil {
		t.Fatal(err)
	}
	return id, workshopSession{res.cookies[0], parsed.CSRF}
}
func TestWorkshopHTTPPostgreSQL(t *testing.T) {
	dsn := os.Getenv("INTEGRATION_DATABASE_URL")
	if dsn == "" {
		t.Skip("No se ejecutó integración: falta INTEGRATION_DATABASE_URL")
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
	permissions := []string{"diagnosticos.asignar", "diagnosticos.supervisar", "clientes.ver", "clientes.crear", "clientes.editar", "vehiculos.ver", "vehiculos.crear", "vehiculos.editar", "recepciones.ver", "recepciones.crear", "diagnosticos.ver", "diagnosticos.crear", "cotizaciones.ver", "cotizaciones.crear", "cotizaciones.editar", "ordenes.ver", "ordenes.crear", "ordenes.editar", "ordenes.eliminar", "ordenes.asignar", "mecanicos.ver", "auxilio.ver", "pagos.ver", "pagos.registrar", "pagos.entregar"}
	_, admin := h.user(t, "manager", "Manager_"+h.suffix, permissions)
	mechanic, mechanicSession := h.user(t, "mechanic", "Mecánico", []string{"ordenes.ver", "ordenes.editar"})
	mechanic2, mechanic2Session := h.user(t, "technician", "Supervisor", []string{"ordenes.ver", "ordenes.editar"})
	aux, auxSession := h.user(t, "helper", "Auxiliar", []string{"auxilio.ver"})
	client := responseID(t, h.call(t, admin, "POST", "/clientes", map[string]any{"name": "Fixture taller", "documentType": "Pasaporte", "documentNumber": "W" + strings.ToUpper(h.suffix), "phone": "+51987654321", "address": "Lima"}, 201))
	otherClient := responseID(t, h.call(t, admin, "POST", "/clientes", map[string]any{"name": "Otro fixture", "documentType": "Pasaporte", "documentNumber": "X" + strings.ToUpper(h.suffix), "phone": "+51987654321", "address": "Lima"}, 201))
	vehicle := responseID(t, h.call(t, admin, "POST", "/vehiculos", map[string]any{"clientId": client, "plate": "W-" + strings.ToUpper(h.suffix[:6]), "brand": "JAC", "model": "T8", "year": 2024, "mileage": "10.00"}, 201))
	service := h.scalar(t, `INSERT INTO servicios(nombre,precio_referencia) VALUES($1,120) RETURNING id_servicio::text`, "Servicio fixture "+h.suffix)
	part := h.scalar(t, `INSERT INTO repuestos(nombre,precio_referencia) VALUES($1,35) RETURNING id_repuesto::text`, "Repuesto fixture "+h.suffix)
	inactive := h.scalar(t, `INSERT INTO servicios(nombre,precio_referencia,activo) VALUES($1,10,false) RETURNING id_servicio::text`, "Inactivo fixture "+h.suffix)
	var reception, diagnosis, quote, order string
	quoteBody := func(state string) map[string]any {
		return map[string]any{"diagnosisId": diagnosis, "status": state, "items": []map[string]any{{"serviceId": service, "quantity": "1", "unitPrice": "120.00"}, {"sparePartId": part, "origin": "Taller", "quantity": "1", "unitPrice": "35.00"}, {"sparePartId": part, "origin": "Cliente", "quantity": "1", "unitPrice": "0.00"}}}
	}
	newOrder := func(t *testing.T) string {
		t.Helper()
		rx := responseID(t, h.call(t, admin, "POST", "/recepciones", map[string]string{"clientId": client, "vehicleId": vehicle, "mileage": "11.25", "attentionType": "Directo", "symptoms": "Otra atención para probar una orden independiente"}, 201))
		body := quoteBody("Aprobada")
		delete(body, "diagnosisId")
		body["receptionId"] = rx
		qid := responseID(t, h.call(t, admin, "POST", "/cotizaciones", body, 201))
		return responseID(t, h.call(t, admin, "POST", "/ordenes", map[string]string{"quoteId": qid, "mechanicId": mechanic}, 201))
	}
	t.Run("recepciones y diagnóstico con restricciones reales", func(t *testing.T) {
		body := map[string]string{"clientId": otherClient, "vehicleId": vehicle, "mileage": "11.25", "symptoms": "Ruido"}
		h.call(t, admin, "POST", "/recepciones", body, 422)
		body["clientId"] = client
		body["mileage"] = "9.00"
		h.call(t, admin, "POST", "/recepciones", body, 422)
		body["mileage"] = "11.25"
		reception = responseID(t, h.call(t, admin, "POST", "/recepciones", body, 201))
		if km := h.scalar(t, `SELECT kilometraje_actual::text FROM vehiculos WHERE id_vehiculo=$1`, vehicle); km != "11.25" {
			t.Fatal(km)
		}
		h.call(t, admin, "GET", "/recepciones?pendienteDiagnostico=true", nil, 200)
		h.call(t, admin, "GET", "/recepciones/"+reception, nil, 200)
		h.call(t, admin, "PATCH", "/recepciones/"+reception+"/mecanico", map[string]any{"technicianId": mechanic, "version": 0}, 200)
		diagnosis = responseID(t, h.call(t, admin, "POST", "/diagnosticos", map[string]string{"status": "Finalizado", "evaluation": "Evaluación técnica de prueba", "recommendations": "Servicios recomendados de prueba", "receptionId": reception, "findings": "Frenos y mantenimiento"}, 201))
		h.call(t, admin, "POST", "/diagnosticos", map[string]string{"receptionId": reception, "findings": "Duplicado"}, 409)
		h.call(t, admin, "POST", "/diagnosticos", map[string]string{"receptionId": "9223372036854775807", "findings": "Sin recepción"}, 404)
		h.call(t, admin, "GET", "/diagnosticos?recepcionId="+reception, nil, 200)
		h.call(t, admin, "GET", "/diagnosticos/"+diagnosis, nil, 200)
	})
	if diagnosis == "" {
		t.Fatal("No se creó diagnóstico")
	}
	t.Run("cotizaciones catálogos y rollback de detalles", func(t *testing.T) {
		for _, path := range []string{"/servicios", "/repuestos"} {
			h.call(t, admin, "GET", path, nil, 200)
		}
		bad := quoteBody("Pendiente")
		bad["items"] = []map[string]string{{"concept": "Libre sin referencia", "quantity": "1", "unitPrice": "10"}}
		h.call(t, admin, "POST", "/cotizaciones", bad, 422)
		bad["items"] = []map[string]string{{"serviceId": inactive, "quantity": "1", "unitPrice": "10"}}
		h.call(t, admin, "POST", "/cotizaciones", bad, 422)
		quote = responseID(t, h.call(t, admin, "POST", "/cotizaciones", quoteBody("Pendiente"), 201))
		h.call(t, admin, "POST", "/ordenes", map[string]string{"quoteId": quote, "mechanicId": mechanic}, 422)
		good := h.call(t, admin, "PUT", "/cotizaciones/"+quote, quoteBody("Aprobada"), 200)
		if dataObject(t, good)["total"] != "155.00" {
			t.Fatal(good)
		}
		bad["items"] = []map[string]string{{"serviceId": service, "quantity": "1", "unitPrice": "2"}, {"serviceId": "9223372036854775807", "quantity": "1", "unitPrice": "1"}}
		h.call(t, admin, "PUT", "/cotizaciones/"+quote, bad, 422)
		if total := h.scalar(t, `SELECT total::text FROM cotizaciones WHERE id_cotizacion=$1`, quote); total != "155.00" {
			t.Fatal("rollback total", total)
		}
		if count := h.scalar(t, `SELECT count(*)::text FROM detalle_cotizacion WHERE id_cotizacion=$1`, quote); count != "3" {
			t.Fatal("rollback lines", count)
		}
		h.call(t, admin, "GET", "/cotizaciones?estado=Aprobada&sinOrden=true", nil, 200)
		h.call(t, admin, "GET", "/cotizaciones/"+quote, nil, 200)
	})
	if quote == "" {
		t.Fatal("No se creó cotización")
	}
	t.Run("creación OT asignaciones y códigos", func(t *testing.T) {
		h.call(t, admin, "POST", "/ordenes", map[string]string{"quoteId": quote, "mechanicId": aux}, 422)
		order = responseID(t, h.call(t, admin, "POST", "/ordenes", map[string]string{"quoteId": quote, "mechanicId": mechanic, "assistantId": aux}, 201))
		h.call(t, admin, "POST", "/ordenes", map[string]string{"quoteId": quote, "mechanicId": mechanic}, 409)
		h.call(t, admin, "PUT", "/cotizaciones/"+quote, quoteBody("Rechazada"), 409)
		h.call(t, admin, "GET", "/ordenes?estado=Pendiente", nil, 200)
		h.call(t, admin, "GET", "/ordenes/"+order, nil, 200)
		h.call(t, admin, "GET", "/mecanicos", nil, 200)
		h.call(t, admin, "GET", "/asignaciones?ordenId="+order, nil, 200)
		h.call(t, auxSession, "GET", "/me/tareas", nil, 200)
		h.call(t, admin, "POST", "/ordenes/"+order+"/pagos", map[string]string{"amount": "9999", "method": "Efectivo"}, 422)
		pristine := newOrder(t)
		h.call(t, admin, "DELETE", "/ordenes/"+pristine, nil, 204)
		h.call(t, admin, "GET", "/ordenes/"+pristine, nil, 404)
	})
	if order == "" {
		t.Fatal("No se creó OT")
	}
	t.Run("permisos por objeto avances pruebas y finalización", func(t *testing.T) {
		h.call(t, mechanic2Session, "GET", "/ordenes/"+order, nil, 404)
		h.call(t, mechanicSession, "PATCH", "/ordenes/"+order, map[string]string{"state": "En reparación", "note": "No puede reasignar", "mechanicId": mechanic2}, 403)
		h.call(t, admin, "PATCH", "/ordenes/"+order, map[string]string{"state": "En reparación", "note": "Reasignación", "mechanicId": mechanic2}, 200)
		h.call(t, mechanicSession, "PATCH", "/ordenes/"+order, map[string]string{"state": "En reparación", "note": "Ya no está asignado"}, 403)
		h.call(t, mechanic2Session, "PATCH", "/ordenes/"+order, map[string]string{"state": "En reparación", "note": "Revisión"}, 200)
		before := h.scalar(t, `SELECT count(*)::text FROM avances_ot WHERE id_orden=$1`, order)
		h.call(t, admin, "PATCH", "/ordenes/"+order, map[string]string{"state": "Finalizada", "note": "Sin prueba"}, 422)
		if after := h.scalar(t, `SELECT count(*)::text FROM avances_ot WHERE id_orden=$1`, order); after != before {
			t.Fatal("failed finalize persisted advance")
		}
		h.call(t, admin, "PATCH", "/ordenes/"+order, map[string]string{"state": "En pruebas", "note": "Ruta fallida", "test": "Ruta", "result": "Fallida"}, 200)
		h.call(t, admin, "PATCH", "/ordenes/"+order, map[string]string{"state": "Finalizada", "note": "No puede finalizar"}, 422)
		h.call(t, admin, "PATCH", "/ordenes/"+order, map[string]string{"state": "En pruebas", "note": "Ruta correcta", "test": "Ruta", "result": "Satisfactoria"}, 200)
		h.call(t, admin, "PATCH", "/ordenes/"+order, map[string]string{"state": "En reparación", "note": "Reparación posterior"}, 200)
		h.call(t, admin, "PATCH", "/ordenes/"+order, map[string]string{"state": "Finalizada", "note": "Prueba antigua"}, 422)
		h.call(t, admin, "PATCH", "/ordenes/"+order, map[string]string{"state": "Finalizada", "note": "Final revisado", "test": strings.Repeat("x", 101), "result": "Satisfactoria"}, 422)
		h.confirmExecution(t, admin, order)
		h.call(t, admin, "PATCH", "/ordenes/"+order, map[string]string{"state": "Finalizada", "note": "Final revisado", "test": "Ruta final", "result": "Satisfactoria"}, 200)
		h.call(t, admin, "PATCH", "/ordenes/"+order, map[string]string{"state": "En reparación", "note": "Reabrir"}, 409)
		h.call(t, admin, "DELETE", "/ordenes/"+order, nil, 409)
	})
	delivery := map[string]any{"kind": "Titular / representante", "method": "Documento", "verified": true}
	t.Run("pagos concurrentes y saldo exacto", func(t *testing.T) {
		h.call(t, admin, "POST", "/ordenes/"+order+"/entrega", delivery, 422)
		var wg sync.WaitGroup
		results := make(chan workshopResponse, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				results <- h.send(admin, "POST", "/ordenes/"+order+"/pagos", map[string]string{"amount": "100", "method": "Efectivo"})
			}()
		}
		wg.Wait()
		close(results)
		success, rejected := 0, 0
		for result := range results {
			if result.err != nil {
				t.Fatal(result.err)
			}
			switch result.status {
			case 201:
				success++
			case 422:
				rejected++
			default:
				t.Fatalf("payment concurrency %d %s", result.status, result.raw)
			}
		}
		if success != 1 || rejected != 1 {
			t.Fatal(success, rejected)
		}
		h.call(t, admin, "POST", "/ordenes/"+order+"/pagos", map[string]string{"amount": "55.01", "method": "Efectivo"}, 422)
		h.call(t, admin, "POST", "/ordenes/"+order+"/pagos", map[string]string{"amount": "55.00", "method": "Transferencia", "comprobante": "QA"}, 201)
		res := h.call(t, admin, "GET", "/ordenes/"+order+"/pagos", nil, 200)
		if dataObject(t, res)["balance"] != "0.00" {
			t.Fatal(res)
		}
	})
	t.Run("entrega contacto autorizado snapshot y concurrencia", func(t *testing.T) {
		h.call(t, admin, "GET", "/ordenes/"+order+"/entrega", nil, 404)
		contactBody := map[string]string{"name": "Receptor fixture", "documentType": "DNI", "documentNumber": "45678123", "phone": "+51987654321", "relation": "Familiar"}
		wrong := responseID(t, h.call(t, admin, "POST", "/clientes/"+otherClient+"/contactos", contactBody, 201))
		right := responseID(t, h.call(t, admin, "POST", "/clientes/"+client+"/contactos", contactBody, 201))
		delivery["kind"] = "Tercero"
		delivery["contactId"] = wrong
		h.call(t, admin, "POST", "/ordenes/"+order+"/entrega", delivery, 422)
		delivery["contactId"] = right
		delivery["verified"] = false
		h.call(t, admin, "POST", "/ordenes/"+order+"/entrega", delivery, 422)
		delivery["verified"] = true
		// Historical client of the reception remains authoritative after ownership transfer.
		h.call(t, admin, "POST", "/vehiculos/"+vehicle+"/transferencias", map[string]any{"newClientId": otherClient, "transferVerified": true}, 200)
		var wg sync.WaitGroup
		results := make(chan workshopResponse, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); results <- h.send(admin, "POST", "/ordenes/"+order+"/entrega", delivery) }()
		}
		wg.Wait()
		close(results)
		success, rejected := 0, 0
		for result := range results {
			if result.err != nil {
				t.Fatal(result.err)
			}
			switch result.status {
			case 201:
				success++
			case 409:
				rejected++
			default:
				t.Fatalf("delivery concurrency %d %s", result.status, result.raw)
			}
		}
		if success != 1 || rejected != 1 {
			t.Fatal(success, rejected)
		}
		h.exec(t, `UPDATE contactos_cliente SET nombre='Cambiado después' WHERE id_contacto=$1`, right)
		res := h.call(t, admin, "GET", "/ordenes/"+order+"/entrega", nil, 200)
		if dataObject(t, res)["name"] != "Receptor fixture" {
			t.Fatal("snapshot lost", res)
		}
		h.call(t, admin, "POST", "/ordenes/"+order+"/pagos", map[string]string{"amount": "1", "method": "Efectivo"}, 422)
		h.call(t, admin, "DELETE", "/ordenes/"+order, nil, 409)
	})
	t.Run("JWT permisos y CSRF obligatorios", func(t *testing.T) {
		for _, path := range []string{"/recepciones", "/diagnosticos", "/cotizaciones", "/ordenes", "/servicios", "/mecanicos", "/ordenes/" + order + "/pagos", "/ordenes/" + order + "/entrega"} {
			h.call(t, workshopSession{}, "GET", path, nil, 401)
		}
		h.call(t, mechanicSession, "POST", "/recepciones", map[string]string{"clientId": client, "vehicleId": vehicle, "mileage": "20", "symptoms": "Sin permiso"}, 403)
		invalid := admin
		invalid.csrf = "incorrecto"
		h.call(t, invalid, "POST", "/cotizaciones", quoteBody("Aprobada"), 403)
		if got := h.scalar(t, `SELECT count(*)::text FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'`); got != "26" {
			t.Fatal(got)
		}
		if got := h.scalar(t, `SELECT count(*)::text FROM pg_constraint WHERE contype='f' AND connamespace='public'::regnamespace`); got != "47" {
			t.Fatal(got)
		}
	})
}

func (h *workshopHarness) confirmExecution(t *testing.T, session workshopSession, order string) {
	record := dataObject(t, h.call(t, session, "GET", "/ordenes/"+order+"/control-tecnico", nil, 200))["record"].(map[string]any)
	var quoteID string
	if err := h.pool.QueryRow(h.ctx, `SELECT id_cotizacion::text FROM ordenes_trabajo WHERE id_orden=$1`, order).Scan(&quoteID); err != nil {
		t.Fatal(err)
	}
	quote := dataObject(t, h.call(t, session, "GET", "/cotizaciones/"+quoteID, nil, 200))
	execution := map[string]any{}
	for _, item := range quote["items"].([]any) {
		execution[item.(map[string]any)["id"].(string)] = map[string]string{"status": "Realizado", "notes": "Ejecución confirmada en fixture aislado"}
	}
	control := record["control"].(map[string]any)
	control["version"] = record["version"]
	control["execution"] = execution
	h.call(t, session, "PUT", "/ordenes/"+order+"/control-tecnico", control, 200)
}
