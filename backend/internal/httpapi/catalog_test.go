package httpapi

import "testing"

func TestOriginalCatalogRelationalCoverage(t *testing.T) {
	s := &server{}
	implemented := map[string]bool{"POST /api/v1/auth/login": true}
	for _, e := range append(append(append(s.endpoints(), s.workshopEndpoints()...), s.relationalEndpoints()...), s.technicalEndpoints()...) {
		key := e.method + " " + e.path
		if implemented[key] {
			t.Fatal("duplicate route", key)
		}
		if e.permission == "" && e.path != "/api/v1/auth/me" && e.path != "/api/v1/auth/logout" {
			t.Fatal("unprotected operation", key)
		}
		implemented[key] = true
	}
	pending := map[string]bool{"GET /api/v1/diagnosticos/{id}/escaneres": true, "POST /api/v1/diagnosticos/{id}/escaneres": true, "GET /api/v1/diagnosticos/{id}/evidencias": true, "POST /api/v1/diagnosticos/{id}/evidencias": true, "GET /api/v1/diagnosticos/{id}/evidencias/{evidenciaId}/archivo": true}
	n := 0
	for _, r := range Routes {
		key := r.Method + " " + r.Path
		if implemented[key] {
			n++
		} else if !pending[key] {
			t.Fatal("relational placeholder remains", key)
		}
	}
	if len(Routes) != 71 || n != 71 || len(implemented) != 86 {
		t.Fatal("unexpected catalog", len(Routes), n, len(implemented))
	}
}
