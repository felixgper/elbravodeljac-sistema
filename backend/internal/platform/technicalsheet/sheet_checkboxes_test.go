package technicalsheet

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestSheetCheckboxStates(t *testing.T) {
	for _, state := range []string{"Bien", "Atención", "Urgente", "Sin evaluar", ""} {
		c := &canvas{}
		g := Group{Label: "Inspección de prueba", Items: []Item{{ID: "test", Label: "Componente"}}}
		sheetInspectionGroup(c, g, map[string]string{"test": state}, 0, 30)
		marks := regexp.MustCompile(`1 0 0 1 ([0-9.]+) ([0-9.]+) Tm \(X\) Tj`).FindAllStringSubmatch(c.String(), -1)
		index := map[string]int{"Bien": 0, "Atención": 1, "Urgente": 2}
		chosen, known := index[state]
		expected := 0
		if known {
			expected = 1
		}
		if len(marks) != expected {
			t.Fatal(state, "incorrect inspection marks", len(marks))
		}
		if known {
			x, _ := strconv.ParseFloat(marks[0][1], 64)
			if x != 120.2+float64(chosen)*13 {
				t.Fatal("Mark in wrong inspection column", state, x)
			}
		}
	}
}
func TestSheetNoImplicitMaterialSelection(t *testing.T) {
	c := &canvas{}
	sheetOptionFlow(c, Catalog.Maintenance[0].Options, nil, 34, 30, 255)
	if strings.Contains(c.String(), "(X) Tj") {
		t.Fatal("Invented selection")
	}
	c = &canvas{}
	sheetOptionFlow(c, Catalog.Maintenance[0].Options, []string{"0", "1"}, 34, 30, 255)
	if strings.Count(c.String(), "(X) Tj") != 2 {
		t.Fatal("Stored multiple selections lost")
	}
	p := &sheetLayout{snapshot: Snapshot{Lines: []Line{{Type: "Repuesto", Concept: "Cerámica"}}}}
	_ = p.next()
	p.maintenanceChecks(Control{PerformedItems: []string{"1"}})
	p.otherServiceChecks(nil)
	if strings.Contains(p.c.String(), "(X) Tj") {
		t.Fatal("Quote or performed line must not imply material/catalog selection")
	}
}
func TestSheetCheckboxExamples(t *testing.T) {
	s := Snapshot{Workshop: Workshop{Name: "EL BRAVO DEL JAC", RUC: "RUC-DEMO", Address: "Dirección del taller de ejemplo"}, Code: "OT-DEMO-019", Client: "Cliente Demostración", Document: "Pasaporte DEMO-001", Phone: "Teléfono de ejemplo", Email: "cliente@example.invalid", Address: "Dirección ficticia de demostración", Plate: "DEMO-19", Vehicle: "Toyota Modelo demostración · 2025", Mileage: "24000.50", Symptoms: "Ruido al frenar", Diagnosis: "Desgaste de pastillas delanteras", Evaluation: "Causa probable: desgaste por uso", Recommendations: "Reemplazar pastillas y verificar frenado", In: "2026-11-03T13:00:00Z", Out: "2026-11-03T20:00:00Z", Mechanic: "Mecánico Demostración", Total: "260.00", Advances: []string{"Reemplazo de pastillas confirmado"}, Tests: []TechnicalTest{{Kind: "Prueba de frenado", Result: "Satisfactoria", Notes: "Verificada por el mecánico"}}, Lines: []Line{{ID: "1", Type: "Servicio", Concept: "Revisión y mantenimiento de frenos", Quantity: "1.00", UnitPrice: "100.00", Discount: "10.00", Subtotal: "90.00"}, {ID: "2", Type: "Repuesto", Concept: "Juego de pastillas delanteras", Quantity: "1.00", UnitPrice: "170.00", Discount: "0.00", Subtotal: "170.00"}}, Control: Control{Inspections: map[string]string{"fugas_0": "Bien", "liquidos_0": "Atención"}, PerformedItems: []string{"1", "2"}, Tires: map[string]string{"frontLeft": "32", "frontRight": "32", "rearLeft": "30", "rearRight": "30"}, ClientNotes: "Solicita revisión de frenos.", TechnicalNotes: "Se verificó el sistema después del reemplazo.", NextMileage: "29000", NextDate: "2027-02-03"}}
	// Presentation sample: only the brake pads actually used in the summary.
	// Multi-oil selections remain covered by the dedicated renderer test.
	s.Control.Maintenance = map[string][]string{"pastillas": {"1"}}
	s.Control.OtherServices = nil
	s.Control.Inspections = map[string]string{"fugas_3": "Bien", "liquidos_3": "Bien", "neumaticos_0": "Bien"}
	save := func(name string, v Snapshot) {
		t.Helper()
		if err := v.Control.Validate(); err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(v)
		pdf, err := Render(raw)
		if err != nil {
			t.Fatal(err)
		}
		text := quotationText(t, pdf)
		expected := 0
		for _, state := range v.Control.Inspections {
			if state == "Bien" || state == "Atención" || state == "Urgente" {
				expected++
			}
		}
		for _, ids := range v.Control.Maintenance {
			expected += len(ids)
		}
		expected += len(v.Control.OtherServices)
		if strings.Count(text, "(X) Tj") != expected {
			t.Fatal("PDF marks do not match the recorded controls", expected, strings.Count(text, "(X) Tj"))
		}
		pages := 2
		if name == "ficha-casillas-extensa.pdf" {
			pages = 3
		}
		if !strings.Contains(string(pdf), "/Count "+strconv.Itoa(pages)+" >>") {
			t.Fatal("Unexpected page count", name, pages)
		}
		if name == "ficha-casillas-completa.pdf" {
			probe := &sheetLayout{snapshot: v}
			_ = probe.next()
			renderSheetControl(probe, v)
			if !strings.Contains(probe.c.String(), "OTROS SERVICIOS") || !strings.Contains(probe.c.String(), "Conformidad del cliente") {
				t.Fatal("Last section orphaned from closing")
			}
		}

		for _, part := range []string{"260.00", "90.00", "170.00", "Conformidad del cliente", "Sin evaluar"} {
			if !strings.Contains(text, part) {
				t.Fatal("Missing preserved sheet data", part)
			}
		}
		if dir := os.Getenv("SHEET_CHECKBOX_ARTIFACT_DIR"); dir != "" {
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, name), pdf, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	save("ficha-casillas.pdf", s)
	stress := s
	stress.Control.Maintenance = map[string][]string{"aceite": {"0", "1"}, "pastillas": {"1"}}
	stress.Control.OtherServices = []string{Catalog.OtherServices[0].ID, Catalog.OtherServices[2].ID}
	stress.Control.Inspections = map[string]string{}
	n := 0
	for _, g := range Catalog.Groups {
		for _, item := range g.Items {
			stress.Control.Inspections[item.ID] = []string{"Bien", "Atención", "Urgente", "Sin evaluar"}[n%4]
			n++
		}
	}
	save("ficha-casillas-multiseleccion.pdf", stress)
	empty := s
	empty.Control = Control{PerformedItems: []string{"1", "2"}}
	save("ficha-casillas-vacias.pdf", empty)
	full := s
	full.Control.Maintenance = map[string][]string{}
	for _, g := range Catalog.Maintenance {
		full.Control.Maintenance[g.ID] = []string{g.Options[0].ID}
	}
	full.Control.OtherServices = nil
	for _, i := range Catalog.OtherServices {
		full.Control.OtherServices = append(full.Control.OtherServices, i.ID)
	}
	full.Control.OtherFluids = "Fluidos registrados en el ejemplo"
	full.Control.Coils = "Bobinas registradas en el ejemplo"
	save("ficha-casillas-completa.pdf", full)
	long := s
	long.Control.TechnicalNotes = strings.Repeat("Observación técnica registrada. ", 35) + "FIN_OBSERVACIONES"
	save("ficha-casillas-extensa.pdf", long)
}
