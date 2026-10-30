package postgres

import "context"

// Metadata only. Historical inactive labels are returned for existing identities;
// creation remains restricted by the active catalogue and PostgreSQL triggers.
func (s *Store) DocumentTypes(ctx context.Context) (any, error) {
	return readJSON(ctx, s.Pool, `SELECT coalesce(jsonb_agg(jsonb_build_object('id',id_tipo_documento::text,'code',codigo,'name',nombre,'active',activo,'documentType',CASE codigo WHEN 'PASAPORTE' THEN 'Pasaporte' WHEN 'DNI' THEN 'DNI' WHEN 'RUC' THEN 'RUC' WHEN 'CE' THEN 'CE' ELSE nombre END,'label',CASE codigo WHEN 'CE' THEN 'Carné de Extranjería' WHEN 'PASAPORTE' THEN 'Pasaporte' ELSE codigo END) ORDER BY CASE codigo WHEN 'DNI' THEN 0 WHEN 'RUC' THEN 1 WHEN 'CE' THEN 2 WHEN 'PASAPORTE' THEN 3 ELSE 4 END,id_tipo_documento),'[]'::jsonb) FROM tipos_documento`)
}
