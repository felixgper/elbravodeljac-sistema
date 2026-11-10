package postgres

import (
	"context"
	_ "embed"
	"fmt"
)

//go:embed migrations/001_revision_funcional.sql
var FunctionalMigration string

//go:embed migrations/002_marcas_vehiculos.sql
var BrandsMigration string

//go:embed migrations/003_atencion_ficha.sql
var AttentionMigration string

//go:embed migrations/004_flujo_diagnosticos.sql
var DiagnosisWorkflowMigration string

//go:embed migrations/005_auditoria_flujo.sql
var AuditMigration string

//go:embed migrations/006_pagos_entrega.sql
var PaymentsMigration string

//go:embed migrations/007_empleados.sql
var EmployeesMigration string

//go:embed migrations/008_tipos_documento.sql
var DocumentTypesMigration string

//go:embed migrations/009_tres_roles.sql
var ThreeRolesMigration string

const consolidatedSchemaMarker = "elbravodeljac:esquema_entrega:v1"

// Migrate only applies historical migrations to installations without the
// consolidated-schema marker. Fresh installations already contain all DDL.
func (s *Store) Migrate(ctx context.Context) error {
	var marker string
	if err := s.Pool.QueryRow(ctx, `SELECT coalesce(obj_description('public'::regnamespace, 'pg_namespace'), '')`).Scan(&marker); err != nil {
		return err
	}
	if marker == consolidatedSchemaMarker {
		var tables, columns, fks int
		err := s.Pool.QueryRow(ctx, `
			SELECT
				(SELECT count(*) FROM information_schema.tables
				 WHERE table_schema='public' AND table_type='BASE TABLE'),
				(SELECT count(*) FROM information_schema.columns c
				 JOIN information_schema.tables t
				   ON c.table_schema=t.table_schema AND c.table_name=t.table_name
				 WHERE t.table_schema='public' AND t.table_type='BASE TABLE'),
				(SELECT count(*) FROM pg_constraint
				 WHERE contype='f' AND connamespace='public'::regnamespace)
		`).Scan(&tables, &columns, &fks)
		if err != nil { return err }
		if tables != 28 || columns != 208 || fks != 50 {
			return fmt.Errorf("esquema consolidado incompleto: tablas=%d columnas=%d fk=%d; no ejecutar migraciones historicas", tables, columns, fks)
		}
		// Schema-only export has no rows. Reproduce the fixed catalog that
		// migration 008 populates for a legacy installation.
		_, err = s.Pool.Exec(ctx, `INSERT INTO public.tipos_documento(codigo,nombre) VALUES
			('DNI','Documento Nacional de Identidad'),
			('RUC','Registro Único de Contribuyentes'),
			('CE','Carné de Extranjería'),
			('PASAPORTE','Pasaporte')
			ON CONFLICT(codigo) DO NOTHING`)
		return err
	}
	_, err := s.Pool.Exec(ctx, FunctionalMigration+"\n"+BrandsMigration+"\n"+AttentionMigration+"\n"+DiagnosisWorkflowMigration+"\n"+AuditMigration+"\n"+PaymentsMigration+"\n"+EmployeesMigration+"\n"+DocumentTypesMigration+"\n"+ThreeRolesMigration)
	return err
}
