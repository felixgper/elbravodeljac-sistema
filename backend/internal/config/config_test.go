package config

import (
	"encoding/base64"
	"testing"
)

func TestSQLServerSelectionIsExplicit(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "")
	t.Setenv("APP_ORIGIN", "http://127.0.0.1:18080")
	t.Setenv("JWT_SECRET_BASE64", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	t.Setenv("COOKIE_SECURE", "false")
	for _, db := range []string{"", "master", "ElBravoDelJac", "elbravodeljac_sqlserver_qa"} {
		t.Run(db, func(t *testing.T) {
			dsn := "sqlserver://localhost?database=" + db
			t.Setenv("SQLSERVER_DSN", dsn)
			got, err := Load()
			if db != "elbravodeljac_sqlserver_qa" {
				if err == nil {
					t.Fatal("accepted an unapproved SQL Server database")
				}
			} else if err != nil || !got.SQLServer || got.DSN != dsn {
				t.Fatalf("SQL Server not selected: %v", err)
			}
		})
	}
	t.Setenv("SQLSERVER_DSN", "sqlserver://localhost?database=elbravodeljac_sqlserver_qa")
	t.Setenv("POSTGRES_DSN", "postgres://localhost/elbravodeljac_test")
	if _, err := Load(); err == nil {
		t.Fatal("accepted two relational engines")
	}
}

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
