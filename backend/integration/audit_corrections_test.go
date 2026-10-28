package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestAuditCorrectionsPostgreSQL(t *testing.T) {
	dsn := os.Getenv("INTEGRATION_DATABASE_URL")
	if dsn == "" {
		t.Skip("Requiere PostgreSQL de pruebas aislado")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_test" {
		t.Fatal("Solo prueba aislada")
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
	server := httptest.NewServer(handler)
	defer server.Close()
	h := workshopHarness{pool: pool, api: server, ctx: ctx, password: rand.Text(), suffix: rand.Text()[:8]}
	permissions := []string{}
	for module, actions := range domain.PermissionPolicy().Modules {
		if module == "usuarios" {
			continue
		}
		for _, action := range actions {
			permissions = append(permissions, module+"."+action)
		}
	}
	_, admin := h.user(t, "auditadmin", "Audit_"+h.suffix, permissions)
	_, reader := h.user(t, "pdfreader", "QuoteReader_"+h.suffix, []string{"cotizaciones.ver"})
	mechanic, worker := h.user(t, "auditmechanic", "Mecánico", []string{"diagnosticos.ver", "diagnosticos.crear", "ordenes.ver", "ordenes.editar", "pagos.ver"})
	client := responseID(t, h.call(t, admin, "POST", "/clientes", map[string]string{"name": "Juan Pérez auditoría " + h.suffix, "documentType": "Pasaporte", "documentNumber": "AU" + h.suffix, "phone": "+51987654321", "address": "Entorno exclusivamente de pruebas"}, 201))
	vehicle := responseID(t, h.call(t, admin, "POST", "/vehiculos", map[string]any{"clientId": client, "plate": "AU" + h.suffix[:5], "brand": "Toyota", "model": "Corolla", "year": 2025, "mileage": "100"}, 201))
	approval := map[string]string{"name": "Juan Pérez", "method": "WhatsApp"}
	delivery := map[string]any{"kind": "Titular / representante", "recipientKind": "Propietario o titular", "method": "Documento", "verified": true, "conformity": true, "conformityNotes": "Entrega conforme documentada por el receptor"}
	for _, kind := range []string{"Directo", "Diagnostico"} {
		t.Run(kind, func(t *testing.T) {
			reception := responseID(t, h.call(t, admin, "POST", "/recepciones", map[string]string{"clientId": client, "vehicleId": vehicle, "attentionType": kind, "technicianId": map[bool]string{true: mechanic, false: ""}[kind == "Diagnostico"], "mileage": "100", "symptoms": "Atención auditada sin datos reales"}, 201))
			quoteBody := map[string]any{"receptionId": reception, "status": "Aprobada", "approval": approval, "items": []map[string]string{{"type": "Servicio", "concept": "Servicio realmente ejecutado", "quantity": "2", "unitPrice": "200"}, {"type": "Repuesto", "partKind": "Original", "origin": "Taller", "concept": "Repuesto no utilizado", "quantity": "1", "unitPrice": "200"}}}
			if kind == "Diagnostico" {
				diagnosis := responseID(t, h.call(t, worker, "POST", "/diagnosticos", map[string]string{"receptionId": reception, "status": "Finalizado", "findings": "Inspección de frenos", "evaluation": "Evaluación registrada", "recommendations": "Servicio autorizado por el cliente"}, 201))
				quoteBody["diagnosisId"] = diagnosis
			}
			invalid := map[string]any{}
			raw, _ := json.Marshal(quoteBody)
			json.Unmarshal(raw, &invalid)
			invalid["approval"] = nil
			h.call(t, admin, "POST", "/cotizaciones", invalid, 422)
			quote := dataObject(t, h.call(t, admin, "POST", "/cotizaciones", quoteBody, 201))
			qid := quote["id"].(string)
			h.checkQuotationPDF(t, admin, qid, "Aprobada")
			h.checkQuotationPDF(t, reader, qid, "Consulta")
			h.call(t, worker, "GET", "/cotizaciones/"+qid+"/pdf", nil, 403)
			h.call(t, workshopSession{}, "GET", "/cotizaciones/"+qid+"/pdf", nil, 401)
			h.call(t, admin, "GET", "/cotizaciones/9223372036854775807/pdf", nil, 404)
			h.call(t, admin, "GET", "/cotizaciones/invalid/pdf", nil, 400)
			quoteBody["version"] = quote["version"]
			quoteBody["items"].([]map[string]string)[0]["concept"] = "Servicio corregido que exige nueva aprobación " + h.suffix
			changed := dataObject(t, h.call(t, admin, "PUT", "/cotizaciones/"+qid, quoteBody, 200))
			if changed["status"] != "Pendiente" || changed["approval"] != nil {
				t.Fatal("Approval was not invalidated", changed)
			}
			h.checkQuotationPDF(t, admin, qid, "Pendiente")
			h.call(t, admin, "POST", "/ordenes", map[string]string{"quoteId": qid, "mechanicId": mechanic}, 422)
			quoteBody["version"] = quote["version"]
			h.call(t, admin, "PUT", "/cotizaciones/"+qid, quoteBody, 422)
			quoteBody["version"] = changed["version"]
			approved := dataObject(t, h.call(t, admin, "PUT", "/cotizaciones/"+qid, quoteBody, 200))
			if approved["approval"] == nil || len(approved["revisionHistory"].([]any)) < 2 {
				t.Fatal("Missing approval revision history")
			}
			oldApprovedAt := approved["approvedAt"]
			quoteBody["version"] = approved["version"]
			approved = dataObject(t, h.call(t, admin, "PUT", "/cotizaciones/"+qid, quoteBody, 200))
			if approved["approvedAt"] == oldApprovedAt || approved["approval"].(map[string]any)["at"] != approved["approvedAt"] {
				t.Fatal("A fresh acceptance must record its actual server time")
			}
			order := responseID(t, h.call(t, admin, "POST", "/ordenes", map[string]string{"quoteId": qid, "mechanicId": mechanic}, 201))
			quoteBody["version"] = approved["version"]
			h.call(t, admin, "PUT", "/cotizaciones/"+qid, quoteBody, 409)
			h.exec(t, `UPDATE ordenes_trabajo SET fecha_creacion=clock_timestamp()-interval '2 months' WHERE id_orden=$1`, order)
			dashboard := dataObject(t, h.call(t, admin, "GET", "/dashboard?period=Mes", nil, 200))
			expectedActive := h.scalar(t, `SELECT count(*)::text FROM ordenes_trabajo WHERE estado IN ('creada','en_reparacion','en_pruebas')`)
			if strconv.Itoa(int(dashboard["activeOrders"].(float64))) != expectedActive {
				t.Fatal("Old active orders were hidden")
			}
			pay := map[string]string{"amount": "200", "method": "Transferencia", "requestId": "audit-" + order}
			concurrent := make(chan workshopResponse, 2)
			for i := 0; i < 2; i++ {
				go func() { concurrent <- h.send(admin, "POST", "/ordenes/"+order+"/pagos", pay) }()
			}
			original := ""
			for i := 0; i < 2; i++ {
				res := <-concurrent
				if res.err != nil || res.status != 201 {
					t.Fatal("Concurrent idempotency", res.status, res.err)
				}
				var body map[string]any
				json.Unmarshal(res.raw, &body)
				key := responseID(t, body)
				if original != "" && original != key {
					t.Fatal("Duplicate concurrent payment")
				}
				original = key
			}

			replay := responseID(t, h.call(t, admin, "POST", "/ordenes/"+order+"/pagos", pay, 201))
			if replay != original {
				t.Fatal("Duplicate retry payment")
			}
			pay["amount"] = "201"
			h.call(t, admin, "POST", "/ordenes/"+order+"/pagos", pay, 409)
			wrong := responseID(t, h.call(t, admin, "POST", "/ordenes/"+order+"/pagos", map[string]string{"amount": "50", "method": "Efectivo", "requestId": "wrong-" + order}, 201))
			correction := map[string]string{"reason": "Pago registrado por error, reversión documentada", "requestId": "reverse-" + order}
			h.call(t, worker, "PATCH", "/ordenes/"+order+"/pagos/"+wrong, correction, 403)
			h.call(t, admin, "PATCH", "/ordenes/"+order+"/pagos/"+wrong, correction, 200)
			h.call(t, admin, "PATCH", "/ordenes/"+order+"/pagos/"+wrong, correction, 200)
			final := map[string]string{"state": "Finalizada", "note": "Cierre auditado", "test": "Prueba posterior a reparación", "result": "Satisfactoria"}
			h.call(t, worker, "PATCH", "/ordenes/"+order, final, 422)
			execution := map[string]any{}
			items := approved["items"].([]any)
			execution[items[0].(map[string]any)["id"].(string)] = map[string]string{"status": "Realizado", "notes": "Ejecución confirmada"}
			execution[items[1].(map[string]any)["id"].(string)] = map[string]string{"status": "No realizado", "notes": "No fue necesario utilizar el repuesto"}
			control := map[string]any{"version": 0, "execution": execution, "technicalNotes": "Inspecciones opcionales no inventadas"}
			h.call(t, worker, "PUT", "/ordenes/"+order+"/control-tecnico", control, 200)
			h.call(t, worker, "PATCH", "/ordenes/"+order, final, 422)
			control["version"] = 1
			control["adjustmentReason"] = "Excluir del cobro el repuesto no utilizado"
			h.call(t, worker, "PUT", "/ordenes/"+order+"/control-tecnico", control, 403)
			h.call(t, admin, "PUT", "/ordenes/"+order+"/control-tecnico", control, 200)
			extraOrderID := ""
			if kind == "Directo" {
				supplementalItems := []map[string]string{}
				for i := 0; i < 24; i++ {
					supplementalItems = append(supplementalItems, map[string]string{"type": "Servicio", "concept": "Servicio adicional autorizado " + strings.Repeat("detallado ", 10), "quantity": "1", "unitPrice": "1"})
				}
				supplement := responseID(t, h.call(t, admin, "POST", "/cotizaciones", map[string]any{"receptionId": reception, "parentQuoteId": qid, "status": "Aprobada", "approval": approval, "items": supplementalItems}, 201))
				extraOrder := responseID(t, h.call(t, admin, "POST", "/ordenes", map[string]string{"quoteId": supplement, "mechanicId": mechanic}, 201))
				h.confirmExecution(t, admin, extraOrder)
				h.call(t, worker, "PATCH", "/ordenes/"+extraOrder, final, 200)
				h.call(t, admin, "POST", "/ordenes/"+extraOrder+"/pagos", map[string]string{"amount": "24", "method": "Efectivo", "requestId": "extra-" + extraOrder}, 201)
				extraOrderID = extraOrder
			}
			h.call(t, worker, "PATCH", "/ordenes/"+order, final, 200)
			balance := dataObject(t, h.call(t, admin, "GET", "/ordenes/"+order+"/pagos", nil, 200))
			if balance["total"] != "400.00" || balance["balance"] != "200.00" {
				t.Fatal("Financial reconciliation", balance)
			}
			h.call(t, admin, "POST", "/ordenes/"+order+"/pagos", map[string]string{"amount": "201", "method": "Efectivo", "requestId": "over-" + order}, 422)
			h.call(t, admin, "POST", "/ordenes/"+order+"/pagos", map[string]string{"amount": "200", "method": "Efectivo", "requestId": "rest-" + order}, 201)
			if extraOrderID != "" {
				h.call(t, admin, "POST", "/ordenes/"+extraOrderID+"/entrega", delivery, 201)
				pdf := h.send(admin, "GET", "/ordenes/"+extraOrderID+"/ficha-tecnica.pdf", nil)
				if pdf.status != 200 || bytes.Contains(pdf.raw, []byte("/Count 2 >>")) {
					t.Fatal("Overflow annex missing")
				}
				saveAuditArtifact(t, "anexos.pdf", pdf.raw)
			}
			if kind == "Diagnostico" {
				contact := responseID(t, h.call(t, admin, "POST", "/clientes/"+client+"/contactos", map[string]string{"name": "Representante físico identificado", "documentType": "DNI", "documentNumber": "12345678", "phone": "+51987654321", "relation": "Representante"}, 201))
				delivery["kind"] = "Tercero"
				delivery["recipientKind"] = "Representante de empresa"
				delivery["contactId"] = contact
			}
			delivery["conformity"] = false
			h.call(t, admin, "POST", "/ordenes/"+order+"/entrega", delivery, 422)
			delivery["conformity"] = true
			h.call(t, admin, "POST", "/ordenes/"+order+"/entrega", delivery, 201)
			pdf := h.send(admin, "GET", "/ordenes/"+order+"/ficha-tecnica.pdf", nil)
			if pdf.status != 200 {
				t.Fatal("PDF failure", pdf.status, string(pdf.raw))
			}
			saveAuditArtifact(t, kind+".pdf", pdf.raw)
			if again := h.send(admin, "GET", "/ordenes/"+order+"/ficha-tecnica.pdf", nil); !bytes.Equal(pdf.raw, again.raw) {
				t.Fatal("PDF not frozen")
			}
			h.call(t, admin, "PATCH", "/ordenes/"+order+"/pagos/"+original, map[string]string{"reason": "Intento posterior a entrega", "requestId": "closed-" + order}, 422)
			fix := map[string]any{"reason": "Corregir medio de pago manteniendo el importe y la entrega", "requestId": "replace-" + order, "replacement": map[string]string{"method": "Tarjeta", "comprobante": "RECT-TEST"}}
			corrected := dataObject(t, h.call(t, admin, "PATCH", "/ordenes/"+order+"/pagos/"+original, fix, 200))
			if corrected["correction"].(map[string]any)["replacementId"] == nil {
				t.Fatal("Missing replacement trace")
			}
			h.call(t, admin, "PATCH", "/ordenes/"+order+"/pagos/"+original, fix, 200)
			if dataObject(t, h.call(t, admin, "GET", "/ordenes/"+order+"/pagos", nil, 200))["balance"] != "0.00" {
				t.Fatal("Closed correction changed balance")
			}
			if again := h.send(admin, "GET", "/ordenes/"+order+"/ficha-tecnica.pdf", nil); !bytes.Equal(pdf.raw, again.raw) {
				t.Fatal("Payment method correction changed frozen PDF")
			}
			h.call(t, admin, "GET", "/clientes/"+client+"/historial", nil, 200)
			h.call(t, admin, "GET", "/vehiculos/"+vehicle+"/historial", nil, 200)
			if dataObject(t, h.call(t, admin, "GET", "/cotizaciones/"+qid, nil, 200))["total"] != "600.00" {
				t.Fatal("Original approved quote overwritten")
			}
		})
	}
	for _, period := range []string{"Semana", "Mes", "Año"} {
		report := dataObject(t, h.call(t, admin, "GET", "/reportes?period="+period, nil, 200))
		expected := h.scalar(t, `SELECT coalesce(sum(monto),0)::numeric(18,2)::text FROM jac_movimientos_pagos WHERE fecha >= $1::timestamptz AND fecha < $2::timestamptz`, report["from"], report["until"])
		foundQuantity := false
		for _, entry := range report["services"].([]any) {
			item := entry.(map[string]any)
			if item["label"] == "Servicio corregido que exige nueva aprobación "+h.suffix {
				foundQuantity = true
				if item["count"] != float64(4) {
					t.Fatal("Executed service quantities", item)
				}
			}
		}
		if !foundQuantity {
			t.Fatal("Executed services absent from report")
		}
		if report["income"] != expected {
			t.Fatal("Net income mismatch", report)
		}
	}

	t.Run("nuevo diseño de cotización persistida", func(t *testing.T) {
		reception := responseID(t, h.call(t, admin, "POST", "/recepciones", map[string]string{"clientId": client, "vehicleId": vehicle, "attentionType": "Directo", "mileage": "100", "symptoms": "Cotización ficticia para revisar maquetación en base test"}, 201))
		body := map[string]any{"receptionId": reception, "status": "Pendiente", "items": []map[string]string{
			{"type": "Servicio", "concept": "Inspección y servicio del sistema de frenos\nRevisión de pastillas delanteras\nRectificado y ajuste de discos", "quantity": "1", "unitPrice": "150", "discount": "15"},
			{"type": "Servicio", "concept": "Revisión preventiva incluida en el servicio anterior", "quantity": "1", "unitPrice": "0", "discount": "0"},
			{"type": "Servicio", "concept": "Mantenimiento de palieres, terminales y bieletas", "quantity": "2", "unitPrice": "95", "discount": "5"},
			{"type": "Repuesto", "partKind": "Original", "origin": "Taller", "concept": "Juego de pastillas delanteras", "quantity": "1", "unitPrice": "220", "discount": "20"},
			{"type": "Repuesto", "partKind": "Alternativo", "origin": "Cliente", "concept": "Filtro de aceite suministrado por el cliente", "quantity": "1", "unitPrice": "0", "discount": "0"},
		}}
		quote := dataObject(t, h.call(t, admin, "POST", "/cotizaciones", body, 201))
		id := quote["id"].(string)
		for _, state := range []string{"Pendiente", "Aprobada", "Rechazada"} {
			if state != "Pendiente" {
				body["status"] = state
				body["version"] = quote["version"]
				body["approval"] = nil
				if state == "Aprobada" {
					body["approval"] = map[string]string{"name": "Cliente de prueba aislada", "method": "WhatsApp"}
				}
				quote = dataObject(t, h.call(t, admin, "PUT", "/cotizaciones/"+id, body, 200))
			}
			if quote["total"] != "515.00" {
				t.Fatal("Los importes originales no concilian", quote["total"])
			}
			h.checkQuotationPDF(t, admin, id, state)
			if dir := os.Getenv("AUDIT_ARTIFACT_DIR"); dir != "" {
				pdf := h.send(admin, "GET", "/cotizaciones/"+id+"/pdf", nil)
				if err := os.WriteFile(filepath.Join(dir, "nuevo-diseno-"+state+".pdf"), pdf.raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
		}
	})
	t.Log("Directo/diagnóstico, revisión/aprobación, ampliación, ejecución/ajuste, anticipos/reintentos/reversión, entrega, PDF/anexos e historial verificados")
}
func saveAuditArtifact(t *testing.T, name string, data []byte) {
	t.Helper()
	if dir := os.Getenv("AUDIT_ARTIFACT_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

// GET exports must not change the quote, its approval, history or detail rows.
func (h *workshopHarness) checkQuotationPDF(t *testing.T, session workshopSession, id, status string) {
	t.Helper()
	query := `SELECT md5(to_jsonb(q)::text||coalesce((SELECT jsonb_agg(to_jsonb(d) ORDER BY d.id_detalle)::text FROM detalle_cotizacion d WHERE d.id_cotizacion=q.id_cotizacion),'')) FROM cotizaciones q WHERE q.id_cotizacion=$1`
	before := h.scalar(t, query, id)
	first := h.send(session, "GET", "/cotizaciones/"+id+"/pdf", nil)
	second := h.send(session, "GET", "/cotizaciones/"+id+"/pdf", nil)
	if first.err != nil || first.status != 200 || !bytes.HasPrefix(first.raw, []byte("%PDF-1.7")) || !bytes.Equal(first.raw, second.raw) {
		t.Fatalf("PDF %s: status %d / %v", status, first.status, first.err)
	}
	if h.scalar(t, query, id) != before {
		t.Fatal("PDF changed persisted quotation")
	}
	if dir := os.Getenv("AUDIT_ARTIFACT_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "cotizacion-"+status+"-"+id+".pdf"), first.raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
