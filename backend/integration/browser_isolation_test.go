package integration

import (
	"context"
	"crypto/subtle"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Only test servers expose this capability. Production code has no such route
// or header. The runner stamps its NEW disposable database before any fixtures.
func requireFreshBrowserRun(t testing.TB, pool *pgxpool.Pool) string {
	t.Helper()
	token := os.Getenv("INTEGRATION_QA_TOKEN")
	if !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(token) {
		t.Fatal("Browser QA requires a newly provisioned database and INTEGRATION_QA_TOKEN")
	}
	var database, actual string
	err := pool.QueryRow(context.Background(), `SELECT current_database(),coalesce(current_setting('jac.qa_run_token',true),'')`).Scan(&database, &actual)
	if err != nil || (database != "elbravodeljac_test" && !regexp.MustCompile(`^jac_restore_\d{14}$`).MatchString(database)) || subtle.ConstantTimeCompare([]byte(token), []byte(actual)) != 1 {
		t.Fatal("Database does not belong to this isolated browser run")
	}
	return token
}

func browserIsolationHandler(next http.Handler, token string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/healthz" {
			w.Header().Set("X-Bravo-Test-Run", token)
		}
		next.ServeHTTP(w, r)
	})
}
