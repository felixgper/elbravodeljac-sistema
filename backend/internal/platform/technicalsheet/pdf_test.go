package technicalsheet

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestControlValidation(t *testing.T) {
	for _, v := range []Control{{Inspections: map[string]string{"fugas_0": "Inventado"}}, {Inspections: map[string]string{"inexistente": "Bien"}}, {Maintenance: map[string][]string{"aceite": {"99"}}}, {Tires: map[string]string{"frontLeft": "NaN"}}, {Tires: map[string]string{"frontLeft": "151"}}, {NextDate: "31/02/2026"}, {TechnicalNotes: strings.Repeat("x", 1201)}} {
		if v.Validate() == nil {
			t.Fatal("Invalid control accepted", v)
		}
	}
	if err := (Control{}).Validate(); err != nil {
		t.Fatal(err)
	}
}
func TestOriginalTwoPagePDF(t *testing.T) {
	s := Snapshot{Workshop: Workshop{Name: "EL BRAVO DEL JAC", RUC: "10400848209", Address: "Calle Urano 381, Los Olivos"}, Code: "OT-TEST", Client: "Cliente de prueba", Document: "DOC-PRUEBA", Phone: "+51987654321", Address: "Dirección de prueba", Plate: "ABC-123", Vehicle: "Toyota Prueba", Mileage: "100.00", Symptoms: "Mantenimiento autorizado", In: "2026-10-27T13:00:00Z", Out: "2026-10-27T15:00:00Z", Mechanic: "Mecánico de prueba", Total: "80.00", Lines: []Line{{ID: "1", Concept: "Servicio realizado", Quantity: "1.00", UnitPrice: "85.00", Discount: "5.00", Subtotal: "80.00"}}, Control: Control{Inspections: map[string]string{"fugas_0": "Bien", "liquidos_0": "Atención"}, PerformedItems: []string{"1"}, Tires: map[string]string{"frontLeft": "32"}, NextMileage: "10000", NextDate: "2027-01-01"}}
	raw, _ := json.Marshal(s)
	pdf, err := Render(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.7")) || !bytes.Contains(pdf, []byte("/Count 2")) {
		t.Fatal("Not a two page PDF")
	}
	if output := os.Getenv("SHEET_ARTIFACT"); output != "" {
		if err = os.WriteFile(output, pdf, 0600); err != nil {
			t.Fatal(err)
		}
	}
	s.Symptoms = strings.Repeat("Texto muy largo ", 1000)
	raw, _ = json.Marshal(s)
	if _, err = Render(raw); err == nil {
		t.Fatal("Silently cropped content")
	}
}
