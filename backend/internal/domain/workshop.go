package domain

import (
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

type ReceptionInput struct {
	TechnicianID  string `json:"technicianId"`
	AttentionType string `json:"attentionType"`
	ClientID      string `json:"clientId"`
	VehicleID     string `json:"vehicleId"`
	Mileage       string `json:"mileage"`
	Symptoms      string `json:"symptoms"`
}

func (v *ReceptionInput) Validate() error {
	if v.TechnicianID != "" {
		if _, err := ID(v.TechnicianID); err != nil {
			return Invalid("technicianId", "Mecánico inválido.")
		}
	}
	if v.AttentionType == "" {
		v.AttentionType = "Diagnostico"
	}
	if v.AttentionType != "Directo" && v.AttentionType != "Diagnostico" {
		return Invalid("attentionType", "Selecciona el tipo de atención.")
	}
	if _, err := ID(v.ClientID); err != nil {
		return Invalid("clientId", "Cliente inválido.")
	}
	if _, err := ID(v.VehicleID); err != nil {
		return Invalid("vehicleId", "Vehículo inválido.")
	}
	n, err := Mileage(v.Mileage)
	if err != nil {
		return err
	}
	v.Mileage = FormatMileage(n)
	return text("symptoms", v.Symptoms, 4000, true)
}

type DiagnosisInput struct {
	Status          string `json:"status"`
	Evaluation      string `json:"evaluation"`
	Recommendations string `json:"recommendations"`
	Notes           string `json:"notes"`
	Version         int64  `json:"version"`
	TechnicianID    string `json:"technicianId"`
	ReceptionID     string `json:"receptionId"`
	Findings        string `json:"findings"`
}

func (v DiagnosisInput) Validate() error {
	if v.TechnicianID != "" {
		if _, err := ID(v.TechnicianID); err != nil {
			return Invalid("technicianId", "Mecánico responsable inválido.")
		}
	}
	if _, err := ID(v.ReceptionID); err != nil {
		return Invalid("receptionId", "Recepción inválida.")
	}
	if v.Status != "" && v.Status != "Borrador" && v.Status != "Finalizado" {
		return Invalid("status", "Estado de evaluación inválido.")
	}
	if v.Version < 0 {
		return Invalid("version", "Versión inválida.")
	}
	for _, value := range []struct{ k, v string }{{"findings", v.Findings}, {"evaluation", v.Evaluation}, {"recommendations", v.Recommendations}, {"notes", v.Notes}} {
		if err := text(value.k, value.v, 4000, v.Status == "Finalizado" && value.k != "notes"); err != nil {
			return err
		}
	}
	return nil
}

// NUMERIC(12,2), transported as a decimal string, never a JSON float.
func Money(value, field string) (int64, error) {
	if !regexp.MustCompile(`^(0|[1-9][0-9]{0,9})(\.[0-9]{1,2})?$`).MatchString(value) {
		return 0, Invalid(field, "Usa un decimal entre 0.00 y 9999999999.99 con hasta dos decimales.")
	}
	p := strings.SplitN(value, ".", 2)
	whole, _ := strconv.ParseInt(p[0], 10, 64)
	fraction := "00"
	if len(p) == 2 {
		fraction = (p[1] + "0")[:2]
	}
	n, _ := strconv.ParseInt(fraction, 10, 64)
	return whole*100 + n, nil
}
func QuoteState(value string) (string, error) {
	switch strings.ToLower(value) {
	case "pendiente", "aprobada", "rechazada":
		return strings.ToLower(value), nil
	}
	return "", Invalid("status", "Estado de cotización inválido.")
}
func OrderState(value string) (string, error) {
	switch strings.ToLower(value) {
	case "pendiente", "creada":
		return "creada", nil
	case "en reparación", "en_reparacion":
		return "en_reparacion", nil
	case "en pruebas", "en_pruebas":
		return "en_pruebas", nil
	case "finalizada":
		return "finalizada", nil
	case "cancelada":
		return "cancelada", nil
	}
	return "", Invalid("state", "Estado de orden inválido.")
}

type QuoteLine struct {
	PartKind    string `json:"partKind"`
	Discount    string `json:"discount"`
	ServiceID   string `json:"serviceId"`
	SparePartID string `json:"sparePartId"`
	Concept     string `json:"concept"`
	Type        string `json:"type"`
	Origin      string `json:"origin"`
	Quantity    string `json:"quantity"`
	UnitPrice   string `json:"unitPrice"`
	Subtotal    string `json:"-"`
}
type QuoteInput struct {
	ReceptionID string      `json:"receptionId"`
	DiagnosisID string      `json:"diagnosisId"`
	Status      string      `json:"status"`
	Items       []QuoteLine `json:"items"`
	Total       string      `json:"-"`
}

func (v *QuoteInput) Validate() error {
	if v.DiagnosisID == "" && v.ReceptionID == "" {
		return Invalid("receptionId", "Se requiere recepción o diagnóstico real.")
	}
	for _, id := range []string{v.DiagnosisID, v.ReceptionID} {
		if id != "" {
			if _, err := ID(id); err != nil {
				return Invalid("receptionId", "Referencia inválida.")
			}
		}
	}
	state, err := QuoteState(v.Status)
	if err != nil {
		return err
	}
	v.Status = state
	if len(v.Items) == 0 || len(v.Items) > 100 {
		return Invalid("items", "Se requieren entre 1 y 100 conceptos.")
	}
	var total int64
	for i := range v.Items {
		line := &v.Items[i]
		manual := line.ServiceID == "" && line.SparePartID == ""
		service := line.ServiceID != "" || (manual && line.Type == "Servicio")
		if line.ServiceID != "" && line.SparePartID != "" {
			return Invalid("items", "No combines servicio y repuesto en una misma referencia.")
		}
		if manual {
			if line.Type != "Servicio" && line.Type != "Repuesto" {
				return Invalid("type", "Selecciona Servicio o Repuesto.")
			}
			if err := text("concept", line.Concept, 4000, true); err != nil {
				return err
			}
		} else {
			ref := line.SparePartID
			if service {
				ref = line.ServiceID
			}
			if _, err := ID(ref); err != nil {
				return Invalid("items", "Referencia de catálogo inválida.")
			}
			if err := text("concept", line.Concept, 4000, false); err != nil {
				return err
			}
		}
		if service {
			if line.Type != "" && line.Type != "Servicio" {
				return Invalid("type", "La referencia corresponde a un servicio.")
			}
			line.Type = "Servicio"
			if line.PartKind != "" {
				return Invalid("partKind", "La mano de obra no tiene tipo de repuesto.")
			}
			if line.Origin != "" && line.Origin != "—" {
				return Invalid("origin", "Un servicio no tiene suministro de repuesto.")
			}
			line.Origin = ""
		} else {
			if line.Type != "" && line.Type != "Repuesto" {
				return Invalid("type", "La referencia corresponde a un repuesto.")
			}
			line.Type = "Repuesto"
			if line.PartKind != "" && line.PartKind != "Original" && line.PartKind != "Alternativo" {
				return Invalid("partKind", "Selecciona Original o Alternativo.")
			}
			if manual && line.PartKind == "" {
				return Invalid("partKind", "El repuesto manual requiere tipo Original o Alternativo.")
			}
			line.Origin = strings.ToUpper(line.Origin)
			if line.Origin != "TALLER" && line.Origin != "CLIENTE" {
				return Invalid("origin", "Indica si el repuesto lo suministra el Taller o el Cliente.")
			}
		}
		quantity, err := Mileage(line.Quantity)
		if err != nil || quantity == 0 {
			return Invalid("quantity", "Cantidad positiva con máximo dos decimales, compatible con NUMERIC(10,2).")
		}
		price, err := Money(line.UnitPrice, "unitPrice")
		if err != nil {
			return err
		}
		if line.Discount == "" {
			line.Discount = "0.00"
		}
		discount, err := Money(line.Discount, "discount")
		if err != nil {
			return err
		}
		if discount > price {
			return Invalid("discount", "El descuento por unidad no puede superar el precio de venta.")
		}
		line.Discount = FormatMileage(discount)
		if line.Origin == "CLIENTE" && price != 0 {
			return Invalid("unitPrice", "El repuesto del cliente debe tener precio cero.")
		}
		// Half-up rounding without overflow: 10-digit quantity * 12-digit price.
		product := new(big.Int).Mul(big.NewInt(quantity), big.NewInt(price-discount))
		product.Add(product, big.NewInt(50))
		product.Div(product, big.NewInt(100))
		if !product.IsInt64() || product.Int64() > 999999999999 {
			return Invalid("subtotal", "Subtotal excede NUMERIC(12,2).")
		}
		total += product.Int64()
		if total > 999999999999 {
			return Invalid("total", "Total excede NUMERIC(12,2).")
		}
		line.Quantity = FormatMileage(quantity)
		line.UnitPrice = FormatMileage(price)
		line.Subtotal = FormatMileage(product.Int64())
	}
	v.Total = FormatMileage(total)
	return nil
}

type CreateOrderInput struct {
	QuoteID     string `json:"quoteId"`
	MechanicID  string `json:"mechanicId"`
	AssistantID string `json:"assistantId"`
}

func (v CreateOrderInput) Validate() error {
	if v.AssistantID != "" && v.AssistantID == v.MechanicID {
		return Invalid("assistantId", "Selecciona personas distintas para mecánico responsable y auxiliar.")
	}
	for field, id := range map[string]string{"quoteId": v.QuoteID, "mechanicId": v.MechanicID} {
		if _, err := ID(id); err != nil {
			return Invalid(field, "ID inválido.")
		}
	}
	if v.AssistantID != "" {
		if _, err := ID(v.AssistantID); err != nil {
			return Invalid("assistantId", "ID inválido.")
		}
	}
	return nil
}

type UpdateOrderInput struct {
	State       string  `json:"state"`
	Note        string  `json:"note"`
	MechanicID  *string `json:"mechanicId"`
	AssistantID *string `json:"assistantId"`
	Test        string  `json:"test"`
	Result      string  `json:"result"`
}

func (v *UpdateOrderInput) Validate() error {
	state, err := OrderState(v.State)
	if err != nil {
		return err
	}
	v.State = state
	if err := text("note", v.Note, 4000, true); err != nil {
		return err
	}
	if err := text("test", v.Test, 100, false); err != nil {
		return err
	}
	if v.MechanicID != nil {
		if _, err := ID(*v.MechanicID); err != nil {
			return Invalid("mechanicId", "ID inválido.")
		}
	}
	if v.AssistantID != nil && *v.AssistantID != "" {
		if _, err := ID(*v.AssistantID); err != nil {
			return Invalid("assistantId", "ID inválido.")
		}
	}
	switch strings.ToLower(v.Result) {
	case "", "no aplica":
		if v.Test != "" {
			return Invalid("result", "Indica el resultado de la prueba.")
		}
		v.Result = ""
	case "satisfactoria", "fallida":
		v.Result = strings.ToLower(v.Result)
		if strings.TrimSpace(v.Test) == "" {
			return Invalid("test", "Indica qué prueba se realizó.")
		}
	default:
		return Invalid("result", "Resultado de prueba inválido.")
	}
	return nil
}

type PaymentInput struct {
	Amount  string `json:"amount"`
	Method  string `json:"method"`
	Receipt string `json:"comprobante"`
}

func (v *PaymentInput) Validate() error {
	n, err := Money(v.Amount, "amount")
	if err != nil {
		return err
	}
	if n == 0 {
		return Invalid("amount", "Monto positivo requerido.")
	}
	v.Amount = FormatMileage(n)
	switch v.Method {
	case "Transferencia", "Efectivo", "Tarjeta":
	default:
		return Invalid("method", "Medio de pago inválido.")
	}
	return text("comprobante", v.Receipt, 50, false)
}

type DeliveryInput struct {
	Kind      string `json:"kind"`
	Method    string `json:"method"`
	Verified  bool   `json:"verified"`
	ContactID string `json:"contactId"`
}

func (v DeliveryInput) Validate() error {
	if !v.Verified {
		return Invalid("verified", "Comprueba identidad y autorización para este retiro.")
	}
	switch v.Method {
	case "Confirmación del titular", "Documento", "Representante facultado":
	default:
		return Invalid("method", "Medio de verificación inválido.")
	}
	switch v.Kind {
	case "Tercero":
		if _, err := ID(v.ContactID); err != nil {
			return Invalid("contactId", "Selecciona contacto vinculado.")
		}
	case "Titular / representante":
		if v.ContactID != "" {
			return Invalid("contactId", "No corresponde contacto para el titular.")
		}
	default:
		return Invalid("kind", "Tipo de receptor inválido.")
	}
	return nil
}

type WorkshopFilter struct {
	Page
	ParentID, State       string
	Pending, WithoutOrder bool
}

type DiagnosisAssignment struct {
	TechnicianID string `json:"technicianId"`
	Version      int64  `json:"version"`
}
