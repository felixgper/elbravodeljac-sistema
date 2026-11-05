package httpapi

import (
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"io"
	"net/http"
	"strconv"
)

func (s *server) technicalEndpoints() []endpoint {
	return []endpoint{
		{"GET", "/api/v1/diagnosticos/{id}/escaneres", "diagnosticos.ver", s.scans},
		{"POST", "/api/v1/diagnosticos/{id}/escaneres", "diagnosticos.crear", s.putScan},
		{"GET", "/api/v1/diagnosticos/{id}/evidencias", "diagnosticos.ver", s.evidenceList},
		{"POST", "/api/v1/diagnosticos/{id}/evidencias", "diagnosticos.crear", s.putEvidence},
		{"GET", "/api/v1/diagnosticos/{id}/evidencias/{evidenciaId}/archivo", "diagnosticos.ver", s.technicalFile},
	}
}
func (s *server) technicalAccess(w http.ResponseWriter, r *http.Request, u domain.User) (string, bool) {
	id, ok := pathID(w, r)
	if !ok {
		return "", false
	}
	if err := s.store.AuthorizeDiagnosis(r.Context(), id, u.ID, r.Method == "POST"); err != nil {
		s.failure(w, err)
		return "", false
	}
	if s.options.Technical == nil {
		writeError(w, 503, "MONGODB_UNAVAILABLE", "MongoDB no está configurado.")
		return "", false
	}
	return id, true
}
func (s *server) scans(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := s.technicalAccess(w, r, u)
	if !ok {
		return
	}
	p, err := page(r)
	if err != nil {
		s.failure(w, err)
		return
	}
	out, err := s.options.Technical.Scans(r.Context(), id, p)
	if err != nil {
		s.failure(w, err)
		return
	}
	respond(w, 200, out)
}
func (s *server) putScan(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := s.technicalAccess(w, r, u)
	if !ok {
		return
	}
	var v domain.ScanInput
	if !decodeLimit(w, r, &v, 256<<10) {
		return
	}
	out, created, err := s.options.Technical.PutScan(r.Context(), id, u.ID, v)
	status := 200
	if created {
		status = 201
	}
	s.result(w, status, out, err)
}
func (s *server) evidenceList(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := s.technicalAccess(w, r, u)
	if !ok {
		return
	}
	p, err := page(r)
	if err != nil {
		s.failure(w, err)
		return
	}
	out, err := s.options.Technical.EvidenceList(r.Context(), id, p)
	if err != nil {
		s.failure(w, err)
		return
	}
	respond(w, 200, out)
}
func evidenceMultipart(w http.ResponseWriter, r *http.Request) (domain.EvidenceInput, bool) {
	var v domain.EvidenceInput
	r.Body = http.MaxBytesReader(w, r.Body, domain.MaxEvidenceBytes+(64<<10))
	reader, err := r.MultipartReader()
	if err != nil {
		writeError(w, 400, "MULTIPART", "Usa multipart/form-data.")
		return v, false
	}
	seen := map[string]bool{}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			writeError(w, 413, "UPLOAD_LIMIT", "Carga inválida o mayor del límite.")
			return v, false
		}
		name := part.FormName()
		if seen[name] || (name != "archivo" && name != "requestId" && name != "descripcion") {
			part.Close()
			writeError(w, 400, "MULTIPART", "Campo desconocido o duplicado.")
			return v, false
		}
		seen[name] = true
		limit := 16001
		if name == "archivo" {
			limit = domain.MaxEvidenceBytes + 1
		} else if name == "requestId" {
			limit = 401
		}
		b, err := io.ReadAll(io.LimitReader(part, int64(limit)))
		part.Close()
		if err != nil {
			writeError(w, 413, "UPLOAD_LIMIT", "No se pudo leer el archivo dentro del límite.")
			return v, false
		}
		switch name {
		case "archivo":
			if len(b) > domain.MaxEvidenceBytes {
				writeError(w, 413, "UPLOAD_LIMIT", "Máximo 10 MiB por archivo.")
				return v, false
			}
			v.Bytes = b
		case "requestId":
			v.RequestID = string(b)
		case "descripcion":
			if len(b) >= 16001 {
				writeError(w, 422, "VALIDATION", "Descripción demasiado grande; máximo 4000 caracteres.")
				return v, false
			}
			v.Description = string(b)
		}
	}
	return v, true
}
func (s *server) putEvidence(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := s.technicalAccess(w, r, u)
	if !ok {
		return
	}
	v, ok := evidenceMultipart(w, r)
	if !ok {
		return
	}
	out, created, err := s.options.Technical.PutEvidence(r.Context(), id, u.ID, v)
	status := 200
	if created {
		status = 201
	}
	if err != nil && status == 200 {
		w.Header().Set("Retry-After", "2")
	}
	s.result(w, status, out, err)
}
func (s *server) technicalFile(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := s.technicalAccess(w, r, u)
	if !ok {
		return
	}
	meta, data, err := s.options.Technical.File(r.Context(), id, r.PathValue("evidenciaId"))
	if err != nil {
		s.failure(w, err)
		return
	}
	ext := map[string]string{"application/pdf": "pdf", "image/png": "png", "image/jpeg": "jpg"}[meta.MIME]
	if ext == "" {
		writeError(w, 503, "FILE_INTEGRITY", "Formato almacenado inválido.")
		return
	}
	w.Header().Set("Content-Type", meta.MIME)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("Content-Disposition", `attachment; filename="evidence-`+meta.ID+`.`+ext+`"`)
	w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'")
	w.WriteHeader(200)
	_, _ = w.Write(data)
}
