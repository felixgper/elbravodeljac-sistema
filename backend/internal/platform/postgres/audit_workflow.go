package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/technicalsheet"
	"github.com/jackc/pgx/v5"
)

func validateExecution(ctx context.Context, tx pgx.Tx, id string) error {
	var version int
	var raw []byte
	if err := tx.QueryRow(ctx, `SELECT flujo_version,control_tecnico FROM ordenes_trabajo WHERE id_orden=$1`, id).Scan(&version, &raw); err != nil {
		return err
	}
	if version == 0 {
		return nil
	} // Historical orders retain their existing workflow.
	var control technicalsheet.Control
	if err := json.Unmarshal(raw, &control); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `SELECT d.id_detalle::text FROM detalle_cotizacion d JOIN ordenes_trabajo o USING(id_cotizacion) WHERE o.id_orden=$1`, id)
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
	if err = tx.QueryRow(ctx, `SELECT jac_importe_orden($1)::text,coalesce((SELECT sum(monto) FROM pagos WHERE id_orden=$1 AND revertido_en IS NULL),0)::text`, id).Scan(&total, &paid); err != nil {
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

	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if _, err = lockOrder(ctx, tx, id); err != nil {
		return nil, err
	}
	var raw []byte
	var reversed bool
	if err = tx.QueryRow(ctx, `SELECT `+paymentJSON+`,p.revertido_en IS NOT NULL FROM pagos p WHERE id_orden=$1 AND id_pago=$2 FOR UPDATE`, id, payment).Scan(&raw, &reversed); err != nil {
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
	out, err := readJSON(ctx, tx, `WITH corrected AS(UPDATE pagos SET revertido_en=clock_timestamp(),rectificacion=jsonb_build_object('by',$3::text,'reason',$4::text,'requestId',$5::text,'at',clock_timestamp()) WHERE id_orden=$1 AND id_pago=$2 RETURNING *) SELECT `+paymentJSON+` FROM corrected p`, id, payment, u.ID, reason, requestID)
	if err != nil {
		return nil, err
	}
	if replacement != nil {
		if err = tx.QueryRow(ctx, `INSERT INTO pagos(id_orden,monto,metodo_pago,comprobante,id_usuario_registro,request_id) VALUES($1,$2,$3,nullif($4,''),$5,$6) RETURNING id_pago::text`, id, replacementInput.Amount, replacementInput.Method, replacementInput.Receipt, u.ID, replacementInput.RequestID).Scan(&replacementID); err != nil {
			return nil, translate(err)
		}
		if _, err = tx.Exec(ctx, `UPDATE pagos SET rectificacion=rectificacion||jsonb_build_object('replacementId',$3::text,'replacementMethod',$4::text,'replacementReceipt',$5::text) WHERE id_orden=$1 AND id_pago=$2`, id, payment, replacementID, replacementInput.Method, replacementInput.Receipt); err != nil {
			return nil, err
		}
		out, err = readJSON(ctx, tx, `SELECT `+paymentJSON+` FROM pagos p WHERE id_pago=$1`, payment)
		if err != nil {
			return nil, err
		}
	}
	return out, translate(tx.Commit(ctx))
}

func (s *Store) AttentionHistory(ctx context.Context, client, vehicle string, u domain.User) (any, error) {
	return readJSON(ctx, s.Pool, `SELECT jsonb_build_object('attentions',coalesce(jsonb_agg(jsonb_build_object('reception',`+receptionJSON+`,'orders',coalesce((SELECT jsonb_agg(jsonb_build_object('id',o.id_orden::text,'code',o.codigo_ot,'state',o.estado,'works',coalesce((SELECT jsonb_agg(jsonb_build_object('concept',dc.descripcion,'type',dc.tipo_concepto,'quantity',dc.cantidad::text,'status',coalesce(o.control_tecnico->'execution'->dc.id_detalle::text->>'status',CASE WHEN coalesce(o.control_tecnico->'performedItems','[]'::jsonb) ? dc.id_detalle::text THEN 'Realizado' ELSE 'Sin confirmar' END),'notes',o.control_tecnico->'execution'->dc.id_detalle::text->>'notes') ORDER BY dc.id_detalle) FROM detalle_cotizacion dc WHERE dc.id_cotizacion=q.id_cotizacion),'[]'::jsonb),'execution',o.control_tecnico->'execution','total',CASE WHEN $3 THEN jac_importe_orden(o.id_orden)::text END,'payments',CASE WHEN $3 THEN coalesce((SELECT jsonb_agg(`+paymentJSON+` ORDER BY p.fecha_pago) FROM pagos p WHERE p.id_orden=o.id_orden),'[]'::jsonb) ELSE '[]'::jsonb END,'delivered',EXISTS(SELECT 1 FROM entregas e WHERE e.id_orden=o.id_orden)) ORDER BY o.id_orden) FROM ordenes_trabajo o JOIN cotizaciones q USING(id_cotizacion) WHERE q.id_recepcion=r.id_recepcion),'[]'::jsonb)) ORDER BY r.fecha_ingreso DESC),'[]'::jsonb)) FROM recepciones r WHERE ($1='' OR r.id_cliente_atencion::text=$1) AND ($2='' OR r.id_vehiculo::text=$2)`, client, vehicle, u.Can("pagos.ver"))
}
