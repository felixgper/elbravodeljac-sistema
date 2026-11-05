package application

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
)

type Workshop interface {
	WorkshopList(context.Context, string, domain.WorkshopFilter, domain.User) (any, error)
	WorkshopItem(context.Context, string, string, domain.User) (any, error)
	CreateReception(context.Context, domain.ReceptionInput, domain.User) (any, error)
	CreateDiagnosis(context.Context, domain.DiagnosisInput, domain.User) (any, error)
	SaveQuote(context.Context, string, domain.QuoteInput, domain.User) (any, error)
	CreateOrder(context.Context, domain.CreateOrderInput, domain.User) (any, error)
	UpdateOrder(context.Context, string, domain.UpdateOrderInput, domain.User) (any, error)
	DeleteOrder(context.Context, string, domain.User) error
	RegisterPayment(context.Context, string, domain.PaymentInput, domain.User) (any, error)
	Deliver(context.Context, string, domain.DeliveryInput, domain.User) (any, error)
}
