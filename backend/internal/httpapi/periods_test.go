package httpapi

import (
	"testing"
	"time"
)

func TestBusinessPeriodLima(t *testing.T) {
	// UTC Jan 1 02:00 still belongs to Dec 31 in Lima.
	now := time.Date(2027, 1, 1, 2, 0, 0, 0, time.UTC)
	for _, tc := range []struct{ p, from, to string }{
		{"Semana", "2026-12-28T00:00:00-05:00", "2027-01-04T00:00:00-05:00"},
		{"Mes", "2026-12-01T00:00:00-05:00", "2027-01-01T00:00:00-05:00"},
		{"Año", "2026-01-01T00:00:00-05:00", "2027-01-01T00:00:00-05:00"},
	} {
		a, b, err := businessPeriod(now, tc.p)
		if err != nil || a.Format(time.RFC3339) != tc.from || b.Format(time.RFC3339) != tc.to {
			t.Fatal(tc, a, b, err)
		}
	}
	if _, _, err := businessPeriod(now, "Trimestre"); err == nil {
		t.Fatal("unknown period accepted")
	}
}
