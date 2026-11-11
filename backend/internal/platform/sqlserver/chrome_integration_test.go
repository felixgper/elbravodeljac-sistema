package sqlserver

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/mongodb"
	mssql "github.com/microsoft/go-mssqldb"
	"github.com/microsoft/go-mssqldb/msdsn"
	"golang.org/x/crypto/bcrypt"
)

type browserRequestKey struct{}
type browserSQLTrace struct{ t *testing.T }

func (l browserSQLTrace) Log(ctx context.Context, category msdsn.Log, msg string) {
	if path, ok := ctx.Value(browserRequestKey{}).(string); ok && category == msdsn.LogErrors && ctx.Err() == nil {
		l.t.Logf("SQL de Chrome %s: %s", path, msg)
	}
}

func browserTraceDSN(t *testing.T, dsn, token string) string {
	t.Helper()
	u, err := url.Parse(dsn)
	if err != nil || u.Hostname() != "127.0.0.1" || u.Port() == "" {
		t.Fatal("Chrome requires a private, loopback SQL Server runner")
	}
	if decoded, err := hex.DecodeString(token); err != nil || len(decoded) != 32 {
		t.Fatal("Chrome requires a valid QA run marker")
	}
	q := u.Query()
	q.Set("log", "1")
	u.RawQuery = q.Encode()
	mssql.SetContextLogger(browserSQLTrace{t})
	t.Cleanup(func() { mssql.SetContextLogger(nil) })
	return u.String()
}

func requireBrowserDatabase(t *testing.T, s *Store, _ string, token string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var actual string
	err := s.db.QueryRowContext(ctx, `SELECT CONVERT(varchar(64),value) FROM sys.extended_properties
WHERE class=0 AND name=N'jac.qa_run_token'`).Scan(&actual)
	if err != nil || actual != token {
		t.Fatal("SQL Server has no matching QA run marker")
	}
}

func exerciseWorkshopChrome(t *testing.T, s *Store, password, manager, token string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	uri := os.Getenv("SQLSERVER_QA_MONGO_URI")
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != "mongodb" || u.Hostname() != "127.0.0.1" || u.Port() == "" {
		t.Fatal("Chrome requires loopback MongoDB from the QA runner")
	}
	mg, err := mongodb.Open(uri, "elbravodeljac_tecnico_test")
	if err != nil {
		t.Fatal(err)
	}
	defer mg.Close(context.Background())
	if err = mg.Init(ctx); err != nil {
		t.Fatal(err)
	}

	suffix := token[:8]
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	seed := func(label, role, job string, active bool) (string, string) {
		t.Helper()
		var user, employee string
		err := s.db.QueryRowContext(ctx, `SET NOCOUNT ON;
DECLARE @employee bigint=NEXT VALUE FOR dbo.empleados_id_empleado_seq;
DECLARE @user bigint=NEXT VALUE FOR dbo.usuarios_id_usuario_seq;
INSERT dbo.empleados(id_empleado,nombre,tipo_empleado,activo) VALUES(@employee,@p1,@p2,@p3);
INSERT dbo.usuarios(id_usuario,nombre,usuario,password_hash,id_empleado) VALUES(@user,@p1,@p4,@p5,@employee);
INSERT dbo.usuario_roles(id_usuario,id_rol) SELECT @user,id_rol FROM dbo.roles WHERE nombre_rol=@p6;
SELECT CONVERT(varchar(20),@user),CONVERT(varchar(20),@employee);`,
			label+" "+suffix, job, active, label+"_"+suffix, string(hash), role).Scan(&user, &employee)
		if err != nil {
			t.Fatal(err)
		}
		return user, employee
	}
	admin, adminEmployee := seed("taller_admin", "Administrador", "Administración", true)
	first, firstEmployee := seed("taller_m1", "Mecánico", "Mecánico", true)
	second, secondEmployee := seed("taller_m2", "Mecánico", "Mecánico", true)
	inactive, inactiveEmployee := seed("taller_inactivo", "Mecánico", "Mecánico", false)
	if _, err = s.db.ExecContext(ctx, `INSERT dbo.clientes(nombre,tipo_documento,numero_documento,telefono,direccion)
SELECT CONCAT(N'Carga QA ',@p1,N' ',g.value),N'Pasaporte',CONCAT(N'LOAD',@p1,g.value),N'+51987654321',N'QA aislado'
FROM GENERATE_SERIES(1,1200) g`, suffix); err != nil {
		t.Fatal(err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	origin := "http://" + ln.Addr().String()
	root := filepath.Join("..", "..", "..", "..")
	handler, err := httpapi.New(s, httpapi.Options{
		Origin: origin, Key: bytes.Repeat([]byte{23}, 32), SQLServer: true, Technical: mg,
		FrontendDir: filepath.Join(root, "dist"),
	})
	if err != nil {
		t.Fatal(err)
	}
	api := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/healthz" {
			w.Header().Set("X-Bravo-Test-Run", token)
		}
		path := r.Method + " " + r.URL.Path
		if r.URL.Path == "/api/v1/frontend/opciones" {
			path += " " + strings.Join([]string{r.URL.Query().Get("module"), r.URL.Query().Get("resource")}, "/")
		}
		handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), browserRequestKey{}, path)))
	}))
	api.Listener.Close()
	api.Listener = ln
	api.Start()
	defer api.Close()

	dir := t.TempDir()
	input, err := json.Marshal(map[string]string{
		"isolationToken": token, "managerUsername": "gerente_qa", "managerId": manager,
		"url": origin, "suffix": suffix, "password": password, "username": "taller_admin_" + suffix,
		"workerUsername": "taller_m1_" + suffix, "otherUsername": "taller_m2_" + suffix,
		"adminId": admin, "adminEmployee": adminEmployee, "mechanicId": first, "mechanicEmployee": firstEmployee,
		"otherId": second, "otherEmployee": secondEmployee, "inactiveId": inactive, "inactiveEmployee": inactiveEmployee,
	})
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(ctx, "node", filepath.Join(root, "checks", "taller-browser.mjs"))
	cmd.Env = append(os.Environ(), "BRAVO_TALLER_QA_OUTPUT="+dir)
	cmd.Stdin = bytes.NewReader(input)
	out, err := cmd.CombinedOutput()
	raw, readErr := os.ReadFile(filepath.Join(dir, "resultado.json"))
	var result struct {
		Passed        bool
		Checks        []string
		OrderID       string
		DirectOrderID string
		ServerErrors  [][]any
	}
	if readErr == nil {
		readErr = json.Unmarshal(raw, &result)
	}
	for _, check := range result.Checks {
		t.Log("PASS:", check)
	}
	if err != nil || readErr != nil || !result.Passed {
		t.Fatalf("Chrome: %v, informe: %v, errores HTTP: %v\n%s", err, readErr, result.ServerErrors, bytes.ReplaceAll(out, []byte(password), []byte("[QA]")))
	}
	for _, order := range []string{result.OrderID, result.DirectOrderID} {
		var archived []byte
		err := s.db.QueryRowContext(ctx, `SELECT ficha_pdf FROM dbo.entregas WHERE id_orden=@p1`, order).Scan(&archived)
		if err != nil || !bytes.HasPrefix(archived, []byte("%PDF-")) {
			t.Fatalf("ficha no archivada en SQL Server: %v", err)
		}
	}
}
