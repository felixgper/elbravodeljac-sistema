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
	"testing"
)

func TestAssignedDiagnosisWorkflow(t *testing.T) {
	dsn := os.Getenv("INTEGRATION_DATABASE_URL")
	if dsn == "" {
		t.Skip("Requiere PostgreSQL aislado")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_test" {
		t.Fatal("Solo pruebas aisladas")
	}
	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	store := &postgres.Store{Pool: pool}
	handler, err := httpapi.New(store, httpapi.Options{Origin: "http://localhost:8080", Key: []byte(rand.Text() + rand.Text())})
	if err != nil {
		t.Fatal(err)
	}
	api := httptest.NewServer(handler)
	defer api.Close()
	h := workshopHarness{pool: pool, api: api, ctx: ctx, password: rand.Text(), suffix: rand.Text()[:8]}
	all := []string{}
	for m, actions := range domain.PermissionPolicy().Modules {
		if m == "usuarios" {
			continue
		}
		for _, a := range actions {
			all = append(all, m+"."+a)
		}
	}
	_, admin := h.user(t, "dispatch", "Dispatch_"+h.suffix, all)
	first, m1 := h.user(t, "diagnoser1", "Mecánico", []string{"inicio.ver", "recepciones.ver", "diagnosticos.ver", "diagnosticos.crear", "ordenes.ver", "ordenes.editar"})
	second, m2 := h.user(t, "diagnoser2", "Mecánico", []string{"inicio.ver", "recepciones.ver", "diagnosticos.ver", "diagnosticos.crear"})
	_, aux := h.user(t, "diagaux", "DiagAux_"+h.suffix, []string{"diagnosticos.ver", "diagnosticos.crear"})
	inactive, _ := h.user(t, "diaginactive", "Mecánico", []string{"diagnosticos.ver"})
	h.exec(t, `UPDATE usuarios SET activo=false WHERE id_usuario=$1`, inactive)
	client := responseID(t, h.call(t, admin, "POST", "/clientes", map[string]string{"name": "Félix Fernández " + h.suffix, "documentType": "Pasaporte", "documentNumber": "DW" + h.suffix, "phone": "+51987654321", "address": "Prueba aislada"}, 201))
	car := responseID(t, h.call(t, admin, "POST", "/vehiculos", map[string]any{"clientId": client, "plate": "DW" + h.suffix[:5], "brand": "Honda", "model": "Prueba", "year": 2025, "mileage": "1"}, 201))
	body := func(tech string) map[string]string {
		return map[string]string{"clientId": client, "vehicleId": car, "mileage": "1", "symptoms": "Evaluación autorizada", "attentionType": "Diagnostico", "technicianId": tech}
	}
	rx := responseID(t, h.call(t, admin, "POST", "/recepciones", body(first), 201))
	other := responseID(t, h.call(t, admin, "POST", "/recepciones", body(second), 201))
	h.call(t, admin, "POST", "/recepciones", body(inactive), 422)
	h.call(t, admin, "POST", "/recepciones", body("9223372036854775807"), 422)
	queue := h.call(t, m1, "GET", "/me/diagnosticos", nil, 200)["data"].([]any)
	if len(queue) != 1 || queue[0].(map[string]any)["reception"].(map[string]any)["id"] != rx {
		t.Fatal("Other worker data leaked", queue)
	}
	h.call(t, m1, "GET", "/recepciones/"+other, nil, 404)
	h.call(t, m1, "PATCH", "/recepciones/"+other+"/mecanico", map[string]any{"technicianId": first, "version": 1}, 403)
	draft := map[string]any{"receptionId": rx, "findings": "Hallazgo guardado", "evaluation": "En revisión", "recommendations": "Pendiente de confirmar", "notes": "Borrador real"}
	saved := dataObject(t, h.call(t, m1, "POST", "/diagnosticos", draft, 201))
	id := saved["id"].(string)
	if saved["status"] != "Borrador" || saved["finishedAt"] != nil {
		t.Fatal(saved)
	}
	h.call(t, m2, "GET", "/diagnosticos/"+id, nil, 404)
	h.call(t, aux, "GET", "/diagnosticos/"+id, nil, 404)
	h.call(t, m1, "GET", "/frontend/opciones?module=diagnosticos&resource=recepciones", nil, 200)
	h.call(t, admin, "POST", "/cotizaciones", map[string]any{"diagnosisId": id, "status": "Aprobada", "items": []map[string]string{{"type": "Servicio", "concept": "No cotizar borrador", "quantity": "1", "unitPrice": "10"}}}, 422)
	draft["version"] = 1
	h.call(t, m2, "PUT", "/diagnosticos/"+id, draft, 403)
	h.call(t, aux, "PUT", "/diagnosticos/"+id, draft, 403)
	h.call(t, admin, "PATCH", "/recepciones/"+rx+"/mecanico", map[string]any{"technicianId": second, "version": 1}, 200)
	h.call(t, m1, "GET", "/diagnosticos/"+id, nil, 404)
	h.call(t, m1, "PUT", "/diagnosticos/"+id, draft, 403)
	reassigned := dataObject(t, h.call(t, m2, "GET", "/diagnosticos/"+id, nil, 200))
	if reassigned["findings"] != "Hallazgo guardado" || reassigned["technicianId"] != second {
		t.Fatal("Draft lost on reassignment", reassigned)
	}
	draft["version"] = 1
	h.call(t, m2, "PUT", "/diagnosticos/"+id, draft, 409)
	draft["version"] = reassigned["version"]
	draft["status"] = "Finalizado"
	draft["evaluation"] = ""
	h.call(t, m2, "PUT", "/diagnosticos/"+id, draft, 422)
	draft["evaluation"] = "Evaluación completada"
	draft["recommendations"] = "Servicio realmente recomendado"
	final := dataObject(t, h.call(t, m2, "PUT", "/diagnosticos/"+id, draft, 200))
	if final["finishedAt"] == nil || final["technicianId"] != second {
		t.Fatal(final)
	}
	h.call(t, admin, "PATCH", "/recepciones/"+rx+"/mecanico", map[string]any{"technicianId": first, "version": 2}, 409)
	h.call(t, m2, "PUT", "/diagnosticos/"+id, draft, 409)
	ready := h.call(t, admin, "GET", "/diagnosticos/pendientes-cotizacion", nil, 200)["data"].([]any)
	found := false
	for _, row := range ready {
		if row.(map[string]any)["diagnosis"].(map[string]any)["id"] == id {
			found = true
		}
	}
	if !found {
		t.Fatal("Final diagnosis missing from admin inbox")
	}
	quoteBody := map[string]any{"diagnosisId": id, "status": "Aprobada", "items": []map[string]string{{"type": "Servicio", "concept": "Servicio autorizado", "quantity": "1", "unitPrice": "10"}}}
	quote := responseID(t, h.call(t, admin, "POST", "/cotizaciones", quoteBody, 201))
	ready = h.call(t, admin, "GET", "/diagnosticos/pendientes-cotizacion", nil, 200)["data"].([]any)
	for _, row := range ready {
		if row.(map[string]any)["diagnosis"].(map[string]any)["id"] == id {
			t.Fatal("Already quoted diagnosis remains pending")
		}
	}
	responseID(t, h.call(t, admin, "POST", "/ordenes", map[string]string{"quoteId": quote, "mechanicId": first}, 201))
	if dataObject(t, h.call(t, admin, "GET", "/diagnosticos/"+id, nil, 200))["technicianId"] != second {
		t.Fatal("OT mechanic overwrote diagnostic mechanic")
	}
	direct := body("")
	direct["attentionType"] = "Directo"
	directRx := responseID(t, h.call(t, admin, "POST", "/recepciones", direct, 201))
	delete(quoteBody, "diagnosisId")
	quoteBody["receptionId"] = directRx
	h.call(t, admin, "POST", "/cotizaciones", quoteBody, 201)
	if h.scalar(t, `SELECT count(*)::text FROM diagnosticos WHERE id_recepcion=$1`, directRx) != "0" {
		t.Fatal("Fictitious diagnosis")
	}
	h.call(t, admin, "POST", "/cotizaciones", quoteBody, 422)
	h.call(t, m1, "GET", "/cotizaciones/atenciones", nil, 403)
	for _, query := range []string{h.suffix, "felix%20fernandez", "DW" + h.suffix, "DW" + h.suffix[:5]} {
		rows := h.call(t, admin, "GET", "/cotizaciones/atenciones?q="+query+"&recepcionId="+rx, nil, 200)["data"].([]any)
		if len(rows) != 1 || rows[0].(map[string]any)["id"] != rx {
			t.Fatal("Normalized remote lookup", query, rows)
		}
	}
	firstPage := h.call(t, admin, "GET", "/cotizaciones/atenciones?q="+h.suffix+"&pageSize=1", nil, 200)["data"].([]any)
	nextPage := h.call(t, admin, "GET", "/cotizaciones/atenciones?q="+h.suffix+"&pageSize=1&page=2", nil, 200)["data"].([]any)
	if len(firstPage) != 1 || len(nextPage) != 1 || firstPage[0].(map[string]any)["id"] == nextPage[0].(map[string]any)["id"] {
		t.Fatal("Lookup pagination")
	}
	t.Log("Assignment, two-worker scope, draft, finalization, reassignment history, quote inbox and direct workflow verified")
}
