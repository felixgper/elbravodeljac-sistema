package sqlserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/application"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/mongodb"
)

type qaSession struct {
	client    *http.Client
	url, csrf string
}

func (s qaSession) send(method, path, contentType string, body io.Reader) (int, []byte, error) {
	r, err := http.NewRequest(method, s.url+path, body)
	if err != nil {
		return 0, nil, err
	}
	r.Header.Set("Origin", s.url)
	r.Header.Set("X-CSRF-Token", s.csrf)
	r.Header.Set("Content-Type", contentType)
	res, err := s.client.Do(r)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	return res.StatusCode, raw, err
}
func (s qaSession) request(method, path string, data any) (int, []byte, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return 0, nil, err
	}
	return s.send(method, path, "application/json", bytes.NewReader(raw))
}
func (s qaSession) call(t *testing.T, method, path string, data any, expected int) []byte {
	t.Helper()
	status, raw, err := s.request(method, "/api/v1"+path, data)
	if err != nil || status != expected {
		t.Fatalf("%s %s: status %d want %d: %s (%v)", method, path, status, expected, raw, err)
	}
	return raw
}
func qaJSON(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	return body
}
func qaData(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	body := qaJSON(t, raw)
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected object: %v", body)
	}
	return data
}
func qaID(t *testing.T, raw []byte) string {
	t.Helper()
	data := qaData(t, raw)
	id, ok := data["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected id: %v", data)
	}
	return id
}

