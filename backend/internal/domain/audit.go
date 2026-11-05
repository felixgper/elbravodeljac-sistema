package domain

type QuoteApproval struct {
	Name   string `json:"name"`
	Method string `json:"method"`
}

func (v QuoteApproval) Validate() error {
	if err := text("approval.name", v.Name, 150, true); err != nil {
		return err
	}
	switch v.Method {
	case "Presencial", "WhatsApp", "Llamada", "Otro":
	default:
		return Invalid("approval.method", "Indica el medio de aprobación.")
	}
	return nil
}

type PaymentCorrection struct {
	RequestID string `json:"requestId"`
	Reason    string `json:"reason"`
}

func (v PaymentCorrection) Validate() error {
	if err := text("requestId", v.RequestID, 100, true); err != nil {
		return err
	}
	return text("reason", v.Reason, 1200, true)
}

type PaymentReplacement struct {
	Method  string `json:"method"`
	Receipt string `json:"comprobante"`
}
