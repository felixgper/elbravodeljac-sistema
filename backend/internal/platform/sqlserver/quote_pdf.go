package sqlserver

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/technicalsheet"
)

// A serializable read keeps the quotation and its details consistent while
// another request edits them. Rendering does not hold database locks.
func (s *Store) QuotePDF(ctx context.Context, id string, u domain.User) ([]byte, error) {
	if !u.Can("cotizaciones.ver") {
		return nil, domain.ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, translate(err)
	}
	defer tx.Rollback()
	out, err := readJSON(ctx, tx, `SELECT JSON_OBJECT('workshop':JSON_QUERY((SELECT * FROM dbo.datos_taller_ficha FOR JSON PATH, WITHOUT_ARRAY_WRAPPER)),
'quote':JSON_QUERY(`+quoteJSON+`),'client':cl.nombre,'document':CONCAT(cl.tipo_documento,N' ',cl.numero_documento),
'phone':cl.telefono,'email':cl.correo,'address':cl.direccion,'plate':v.placa,
'vehicle':v.marca+N' '+v.modelo+N' · '+COALESCE(CONVERT(varchar(10),v.anio),N'No registrado'))
FROM dbo.cotizaciones q JOIN dbo.recepciones r ON r.id_recepcion=q.id_recepcion
JOIN dbo.clientes cl ON cl.id_cliente=r.id_cliente_atencion JOIN dbo.vehiculos v ON v.id_vehiculo=r.id_vehiculo
WHERE q.id_cotizacion=@p1`, id)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, translate(err)
	}
	pdf, err := technicalsheet.RenderQuotation(out.(json.RawMessage))
	if err != nil {
		return nil, domain.Invalid("pdf", err.Error())
	}
	return pdf, nil
}