func exerciseWorkshopHTTP(t *testing.T, s *Store, password, client, vehicle string) {
	t.Helper()
	var technical application.Technical
	if uri := os.Getenv("SQLSERVER_QA_MONGO_URI"); uri != "" {
		mg, err := mongodb.Open(uri, "elbravodeljac_tecnico_test")
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		t.Cleanup(func() { _ = mg.Close(context.Background()) })
		if err = mg.Init(ctx); err != nil {
			t.Fatal(err)
		}
		technical = mg
	}
	srv := httptest.NewUnstartedServer(nil)
	origin := "http://" + srv.Listener.Addr().String()
	h, err := httpapi.New(s, httpapi.Options{Origin: origin, Key: bytes.Repeat([]byte{19}, 32), Technical: technical})
	if err != nil {
		srv.Close()
		t.Fatal(err)
	}
	srv.Config.Handler = h
	srv.Start()
	defer srv.Close()
	session := func() qaSession {
		jar, _ := cookiejar.New(nil)
		return qaSession{client: &http.Client{Jar: jar, Timeout: 15 * time.Second}, url: srv.URL}
	}
	login := func(name string) qaSession {
		t.Helper()
		u := session()
		login := qaJSON(t, u.call(t, "POST", "/auth/login", map[string]string{"username": name, "password": password}, 200))
		u.csrf = login["csrfToken"].(string)
		u.call(t, "GET", "/auth/me", nil, 200)
		return u
	}
	anon := session()
	anon.call(t, "GET", "/usuarios", nil, 401)
	manager := login("gerente_qa")
	var managerRole, mechanicRole, adminRole string
	for _, role := range []struct {
		name string
		out  *string
	}{{"Gerente", &managerRole}, {"Mecánico", &mechanicRole}, {"Administrador", &adminRole}} {
		if err := s.db.QueryRow("SELECT CONVERT(varchar(20),id_rol) FROM roles WHERE nombre_rol=@p1", role.name).Scan(role.out); err != nil {
			t.Fatal(err)
		}
	}
	mechanicID := qaID(t, manager.call(t, "POST", "/usuarios", map[string]string{"name": "Mecánico QA", "username": "mecanico_qa", "password": password, "roleId": mechanicRole}, 201))
	manager.call(t, "POST", "/usuarios", map[string]string{"name": "Otro Mecánico QA", "username": "otro_qa", "password": password, "roleId": mechanicRole}, 201)
	manager.call(t, "POST", "/usuarios", map[string]string{"name": "Administrador QA", "username": "administrador_qa", "password": password, "roleId": adminRole}, 201)
	worker, other, admin := login("mecanico_qa"), login("otro_qa"), login("administrador_qa")
	worker.call(t, "GET", "/usuarios", nil, 403)
	csrf := manager
	csrf.csrf = ""
	csrf.call(t, "POST", "/recepciones", map[string]string{}, 403)
	for _, p := range []string{"/usuarios", "/empleados", "/roles", "/permisos", "/roles/" + managerRole + "/permisos", "/proveedores", "/auxilios", "/unidades-auxilio", "/servicios", "/repuestos", "/dashboard", "/reportes", "/recepciones", "/diagnosticos", "/cotizaciones", "/ordenes", "/cotizaciones/atenciones", "/diagnosticos/pendientes-cotizacion"} {
		manager.call(t, "GET", p, nil, 200)
	}
	for _, attention := range []string{"Directo", "Diagnostico"} {
		t.Run(attention, func(t *testing.T) {
			reception := qaID(t, manager.call(t, "POST", "/recepciones", map[string]string{"clientId": client, "vehicleId": vehicle, "mileage": "12", "attentionType": attention, "symptoms": "Ruido de prueba"}, 201))
			manager.call(t, "GET", "/recepciones/"+reception, nil, 200)
			quoteBody := map[string]any{"receptionId": reception, "status": "Pendiente", "items": []map[string]string{{"type": "Servicio", "concept": "Revisión sintética", "quantity": "1", "unitPrice": "100.25"}}}
			if attention == "Diagnostico" {
				manager.call(t, "POST", "/cotizaciones", quoteBody, 422)
				manager.call(t, "PATCH", "/recepciones/"+reception+"/mecanico", map[string]any{"technicianId": mechanicID, "version": 0}, 200)
				worker.call(t, "GET", "/me/diagnosticos", nil, 200)
				draftBody := map[string]any{"status": "Borrador", "receptionId": reception, "technicalDiagnosis": "Prueba técnica QA"}
				draft := qaData(t, worker.call(t, "POST", "/diagnosticos", draftBody, 201))
				diagnosis := draft["id"].(string)
				manager.call(t, "POST", "/cotizaciones", quoteBody, 422)
				other.call(t, "GET", "/diagnosticos/"+diagnosis, nil, 404)
				draftBody["status"] = "Finalizado"
				draftBody["version"] = draft["version"]
				draftBody["recommendations"] = "Revisión sintética autorizada"
				worker.call(t, "PUT", "/diagnosticos/"+diagnosis, draftBody, 200)
				worker.call(t, "PUT", "/diagnosticos/"+diagnosis, draftBody, 409)
				quoteBody["diagnosisId"] = diagnosis
				t.Run("MongoDB_GridFS", func(t *testing.T) {
					if technical == nil {
						t.Skip("requires isolated MongoDB QA")
					}
					base := "/diagnosticos/" + diagnosis
					scan := map[string]any{"requestId": "qa-scan", "equipo": map[string]string{"nombre": "Scanner QA", "version": "1"}, "dtcs": []map[string]string{{"codigo": "P0300"}}}
					id := qaID(t, worker.call(t, "POST", base+"/escaneres", scan, 201))
					if got := qaID(t, worker.call(t, "POST", base+"/escaneres", scan, 200)); got != id {
						t.Fatal("scan replay changed id")
					}
					other.call(t, "GET", base+"/escaneres", nil, 403)
					content := []byte("%PDF-1.4\n% synthetic QA evidence\n%%EOF\n")
					var body bytes.Buffer
					form := multipart.NewWriter(&body)
					if err := form.WriteField("requestId", "qa-gridfs"); err != nil {
						t.Fatal(err)
					}
					part, err := form.CreateFormFile("archivo", "evidencia.pdf")
					if err != nil {
						t.Fatal(err)
					}
					if _, err = part.Write(content); err != nil {
						t.Fatal(err)
					}
					if err = form.Close(); err != nil {
						t.Fatal(err)
					}
					status, raw, err := worker.send("POST", "/api/v1"+base+"/evidencias", form.FormDataContentType(), bytes.NewReader(body.Bytes()))
					if err != nil || status != 201 {
						t.Fatalf("evidence: %d %s %v", status, raw, err)
					}
					evidence := qaID(t, raw)
					if got := worker.call(t, "GET", base+"/evidencias/"+evidence+"/archivo", nil, 200); !bytes.Equal(got, content) {
						t.Fatal("GridFS file changed")
					}
				})
			}
			quote := qaData(t, manager.call(t, "POST", "/cotizaciones", quoteBody, 201))
			qid := quote["id"].(string)
			pdf := manager.call(t, "GET", "/cotizaciones/"+qid+"/pdf", nil, 200)
			if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
				t.Fatal("missing quote PDF")
			}
			orderBody := map[string]string{"quoteId": qid, "mechanicId": mechanicID}
			manager.call(t, "POST", "/ordenes", orderBody, 422)
			quoteBody["status"] = "Aprobada"
			quoteBody["version"] = quote["version"]
			quoteBody["approval"] = map[string]string{"name": "Cliente QA", "method": "WhatsApp"}
			approved := qaData(t, manager.call(t, "PUT", "/cotizaciones/"+qid, quoteBody, 200))
			if approved["status"] != "Aprobada" || approved["approval"] == nil {
				t.Fatalf("missing approval %v", approved)
			}
			order := qaID(t, manager.call(t, "POST", "/ordenes", orderBody, 201))
			orderPath := "/ordenes/" + order
			worker.call(t, "GET", orderPath, nil, 200)
			other.call(t, "GET", orderPath, nil, 404)
			worker.call(t, "GET", orderPath+"/pagos", nil, 403)
			manager.call(t, "PATCH", orderPath, map[string]string{"state": "En reparación", "note": "Servicio QA"}, 200)
			line := approved["items"].([]any)[0].(map[string]any)["id"].(string)
			control := map[string]any{"version": 0, "execution": map[string]any{line: map[string]string{"status": "Realizado"}}}
			worker.call(t, "PUT", orderPath+"/control-tecnico", control, 200)
			worker.call(t, "PUT", orderPath+"/control-tecnico", control, 409)
			worker.call(t, "PATCH", orderPath, map[string]string{"state": "Finalizada", "note": "Prueba QA", "test": "Ruta", "result": "Satisfactoria"}, 200)
			delivery := map[string]any{"kind": "Titular / representante", "recipientKind": "Propietario o titular", "method": "Documento", "verified": true, "conformity": true}
			manager.call(t, "POST", orderPath+"/entrega", delivery, 422)
			payment := func(id, amount string) map[string]string {
				return map[string]string{"requestId": id, "amount": amount, "method": "Efectivo"}
			}
			pay := payment("partial-"+order, "60")
			type result struct {
				status int
				raw    []byte
				err    error
			}
			start := make(chan struct{})
			results := make(chan result, 2)
			for range 2 {
				go func() {
					<-start
					st, b, e := manager.request("POST", "/api/v1"+orderPath+"/pagos", pay)
					results <- result{st, b, e}
				}()
			}
			close(start)
			one, two := <-results, <-results
			for _, v := range []result{one, two} {
				if v.err != nil || v.status != 201 {
					t.Fatalf("concurrent payment: %d %s %v", v.status, v.raw, v.err)
				}
			}
			if qaID(t, one.raw) != qaID(t, two.raw) {
				t.Fatal("duplicate payment after concurrent retry")
			}
			balance := qaData(t, manager.call(t, "GET", orderPath+"/pagos", nil, 200))
			if balance["balance"] != "40.25" {
				t.Fatalf("unexpected balance %v", balance)
			}
			manager.call(t, "POST", orderPath+"/pagos", payment("excess-"+order, "40.26"), 422)
			manager.call(t, "POST", orderPath+"/entrega", delivery, 422)
			finalID := qaID(t, manager.call(t, "POST", orderPath+"/pagos", payment("final-"+order, "40.25"), 201))
			manager.call(t, "POST", orderPath+"/entrega", delivery, 201)
			first := manager.call(t, "GET", orderPath+"/ficha-tecnica.pdf", nil, 200)
			if !bytes.HasPrefix(first, []byte("%PDF-")) {
				t.Fatal("missing delivery PDF")
			}
			fix := map[string]any{"reason": "Corregir método de pago de QA", "requestId": "replace-" + order, "replacement": map[string]string{"method": "Tarjeta", "comprobante": "RECT-QA"}}
			manager.call(t, "PATCH", orderPath+"/pagos/"+finalID, fix, 200)
			manager.call(t, "PATCH", orderPath+"/pagos/"+finalID, fix, 200)
			if second := manager.call(t, "GET", orderPath+"/ficha-tecnica.pdf", nil, 200); !bytes.Equal(first, second) {
				t.Fatal("archived PDF changed")
			}
			if got := qaData(t, manager.call(t, "GET", orderPath+"/pagos", nil, 200))["balance"]; got != "0.00" {
				t.Fatal(fmt.Sprint(got))
			}
			manager.call(t, "GET", "/dashboard", nil, 200)
			manager.call(t, "GET", "/reportes", nil, 200)
		})
	}
	admin.call(t, "GET", "/dashboard", nil, 200)
}
