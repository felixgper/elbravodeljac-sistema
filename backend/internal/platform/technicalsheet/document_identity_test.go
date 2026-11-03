package technicalsheet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These are fictitious, in-memory examples. No account or database is needed.
func TestBusinessDocumentIdentity(t *testing.T) {
	s := Snapshot{Workshop: Workshop{Name: "EL BRAVO DEL JAC", RUC: "RUC-DEMO", Address: "Dirección del taller de ejemplo"}, Code: "OT-DEMO-019", Client: "Cliente Demostración", Document: "Pasaporte DEMO-001", Phone: "Teléfono de ejemplo", Email: "cliente@example.invalid", Address: "Dirección ficticia de demostración", Plate: "DEMO-19", Vehicle: "Toyota Modelo demostración · 2025", Mileage: "24000.50", Symptoms: "Ruido al frenar", Diagnosis: "Desgaste de pastillas delanteras", Evaluation: "Causa probable: desgaste por uso", Recommendations: "Reemplazar pastillas y verificar frenado", In: "2026-11-03T13:00:00Z", Out: "2026-11-03T20:00:00Z", Mechanic: "Mecánico Demostración", Total: "260.00", Advances: []string{"Reemplazo de pastillas confirmado"}, Tests: []TechnicalTest{{Kind: "Prueba de frenado", Result: "Satisfactoria", Notes: "Verificada por el mecánico"}}, Lines: []Line{{ID: "1", Type: "Servicio", Concept: "Revisión y mantenimiento de frenos", Quantity: "1.00", UnitPrice: "100.00", Discount: "10.00", Subtotal: "90.00"}, {ID: "2", Type: "Repuesto", Concept: "Juego de pastillas delanteras", Quantity: "1.00", UnitPrice: "170.00", Discount: "0.00", Subtotal: "170.00"}}, Control: Control{Inspections: map[string]string{"fugas_0": "Bien", "liquidos_0": "Atención"}, PerformedItems: []string{"1", "2"}, Tires: map[string]string{"frontLeft": "32", "frontRight": "32", "rearLeft": "30", "rearRight": "30"}, ClientNotes: "Solicita revisión de frenos.", TechnicalNotes: "Se verificó el sistema después del reemplazo.", NextMileage: "29000", NextDate: "2027-02-03"}}
	output := func(name string, pdf []byte) {
		t.Helper()
		if dir := os.Getenv("DOCUMENT_ARTIFACT_DIR"); dir != "" {
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, name), pdf, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	raw, _ := json.Marshal(s)
	before := append([]byte(nil), raw...)
	pdf, err := Render(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, raw) {
		t.Fatal("Rendering changed the snapshot")
	}
	text := quotationText(t, pdf)
	for _, wanted := range []string{"Cliente Demostraci", "DEMO-19", "Desgaste de pastillas", "Causa probable", "Reemplazar pastillas", "Reemplazo de pastillas confirmado", "Prueba de frenado", "Satisfactoria", "90.00", "170.00", "260.00", "Sin evaluar", "Conformidad del cliente"} {
		if !strings.Contains(text, wanted) {
			t.Fatal("Missing sheet data", wanted)
		}
	}
	if bytes.Contains(pdf, []byte("/DCTDecode")) || !bytes.Contains(pdf, []byte("/Width 410 /Height 430")) {
		t.Fatal("Legacy page images instead of official emblem")
	}
	output("ficha-tecnica-diseno19.pdf", pdf)
	// The simple review sample also uses only fictitious workshop/contact data.
	basic := s
	basic.Diagnosis = ""
	basic.Evaluation = ""
	basic.Recommendations = ""
	basic.DiagnosisNotes = ""
	basic.Tests = nil
	basic.Advances = nil
	basic.Lines = s.Lines[:1]
	basic.Total = "90.00"
	basic.Control = Control{PerformedItems: []string{"1"}}
	raw, _ = json.Marshal(basic)
	pdf, err = Render(raw)
	if err != nil {
		t.Fatal(err)
	}
	output("ficha-basica.pdf", pdf)

	q := Quotation{Workshop: quotationWorkshop{Workshop: s.Workshop}, Client: s.Client, Document: s.Document, Phone: s.Phone, Email: s.Email, Address: s.Address, Plate: s.Plate, Vehicle: s.Vehicle}
	q.Quote.ID = "DEMO-019"
	q.Quote.Version = 2
	q.Quote.ReceptionID = "DEMO-RX"
	q.Quote.CreatedAt = s.In
	q.Quote.Total = s.Total
	for _, l := range s.Lines {
		q.Quote.Items = append(q.Quote.Items, QuotationItem{Line: l, Type: l.Type})
	}
	for _, status := range []string{"Pendiente", "Aprobada"} {
		q.Quote.Status = status
		raw, _ = json.Marshal(q)
		pdf, err = RenderQuotation(raw)
		if err != nil {
			t.Fatal(err)
		}
		text = quotationText(t, pdf)
		for _, wanted := range []string{status, "DEMO-019", "SERVICIOS", "REPUESTOS", "260.00", "10.00", "cliente@example.invalid"} {
			if !strings.Contains(text, wanted) {
				t.Fatal("Missing quote data", wanted)
			}
		}
		output("cotizacion-"+strings.ToLower(status)+"-diseno19.pdf", pdf)
	}

	// Complete controls are explicitly populated ONLY in this fictitious fixture.
	raw, _ = json.Marshal(s)
	var complete Snapshot
	_ = json.Unmarshal(raw, &complete)
	complete.Control.Inspections = map[string]string{}
	n := 0
	for _, g := range Catalog.Groups {
		for _, item := range g.Items {
			complete.Control.Inspections[item.ID] = []string{"Bien", "Atención", "Urgente"}[n%3]
			n++
		}
	}
	complete.Control.Maintenance = map[string][]string{}
	for _, g := range Catalog.Maintenance {
		if len(g.Options) > 0 {
			complete.Control.Maintenance[g.ID] = []string{g.Options[0].ID}
		}
	}
	for _, item := range Catalog.OtherServices {
		complete.Control.OtherServices = append(complete.Control.OtherServices, item.ID)
	}
	complete.Control.OtherFluids = "Fluidos registrados en el ejemplo"
	complete.Control.Coils = "Bobinas registradas en el ejemplo"
	if err := complete.Control.Validate(); err != nil {
		t.Fatal(err)
	}
	raw, _ = json.Marshal(complete)
	pdf, err = Render(raw)
	if err != nil {
		t.Fatal(err)
	}
	text = quotationText(t, pdf)
	if !strings.Contains(text, "Urgente") || !strings.Contains(text, "Fluidos registrados") || !strings.Contains(text, "Bobinas registradas") {
		t.Fatal("Complete controls lost data")
	}
	if !bytes.Contains(pdf, []byte("/Count 2 >>")) {
		t.Fatal("Complete ordinary controls must not orphan a nearly empty third page")
	}
	output("ficha-control-completo.pdf", pdf)
	multi := q
	multi.Quote.ID = "DEMO-020"
	multi.Quote.Status = "Pendiente"
	multi.Quote.Total = "780.00"
	multi.Quote.Items = append(append(append([]QuotationItem{}, q.Quote.Items...), q.Quote.Items...), q.Quote.Items...)
	for i := range multi.Quote.Items {
		multi.Quote.Items[i].ID = fmt.Sprint(i + 1)
	}
	raw, _ = json.Marshal(multi)
	pdf, err = RenderQuotation(raw)
	if err != nil {
		t.Fatal(err)
	}
	output("cotizacion-varios-conceptos.pdf", pdf)
	multiSheet := s
	multiSheet.Lines = append(append(append([]Line{}, s.Lines...), s.Lines...), s.Lines...)
	multiSheet.Total = "780.00"
	multiSheet.Control.PerformedItems = nil
	for i := range multiSheet.Lines {
		multiSheet.Lines[i].ID = fmt.Sprint(i + 1)
		multiSheet.Control.PerformedItems = append(multiSheet.Control.PerformedItems, multiSheet.Lines[i].ID)
	}
	raw, _ = json.Marshal(multiSheet)
	pdf, err = Render(raw)
	if err != nil {
		t.Fatal(err)
	}
	output("ficha-varios-trabajos.pdf", pdf)
	// Exercise long words, all content tails and many rows without shrinking type.
	s.Client = strings.Repeat("NombreExtenso ", 20) + "FIN_CLIENTE"
	s.Diagnosis = strings.Repeat("Diagnóstico extenso con causa probable. ", 180) + "FIN_DIAGNOSTICO"
	s.Control.TechnicalNotes = strings.Repeat("Observación extensa del trabajo. ", 35) + "FIN_OBSERVACIONES"
	s.Lines = nil
	for n := 0; n < 45; n++ {
		s.Lines = append(s.Lines, Line{ID: fmt.Sprint(n), Type: "Servicio", Concept: fmt.Sprintf("CONCEPTO_%02d ", n) + strings.Repeat("Mantenimiento autorizado ", 6), Quantity: "1.00", UnitPrice: "10.00", Discount: "0.00", Subtotal: "10.00"})
	}
	s.Total = "450.00"
	raw, _ = json.Marshal(s)
	pdf, err = Render(raw)
	if err != nil {
		t.Fatal(err)
	}
	text = quotationText(t, pdf)
	for _, wanted := range []string{"FIN_CLIENTE", "FIN_DIAGNOSTICO", "FIN_OBSERVACIONES", "CONCEPTO_44", "450.00"} {
		if !strings.Contains(text, wanted) {
			t.Fatal("Long sheet lost data", wanted)
		}
	}
	if bytes.Contains(pdf, []byte("/Count 2 >>")) {
		t.Fatal("Long sheet must paginate")
	}
	output("ficha-tecnica-extensa-diseno19.pdf", pdf)
}
