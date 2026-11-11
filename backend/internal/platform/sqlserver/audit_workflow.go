package sqlserver

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/technicalsheet"
)

func validateExecution(ctx context.Context, tx *sql.Tx, id string) error {
	var version int
	var raw []byte
	if err := tx.QueryRowContext(ctx, `SELECT flujo_version,control_tecnico FROM ordenes_trabajo WHERE id_orden=@p1`, id).Scan(&version, &raw); err != nil {
		return err
	}
	if version == 0 {
		return nil
	} // Historical orders retain their existing workflow.
	var control technicalsheet.Control
	if err := json.Unmarshal(raw, &control); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, `SELECT CONVERT(varchar(20),d.id_detalle) FROM detalle_cotizacion d JOIN ordenes_trabajo o ON o.id_cotizacion=d.id_cotizacion WHERE o.id_orden=@p1`, id)
	if err != nil {
		return err
	}
	unperformed := false
	for rows.Next() {
		var key string
		if err = rows.Scan(&key); err != nil {
			rows.Close()
			return err
		}
		entry := control.Execution[key]
		if entry.Status != "Realizado" && entry.Status != "No realizado" {
			rows.Close()
			return domain.Invalid("execution", "Concilia todos los conceptos en Control técnico antes de finalizar.")
		}
		if entry.Status == "No realizado" {
			unperformed = true
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if unperformed && control.Reconciliation == nil {
		return domain.Invalid("reconciliation", "La administradora debe autorizar el ajuste por trabajos no realizados.")
	}
	var paid, total string
	if err = tx.QueryRowContext(ctx, `SELECT CONVERT(varchar(42),dbo.jac_importe_orden(@p1)),CONVERT(varchar(42),CAST(COALESCE((SELECT SUM(monto) FROM pagos WHERE id_orden=@p1 AND revertido_en IS NULL),0) AS decimal(38,2)))`, id).Scan(&total, &paid); err != nil {
		return err
	}
	t, _ := domain.Money(total, "total")
	p, _ := domain.Money(paid, "paid")
	if p > t {
		return domain.Invalid("balance", "Revierte o devuelve el exceso cobrado antes de finalizar.")
	}
	return nil
}

func (s *Store) ReversePayment(ctx context.Context, id, payment, requestID, reason string, replacement *domain.PaymentReplacement, u domain.User) (any, error) {
	if !u.Can("pagos.rectificar") {
		return nil, domain.ErrForbidden
	}
	if err := (domain.PaymentCorrection{RequestID: requestID, Reason: reason}).Validate(); err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err = lockOrder(ctx, tx, id); err != nil {
		return nil, err
	}
	var raw []byte
	var reversed bool
	if err = tx.QueryRowContext(ctx, `SELECT `+paymentJSON+`,CONVERT(bit,CASE WHEN p.revertido_en IS NOT NULL THEN 1 ELSE 0 END) FROM dbo.pagos p WITH (UPDLOCK,HOLDLOCK) WHERE id_orden=@p1 AND id_pago=@p2`, id, payment).Scan(&raw, &reversed); err != nil {
		return nil, translate(err)
	}
	if reversed {
		var v struct {
			Correction struct {
				RequestID          string `json:"requestId"`
				Reason             string `json:"reason"`
				ReplacementID      string `json:"replacementId"`
				ReplacementMethod  string `json:"replacementMethod"`
				ReplacementReceipt string `json:"replacementReceipt"`
			} `json:"correction"`
		}
		json.Unmarshal(raw, &v)
		sameReplacement := (replacement == nil && v.Correction.ReplacementID == "") || (replacement != nil && v.Correction.ReplacementID != "" && v.Correction.ReplacementMethod == replacement.Method && v.Correction.ReplacementReceipt == replacement.Receipt)
		if v.Correction.RequestID == requestID && v.Correction.Reason == reason && sameReplacement {
			return json.RawMessage(raw), nil
		}
		return nil, domain.ErrConflict
	}
	closed, err := orderClosed(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if closed && replacement == nil {
		return nil, domain.Invalid("orderId", "Una entrega cerrada requiere sustitución por el mismo importe; no se puede dejar saldo pendiente.")
	}
	var replacementInput domain.PaymentInput
	var replacementID string
	if replacement != nil {
		var original struct {
			Amount string `json:"amount"`
		}
		json.Unmarshal(raw, &original)
		replacementInput = domain.PaymentInput{Amount: original.Amount, Method: replacement.Method, Receipt: replacement.Receipt, RequestID: fmt.Sprintf("%x", sha256.Sum256([]byte("replacement:"+id+":"+requestID)))}
		if err = replacementInput.Validate(); err != nil {
			return nil, err
		}
	}
	out, err := readJSON(ctx, tx, `UPDATE dbo.pagos SET revertido_en=TODATETIMEOFFSET(SYSUTCDATETIME(),'+00:00'),
rectificacion=JSON_OBJECT('by':@p3,'reason':@p4,'requestId':@p5,'at':SYSUTCDATETIME())
WHERE id_orden=@p1 AND id_pago=@p2;
SELECT `+paymentJSON+` FROM dbo.pagos p WHERE p.id_pago=@p2;`, id, payment, u.ID, reason, requestID)
	if err != nil {
		return nil, err
	}
	if replacement != nil {
		if err = tx.QueryRowContext(ctx, `INSERT dbo.pagos(id_orden,monto,metodo_pago,comprobante,id_usuario_registro,request_id) OUTPUT CONVERT(varchar(20),inserted.id_pago) VALUES(@p1,@p2,@p3,nullif(@p4,N''),@p5,@p6)`, id, replacementInput.Amount, replacementInput.Method, replacementInput.Receipt, u.ID, replacementInput.RequestID).Scan(&replacementID); err != nil {
			return nil, translate(err)
		}
		if _, err = tx.ExecContext(ctx, `UPDATE pagos SET rectificacion=JSON_MODIFY(JSON_MODIFY(JSON_MODIFY(rectificacion,'$.replacementId',@p3),'$.replacementMethod',@p4),'$.replacementReceipt',@p5) WHERE id_orden=@p1 AND id_pago=@p2`, id, payment, replacementID, replacementInput.Method, replacementInput.Receipt); err != nil {
			return nil, err
		}
		out, err = readJSON(ctx, tx, `SELECT `+paymentJSON+` FROM pagos p WHERE id_pago=@p1`, payment)
		if err != nil {
			return nil, err
		}
	}
	return out, translate(tx.Commit())
}

func (s *Store) AttentionHistory(ctx context.Context, client, vehicle string, u domain.User) (any, error) {
	return readJSON(ctx, s.db, `SELECT JSON_OBJECT('attentions':JSON_QUERY((
SELECT JSON_QUERY(`+receptionJSON+`) AS reception,
 JSON_QUERY((SELECT CONVERT(varchar(20),o.id_orden) AS id,o.codigo_ot AS code,o.estado AS state,
  JSON_QUERY((SELECT dc.descripcion AS concept,dc.tipo_concepto AS type,CONVERT(varchar(32),dc.cantidad) AS quantity,
   COALESCE(JSON_VALUE(o.control_tecnico,CONCAT(N'$.execution."',dc.id_detalle,N'".status')),
    CASE WHEN EXISTS(SELECT 1 FROM OPENJSON(o.control_tecnico,N'$.performedItems') x WHERE x.value=CONVERT(varchar(20),dc.id_detalle))
    THEN N'Realizado' ELSE N'Sin confirmar' END) AS status,
   JSON_VALUE(o.control_tecnico,CONCAT(N'$.execution."',dc.id_detalle,N'".notes')) AS notes
   FROM dbo.detalle_cotizacion dc WHERE dc.id_cotizacion=q.id_cotizacion ORDER BY dc.id_detalle FOR JSON PATH, INCLUDE_NULL_VALUES)) AS works,
  JSON_QUERY(o.control_tecnico,N'$.execution') AS execution,
  CASE WHEN @p3=1 THEN CONVERT(varchar(42),dbo.jac_importe_orden(o.id_orden)) END AS total,
  JSON_QUERY(CASE WHEN @p3=1 THEN (SELECT `+paymentColumns+` FROM dbo.pagos p
    WHERE p.id_orden=o.id_orden ORDER BY p.fecha_pago,p.id_pago FOR JSON PATH, INCLUDE_NULL_VALUES) ELSE N'[]' END) AS payments,
  CONVERT(bit,CASE WHEN EXISTS(SELECT 1 FROM dbo.entregas e WHERE e.id_orden=o.id_orden) THEN 1 ELSE 0 END) AS delivered
  FROM dbo.ordenes_trabajo o JOIN dbo.cotizaciones q ON q.id_cotizacion=o.id_cotizacion
  WHERE q.id_recepcion=r.id_recepcion ORDER BY o.id_orden FOR JSON PATH, INCLUDE_NULL_VALUES)) AS orders
FROM dbo.recepciones r WHERE (@p1=N'' OR r.id_cliente_atencion=TRY_CONVERT(bigint,@p1))
 AND (@p2=N'' OR r.id_vehiculo=TRY_CONVERT(bigint,@p2))
ORDER BY r.fecha_ingreso DESC,r.id_recepcion DESC FOR JSON PATH, INCLUDE_NULL_VALUES)))`, client, vehicle, u.Can("pagos.ver"))
}
