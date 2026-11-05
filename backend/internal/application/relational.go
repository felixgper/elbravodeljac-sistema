package application

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"time"
)

type Relational interface {
	RelationalList(context.Context, string, string, domain.Page) (any, error)
	RelationalItem(context.Context, string, string) (any, error)
	SaveUser(context.Context, string, domain.UserInput, domain.User) (any, error)
	DeleteUser(context.Context, string, domain.User) error
	SetRolePermissions(context.Context, string, domain.MatrixInput, bool, domain.User) (any, error)
	SaveProvider(context.Context, string, domain.ProviderInput) (any, error)
	DeleteProvider(context.Context, string) error
	SaveCatalog(context.Context, string, domain.CatalogInput) (any, error)
	SaveAuxilio(context.Context, string, domain.AuxilioInput, domain.User) (any, error)
	LinkProviderPart(context.Context, string, string, domain.PartLinkInput, bool) (any, error)
	Statistics(context.Context, bool, time.Time, time.Time) (any, error)
}
