package domain

import "testing"

func TestManualQuoteDiscountAndPartKind(t *testing.T) {
	v := QuoteInput{DiagnosisID: "1", Status: "Aprobada", Items: []QuoteLine{
		{Concept: "Servicio negociado", Type: "Servicio", Quantity: "1.25", UnitPrice: "100.00", Discount: "10.00"},
		{Concept: "Repuesto manual", Type: "Repuesto", PartKind: "Alternativo", Origin: "Taller", Quantity: "2", UnitPrice: "50.00", Discount: "5.00"},
	}}
	if err := v.Validate(); err != nil || v.Total != "202.50" {
		t.Fatal(v, err)
	}
	for _, line := range []QuoteLine{
		{Concept: "", Type: "Servicio", Quantity: "1", UnitPrice: "1"},
		{Concept: "Manual", Type: "Repuesto", Origin: "Taller", Quantity: "1", UnitPrice: "1"},
		{Concept: "Manual", Type: "Servicio", PartKind: "Original", Quantity: "1", UnitPrice: "1"},
		{Concept: "Manual", Type: "Servicio", Quantity: "1", UnitPrice: "1", Discount: "1.01"},
		{Concept: "Manual", Type: "Servicio", Quantity: "1", UnitPrice: "1", Discount: "-1"},
	} {
		v.Items = []QuoteLine{line}
		if v.Validate() == nil {
			t.Fatal("invalid manual quote accepted", line)
		}
	}
}
