package sqlserver

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func TestSQLServer2025(t *testing.T) {
	dsn := os.Getenv("SQLSERVER_QA_DSN")
	if dsn == "" {
		t.Skip("requires isolated SQL Server 2025 QA")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	s, err := Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var db string
	var version, edition int
	err = s.db.QueryRowContext(ctx, `SELECT DB_NAME(),CONVERT(int,SERVERPROPERTY('ProductMajorVersion')),CONVERT(int,SERVERPROPERTY('EditionID'))`).Scan(&db, &version, &edition)
	if err != nil {
		t.Fatal(err)
	}
	if db != "elbravodeljac_sqlserver_qa" || version != 17 || edition != -1785266663 {
		t.Fatal("expected isolated SQL Server 2025 Standard Developer database")
	}
	const password = "Synthetic-QA-only-Pass9!"
	if err := s.Provision(ctx, db, "gerente_qa", password); err != nil {
		t.Fatal(err)
	}
	if err := s.Provision(ctx, db, "gerente_otro", password); err == nil {
		t.Fatal("provisioning accepted an occupied database")
	}
	user, err := s.FindLogin(ctx, "gerente_qa")
	if err != nil {
		t.Fatal(err)
	}
	if !user.Active || !user.Can("usuarios.roles") || len(user.Roles) != 1 || user.Roles[0] != "Gerente" {
		t.Fatalf("unexpected identity: %v %v", user.Roles, user.Permissions)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.User(ctx, user.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DocumentTypes(ctx); err != nil {
		t.Fatal(err)
	}
	brands, err := s.Brands(ctx)
	if err != nil || len(brands) != 34 {
		t.Fatalf("brands=%d err=%v", len(brands), err)
	}

	a, err := s.CreateClient(ctx, domain.ClientInput{Name: "Lucía QA 🚘", DocumentType: "DNI", DocumentNumber: "12345678", Phone: "987654321", Address: "Av. QA", Email: "qa@example.invalid"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.CreateClient(ctx, domain.ClientInput{Name: "Contacto QA", DocumentType: "Carné de Extranjería", DocumentNumber: "X10000", Phone: "987654321", Address: "Av. QA"})
	if err != nil || b.DocumentType != "CE" {
		t.Fatalf("client=%+v err=%v", b, err)
	}
	if _, err := s.CreateClient(ctx, a.ClientInput); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected duplicate conflict, got %v", err)
	}
	if _, err := s.PatchClient(ctx, a.ID, map[string]string{"address": "Otra dirección"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateContact(ctx, a.ID, domain.ContactInput{Name: "Contacto QA", DocumentType: "DNI", DocumentNumber: "87654321", Phone: "987654321", Relation: "Contacto"}); err != nil {
		t.Fatal(err)
	}
	if contacts, err := s.Contacts(ctx, a.ID); err != nil || len(contacts) != 1 {
		t.Fatalf("contacts=%v err=%v", contacts, err)
	}

	const plate = "QA-1234"
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, p := range []string{plate, "QA1234"} {
		wg.Go(func() {
			<-start
			_, err := s.CreateVehicle(ctx, domain.VehicleInput{ClientID: a.ID, Plate: p, Brand: "Citroen", Model: "QA", Year: 2025, Mileage: "10.50"}, user.ID)
			results <- err
		})
	}
	close(start)
	wg.Wait()
	close(results)
	created, conflicts := 0, 0
	for err := range results {
		switch {
		case err == nil:
			created++
		case errors.Is(err, domain.ErrConflict):
			conflicts++
		default:
			t.Fatal(err)
		}
	}
	if created != 1 || conflicts != 1 {
		t.Fatalf("created=%d conflicts=%d", created, conflicts)
	}
	vehicles, err := s.Vehicles(ctx, domain.Page{Number: 1, Size: 10, ClientID: a.ID})
	if err != nil || len(vehicles.Data) != 1 {
		t.Fatalf("vehicles=%v err=%v", vehicles, err)
	}
	v := vehicles.Data[0]
	if v.Brand != "Citroën" || v.Mileage != "10.50" {
		t.Fatalf("vehicle=%+v", v)
	}
	list, err := s.Clients(ctx, domain.Page{Number: 1, Size: 1, Query: "Lucia"})
	if err != nil || list.Meta.Total != 1 || len(list.Data) != 1 || len(list.Data[0].Vehicles) != 1 {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	if _, err := s.TransferVehicle(ctx, v.ID, domain.Transfer{NewClientID: b.ID, Verified: true, Observation: "Transferencia QA"}, user.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.VehicleHistory(ctx, v.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteVehicle(ctx, v.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ownership history conflict: %v", err)
	}
	exerciseWorkshopHTTP(t, s, password, b.ID, v.ID)
}
