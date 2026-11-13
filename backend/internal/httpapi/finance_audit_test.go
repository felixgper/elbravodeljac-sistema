package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/technicalsheet"
	"golang.org/x/crypto/bcrypt"
)

type financeAuditStore struct{ stubStore }

var auditOrder = json.RawMessage(`{"id":"1","amountDue":"100.25","state":"En reparación","assignments":[{"userId":"1"}]}`)
var auditQuote = json.RawMessage(`{"id":"2","total":"100.25","diagnosisId":"3","status":"Aprobada","items":[{"concept":"Revisión","quantity":"1.00","serviceId":"4","unitPrice":"100.25","discount":"0.00","subtotal":"100.25"},{"concept":"Filtro","quantity":"2.00","sparePartId":"5"}],"approval":{"method":"WhatsApp","total":"100.25","items":[{"unitPrice":"100.25"}]},"revisionHistory":[{"total":"100.25","items":[{"precio_unitario":"100.25","descuento_unitario":"0.00"}]}]}`)

func (*financeAuditStore) WorkshopList(context.Context, string, domain.WorkshopFilter, domain.User) (any, error) {
	return map[string]any{"data": []json.RawMessage{auditOrder}, "meta": map[string]int{"total": 1}}, nil
}

func (*financeAuditStore) WorkshopItem(_ context.Context, resource, _ string, _ domain.User) (any, error) {
	if resource == "pagos" {
		return json.RawMessage(`{"total":"100.25","paid":"60.00","balance":"40.25"}`), nil
	}
	if resource == "cotizaciones" {
		return auditQuote, nil
	}
	return auditOrder, nil
}

func (*financeAuditStore) UpdateOrder(context.Context, string, domain.UpdateOrderInput, domain.User) (any, error) {
	return auditOrder, nil
}

func (*financeAuditStore) FrontendOptions(_ context.Context, _, resource string, _ domain.Page, _ domain.User) (any, error) {
	value := auditOrder
	if resource == "cotizaciones" {
		value = auditQuote
	}
	return map[string]any{"data": []json.RawMessage{value}}, nil
}

func (*financeAuditStore) Inspection(context.Context, string, domain.User) (any, error) {
	panic("unexpected inspection")
}
func (*financeAuditStore) SaveInspection(context.Context, string, technicalsheet.Control, domain.User) (any, error) {
	panic("unexpected inspection update")
}
func (*financeAuditStore) SheetPDF(context.Context, string, domain.User) ([]byte, error) {
	return []byte("%PDF-synthetic"), nil
}

func (*financeAuditStore) Statistics(context.Context, bool, time.Time, time.Time) (any, error) {
	return map[string]any{"monthlyIncome": "100.25", "monthlyIncomeCents": "10025", "income": []int{10025}, "activeOrders": 1}, nil
}

