package domain

import "testing"

func TestQuoteExactMoneyAndReferences(t *testing.T) {
	v := QuoteInput{DiagnosisID: "1", Status: "Aprobada", Items: []QuoteLine{{ServiceID: "1", Quantity: "1.25", UnitPrice: "0.10"}}}
	if err := v.Validate(); err != nil {
		t.Fatal(err)
	}
	if v.Total != "0.13" {
		t.Fatal("half-up", v.Total)
	}
	for _, line := range []QuoteLine{
		{Quantity: "1", UnitPrice: "1"},
		{ServiceID: "1", SparePartID: "1", Quantity: "1", UnitPrice: "1"},
		{SparePartID: "1", Origin: "Cliente", Quantity: "1", UnitPrice: "1"},
		{ServiceID: "1", Quantity: "99999999.99", UnitPrice: "9999999999.99"},
		{ServiceID: "1", Quantity: "1.001", UnitPrice: "1"},
	} {
		v.Items = []QuoteLine{line}
		if err := v.Validate(); err == nil {
			t.Fatal("invalid line accepted", line)
		}
	}
	v.Items = []QuoteLine{{SparePartID: "1", Origin: "Cliente", Quantity: "1", UnitPrice: "0"}}
	if err := v.Validate(); err != nil || v.Total != "0.00" {
		t.Fatal(v.Total, err)
	}
}
func TestOrderAndPaymentValidation(t *testing.T) {
	v := UpdateOrderInput{State: "Finalizada", Note: "Control", Test: "Ruta", Result: "Satisfactoria"}
	if err := v.Validate(); err != nil {
		t.Fatal(err)
	}
	v.Test = ""
	if err := v.Validate(); err == nil {
		t.Fatal("result without test")
	}
	for _, amount := range []string{"0", "-1", "1.001", "NaN", "10000000000.00"} {
		p := PaymentInput{Amount: amount, Method: "Efectivo"}
		if err := p.Validate(); err == nil {
			t.Fatal(amount)
		}
	}
	d := DeliveryInput{Kind: "Tercero", Verified: true, Method: "Documento"}
	if err := d.Validate(); err == nil {
		t.Fatal("contact missing")
	}
}
