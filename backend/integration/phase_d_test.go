package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	mongoadapter "github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/mongodb"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/bson"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func technicalUpload(h *workshopHarness, s workshopSession, diagnosis, request, description string, data []byte) workshopResponse {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("requestId", request)
	_ = w.WriteField("descripcion", description)
	part, err := w.CreateFormFile("archivo", "untrusted.bin")
	if err != nil {
		return workshopResponse{err: err}
	}
	if _, err = part.Write(data); err != nil {
		return workshopResponse{err: err}
	}
	if err = w.Close(); err != nil {
		return workshopResponse{err: err}
	}
	r, err := http.NewRequest("POST", h.api.URL+"/api/v1/diagnosticos/"+diagnosis+"/evidencias", &body)
	if err != nil {
		return workshopResponse{err: err}
	}
	r.Header.Set("Content-Type", w.FormDataContentType())
	r.Header.Set("Origin", "http://localhost:8080")
	r.Header.Set("X-CSRF-Token", s.csrf)
	if s.cookie != nil {
		r.AddCookie(s.cookie)
	}
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return workshopResponse{err: err}
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	return workshopResponse{res.StatusCode, b, res.Cookies(), err}
}
func uploadID(t *testing.T, res workshopResponse, status int) string {
	t.Helper()
	if res.err != nil || res.status != status {
		t.Fatalf("upload %d expected %d: %s %v", res.status, status, res.raw, res.err)
	}
	var body map[string]any
	decodeTest(t, res.raw, &body)
	return responseID(t, body)
}
func TestPhaseDMongoDBAndPostgreSQL(t *testing.T) {
	dsn, uri, name := os.Getenv("INTEGRATION_DATABASE_URL"), os.Getenv("INTEGRATION_MONGODB_URI"), os.Getenv("INTEGRATION_MONGODB_DATABASE")
	if dsn == "" || uri == "" || name == "" {
		t.Skip("Integración D no ejecutada: requiere PostgreSQL y MongoDB aislados")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_test" || name != "elbravodeljac_tecnico_test" {
		t.Fatal("Solo bases exactas de prueba")
	}
	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	mg, err := mongoadapter.Open(uri, name)
	if err != nil {
		t.Fatal(err)
	}
	defer mg.Close(ctx)
	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	err = mg.Init(initCtx)
	cancel()
	if err != nil {
		t.Fatalf("MongoDB real Init: %v", err)
	}
	key := make([]byte, 32)
	if _, err = rand.Read(key); err != nil {
		t.Fatal(err)
	}
	handler, err := httpapi.New(&postgres.Store{Pool: pool}, httpapi.Options{Origin: "http://localhost:8080", Key: key, Technical: mg})
	if err != nil {
		t.Fatal(err)
	}
	api := httptest.NewServer(handler)
	defer api.Close()
	h := workshopHarness{pool: pool, api: api, ctx: ctx, password: rand.Text(), suffix: rand.Text()[:10]}
	_, creator := h.user(t, "dcreator", "DCreator_"+h.suffix, []string{"diagnosticos.ver", "diagnosticos.crear", "clientes.crear", "vehiculos.crear", "recepciones.crear"})
	_, other := h.user(t, "dother", "DOther_"+h.suffix, []string{"diagnosticos.ver", "diagnosticos.crear"})
	_, manager := h.user(t, "dmanager", "DManager_"+h.suffix, []string{"diagnosticos.ver", "diagnosticos.crear", "ordenes.asignar"})
	client := responseID(t, h.call(t, creator, "POST", "/clientes", map[string]string{"name": "Diagnóstico fixture", "documentType": "Pasaporte", "documentNumber": "D" + strings.ToUpper(h.suffix), "phone": "+51987654321", "address": "Lima"}, 201))
	vehicle := responseID(t, h.call(t, creator, "POST", "/vehiculos", map[string]any{"clientId": client, "plate": "D-" + strings.ToUpper(h.suffix[:6]), "brand": "JAC", "model": "T8", "year": 2024, "mileage": "10"}, 201))
	makeDiagnosis := func(s workshopSession) string {
		r := responseID(t, h.call(t, creator, "POST", "/recepciones", map[string]string{"clientId": client, "vehicleId": vehicle, "mileage": "10", "symptoms": "Scanner fixture"}, 201))
		return responseID(t, h.call(t, s, "POST", "/diagnosticos", map[string]string{"receptionId": r, "findings": "Evaluación técnica"}, 201))
	}
	diagnosis := makeDiagnosis(creator)
	otherDiagnosis := makeDiagnosis(other)
	base := "/diagnosticos/" + diagnosis
	scan := map[string]any{"requestId": "scan-" + h.suffix, "equipo": map[string]string{"nombre": "Scanner fixture", "version": "1"}, "dtcs": []map[string]string{{"codigo": "P0300"}}, "mediciones": map[string]any{"rpm": float64(1234), "temperatura": map[string]any{"valor": float64(85.5)}}}
	t.Run("scanner BSON validadores reintentos y concurrencia", func(t *testing.T) {
		id := responseID(t, h.call(t, creator, "POST", base+"/escaneres", scan, 201))
		if again := responseID(t, h.call(t, creator, "POST", base+"/escaneres", scan, 200)); again != id {
			t.Fatal("duplicate")
		}
		scan["equipo"] = map[string]string{"nombre": "Cambio"}
		h.call(t, creator, "POST", base+"/escaneres", scan, 409)
		scan["equipo"] = map[string]string{"nombre": "Scanner fixture", "version": "1"}
		list := h.call(t, creator, "GET", base+"/escaneres?pageSize=1", nil, 200)
		if list["meta"].(map[string]any)["total"] != float64(1) {
			t.Fatal(list)
		}
		item := list["data"].([]any)[0].(map[string]any)
		if item["mediciones"].(map[string]any)["rpm"] != float64(1234) {
			t.Fatal("BSON changed")
		}
		concurrent := map[string]any{}
		for k, v := range scan {
			concurrent[k] = v
		}
		concurrent["requestId"] = "parallel-" + h.suffix
		var wg sync.WaitGroup
		results := make(chan workshopResponse, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); results <- h.send(creator, "POST", base+"/escaneres", concurrent) }()
		}
		wg.Wait()
		close(results)
		created := 0
		for res := range results {
			if res.err != nil || (res.status != 200 && res.status != 201) {
				t.Fatalf("parallel scan %d %s %v", res.status, res.raw, res.err)
			}
			if res.status == 201 {
				created++
			}
		}
		if created != 1 {
			t.Fatal(created)
		}
		if _, err := mg.DB.Collection("resultados_escaner").InsertOne(ctx, bson.M{"diagnosticoId": diagnosis}); err == nil {
			t.Fatal("invalid BSON accepted")
		}
	})
	var buffer bytes.Buffer
	if err = png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, 3, 3))); err != nil {
		t.Fatal(err)
	}
	data := buffer.Bytes()
	var evidence string
	t.Run("GridFS validación descarga e idempotencia", func(t *testing.T) {
		evidence = uploadID(t, technicalUpload(&h, creator, diagnosis, "file-"+h.suffix, "Evidencia fixture", data), 201)
		if id := uploadID(t, technicalUpload(&h, creator, diagnosis, "file-"+h.suffix, "Evidencia fixture", data), 200); id != evidence {
			t.Fatal("duplicate evidence")
		}
		if res := technicalUpload(&h, creator, diagnosis, "file-"+h.suffix, "Cambio", data); res.status != 409 {
			t.Fatal(res.status)
		}
		file := h.send(creator, "GET", base+"/evidencias/"+evidence+"/archivo", nil)
		if file.err != nil || file.status != 200 || !bytes.Equal(file.raw, data) {
			t.Fatalf("download %d %v", file.status, file.err)
		}
		h.call(t, creator, "GET", base+"/evidencias", nil, 200)
		for _, bad := range [][]byte{[]byte("<svg/>"), []byte("%PDF-1.7 incomplete")} {
			if res := technicalUpload(&h, creator, diagnosis, "bad-"+rand.Text(), "", bad); res.status != 422 {
				t.Fatal(res.status)
			}
		}
		if res := technicalUpload(&h, creator, diagnosis, "large-"+h.suffix, "", make([]byte, domain.MaxEvidenceBytes+1)); res.status != 413 {
			t.Fatal(res.status)
		}
		uploadID(t, technicalUpload(&h, creator, diagnosis, "pdf-"+h.suffix, "PDF fixture", []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<<>>\n%%EOF\n")), 201)
	})
	t.Run("autorización PostgreSQL JWT CSRF y aislamiento", func(t *testing.T) {
		h.call(t, other, "GET", base+"/escaneres", nil, 403)
		h.call(t, other, "POST", base+"/escaneres", scan, 403)
		h.call(t, manager, "GET", base+"/evidencias", nil, 200)
		h.call(t, other, "GET", "/diagnosticos/"+otherDiagnosis+"/evidencias/"+evidence+"/archivo", nil, 404)
		h.call(t, creator, "GET", "/diagnosticos/9223372036854775807/escaneres", nil, 404)
		h.call(t, workshopSession{}, "GET", base+"/evidencias", nil, 401)
		invalid := creator
		invalid.csrf = "wrong"
		h.call(t, invalid, "POST", base+"/escaneres", scan, 403)
	})
	t.Run("fallo MongoDB tras subir GridFS y compensación", func(t *testing.T) {
		failpoint := bson.D{{Key: "configureFailPoint", Value: "failCommand"}, {Key: "mode", Value: bson.M{"times": 1}}, {Key: "data", Value: bson.M{"failCommands": bson.A{"update"}, "errorCode": 121}}}
		if err := mg.Client.Database("admin").RunCommand(ctx, failpoint).Err(); err != nil {
			t.Fatal("enableTestCommands requerido en instancia aislada", err)
		}
		defer mg.Client.Database("admin").RunCommand(ctx, bson.D{{Key: "configureFailPoint", Value: "failCommand"}, {Key: "mode", Value: "off"}})
		res := technicalUpload(&h, creator, diagnosis, "fail-"+h.suffix, "Fallo metadata", data)
		if res.status != 503 {
			t.Fatalf("fault %d %s", res.status, res.raw)
		}
		count, err := mg.DB.Collection("evidencias_tecnicas").CountDocuments(ctx, bson.M{"diagnosticoId": diagnosis, "requestId": "fail-" + h.suffix})
		if err != nil || count != 0 {
			t.Fatal("reservation leak", count, err)
		}
		uploadID(t, technicalUpload(&h, creator, diagnosis, "fail-"+h.suffix, "Fallo metadata", data), 201)
	})
	t.Run("evidencia concurrente reserva vencida reconciliación", func(t *testing.T) {
		req := "parallel-file-" + h.suffix
		var wg sync.WaitGroup
		results := make(chan workshopResponse, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); results <- technicalUpload(&h, creator, diagnosis, req, "Concurrente", data) }()
		}
		wg.Wait()
		close(results)
		created := 0
		for res := range results {
			if res.err != nil || (res.status != 201 && res.status != 200 && res.status != 409) {
				t.Fatalf("concurrent file %d %s %v", res.status, res.raw, res.err)
			}
			if res.status == 201 {
				created++
			}
		}
		if created != 1 {
			t.Fatal(created)
		}
		uploadID(t, technicalUpload(&h, creator, diagnosis, req, "Concurrente", data), 200)
		var old struct {
			FileID string `bson:"gridfsFileId"`
		}
		if err := mg.DB.Collection("evidencias_tecnicas").FindOne(ctx, bson.M{"diagnosticoId": diagnosis, "requestId": req}).Decode(&old); err != nil {
			t.Fatal(err)
		}
		if _, err := mg.DB.Collection("evidencias_tecnicas").UpdateOne(ctx, bson.M{"diagnosticoId": diagnosis, "requestId": req}, bson.M{"$set": bson.M{"state": "uploading", "leaseUntil": time.Now().Add(-20 * time.Minute)}}); err != nil {
			t.Fatal(err)
		}
		uploadID(t, technicalUpload(&h, creator, diagnosis, req, "Concurrente", data), 201)
		oldID, err := bson.ObjectIDFromHex(old.FileID)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = mg.DB.Collection("archivos_tecnicos.files").UpdateOne(ctx, bson.M{"_id": oldID}, bson.M{"$set": bson.M{"uploadDate": time.Now().Add(-20 * time.Minute)}}); err != nil {
			t.Fatal(err)
		}
		if n, err := mg.Reconcile(ctx, true); err != nil || n < 1 {
			t.Fatal("dry-run", n, err)
		}
		if n, err := mg.Reconcile(ctx, false); err != nil || n < 1 {
			t.Fatal("repair", n, err)
		}
		if n, err := mg.DB.Collection("archivos_tecnicos.chunks").CountDocuments(ctx, bson.M{"files_id": oldID}); err != nil || n != 0 {
			t.Fatal("orphan chunks", n, err)
		}
		file := h.send(creator, "GET", base+"/evidencias/"+evidence+"/archivo", nil)
		if file.status != 200 || !bytes.Equal(file.raw, data) {
			t.Fatal("reconcile damaged ready file", file.status)
		}
	})
	t.Run("MongoDB no disponible no afecta PostgreSQL", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		address := listener.Addr().String()
		listener.Close()
		down, err := mongoadapter.Open("mongodb://"+address+"/?directConnection=true", name)
		if err != nil {
			t.Fatal(err)
		}
		defer down.Close(ctx)
		handler, err := httpapi.New(&postgres.Store{Pool: pool}, httpapi.Options{Origin: "http://localhost:8080", Key: key, Technical: down})
		if err != nil {
			t.Fatal(err)
		}
		server := httptest.NewServer(handler)
		defer server.Close()
		failed := workshopHarness{api: server}
		res := failed.send(workshopSession{}, "POST", "/auth/login", map[string]string{"username": "dcreator_" + h.suffix, "password": h.password})
		if res.status != 200 {
			t.Fatal("SQL login", res.status)
		}
		var parsed struct {
			CSRF string `json:"csrfToken"`
		}
		decodeTest(t, res.raw, &parsed)
		session := workshopSession{res.cookies[0], parsed.CSRF}
		failed.call(t, session, "GET", base+"/escaneres", nil, 503)
		failed.call(t, session, "GET", base, nil, 200)
	})
	if count := h.scalar(t, `SELECT count(*)::text FROM diagnosticos WHERE id_diagnostico=$1`, diagnosis); count != "1" {
		t.Fatal("SQL diagnosis affected", count)
	}
}