// Characterizes the HTTP boundary using synthetic projections, independently of SQL.
func TestMechanicFinancialBoundaryAudit(t *testing.T) {
	const origin = "https://audit.example"
	hash, err := bcrypt.GenerateFromPassword([]byte("audit-fixture-only"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	store := &financeAuditStore{stubStore{user: domain.User{ID: "1", Active: true, PasswordHash: string(hash), Permissions: []string{"inicio.ver", "ordenes.ver", "auxilio.ver", "ordenes.editar"}}}}
	h, err := New(store, Options{Origin: origin, Key: []byte(strings.Repeat("a", 32)), SecureCookies: true, SQLServer: true})
	if err != nil {
		t.Fatal(err)
	}
	login := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"mecanico_qa","password":"audit-fixture-only"}`))
	login.Header.Set("Content-Type", "application/json")
	login.Header.Set("Origin", origin)
	auth := httptest.NewRecorder()
	h.ServeHTTP(auth, login)
	if auth.Code != 200 {
		t.Fatalf("login: %d", auth.Code)
	}
	cookie := auth.Result().Cookies()[0]
	var loginResult struct {
		CSRF string `json:"csrfToken"`
	}
	if err := json.Unmarshal(auth.Body.Bytes(), &loginResult); err != nil {
		t.Fatal(err)
	}
	orderFields := []string{`"state":"En reparación"`, `"assignments"`}
	quoteFields := []string{`"concept":"Revisión"`, `"quantity":"1.00"`, `"serviceId":"4"`, `"sparePartId":"5"`, `"diagnosisId":"3"`, `"status":"Aprobada"`, `"method":"WhatsApp"`}
	orderDenied := []string{`"amountDue"`, `100.25`}
	quoteDenied := []string{`100.25`, `"unitPrice"`, `"discount"`, `"subtotal"`, `"revisionHistory"`, `"precio_unitario"`, `"descuento_unitario"`}

	for _, tc := range []struct {
		path               string
		status             int
		includes, excludes []string
	}{
		{"/api/v1/ordenes", 200, orderFields, orderDenied},
		{"/api/v1/me/tareas", 200, orderFields, orderDenied},
		{"/api/v1/ordenes/1", 200, orderFields, orderDenied},
		{"/api/v1/frontend/opciones?module=auxilio&resource=ordenes", 200, orderFields, orderDenied},
		{"/api/v1/frontend/opciones?module=ordenes&resource=cotizaciones", 200, quoteFields, quoteDenied},
		{"/api/v1/frontend/opciones?module=auxilio&resource=cotizaciones", 200, quoteFields, quoteDenied},
		{"/api/v1/dashboard", 200, []string{`"financialVisible":false`, `"activeOrders":1`}, []string{`"monthlyIncome"`, `"monthlyIncomeCents"`, `"income"`}},
		{"/api/v1/ordenes/1/pagos", 403, nil, nil},
		{"/api/v1/ordenes/1/ficha-tecnica.pdf", 403, nil, nil},
		{"/api/v1/cotizaciones/1", 403, nil, nil},
		{"/api/v1/reportes", 403, nil, nil},
	} {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.AddCookie(cookie)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("got %d want %d", w.Code, tc.status)
			}
			for _, field := range tc.includes {
				if !strings.Contains(w.Body.String(), field) {
					t.Errorf("missing %s", field)
				}
			}
			for _, field := range tc.excludes {
				if strings.Contains(w.Body.String(), field) {
					t.Errorf("unexpected %s", field)
				}
			}
		})
	}
	check := func(method, path string) string {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(`{"state":"En reparación","note":"QA"}`))
		r.AddCookie(cookie)
		r.Header.Set("Origin", origin)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-CSRF-Token", loginResult.CSRF)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("%s %s: %d", method, path, w.Code)
		}
		return w.Body.String()
	}
	if body := check(http.MethodPatch, "/api/v1/ordenes/1"); strings.Contains(body, `"amountDue"`) || !strings.Contains(body, `"state"`) {
		t.Fatal(body)
	}
	status := func(path string) int {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.AddCookie(cookie)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code
	}

	base := append([]string(nil), store.user.Permissions...)
	store.user.Permissions = append(base, "pagos.ver")
	if got := status("/api/v1/ordenes/1/ficha-tecnica.pdf"); got != http.StatusForbidden {
		t.Errorf("payment-only PDF: got %d", got)
	}
	if body := check(http.MethodGet, "/api/v1/ordenes/1"); !strings.Contains(body, `"amountDue":"100.25"`) {
		t.Fatal(body)
	}
	if body := check(http.MethodGet, "/api/v1/ordenes/1/pagos"); !strings.Contains(body, `"balance":"40.25"`) {
		t.Fatal(body)
	}
	if body := check(http.MethodGet, "/api/v1/frontend/opciones?module=pagos&resource=cotizaciones"); strings.Contains(body, "100.25") {
		t.Fatal(body)
	}
	store.user.Permissions = append(base, "cotizaciones.ver")
	if body := check(http.MethodGet, "/api/v1/frontend/opciones?module=ordenes&resource=cotizaciones"); !strings.Contains(body, `"unitPrice":"100.25"`) || !strings.Contains(body, `"revisionHistory"`) {
		t.Fatal(body)
	}
	if body := check(http.MethodGet, "/api/v1/ordenes/1"); strings.Contains(body, `"amountDue"`) {
		t.Fatal(body)
	}

	if got := status("/api/v1/ordenes/1/ficha-tecnica.pdf"); got != http.StatusForbidden {
		t.Errorf("quote-only PDF: %d", got)
	}
	store.user.Permissions = append(base, "pagos.ver", "cotizaciones.ver")
	if got := status("/api/v1/ordenes/1/ficha-tecnica.pdf"); got != http.StatusOK {
		t.Errorf("authorized PDF: %d", got)
	}

}
