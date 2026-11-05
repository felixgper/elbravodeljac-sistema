package domain

import (
	"strings"
	"testing"
)

func TestEmployeeValidation(t *testing.T) {
	name, kind := "Trabajador", "Mecánico"
	if err := (EmployeeInput{Name: &name, Type: &kind}).Validate(true); err != nil {
		t.Fatal(err)
	}
	for _, v := range []string{"", "   ", strings.Repeat("ñ", 151)} {
		if err := (EmployeeInput{Name: &v, Type: &kind}).Validate(true); err == nil {
			t.Fatal("Nombre inválido aceptado")
		}
	}
	long := strings.Repeat("x", 41)
	if err := (EmployeeInput{Name: &name, Type: &long}).Validate(true); err == nil {
		t.Fatal("Cargo demasiado largo aceptado")
	}
	if err := (EmployeeInput{}).Validate(true); err == nil {
		t.Fatal("Alta incompleta")
	}
	active := false
	if err := (EmployeeInput{Active: &active}).Validate(false); err != nil {
		t.Fatal(err)
	}
}
