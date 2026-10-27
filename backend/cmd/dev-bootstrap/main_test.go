package main

import "testing"

func TestDevelopmentTargetRejectsRealDatabase(t *testing.T) {
	for _, tc := range []struct{ dsn, confirm string }{
		{"postgres://postgres@127.0.0.1:5432/ElBravoDelJac", "elbravodeljac_dev"},
		{"postgres://postgres@postgres-dev/elbravodeljac_test", "elbravodeljac_dev"},
		{"postgres://postgres@postgres-dev/elbravodeljac_dev", ""},
	} {
		if _, err := developmentTarget(tc.dsn, tc.confirm); err == nil {
			t.Fatal("unsafe bootstrap target accepted")
		}
	}
	if _, err := developmentTarget("postgres://postgres@postgres-dev/elbravodeljac_dev", "elbravodeljac_dev"); err != nil {
		t.Fatal(err)
	}
}
