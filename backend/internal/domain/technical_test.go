package domain

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

func TestTechnicalValidation(t *testing.T) {
	v := ScanInput{RequestID: "request-1", Equipment: ScannerEquipment{Name: "Equipo"}, DTCs: []DTC{}, Measurements: map[string]any{"rpm": float64(1000)}}
	if err := v.Validate(); err != nil {
		t.Fatal(err)
	}
	v.Measurements = map[string]any{"$where": "invalid"}
	if v.Validate() == nil {
		t.Fatal("reserved key")
	}
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	e := EvidenceInput{RequestID: "file-1", Bytes: b.Bytes()}
	if err := e.Validate(); err != nil || e.MIME != "image/png" {
		t.Fatal(e.MIME, err)
	}
	for _, bad := range [][]byte{[]byte("<svg><script/></svg>"), []byte("MZ executable"), nil, []byte("%PDF-1.7 incomplete"), make([]byte, MaxEvidenceBytes+1)} {
		e.Bytes = bad
		if e.Validate() == nil {
			t.Fatal("accepted invalid file")
		}
	}
}
