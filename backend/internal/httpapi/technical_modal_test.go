package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/application"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type technicalModalStore struct {
	stubStore
	closed, race bool
}

func (s *technicalModalStore) AuthorizeDiagnosis(_ context.Context, id, actor string, write bool) error {
	if id != "1" || actor != s.user.ID {
		return domain.ErrForbidden
	}
	if write && s.closed {
		return domain.Invalid("diagnosisId", "Diagnóstico finalizado: solo consulta.")
	}
	return nil
}
func (s *technicalModalStore) WithDiagnosisTechnicalWrite(ctx context.Context, id, actor string, f func() error) error {
	if s.race {
		s.closed = true
	}
	if err := s.AuthorizeDiagnosis(ctx, id, actor, true); err != nil {
		return err
	}
	return f()
}

type technicalModalData struct {
	application.Technical
	scans map[string]domain.Scan
	files map[string]domain.EvidenceInput
}

func (m *technicalModalData) PutScan(_ context.Context, id, actor string, v domain.ScanInput) (domain.Scan, bool, error) {
	if err := v.Validate(); err != nil {
		return domain.Scan{}, false, err
	}
	if s, ok := m.scans[v.RequestID]; ok {
		return s, false, nil
	}
	s := domain.Scan{ID: "scan", DiagnosisID: id, ActorID: actor, RequestID: v.RequestID, Equipment: v.Equipment, DTCs: v.DTCs, Measurements: v.Measurements}
	m.scans[v.RequestID] = s
	return s, true, nil
}
func (m *technicalModalData) Scans(context.Context, string, domain.Page) (domain.List[domain.Scan], error) {
	rows := []domain.Scan{}
	for _, v := range m.scans {
		rows = append(rows, v)
	}
	return domain.List[domain.Scan]{Data: rows}, nil
}
func (m *technicalModalData) PutEvidence(_ context.Context, id, actor string, v domain.EvidenceInput) (domain.Evidence, bool, error) {
	if err := v.Validate(); err != nil {
		return domain.Evidence{}, false, err
	}
	_, exists := m.files[v.RequestID]
	m.files[v.RequestID] = v
	return domain.Evidence{ID: "file", DiagnosisID: id, SHA256: v.SHA256, MIME: v.MIME, RequestID: v.RequestID}, !exists, nil
}
func (m *technicalModalData) File(context.Context, string, string) (domain.Evidence, []byte, error) {
	v := m.files["file-request"]
	return domain.Evidence{ID: "file", MIME: v.MIME, SHA256: v.SHA256}, v.Bytes, nil
}
func TestTechnicalModalHTTPGuards(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("unit-only-password"), bcrypt.MinCost)
	s := &technicalModalStore{stubStore: stubStore{user: domain.User{ID: "1", Active: true, PasswordHash: string(hash), Permissions: []string{"diagnosticos.ver", "diagnosticos.crear"}}}}
	m := &technicalModalData{scans: map[string]domain.Scan{}, files: map[string]domain.EvidenceInput{}}
	h, err := New(s, Options{Origin: "https://unit.test", Key: []byte(strings.Repeat("k", 32)), Technical: m})
	if err != nil {
		t.Fatal(err)
	}
	var cookies []*http.Cookie
	csrf := ""
	call := func(method, path string, body []byte, ctype string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, bytes.NewReader(body))
		r.Header.Set("Origin", "https://unit.test")
		r.Header.Set("Content-Type", ctype)
		r.Header.Set("X-CSRF-Token", csrf)
		for _, c := range cookies {
			r.AddCookie(c)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	base := "/api/v1/diagnosticos/1"
	if w := call("POST", base+"/escaneres", []byte(`{}`), "application/json"); w.Code != 401 {
		t.Fatal("anonymous write", w.Code)
	}
	login := call("POST", "/api/v1/auth/login", []byte(`{"username":"unit","password":"unit-only-password"}`), "application/json")
	if login.Code != 200 {
		t.Fatal(login.Code)
	}
	cookies = login.Result().Cookies()
	var v struct {
		CSRF string `json:"csrfToken"`
	}
	json.Unmarshal(login.Body.Bytes(), &v)
	csrf = v.CSRF
	scan := []byte(`{"requestId":"scan-request","equipo":{"nombre":"Equipo QA","version":""},"dtcs":[{"codigo":"P0300","descripcion":"QA"}],"mediciones":{}}`)
	for _, status := range []int{201, 200} {
		if w := call("POST", base+"/escaneres", scan, "application/json"); w.Code != status {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if len(m.scans) != 1 {
		t.Fatal("duplicate scan")
	}
	file := func(data []byte) *httptest.ResponseRecorder {
		var b bytes.Buffer
		f := multipart.NewWriter(&b)
		f.WriteField("requestId", "file-request")
		part, _ := f.CreateFormFile("archivo", "qa.png")
		part.Write(data)
		f.Close()
		return call("POST", base+"/evidencias", b.Bytes(), f.FormDataContentType())
	}
	if w := file([]byte("not an image")); w.Code != 422 || len(m.files) != 0 {
		t.Fatal("invalid content accepted", w.Code)
	}
	var b bytes.Buffer
	png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	if w := file(b.Bytes()); w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := file(b.Bytes()); w.Code != 200 {
		t.Fatal("replay", w.Code)
	}
	if w := call("GET", base+"/evidencias/file/archivo", nil, ""); w.Code != 200 || !bytes.Equal(w.Body.Bytes(), b.Bytes()) {
		t.Fatal("file changed", w.Code)
	}
	s.race = true
	if w := call("POST", base+"/escaneres", scan, "application/json"); w.Code != 422 {
		t.Fatal("finalization race accepted", w.Code)
	}
	if w := call("GET", base+"/escaneres", nil, ""); w.Code != 200 {
		t.Fatal("closed read denied", w.Code)
	}
	s.user.Permissions = nil
	if w := call("GET", base+"/escaneres", nil, ""); w.Code != 403 {
		t.Fatal("permission bypass", w.Code)
	}
}
