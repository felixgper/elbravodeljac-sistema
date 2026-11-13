package sqlserver

import "testing"

// Uses only the synthetic order and accounts provisioned by TestSQLServer2025.
func auditAssignedOrderFinancialProjection(t *testing.T, manager, worker, other qaSession, order, quote, amount string) {
	t.Helper()
	path := "/ordenes/" + order
	if data := qaData(t, manager.call(t, "GET", path, nil, 200)); data["amountDue"] != amount {
		t.Errorf("authorized amountDue: %v", data["amountDue"])
	}
	if data := qaData(t, worker.call(t, "GET", path, nil, 200)); data["state"] == nil {
		t.Error("assigned order lost its technical state")
	} else if _, exists := data["amountDue"]; exists {
		t.Error("assigned order contains amountDue")
	}

	for _, endpoint := range []string{"/ordenes", "/me/tareas", "/frontend/opciones?module=auxilio&resource=ordenes"} {
		entry := auditFindRow(t, worker.call(t, "GET", endpoint, nil, 200), order)
		if _, exists := entry["amountDue"]; exists {
			t.Errorf("%s: restricted amountDue", endpoint)
		}
	}
	for _, module := range []string{"ordenes", "auxilio"} {
		endpoint := "/frontend/opciones?module=" + module + "&resource=cotizaciones"
		entry := auditFindRow(t, worker.call(t, "GET", endpoint, nil, 200), quote)
		for _, field := range []string{"total", "revisionHistory"} {
			if _, ok := entry[field]; ok {
				t.Errorf("%s: restricted %s", endpoint, field)
			}
		}
		if approval, ok := entry["approval"].(map[string]any); ok {
			for _, field := range []string{"total", "items"} {
				if _, ok := approval[field]; ok {
					t.Errorf("%s: restricted approval.%s", endpoint, field)
				}
			}
		}
		items, ok := entry["items"].([]any)
		if !ok || len(items) == 0 {
			t.Fatalf("%s: missing quote items", endpoint)
		}
		item, ok := items[0].(map[string]any)
		if !ok {
			t.Fatalf("%s: invalid quote item", endpoint)
		}
		if item["concept"] == nil || item["quantity"] == nil {
			t.Error("missing technical item fields")
		}
		for _, field := range []string{"unitPrice", "discount", "subtotal"} {
			if _, ok := item[field]; ok {
				t.Errorf("%s: restricted %s", endpoint, field)
			}
		}
	}
	other.call(t, "GET", path, nil, 404)
	worker.call(t, "GET", path+"/pagos", nil, 403)
	worker.call(t, "GET", "/cotizaciones/"+quote, nil, 403)
	worker.call(t, "GET", "/reportes", nil, 403)
}

func auditFindRow(t *testing.T, raw []byte, id string) map[string]any {
	t.Helper()
	rows, ok := qaJSON(t, raw)["data"].([]any)
	if !ok {
		t.Fatal("expected paginated array")
	}
	for _, row := range rows {
		if item, ok := row.(map[string]any); ok && item["id"] == id {
			return item
		}
	}
	t.Fatalf("synthetic row %s missing", id)
	return nil
}
