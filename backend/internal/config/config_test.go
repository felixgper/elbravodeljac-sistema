package config

import (
	"encoding/base64"
	"net/netip"
	"strings"
	"testing"
)

func TestSurveyTrustedProxyIPs(t *testing.T) {
	for _, value := range []string{"*", "0.0.0.0/0", "172.19.0.0/16", "caddy", "127.0.0.1:8080", "0.0.0.0", "::", "ff02::1", "fe80::1%eth0", "127.0.0.1,", strings.Repeat("127.0.0.1,", 16) + "127.0.0.1"} {
		if _, err := surveyProxyIPs(value); err == nil {
			t.Fatalf("unsafe proxy configuration accepted: %q", value)
		}
	}
	if ips, err := surveyProxyIPs(" "); err != nil || len(ips) != 0 {
		t.Fatal("proxy trust must be disabled by default")
	}
	ips, err := surveyProxyIPs("172.19.0.3, ::ffff:127.0.0.1, ::1")
	if err != nil || len(ips) != 3 || ips[1] != netip.MustParseAddr("127.0.0.1") {
		t.Fatalf("explicit IP parsing: %v", err)
	}
	// Validate the wiring through the real configuration loader as well.
	t.Setenv("POSTGRES_DSN", "")
	t.Setenv("SQLSERVER_DSN", "sqlserver://localhost?database=elbravodeljac_sqlserver_qa")
	t.Setenv("APP_ORIGIN", "https://workshop.test")
	t.Setenv("JWT_SECRET_BASE64", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	t.Setenv("COOKIE_SECURE", "true")
	t.Setenv("SURVEY_TRUSTED_PROXY_IPS", "172.19.0.3")
	cfg, err := Load()
	if err != nil || len(cfg.SurveyTrustedProxyIPs) != 1 || cfg.SurveyTrustedProxyIPs[0] != ips[0] {
		t.Fatalf("proxy configuration not propagated: %v", err)
	}
	t.Setenv("SURVEY_TRUSTED_PROXY_IPS", "0.0.0.0/0")
	if _, err = Load(); err == nil {
		t.Fatal("unsafe proxy environment accepted by loader")
	}
}

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
