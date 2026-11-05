package technicalsheet

import (
	"encoding/json"
	"strings"
)

// Quotation is a read-only snapshot. Monetary values come from PostgreSQL;
// rendering never recomputes discounts, subtotals, or the authorized amount.
type Quotation struct {
	Workshop quotationWorkshop `json:"workshop"`
	Client   string            `json:"client"`
	Document string            `json:"document"`
	Phone    string            `json:"phone"`
	Email    string            `json:"email"`
	Address  string            `json:"address"`
	Plate    string            `json:"plate"`
	Vehicle  string            `json:"vehicle"`
	Quote    struct {
		ID            string `json:"id"`
		Version       int    `json:"version"`
		ParentQuoteID string `json:"parentQuoteId"`
		ReceptionID   string `json:"receptionId"`
		CreatedAt     string `json:"createdAt"`
		Status        string `json:"status"`
		Total         string `json:"total"`
		Approval      *struct {
			Name    string `json:"name"`
			Method  string `json:"method"`
			At      string `json:"at"`
			Version int    `json:"version"`
		} `json:"approval"`
		Items []QuotationItem `json:"items"`
	} `json:"quote"`
}

// RenderQuotation formats a read-only persisted quotation, independent of the
// automotive technical sheet and without changing any financial values.
func RenderQuotation(raw []byte) ([]byte, error) {
	var snapshot Quotation
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, err
	}
	return renderQuotationLayout(snapshot)
}

// Conservative Helvetica/Bold glyph widths: even an unbroken sequence of W's
// fits its column. Keep every rune; newlines and long words remain printable.
func quoteGlyph(r rune) float64 {
	switch {
	case strings.ContainsRune(" ilI.,:;!'|", r):
		return .36
	case r > 127 || strings.ContainsRune("MW@%mw", r):
		return 1.02
	default:
		return .75
	}
}
func quoteWrap(text string, size, max float64) []string {
	var out []string
	for _, paragraph := range strings.Split(text, "\n") {
		runes := []rune(paragraph)
		if len(runes) == 0 {
			out = append(out, "")
			continue
		}
		for start := 0; start < len(runes); {
			end := start
			lastSpace := -1
			used := 0.0
			for end < len(runes) && used+quoteGlyph(runes[end])*size <= max {
				used += quoteGlyph(runes[end]) * size
				if runes[end] == ' ' {
					lastSpace = end
				}
				end++
			}
			if end < len(runes) && lastSpace > start {
				end = lastSpace + 1
			}
			if end == start {
				end++
			}
			out = append(out, string(runes[start:end]))
			start = end
		}
	}
	return out
}

// The configured workshop view may expose a phone; absent values are not invented.
type quotationWorkshop struct {
	Workshop
	Phone string `json:"phone"`
}
type QuotationItem struct {
	Line
	Type     string `json:"type"`
	Origin   string `json:"origin"`
	PartKind string `json:"partKind"`
}
