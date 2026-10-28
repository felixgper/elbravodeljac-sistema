package technicalsheet

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func quotationText(t *testing.T, pdf []byte) string {
	t.Helper()
	var text strings.Builder
	for {
		start := bytes.Index(pdf, []byte("stream\n"))
		if start < 0 {
			break
		}
		pdf = pdf[start+7:]
		end := bytes.Index(pdf, []byte("\nendstream"))
		if end < 0 {
			t.Fatal("Invalid PDF stream")
		}
		r, err := zlib.NewReader(bytes.NewReader(pdf[:end]))
		if err != nil {
			t.Fatal(err)
		}
		raw, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		text.Write(raw)
		pdf = pdf[end+10:]
	}
	return text.String()
}

func TestQuotationPDF(t *testing.T) {
	for _, status := range []string{"Pendiente", "Aprobada", "Rechazada"} {
		t.Run(status, func(t *testing.T) {
			approval := any(nil)
			if status == "Aprobada" {
				approval = map[string]any{"name": "Cliente prueba", "method": "WhatsApp", "at": "2026-10-27T15:00:00Z", "version": 8}
			}
			snapshot := map[string]any{"workshop": Workshop{Name: "EL BRAVO DEL JAC"}, "client": "Cliente real de prueba", "document": "Pasaporte TEST", "plate": "PDF-TEST", "vehicle": "Toyota prueba", "quote": map[string]any{"id": "159", "version": 12, "createdAt": "2026-10-27T15:00:00Z", "status": status, "total": "185.00", "approval": approval, "items": []map[string]string{{"concept": "Servicio presupuestado", "type": "Servicio", "quantity": "2.00", "unitPrice": "100.00", "discount": "7.50", "subtotal": "185.00"}}}}
			raw, _ := json.Marshal(snapshot)
			pdf, err := RenderQuotation(raw)
			if err != nil {
				t.Fatal(err)
			}
			text := quotationText(t, pdf)
			for _, value := range []string{status, "Servicio presupuestado", "185.00", "7.50", "VERSI", "12", "PDF-TEST"} {
				if !strings.Contains(text, value) {
					t.Fatal("Missing", value)
				}
			}
			if bytes.Contains(pdf, []byte("/Subtype /Image")) {
				t.Fatal("Quote must not use automotive sheet background")
			}
			if status == "Aprobada" && !strings.Contains(text, "versi\xf3n 8") {
				t.Fatal("Approved version must remain distinct from current revision")
			}
			if status != "Aprobada" && strings.Contains(text, "WhatsApp") {
				t.Fatal("Invented approval")
			}
		})
	}
}

func TestQuotationLongContent(t *testing.T) {
	items := []map[string]string{}
	for i := 0; i < 65; i++ {
		items = append(items, map[string]string{"concept": "Trabajo de mantenimiento preventivo y revisión técnica completa " + strings.Repeat("W", 120), "type": "Servicio", "quantity": "999999.99", "unitPrice": "999999999.99", "discount": "0.00", "subtotal": "9999999999.99"})
	}
	items[0]["concept"] = strings.Repeat("Inspección extensa ", 180) + "FIN_PRIMER_CONCEPTO"
	items[64]["concept"] = "ULTIMO_CONCEPTO_CONSERVADO"
	raw, _ := json.Marshal(map[string]any{"workshop": Workshop{Name: "EL BRAVO DEL JAC"}, "quote": map[string]any{"id": "160", "version": 1, "status": "Pendiente", "total": "9999999999.99", "items": items}})
	pdf, err := RenderQuotation(raw)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(pdf, []byte("/Count 1 >>")) {
		t.Fatal("Long quote was not paginated")
	}
	text := quotationText(t, pdf)
	for _, last := range []string{"FIN_PRIMER_CONCEPTO", "ULTIMO_CONCEPTO_CONSERVADO", "TOTAL COTIZADO"} {
		if !strings.Contains(text, last) {
			t.Fatal("Lost content:", last)
		}
	}
	if dir := os.Getenv("QUOTATION_ARTIFACT_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "cotizacion-multipagina.pdf"), pdf, 0600); err != nil {
			t.Fatal(err)
		}
	}
	raw = []byte(`{"quote":{"items":[{"concept":"Trabajo 🚗"}]}}`)
	if _, err = RenderQuotation(raw); err == nil {
		t.Fatal("Unrepresentable characters must fail explicitly rather than disappear")
	}
	for _, line := range quoteWrap(strings.Repeat("W", 300), 10, 252) {
		if len(line) > 24 {
			t.Fatal("Wide glyphs overflow description column")
		}
	}
}

func TestQuotationLegacyApproval(t *testing.T) {
	pdf, err := RenderQuotation([]byte(`{"quote":{"id":"161","version":3,"status":"Aprobada","approval":{},"total":"85.00"}}`))
	if err != nil {
		t.Fatal(err)
	}
	text := quotationText(t, pdf)
	if !strings.Contains(text, "no registrada") || strings.Contains(text, "versi\xf3n 0") {
		t.Fatal("Missing metadata must not become a fictitious approved revision")
	}
}
