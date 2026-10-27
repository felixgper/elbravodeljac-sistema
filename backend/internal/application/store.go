package application

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
)

// Store is the application port. SQL and transaction details stay in the adapter.
type Store interface {
	Ping(context.Context) error
	FindLogin(context.Context, string) (domain.User, error)
	User(context.Context, string) (domain.User, error)
	Clients(context.Context, domain.Page) (domain.List[domain.Client], error)
	Client(context.Context, string) (domain.Client, error)
	CreateClient(context.Context, domain.ClientInput) (domain.Client, error)
	PatchClient(context.Context, string, map[string]string) (domain.Client, error)
	DeleteClient(context.Context, string) error
	Contacts(context.Context, string) ([]domain.Contact, error)
	CreateContact(context.Context, string, domain.ContactInput) (domain.Contact, error)
	Vehicles(context.Context, domain.Page) (domain.List[domain.Vehicle], error)
	Vehicle(context.Context, string) (domain.Vehicle, error)
	CreateVehicle(context.Context, domain.VehicleInput, string) (domain.Vehicle, error)
	PatchVehicle(context.Context, string, domain.VehiclePatch) (domain.Vehicle, error)
	DeleteVehicle(context.Context, string) error
	TransferVehicle(context.Context, string, domain.Transfer, string) (domain.Vehicle, error)
	VehicleHistory(context.Context, string) (any, error)
}
