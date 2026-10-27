package domain

import "testing"

func TestPermissionMatrixAndSQLLimits(t *testing.T) {
	for _, matrix := range []Matrix{{"fake": {"ver"}}, {"clientes": {"crear"}}, {"clientes": {"ver", "ver"}}, {"usuarios": {"ver", "root"}}} {
		if _, err := matrix.Codes(); err == nil {
			t.Fatal("invalid matrix accepted", matrix)
		}
	}
	for role, matrix := range PermissionPolicy().Defaults {
		if _, err := matrix.Codes(); err != nil {
			t.Fatal(role, err)
		}
	}
	parts := "Especialidad que no tiene columna"
	if err := (ProviderInput{Parts: &parts}).Validate(); err == nil {
		t.Fatal("parts discarded")
	}
	v := CatalogInput{Name: "Repuesto", ReferencePrice: "1.001", Stock: "0"}
	if v.Validate("repuestos") == nil {
		t.Fatal("fraction silently rounded")
	}
	v.ReferencePrice = "1.25"
	v.Stock = "-1"
	if v.Validate("repuestos") == nil {
		t.Fatal("negative stock")
	}
	u := UserInput{}
	if u.Validate(true) == nil {
		t.Fatal("required user fields")
	}
}
