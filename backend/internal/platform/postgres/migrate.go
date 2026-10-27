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

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.Pool.Exec(ctx, FunctionalMigration+"\n"+BrandsMigration+"\n"+AttentionMigration)
	return err
}
