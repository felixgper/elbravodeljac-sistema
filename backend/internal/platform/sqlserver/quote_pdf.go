package sqlserver

import (
	"context"
	"encoding/json"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/technicalsheet"
)

// One statement provides a consistent snapshot without updating the quote,
// approval, revision history or any cached automotive sheet.
func (s *Store) QuotePDF(ctx context.Context, id string, u domain.User) ([]byte, error) {
	if !u.Can("cotizaciones.ver") {
		return nil, domain.ErrForbidden
	}
	out, err := readJSON(ctx, s.db, `SELECT JSON_OBJECT('workshop':JSON_QUERY((SELECT * FROM dbo.datos_taller_ficha FOR JSON PATH, WITHOUT_ARRAY_WRAPPER)),
'quote':JSON_QUERY(`+quoteJSON+`),'client':cl.nombre,'document':CONCAT(cl.tipo_documento,N' ',cl.numero_documento),
'phone':cl.telefono,'email':cl.correo,'address':cl.direccion,'plate':v.placa,
'vehicle':v.marca+N' '+v.modelo+N' · '+COALESCE(CONVERT(varchar(10),v.anio),N'No registrado'))
FROM dbo.cotizaciones q JOIN dbo.recepciones r ON r.id_recepcion=q.id_recepcion
JOIN dbo.clientes cl ON cl.id_cliente=r.id_cliente_atencion JOIN dbo.vehiculos v ON v.id_vehiculo=r.id_vehiculo
WHERE q.id_cotizacion=@p1`, id)
	if err != nil {
		return nil, err
	}
	pdf, err := technicalsheet.RenderQuotation(out.(json.RawMessage))
	if err != nil {
		return nil, domain.Invalid("pdf", err.Error())
	}
	return pdf, nil
}
