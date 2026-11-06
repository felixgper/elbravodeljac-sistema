package technicalsheet

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// These graphic coordinates/colors are the approved print contract, independent
// of the shared style constants. The former white Docker quotation must fail it.
func TestApprovedBurgundyPDFLayout(t *testing.T) {
	assertBrand := func(t *testing.T, raw []byte, render func([]byte) ([]byte, error)) string {
		t.Helper()
		pdf, err := render(raw)
		if err != nil {
			t.Fatal(err)
		}
		text := quotationText(t, pdf)
		pages := strings.Count(text, "34.00 737.25 527.00 77.00 re f Q")
		if pages == 0 || strings.Count(text, "q 0.25 0.035 0.085 rg 0.65 0.61 0.62 RG 0.6 w 34.00 737.25 527.00 77.00 re f Q") != pages || strings.Count(text, "34.00 26.25 527.00 24.00 re f Q") != pages {
			t.Fatal("Approved burgundy header/footer missing on a document page")
		}
		return text
	}
	for _, status := range []string{"Pendiente", "Aprobada", "Rechazada", "Multipágina"} {
		t.Run("cotizacion_"+status, func(t *testing.T) {
			q := Quotation{Client: "Cliente ficticio", Plate: "DEMO-19"}
			q.Workshop.Name = "EL BRAVO DEL JAC"
			q.Quote.ID, q.Quote.Version, q.Quote.Status = "DEMO", 1, status
			count := 1
			if status == "Multipágina" {
				q.Quote.Status, count = "Pendiente", 40
			}
			q.Quote.Total = fmt.Sprintf("%d.00", count*10)
			for i := 0; i < count; i++ {
				q.Quote.Items = append(q.Quote.Items, QuotationItem{Type: "Servicio", Line: Line{ID: fmt.Sprint(i + 1), Concept: "Servicio ficticio autorizado", Quantity: "1", UnitPrice: "10.00", Discount: "0.00", Subtotal: "10.00"}})
			}
			raw, _ := json.Marshal(q)
			text := assertBrand(t, raw, RenderQuotation)
			if status == "Multipágina" && strings.Count(text, "34.00 737.25 527.00 77.00 re f Q") < 2 {
				t.Fatal("Multi-page fixture did not exercise pagination")
			}
		})
	}
	for _, evaluated := range []bool{false, true} {
		t.Run(fmt.Sprintf("ficha_evaluada_%t", evaluated), func(t *testing.T) {
			s := Snapshot{Code: "OT-DEMO", Client: "Cliente ficticio", Plate: "DEMO-19", Total: "10.00", Workshop: Workshop{Name: "EL BRAVO DEL JAC"}, Lines: []Line{{ID: "1", Type: "Servicio", Concept: "Trabajo realmente realizado en el fixture", Quantity: "1", UnitPrice: "10.00", Discount: "0.00", Subtotal: "10.00"}}, Control: Control{PerformedItems: []string{"1"}, Inspections: map[string]string{}}}
			if evaluated {
				for i, g := range Catalog.Groups {
					s.Control.Inspections[g.Items[0].ID] = []string{"Bien", "Atención", "Urgente"}[i%3]
				}
			}
			raw, _ := json.Marshal(s)
			text := assertBrand(t, raw, Render)
			for _, label := range []string{"Fugas de fluidos", "Revisi\xf3n de l\xedquidos", "Neum\xe1ticos", "(B) Tj", "(A) Tj", "(U) Tj", "Trabajo realmente realizado"} {
				if !strings.Contains(text, label) {
					t.Fatal("Approved inspection/work content missing", label)
				}
			}
			marks := 0
			if evaluated {
				marks = len(Catalog.Groups)
			}
			if strings.Count(text, "(X) Tj") != marks {
				t.Fatal("Inspection marks differ from recorded values")
			}
		})
	}
}
