package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"net/http"
	"strings"
	"time"
)

const MaxEvidenceBytes = 10 << 20

type ScannerEquipment struct {
	Name    string `json:"nombre" bson:"nombre"`
	Version string `json:"version" bson:"version"`
}
type DTC struct {
	Code        string `json:"codigo" bson:"codigo"`
	Description string `json:"descripcion" bson:"descripcion"`
}
type ScanInput struct {
	RequestID    string           `json:"requestId" bson:"requestId"`
	CapturedAt   *time.Time       `json:"capturadoEn,omitempty" bson:"-"`
	Equipment    ScannerEquipment `json:"equipo" bson:"equipo"`
	DTCs         []DTC            `json:"dtcs" bson:"dtcs"`
	Measurements map[string]any   `json:"mediciones" bson:"mediciones"`
}

func ValidateRequestID(v string) error { return text("requestId", v, 100, true) }
func validateMeasurement(v any, depth int) bool {
	if depth > 8 {
		return false
	}
	switch x := v.(type) {
	case nil, bool:
		return true
	case string:
		return len(x) <= 4000
	case float64:
		return !math.IsNaN(x) && !math.IsInf(x, 0)
	case []any:
		if len(x) > 1000 {
			return false
		}
		for _, v := range x {
			if !validateMeasurement(v, depth+1) {
				return false
			}
		}
		return true
	case map[string]any:
		if len(x) > 100 {
			return false
		}
		for k, v := range x {
			if k == "" || len(k) > 100 || strings.HasPrefix(k, "$") || strings.ContainsAny(k, ".\x00") || !validateMeasurement(v, depth+1) {
				return false
			}
		}
		return true
	}
	return false
}
func (v *ScanInput) Validate() error {
	if err := ValidateRequestID(v.RequestID); err != nil {
		return err
	}
	if err := text("equipo.nombre", v.Equipment.Name, 120, true); err != nil {
		return err
	}
	if err := text("equipo.version", v.Equipment.Version, 80, false); err != nil {
		return err
	}
	if v.CapturedAt != nil {
		at := v.CapturedAt.UTC().Truncate(time.Millisecond)
		if at.After(time.Now().Add(5*time.Minute)) || at.Year() < 2000 {
			return Invalid("capturadoEn", "Fecha de captura fuera de rango.")
		}
		v.CapturedAt = &at
	}
	if v.DTCs == nil || len(v.DTCs) > 1000 {
		return Invalid("dtcs", "Envía un arreglo de hasta 1000 códigos.")
	}
	for _, d := range v.DTCs {
		if err := text("codigo", d.Code, 30, true); err != nil {
			return err
		}
		if err := text("descripcion", d.Description, 1000, false); err != nil {
			return err
		}
	}
	if v.Measurements == nil || !validateMeasurement(v.Measurements, 0) {
		return Invalid("mediciones", "Objeto requerido; profundidad máxima 8 y sin claves MongoDB reservadas.")
	}
	b, err := json.Marshal(v.Measurements)
	if err != nil || len(b) > 64<<10 {
		return Invalid("mediciones", "Máximo 64 KiB de mediciones.")
	}
	return nil
}

type Scan struct {
	ID            string           `json:"id" bson:"-"`
	DiagnosisID   string           `json:"diagnosticoId" bson:"diagnosticoId"`
	ActorID       string           `json:"capturadoPor" bson:"capturadoPor"`
	CapturedAt    time.Time        `json:"capturadoEn" bson:"capturadoEn"`
	CreatedAt     time.Time        `json:"registradoEn" bson:"registradoEn"`
	SchemaVersion int32            `json:"schemaVersion" bson:"schemaVersion"`
	RequestID     string           `json:"requestId" bson:"requestId"`
	Equipment     ScannerEquipment `json:"equipo" bson:"equipo"`
	DTCs          []DTC            `json:"dtcs" bson:"dtcs"`
	Measurements  map[string]any   `json:"mediciones" bson:"mediciones"`
}
type EvidenceInput struct {
	RequestID   string `json:"requestId"`
	Description string `json:"descripcion"`
	Bytes       []byte `json:"-"`
	MIME        string `json:"-"`
	SHA256      string `json:"-"`
}

func (v *EvidenceInput) Validate() error {
	if err := ValidateRequestID(v.RequestID); err != nil {
		return err
	}
	if err := text("descripcion", v.Description, 4000, false); err != nil {
		return err
	}
	if len(v.Bytes) == 0 || len(v.Bytes) > MaxEvidenceBytes {
		return Invalid("archivo", "Archivo requerido de máximo 10 MiB.")
	}
	v.MIME = http.DetectContentType(v.Bytes)
	switch v.MIME {
	case "image/png", "image/jpeg":
		cfg, format, err := image.DecodeConfig(bytes.NewReader(v.Bytes))
		if err != nil || cfg.Width < 1 || cfg.Height < 1 || int64(cfg.Width)*int64(cfg.Height) > 40_000_000 {
			return Invalid("archivo", "Imagen inválida o mayor de 40 megapíxeles.")
		}
		if (v.MIME == "image/png" && format != "png") || (v.MIME == "image/jpeg" && format != "jpeg") {
			return Invalid("archivo", "Formato de imagen inconsistente.")
		}
		if _, _, err = image.Decode(bytes.NewReader(v.Bytes)); err != nil {
			return Invalid("archivo", "Imagen incompleta o corrupta.")
		}
	case "application/pdf":
		trim := bytes.TrimSpace(v.Bytes)
		if !bytes.HasPrefix(trim, []byte("%PDF-")) || !bytes.HasSuffix(trim, []byte("%%EOF")) {
			return Invalid("archivo", "PDF incompleto o inválido.")
		}
	default:
		return Invalid("archivo", "Solo se permiten PDF, JPEG y PNG; se comprueba contenido, no extensión.")
	}
	sum := sha256.Sum256(v.Bytes)
	v.SHA256 = hex.EncodeToString(sum[:])
	return nil
}

type Evidence struct {
	ID            string    `json:"id" bson:"-"`
	DiagnosisID   string    `json:"diagnosticoId" bson:"diagnosticoId"`
	ActorID       string    `json:"capturadoPor" bson:"capturadoPor"`
	CapturedAt    time.Time `json:"capturadoEn" bson:"capturadoEn"`
	CreatedAt     time.Time `json:"registradoEn" bson:"registradoEn"`
	SchemaVersion int32     `json:"schemaVersion" bson:"schemaVersion"`
	RequestID     string    `json:"requestId" bson:"requestId"`
	Description   string    `json:"descripcion" bson:"descripcion"`
	MIME          string    `json:"mimeType" bson:"mimeType"`
	Size          int64     `json:"sizeBytes" bson:"sizeBytes"`
	SHA256        string    `json:"sha256" bson:"sha256"`
	GridFSID      string    `json:"gridfsFileId" bson:"gridfsFileId"`
}
