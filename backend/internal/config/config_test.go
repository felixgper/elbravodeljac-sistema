package config

import "testing"

func TestRejectRealDatabaseBeforeConnecting(t *testing.T) {
	for _, dsn := range []string{"", "postgres://localhost/ElBravoDelJac", "dbname=elbravodeljac", "postgres://localhost/postgres"} {
		if ValidateDSN(dsn) == nil {
			t.Fatal("unsafe DB accepted", dsn)
		}
	}
	if err := ValidateDSN("postgres://localhost/elbravodeljac_test"); err != nil {
		t.Fatal(err)
	}
}
