package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func TestPhaseAWithPostgreSQL(t *testing.T) {
	dsn := os.Getenv("INTEGRATION_DATABASE_URL")
	if dsn == "" {
		t.Skip("INTEGRATION_DATABASE_URL no configurada; no se ejecutaron pruebas PostgreSQL")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal("DSN de prueba inválido")
	}
	if cfg.ConnConfig.Database != "elbravodeljac_test" {
		t.Fatal("La integración solo permite la base exacta elbravodeljac_test")
	}
	ctx := context.Background()
	cfg.ConnConfig.RuntimeParams["timezone"] = "UTC"
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err = pool.Ping(ctx); err != nil {
		t.Fatal("No se pudo conectar a PostgreSQL de pruebas", err)
	}
	var tables, fks int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'`).Scan(&tables); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM pg_constraint WHERE contype='f' AND connamespace='public'::regnamespace`).Scan(&fks); err != nil {
		t.Fatal(err)
	}
	if tables != 26 || fks != 44 {
		t.Fatalf("Esquema inesperado: %d tablas, %d FK", tables, fks)
	}
	random := make([]byte, 16)
	if _, err = rand.Read(random); err != nil {
		t.Fatal(err)
	}
	suffix := hex.EncodeToString(random)
	passwordBytes := make([]byte, 24)
	if _, err = rand.Read(passwordBytes); err != nil {
		t.Fatal(err)
	}
	password := hex.EncodeToString(passwordBytes)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	var userID, roleID, extraRoleID string
	if err = pool.QueryRow(ctx, `INSERT INTO usuarios(nombre,usuario,password_hash) VALUES('Prueba API',$1,$2) RETURNING id_usuario::text`, "test_"+suffix, string(hash)).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO roles(nombre_rol) VALUES($1) RETURNING id_rol::text`, "test_"+suffix).Scan(&roleID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO roles(nombre_rol) VALUES($1) RETURNING id_rol::text`, "extra_"+suffix).Scan(&extraRoleID); err != nil {
		t.Fatal(err)
	}
	exec(`INSERT INTO usuario_roles(id_usuario,id_rol) VALUES($1,$2),($1,$3)`, userID, roleID, extraRoleID)
	// INSERT permission codes only in this isolated test DB; never change existing grants.
	for _, code := range []string{"clientes.ver", "clientes.crear", "clientes.editar", "clientes.eliminar", "vehiculos.ver", "vehiculos.crear", "vehiculos.editar", "vehiculos.eliminar", "pagos.entregar"} {
		exec(`INSERT INTO permisos(codigo) VALUES($1) ON CONFLICT(codigo) DO NOTHING`, code)
		exec(`INSERT INTO rol_permisos(id_rol,id_permiso) SELECT $1,id_permiso FROM permisos WHERE codigo=$2`, roleID, code)
	}
	key := make([]byte, 32)
	if _, err = rand.Read(key); err != nil {
		t.Fatal(err)
	}
	h, err := httpapi.New(&postgres.Store{Pool: pool}, httpapi.Options{Origin: "http://localhost:8080", Key: key})
	if err != nil {
		t.Fatal(err)
	}
	api := httptest.NewServer(h)
	defer api.Close()
	type response struct {
		status  int
		body    []byte
		header  http.Header
		cookies []*http.Cookie
	}
	var cookie *http.Cookie
	csrf := ""
	call := func(method, path string, body any, auth bool) response {
		var content io.Reader
		if body != nil {
			b, e := json.Marshal(body)
			if e != nil {
				t.Fatal(e)
			}
			content = bytes.NewReader(b)
		}
		r, e := http.NewRequest(method, api.URL+path, content)
		if e != nil {
			t.Fatal(e)
		}
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", "http://localhost:8080")
		if auth && cookie != nil {
			r.AddCookie(cookie)
			r.Header.Set("X-CSRF-Token", csrf)
		}
		res, e := http.DefaultClient.Do(r)
		if e != nil {
			t.Fatal(e)
		}
		defer res.Body.Close()
		b, e := io.ReadAll(res.Body)
		if e != nil {
			t.Fatal(e)
		}
		return response{res.StatusCode, b, res.Header, res.Cookies()}
	}
	expect := func(res response, status int) response {
		t.Helper()
		if res.status != status {
			t.Fatalf("HTTP %d esperado %d: %s", res.status, status, res.body)
		}
		return res
	}
	dataID := func(res response) string {
		t.Helper()
		var payload struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(res.body, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.Data.ID == "" {
			t.Fatalf("missing id: %s", res.body)
		}
		return payload.Data.ID
	}
	clientPayload := func(number string) map[string]any {
		return map[string]any{"name": "Cliente de prueba", "documentType": "Pasaporte", "documentNumber": number, "phone": "+51987654321", "address": "Lima", "email": ""}
	}
	vehiclePayload := func(owner, plate string) map[string]any {
		return map[string]any{"clientId": owner, "plate": plate, "brand": "JAC", "model": "T8", "year": 2024, "mileage": "2023.50", "notes": "Prueba aislada"}
	}

	t.Run("authentication", func(t *testing.T) {
		expect(call("GET", "/api/v1/clientes", nil, false), 401)
		expect(call("POST", "/api/v1/auth/login", map[string]string{"username": "test_" + suffix, "password": "incorrecta"}, false), 401)
		res := expect(call("POST", "/api/v1/auth/login", map[string]string{"username": "test_" + suffix, "password": password}, false), 200)
		if len(res.cookies) != 1 {
			t.Fatal("missing cookie")
		}
		cookie = res.cookies[0]
		if !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
			t.Fatal("cookie flags")
		}
		var payload struct {
			CSRF string `json:"csrfToken"`
		}
		if err := json.Unmarshal(res.body, &payload); err != nil {
			t.Fatal(err)
		}
		csrf = payload.CSRF
		if csrf == "" || strings.Contains(string(res.body), string(hash)) || strings.Contains(string(res.body), password) {
			t.Fatal("credentials/CSRF response")
		}
		expect(call("GET", "/api/v1/auth/me", nil, true), 200)
		if !strings.Contains(string(res.body), "extra_"+suffix) {
			t.Fatal("additional role lost")
		}
		expect(call("GET", "/readyz", nil, false), 200)
	})
	if cookie == nil {
		t.Fatal("authentication prerequisite failed")
	}
	var clientID, vehicleID, secondID string
	t.Run("clients CRUD lengths duplicate FK pagination", func(t *testing.T) {
		clientID = dataID(expect(call("POST", "/api/v1/clientes", clientPayload("A"+suffix[:12]), true), 201))
		expect(call("POST", "/api/v1/clientes", clientPayload("A"+suffix[:12]), true), 409)
		expect(call("GET", "/api/v1/clientes/"+clientID, nil, true), 200)
		expect(call("PATCH", "/api/v1/clientes/"+clientID, map[string]string{"name": "Actualizado"}, true), 200)
		expect(call("PATCH", "/api/v1/clientes/"+clientID, map[string]string{"name": strings.Repeat("ñ", 151)}, true), 422)
		res := expect(call("GET", "/api/v1/clientes?q=Actualizado&pageSize=1", nil, true), 200)
		if !strings.Contains(string(res.body), clientID) {
			t.Fatal("search missing")
		}
		expect(call("GET", "/api/v1/clientes?pageSize=101", nil, true), 422)
		expect(call("GET", "/api/v1/clientes/0", nil, true), 400)
		expect(call("GET", "/api/v1/clientes/9223372036854775807", nil, true), 404)
		expect(call("PATCH", "/api/v1/clientes/"+clientID, map[string]string{"id": "2"}, true), 422)
		secondID = dataID(expect(call("POST", "/api/v1/clientes", clientPayload("B"+suffix[:12]), true), 201))
		third := dataID(expect(call("POST", "/api/v1/clientes", clientPayload("C"+suffix[:12]), true), 201))
		expect(call("DELETE", "/api/v1/clientes/"+third, nil, true), 204)
		contact := map[string]string{"name": "Contacto QA", "documentType": "DNI", "documentNumber": "45678123", "phone": "+51987654321", "relation": "Conductor"}
		expect(call("POST", "/api/v1/clientes/"+clientID+"/contactos", contact, true), 201)
		expect(call("GET", "/api/v1/clientes/"+clientID+"/contactos", nil, true), 200)
		expect(call("DELETE", "/api/v1/clientes/"+clientID, nil, true), 409)
	})
	if clientID == "" || secondID == "" {
		t.Fatal("client prerequisite failed")
	}
	plate := "T-" + strings.ToUpper(suffix[:8])
	t.Run("vehicles transactions history constraints", func(t *testing.T) {
		vehicleID = dataID(expect(call("POST", "/api/v1/vehiculos", vehiclePayload(clientID, plate), true), 201))
		expect(call("POST", "/api/v1/vehiculos", vehiclePayload(clientID, strings.ToLower(plate)), true), 409)
		expect(call("POST", "/api/v1/vehiculos", vehiclePayload("9223372036854775807", "Z-"+suffix[:8]), true), 409)
		expect(call("GET", "/api/v1/vehiculos/"+vehicleID, nil, true), 200)
		expect(call("GET", "/api/v1/vehiculos?clientId="+clientID, nil, true), 200)
		expect(call("PATCH", "/api/v1/vehiculos/"+vehicleID, map[string]string{"mileage": "2024.25"}, true), 200)
		expect(call("PATCH", "/api/v1/vehiculos/"+vehicleID, map[string]string{"mileage": "2023.00"}, true), 422)
		expect(call("PATCH", "/api/v1/vehiculos/"+vehicleID, map[string]string{"mileage": "2024.251"}, true), 422)
		expect(call("PATCH", "/api/v1/vehiculos/"+vehicleID, map[string]string{"plate": "NEW-123"}, true), 400)
		expect(call("POST", "/api/v1/vehiculos/"+vehicleID+"/transferencias", map[string]any{"newClientId": secondID, "transferVerified": false}, true), 422)
		expect(call("POST", "/api/v1/vehiculos/"+vehicleID+"/transferencias", map[string]any{"newClientId": "9223372036854775807", "transferVerified": true}, true), 409)
		var current int
		_ = pool.QueryRow(ctx, `SELECT count(*) FROM historial_propietarios WHERE id_vehiculo=$1 AND fecha_fin IS NULL`, vehicleID).Scan(&current)
		if current != 1 {
			t.Fatal("failed transfer did not roll back")
		}
		expect(call("POST", "/api/v1/vehiculos/"+vehicleID+"/transferencias", map[string]any{"newClientId": secondID, "transferVerified": true}, true), 200)
		res := expect(call("GET", "/api/v1/vehiculos/"+vehicleID+"/historial", nil, true), 200)
		if !strings.Contains(string(res.body), secondID) {
			t.Fatal("missing history")
		}
		expect(call("DELETE", "/api/v1/vehiculos/"+vehicleID, nil, true), 409)
		// A vehicle with operational references must retain its initial history on failed deletion.
		scratch := dataID(expect(call("POST", "/api/v1/vehiculos", vehiclePayload(clientID, "R-"+suffix[:8]), true), 201))
		var reception string
		if err := pool.QueryRow(ctx, `INSERT INTO recepciones(id_vehiculo,id_cliente_atencion,id_recepcionista,sintomas) VALUES($1,$2,$3,'QA') RETURNING id_recepcion::text`, scratch, clientID, userID).Scan(&reception); err != nil {
			t.Fatal(err)
		}
		expect(call("DELETE", "/api/v1/vehiculos/"+scratch, nil, true), 409)
		_ = pool.QueryRow(ctx, `SELECT count(*) FROM historial_propietarios WHERE id_vehiculo=$1`, scratch).Scan(&current)
		if current != 1 {
			t.Fatal("history removed despite FK failure")
		}
		exec(`DELETE FROM recepciones WHERE id_recepcion=$1`, reception)
		expect(call("DELETE", "/api/v1/vehiculos/"+scratch, nil, true), 204)
	})
	if vehicleID == "" {
		t.Fatal("vehicle prerequisite failed")
	}
	t.Run("concurrent mileage updates serialize", func(t *testing.T) {
		var wg sync.WaitGroup
		codes := make(chan int, 2)
		for _, km := range []string{"3000.00", "4000.00"} {
			wg.Add(1)
			go func(value string) {
				defer wg.Done()
				codes <- call("PATCH", "/api/v1/vehiculos/"+vehicleID, map[string]string{"mileage": value}, true).status
			}(km)
		}
		wg.Wait()
		close(codes)
		for code := range codes {
			if code != 200 && code != 422 {
				t.Fatal("concurrent update", code)
			}
		}
		var mileage string
		if err := pool.QueryRow(ctx, `SELECT kilometraje_actual::text FROM vehiculos WHERE id_vehiculo=$1`, vehicleID).Scan(&mileage); err != nil {
			t.Fatal(err)
		}
		if mileage != "4000.00" {
			t.Fatal("lost/decreasing mileage", mileage)
		}
	})
	t.Run("CSRF and permission changes apply immediately", func(t *testing.T) {
		saved := csrf
		csrf = "invalid"
		expect(call("PATCH", "/api/v1/clientes/"+clientID, map[string]string{"name": "Denied"}, true), 403)
		csrf = saved
		exec(`DELETE FROM rol_permisos WHERE id_rol=$1 AND id_permiso=(SELECT id_permiso FROM permisos WHERE codigo='clientes.ver')`, roleID)
		expect(call("GET", "/api/v1/clientes", nil, true), 403)
		var roles int
		_ = pool.QueryRow(ctx, `SELECT count(*) FROM usuario_roles WHERE id_usuario=$1`, userID).Scan(&roles)
		if roles != 2 {
			t.Fatal("role assignments changed")
		}
		expect(call("POST", "/api/v1/auth/logout", nil, true), 204)
		expect(call("GET", "/api/v1/auth/me", nil, true), 401)
	})
	t.Log(fmt.Sprintf("Esquema verificado: %d tablas y %d FK; HTTP + PostgreSQL reales", tables, fks))
}
