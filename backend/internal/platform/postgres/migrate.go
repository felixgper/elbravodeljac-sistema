package postgres

import (
	"context"
	_ "embed"
)

//go:embed migrations/001_revision_funcional.sql
var FunctionalMigration string

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.Pool.Exec(ctx, FunctionalMigration)
	return err
}
