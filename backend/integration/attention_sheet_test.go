package integration

import (
	"bytes"
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

func TestAttentionInspectionAndOriginalSheet(t *testing.T) {
	dsn := os.Getenv("INTEGRATION_DATABASE_URL")
	if dsn == "" {
		t.Skip("Requiere PostgreSQL aislado")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_test" {
		t.Fatal("Solo base de prueba")
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
	_, manager := h.user(t, "sheetmanager", "SheetManager_"+h.suffix, all)
	technician, mechanic := h.user(t, "sheetmechanic", "Mecánico", []string{"ordenes.ver", "ordenes.editar", "diagnosticos.ver", "diagnosticos.crear"})
	_, stranger := h.user(t, "sheetstranger", "SheetStranger_"+h.suffix, []string{"ordenes.ver", "ordenes.editar"})
	client := responseID(t, h.call(t, manager, "POST", "/clientes", map[string]string{"name": "Atención prueba " + h.suffix, "documentType": "Pasaporte", "documentNumber": "AT" + h.suffix, "phone": "+51987654321", "address": "Prueba aislada"}, 201))
	vehicle := responseID(t, h.call(t, manager, "POST", "/vehiculos", map[string]any{"clientId": client, "plate": "AT" + h.suffix[:5], "brand": "Toyota", "model": "Prueba", "year": 2025, "mileage": "100"}, 201))
	for _, attention := range []string{"Directo", "Diagnostico"} {
		t.Run(attention, func(t *testing.T) {
			reception := responseID(t, h.call(t, manager, "POST", "/recepciones", map[string]string{"clientId": client, "vehicleId": vehicle, "mileage": "100", "symptoms": "Mantenimiento autorizado", "attentionType": attention}, 201))
			source := map[string]any{"receptionId": reception, "status": "Pendiente", "items": []map[string]string{{"concept": "Servicio efectivamente realizado", "type": "Servicio", "quantity": "1", "unitPrice": "85", "discount": "5"}}}
			if attention == "Directo" {
				h.call(t, mechanic, "POST", "/diagnosticos", map[string]string{"receptionId": reception, "findings": "No debe crearse"}, 422)
			} else {
				h.call(t, manager, "POST", "/cotizaciones", source, 422)
				h.call(t, manager, "PATCH", "/recepciones/"+reception+"/mecanico", map[string]any{"technicianId": technician, "version": 0}, 200)
				diagnosis := responseID(t, h.call(t, manager, "POST", "/diagnosticos", map[string]string{"status": "Finalizado", "evaluation": "Evaluación técnica de prueba", "recommendations": "Servicios recomendados de prueba", "receptionId": reception, "findings": "Diagnóstico técnico real", "technicianId": technician}, 201))
				source["diagnosisId"] = diagnosis
				if got := dataObject(t, h.call(t, manager, "GET", "/diagnosticos/"+diagnosis, nil, 200))["technicianId"]; got != technician {
					t.Fatal("No responsible mechanic", got)
				}
			}
			quote := responseID(t, h.call(t, manager, "POST", "/cotizaciones", source, 201))
			h.call(t, manager, "POST", "/ordenes", map[string]string{"quoteId": quote, "mechanicId": technician}, 422)
			source["status"] = "Aprobada"
			approved := dataObject(t, h.call(t, manager, "PUT", "/cotizaciones/"+quote, source, 200))
			if attention == "Directo" && approved["diagnosisId"] != nil {
				t.Fatal("Synthetic diagnosis", approved)
			}
			order := responseID(t, h.call(t, manager, "POST", "/ordenes", map[string]string{"quoteId": quote, "mechanicId": technician}, 201))
			h.call(t, manager, "GET", "/ordenes/"+order+"/ficha-tecnica.pdf", nil, 404)
			h.call(t, stranger, "GET", "/ordenes/"+order+"/control-tecnico", nil, 404)
			control := map[string]any{"version": 0, "inspections": map[string]string{"fugas_0": "Bien", "liquidos_0": "Atención", "neumaticos_0": "Sin evaluar"}, "maintenance": map[string][]string{"aceite": {"0"}}, "tires": map[string]string{"frontLeft": "32"}, "technicalNotes": "Trabajo comprobado en orden", "nextMileage": "10000", "nextDate": "2027-01-01", "performedItems": []string{approved["items"].([]any)[0].(map[string]any)["id"].(string)}}
			h.call(t, stranger, "PUT", "/ordenes/"+order+"/control-tecnico", control, 403)
			saved := dataObject(t, h.call(t, mechanic, "PUT", "/ordenes/"+order+"/control-tecnico", control, 200))["record"].(map[string]any)
			if saved["version"] != float64(1) {
				t.Fatal(saved)
			}
			h.call(t, mechanic, "PUT", "/ordenes/"+order+"/control-tecnico", control, 409)
			h.call(t, manager, "PATCH", "/ordenes/"+order, map[string]string{"state": "En reparación", "note": "Servicio confirmado"}, 200)
			h.call(t, manager, "PATCH", "/ordenes/"+order, map[string]string{"state": "Finalizada", "note": "Entrega técnica", "test": "Ruta", "result": "Satisfactoria"}, 200)
			control["version"] = 1
			h.call(t, mechanic, "PUT", "/ordenes/"+order+"/control-tecnico", control, 409)
			h.call(t, manager, "POST", "/ordenes/"+order+"/entrega", map[string]any{"kind": "Titular / representante", "method": "Documento", "verified": true}, 422)
			h.call(t, manager, "POST", "/ordenes/"+order+"/pagos", map[string]string{"amount": "80", "method": "Efectivo"}, 201)
			h.call(t, manager, "POST", "/ordenes/"+order+"/entrega", map[string]any{"kind": "Titular / representante", "method": "Documento", "verified": true}, 201)
			if h.scalar(t, `SELECT count(*)::text FROM entregas WHERE id_orden=$1 AND ficha_pdf IS NOT NULL`, order) != "1" {
				t.Fatal("PDF not generated on delivery")
			}
			pdf := h.send(manager, "GET", "/ordenes/"+order+"/ficha-tecnica.pdf", nil)
			if pdf.status != 200 || !bytes.HasPrefix(pdf.raw, []byte("%PDF-1.7")) || !bytes.Contains(pdf.raw, []byte("/Count 2")) {
				t.Fatalf("PDF failed %d %s", pdf.status, pdf.raw)
			}
			h.exec(t, `UPDATE clientes SET nombre=$2 WHERE id_cliente=$1`, client, "Cliente actualizado después de entrega")
			again := h.send(manager, "GET", "/ordenes/"+order+"/ficha-tecnica.pdf", nil)
			if !bytes.Equal(pdf.raw, again.raw) {
				t.Fatal("PDF changed after delivery")
			}
			h.call(t, mechanic, "GET", "/ordenes/"+order+"/ficha-tecnica.pdf", nil, 403)
			history := dataObject(t, h.call(t, manager, "GET", "/vehiculos/"+vehicle+"/historial", nil, 200))
			if len(history["receptions"].([]any)) == 0 {
				t.Fatal(history)
			}
			if attention == "Directo" && h.scalar(t, `SELECT count(*)::text FROM diagnosticos WHERE id_recepcion=$1`, reception) != "0" {
				t.Fatal("Fictional diagnosis")
			}
			dir := "../.cache/ficha-e2e"
			os.MkdirAll(dir, 0700)
			if err = os.WriteFile(dir+"/"+attention+".pdf", pdf.raw, 0600); err != nil {
				t.Fatal(err)
			}
		})
	}
	t.Run("contenido_excesivo_no_recorta_ni_pierde_entrega", func(t *testing.T) {
		symptoms := strings.Repeat("Síntoma extenso ", 100)
		rx := responseID(t, h.call(t, manager, "POST", "/recepciones", map[string]string{"clientId": client, "vehicleId": vehicle, "mileage": "100", "symptoms": symptoms, "attentionType": "Directo"}, 201))
		quote := responseID(t, h.call(t, manager, "POST", "/cotizaciones", map[string]any{"receptionId": rx, "status": "Aprobada", "items": []map[string]string{{"concept": "Servicio autorizado", "type": "Servicio", "quantity": "1", "unitPrice": "10"}}}, 201))
		order := responseID(t, h.call(t, manager, "POST", "/ordenes", map[string]string{"quoteId": quote, "mechanicId": technician}, 201))
		h.call(t, manager, "PATCH", "/ordenes/"+order, map[string]string{"state": "En reparación", "note": "Ejecutado"}, 200)
		h.call(t, manager, "PATCH", "/ordenes/"+order, map[string]string{"state": "Finalizada", "note": "Verificado", "test": "Ruta", "result": "Satisfactoria"}, 200)
		h.call(t, manager, "POST", "/ordenes/"+order+"/pagos", map[string]string{"amount": "10", "method": "Efectivo"}, 201)
		out := dataObject(t, h.call(t, manager, "POST", "/ordenes/"+order+"/entrega", map[string]any{"kind": "Titular / representante", "method": "Documento", "verified": true}, 201))
		if out["sheetReady"] != false {
			t.Fatal("False PDF success", out)
		}
		h.call(t, manager, "GET", "/ordenes/"+order+"/ficha-tecnica.pdf", nil, 422)
		if saved := h.scalar(t, `SELECT ficha_snapshot->>'symptoms' FROM entregas WHERE id_orden=$1`, order); saved != symptoms {
			t.Fatal("Snapshot truncated")
		}
	})
	for _, period := range []string{"Semana", "Mes", "Año"} {
		report := dataObject(t, h.call(t, manager, "GET", "/reportes?period="+period, nil, 200))
		n := len(report["months"].([]any))
		if period == "Semana" && n != 7 || period == "Año" && n != 12 || period == "Mes" && (n < 28 || n > 31) {
			t.Fatal("Wrong buckets", period, n)
		}
		expected := h.scalar(t, `SELECT coalesce(sum(monto),0)::numeric(18,2)::text FROM pagos WHERE fecha_pago>=$1::timestamptz AND fecha_pago<$2::timestamptz`, report["from"], report["until"])
		if report["income"] != expected {
			t.Fatal("Report does not reconcile", report, expected)
		}
	}
	// Every original FK in the approved export must remain; names are test fixtures.
	for _, name := range []string{"fk_asignaciones_ot_id_orden", "fk_asignaciones_ot_id_usuario", "fk_auxilios_id_cliente_solicitante", "fk_auxilios_id_recepcion", "fk_auxilios_id_tecnico", "fk_auxilios_id_unidad_auxilio", "fk_auxilios_id_usuario_registro", "fk_auxilios_id_vehiculo", "fk_avances_ot_id_orden", "fk_avances_ot_id_usuario", "fk_contactos_cliente_id_cliente", "fk_cotizaciones_id_diagnostico", "fk_detalle_cotizacion_id_cotizacion", "fk_detalle_cotizacion_id_repuesto", "fk_detalle_cotizacion_id_servicio", "fk_diagnosticos_id_recepcion", "fk_diagnosticos_id_tecnico", "fk_entregas_id_contacto_receptor", "fk_entregas_id_orden", "fk_entregas_id_usuario_entrega", "fk_entregas_id_usuario_verificador", "fk_historial_cliente", "fk_historial_estados_ot_id_orden", "fk_historial_estados_ot_id_usuario", "fk_historial_usuario", "fk_historial_vehiculo", "fk_ordenes_trabajo_id_cotizacion", "fk_ordenes_trabajo_id_usuario_creacion", "fk_pagos_id_orden", "fk_pagos_id_usuario_registro", "fk_pruebas_ot_id_orden", "fk_pruebas_ot_id_usuario", "fk_recepciones_id_cliente_atencion", "fk_recepciones_id_recepcionista", "fk_recepciones_id_vehiculo", "fk_repuesto_proveedor_id_proveedor", "fk_repuesto_proveedor_id_repuesto", "fk_rol_permisos_permiso", "fk_rol_permisos_rol", "fk_usuario_roles_rol", "fk_usuario_roles_usuario", "fk_vehiculos_cliente"} {
		if h.scalar(t, `SELECT count(*)::text FROM pg_constraint WHERE conname=$1`, name) != "1" {
			t.Fatal("Missing original FK", name)
		}
	}
	t.Log("Atención directa/técnica, controles, PDF inmutable, permisos e ingresos verificados")
}
