package postgres

import (
	"context"
	_ "embed"
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

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.Pool.Exec(ctx, FunctionalMigration+"\n"+BrandsMigration+"\n"+AttentionMigration+"\n"+DiagnosisWorkflowMigration+"\n"+AuditMigration+"\n"+PaymentsMigration+"\n"+EmployeesMigration)
	return err
}
