package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"net"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	mgstore "github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/mongodb"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Historical scenarios are preserved and now run against a harness-owned API.
func TestLegacyBrowserRegressionsIsolated(t *testing.T) {
	if os.Getenv("LEGACY_BROWSER_E2E") != "1" {
		t.Skip("Requires new isolated QA and LEGACY_BROWSER_E2E=1")
	}
	cfg, err := pgxpool.ParseConfig(os.Getenv("INTEGRATION_DATABASE_URL"))
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_test" || cfg.ConnConfig.Host != "127.0.0.1" {
		t.Fatal("Only new loopback PostgreSQL QA")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	token := requireFreshBrowserRun(t, pool)
	mg, err := mgstore.Open(os.Getenv("INTEGRATION_MONGODB_URI"), os.Getenv("INTEGRATION_MONGODB_DATABASE"))
	if err != nil {
		t.Fatal(err)
	}
	defer mg.Close(context.Background())
	if err = mg.Init(ctx); err != nil {
		t.Fatal(err)
	}
	for _, script := range []string{"brands-browser.mjs", "diagnosis-browser.mjs", "quotes-browser.mjs", "review-browser.mjs", "attention-browser.mjs", "live-docker.mjs"} {
		t.Run(script, func(t *testing.T) {
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			origin := "http://" + ln.Addr().String()
			handler, err := httpapi.New(&postgres.Store{Pool: pool}, httpapi.Options{Origin: origin, Key: []byte(rand.Text() + rand.Text()), Technical: mg, FrontendDir: filepath.Join("..", "..", "dist")})
			if err != nil {
				t.Fatal(err)
			}
			api := httptest.NewUnstartedServer(browserIsolationHandler(handler, token))
			api.Listener = ln
			api.Start()
			defer api.Close()
			h := workshopHarness{pool: pool, api: api, ctx: ctx, origin: origin, password: rand.Text(), suffix: rand.Text()[:10]}
			permissions := []string{}
			for module, actions := range domain.PermissionPolicy().Modules {
				for _, action := range actions {
					permissions = append(permissions, module+"."+action)
				}
			}
			h.user(t, "legacybrowser", "LegacyBrowser_"+h.suffix, permissions)
			fixture, _ := json.Marshal(map[string]string{"url": origin, "isolationToken": token, "username": "legacybrowser_" + h.suffix, "password": h.password, "roleId": h.scalar(t, `SELECT id_rol::text FROM roles WHERE nombre_rol=$1`, "LegacyBrowser_"+h.suffix)})
			scripts := []string{script}
			if script == "live-docker.mjs" {
				scripts = append(scripts, "read-persisted-docker.mjs")
			}
			for _, entry := range scripts {
				cmd := exec.CommandContext(ctx, "node", filepath.Join("..", "..", "checks", entry))
				cmd.Stdin = bytes.NewReader(fixture)
				out, err := cmd.CombinedOutput()
				dir := filepath.Join("..", ".cache", "legacy-browser")
				_ = os.MkdirAll(dir, 0700)
				_ = os.WriteFile(filepath.Join(dir, entry+".log"), out, 0600)
				if err != nil {
					t.Fatalf("Isolated historical scenario %s failed; see ignored log", script)
				}
				t.Log(string(out))
			}
		})
	}
}
