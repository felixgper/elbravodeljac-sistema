package sqlserver

import "context"

// Serialize Mongo writes with SaveDiagnosis without changing relational data.
// Lock order matches finalization and reception reassignment. Mongo request IDs
// remain authoritative if a commit acknowledgement is lost; no distributed tx.
func (s *Store) WithDiagnosisTechnicalWrite(ctx context.Context, id, actor string, action func() error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return translate(err)
	}
	defer tx.Rollback()
	var reception, state string
	if err = tx.QueryRowContext(ctx, `SELECT CONVERT(varchar(20),id_recepcion) FROM dbo.diagnosticos WHERE id_diagnostico=@p1`, id).Scan(&reception); err != nil {
		return translate(err)
	}
	if err = tx.QueryRowContext(ctx, `SELECT CONVERT(varchar(20),id_recepcion) FROM dbo.recepciones WITH (UPDLOCK,HOLDLOCK) WHERE id_recepcion=@p1`, reception).Scan(&reception); err != nil {
		return translate(err)
	}
	if err = tx.QueryRowContext(ctx, `SELECT estado FROM dbo.diagnosticos WITH (UPDLOCK,HOLDLOCK) WHERE id_diagnostico=@p1`, id).Scan(&state); err != nil {
		return translate(err)
	}
	if err = s.AuthorizeDiagnosis(ctx, id, actor, true); err != nil {
		return err
	}
	if err = action(); err != nil {
		return err
	}
	return translate(tx.Commit())
}
