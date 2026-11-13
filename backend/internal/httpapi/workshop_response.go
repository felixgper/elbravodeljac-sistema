package httpapi

import (
	"encoding/json"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
)

// workshopResponse applies the monetary field policy to order and quotation
// representations, including those returned by auxiliary lookups.
func workshopResponse(value any, resource string, list bool, user domain.User) (any, error) {
	var project func(map[string]json.RawMessage) (map[string]json.RawMessage, error)
	switch resource {
	case "ordenes", "tareas":
		if user.Can("pagos.ver") {
			return value, nil
		}
		project = func(row map[string]json.RawMessage) (map[string]json.RawMessage, error) {
			delete(row, "amountDue")
			return row, nil
		}
	case "cotizaciones":
		if user.Can("cotizaciones.ver") {
			return value, nil
		}
		project = technicalQuote
	default:
		return value, nil
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, err
	}
	if !list {
		return project(object)
	}
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(object["data"], &rows); err != nil {
		return nil, err
	}
	for i, row := range rows {
		rows[i], err = project(row)
		if err != nil {
			return nil, err
		}
	}
	object["data"], err = json.Marshal(rows)
	return object, err
}

func technicalQuote(row map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	// Approval snapshots and revision history can include the entire priced
	// quote. The technical response uses an allowlist at each retained level.
	out := selectFields(row, "id", "clientName", "diagnosisId", "receptionId", "orderId", "orderCode", "status", "createdAt", "approvedAt", "version", "parentQuoteId")
	if approval, ok := row["approval"]; ok {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(approval, &fields); err != nil {
			return nil, err
		}
		if fields != nil {
			fields = selectFields(fields, "name", "method", "by", "version")
		}
		raw, err := json.Marshal(fields)
		if err != nil {
			return nil, err
		}
		out["approval"] = raw
	}
	if raw, ok := row["items"]; ok {
		var items []map[string]json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return nil, err
		}
		for i, item := range items {
			items[i] = selectFields(item, "id", "serviceId", "sparePartId", "concept", "type", "partKind", "origin", "quantity")
		}
		raw, err := json.Marshal(items)
		if err != nil {
			return nil, err
		}
		out["items"] = raw
	}
	return out, nil
}

func selectFields(row map[string]json.RawMessage, keys ...string) map[string]json.RawMessage {
	out := make(map[string]json.RawMessage, len(keys))
	for _, key := range keys {
		if value, ok := row[key]; ok {
			out[key] = value
		}
	}
	return out
}
