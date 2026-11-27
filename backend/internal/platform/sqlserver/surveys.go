package sqlserver

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
)

// No schema change or business write: only a completed, verified delivery can
// generate an invitation. Financial rules remain in Deliver, not duplicated here.
func (s *Store) SurveyDelivery(ctx context.Context, orderID string) (domain.SurveyDelivery, error) {
	var v domain.SurveyDelivery
	err := s.db.QueryRowContext(ctx, `SELECT CONVERT(varchar(20),e.id_entrega),CONVERT(varchar(20),e.id_orden)
FROM dbo.entregas e JOIN dbo.ordenes_trabajo o ON o.id_orden=e.id_orden
WHERE e.id_orden=@p1 AND o.estado='finalizada' AND e.autorizacion_verificada=1 AND e.fecha_entrega IS NOT NULL`, orderID).Scan(&v.ID, &v.OrderID)
	return v, translate(err)
}
