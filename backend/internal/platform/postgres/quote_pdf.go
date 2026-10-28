package postgres

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
	out, err := readJSON(ctx, s.Pool, `SELECT jsonb_build_object('workshop',(SELECT to_jsonb(t) FROM datos_taller_ficha t),'quote',`+quoteJSON+`,'client',cl.nombre,'document',cl.tipo_documento||' '||cl.numero_documento,'phone',cl.telefono,'email',cl.correo,'address',cl.direccion,'plate',v.placa,'vehicle',v.marca||' '||v.modelo||' · '||coalesce(v.anio::text,'No registrado')) FROM cotizaciones q JOIN recepciones r ON r.id_recepcion=q.id_recepcion JOIN clientes cl ON cl.id_cliente=r.id_cliente_atencion JOIN vehiculos v USING(id_vehiculo) WHERE q.id_cotizacion=$1`, id)
	if err != nil {
		return nil, err
	}
	pdf, err := technicalsheet.RenderQuotation(out.(json.RawMessage))
	if err != nil {
		return nil, domain.Invalid("pdf", err.Error())
	}
	return pdf, nil
}
