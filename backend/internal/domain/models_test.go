package domain

import (
	"strings"
	"testing"
)

func TestDecimalAndIDs(t *testing.T) {
	for _, v := range []string{"-1", "1.001", "1,25", "NaN", "100000000.00", "01.00"} {
		if _, err := Mileage(v); err == nil {
			t.Fatalf("accepted %s", v)
		}
	}
	for v, want := range map[string]int64{"0": 0, "2023.5": 202350, "99999999.99": 9999999999} {
		got, err := Mileage(v)
		if err != nil || got != want {
			t.Fatal(v, got, err)
		}
	}
	for _, v := range []string{"0", "-1", "c1", "01", "9223372036854775808"} {
		if _, err := ID(v); err == nil {
			t.Fatal(v)
		}
	}
}
func TestPostgresLengthsWithoutTruncation(t *testing.T) {
	v := ClientInput{Name: strings.Repeat("ñ", 150), DocumentType: "DNI", DocumentNumber: "12345678", Phone: "1234567", Address: "Lima"}
	if err := v.Validate(); err != nil {
		t.Fatal(err)
	}
	v.Name += "ñ"
	if err := v.Validate(); err == nil {
		t.Fatal("overlong name accepted")
	}
	if len([]rune(v.Name)) != 151 {
		t.Fatal("truncated")
	}
	v.Name = "Test"
	v.Email = strings.Repeat("a", 151)
	if err := v.Validate(); err == nil {
		t.Fatal("overlong email accepted")
	}
	v.Email = ""
	v.Phone = strings.Repeat("1", 21)
	if err := v.Validate(); err == nil {
		t.Fatal("overlong phone accepted")
	}
}
