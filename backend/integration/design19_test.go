package integration

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestDesign19DiagnosisCompatibilityAndFinance(t *testing.T) {
	dsn := os.Getenv("INTEGRATION_DATABASE_URL")
	if dsn == "" {
		t.Skip("Requiere PostgreSQL QA")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_test" {
		t.Fatal("Solo PostgreSQL QA")
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	handler, err := httpapi.New(&postgres.Store{Pool: pool}, httpapi.Options{Origin: "http://localhost:8080", Key: []byte(rand.Text() + rand.Text())})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	h := workshopHarness{pool: pool, ctx: context.Background(), api: server, password: rand.Text(), suffix: rand.Text()[:8]}
	_, admin := h.user(t, "design_admin", "Design_"+h.suffix, []string{"inicio.ver", "reportes.ver", "clientes.crear", "vehiculos.crear", "recepciones.crear", "diagnosticos.asignar", "diagnosticos.supervisar", "diagnosticos.ver"})
	worker, mechanic := h.user(t, "design_mechanic", "Mecánico", []string{"diagnosticos.crear", "diagnosticos.ver"})
	_, panel := h.user(t, "design_panel", "Panel_"+h.suffix, []string{"inicio.ver", "diagnosticos.ver", "diagnosticos.crear"})
	unauthorized := dataObject(t, h.call(t, panel, "GET", "/dashboard", nil, 200))
	for _, key := range []string{"monthlyIncome", "monthlyIncomeCents", "income"} {
		if _, ok := unauthorized[key]; ok {
			t.Fatal("Finanzas expuestas a cuenta sin reportes.ver", key)
		}
	}
	if unauthorized["financialVisible"] != false {
		t.Fatal("Falta señal de restricción financiera")
	}
	h.call(t, panel, "GET", "/reportes", nil, 403)
	authorized := dataObject(t, h.call(t, admin, "GET", "/dashboard?period=Semana&desde=2025-02-03T05:00:00Z&hasta=2025-02-10T05:00:00Z", nil, 200))
	if authorized["monthlyIncomeCents"] == nil || !strings.HasPrefix(authorized["from"].(string), "2025-02-03") {
		t.Fatal("Finanzas autorizadas o filtro histórico inválidos")
	}
	client := responseID(t, h.call(t, admin, "POST", "/clientes", map[string]string{"name": "Diseño QA " + h.suffix, "documentType": "Pasaporte", "documentNumber": "D19" + h.suffix, "phone": "987654321", "address": "QA aislado"}, 201))
	vehicle := responseID(t, h.call(t, admin, "POST", "/vehiculos", map[string]any{"clientId": client, "plate": "D19" + h.suffix[:4], "brand": "Toyota", "model": "QA", "year": 2025, "mileage": "10.25"}, 201))
	reception := func() string {
		return responseID(t, h.call(t, admin, "POST", "/recepciones", map[string]string{"clientId": client, "vehicleId": vehicle, "mileage": "10.25", "attentionType": "Diagnostico", "technicianId": worker, "symptoms": "Ruido en frenos QA"}, 201))
	}
	rx := reception()
	draft := dataObject(t, h.call(t, mechanic, "POST", "/diagnosticos", map[string]any{"receptionId": rx, "technicalDiagnosis": "Desgaste probable de pastillas", "status": "Borrador"}, 201))
	id := draft["id"].(string)
	data := map[string]any{"receptionId": rx, "technicalDiagnosis": "Desgaste de pastillas y ruido al frenar", "recommendations": "Reemplazar pastillas y comprobar discos", "status": "Finalizado", "version": draft["version"]}
	h.call(t, panel, "PUT", "/diagnosticos/"+id, data, 403)
	final := dataObject(t, h.call(t, mechanic, "PUT", "/diagnosticos/"+id, data, 200))
	if final["evaluation"] != "" || final["findings"] != data["technicalDiagnosis"] {
		t.Fatal("El contrato nuevo no debe inventar una segunda evaluación")
	}
	h.call(t, mechanic, "PUT", "/diagnosticos/"+id, data, 409)
	legacyRX := reception()
	old := dataObject(t, h.call(t, mechanic, "POST", "/diagnosticos", map[string]any{"receptionId": legacyRX, "findings": "Hallazgo anterior distinto", "evaluation": "Causa anterior distinta", "status": "Borrador"}, 201))
	oldID := old["id"].(string)
	edited := dataObject(t, h.call(t, mechanic, "PUT", "/diagnosticos/"+oldID, map[string]any{"receptionId": legacyRX, "technicalDiagnosis": "Texto revisado visible", "recommendations": "Trabajo recomendado", "status": "Finalizado", "version": old["version"]}, 200))
	raw, _ := json.Marshal(edited["history"])
	var events []map[string]any
	if json.Unmarshal(raw, &events) != nil {
		t.Fatal("Historial inválido")
	}
	previous := events[len(events)-1]["previous"].(map[string]any)
	if previous["findings"] != "Hallazgo anterior distinto" || previous["evaluation"] != "Causa anterior distinta" {
		t.Fatal("Se perdió el par histórico anterior")
	}
	h.call(t, mechanic, "POST", "/diagnosticos", map[string]any{"receptionId": reception(), "technicalDiagnosis": "", "recommendations": "Trabajo", "status": "Finalizado"}, 422)
	h.call(t, mechanic, "POST", "/diagnosticos", map[string]any{"receptionId": reception(), "technicalDiagnosis": "Diagnóstico", "findings": "Ambiguo", "recommendations": "Trabajo", "status": "Finalizado"}, 422)
}
