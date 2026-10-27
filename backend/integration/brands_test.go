package integration

import (
	"context"
	"crypto/rand"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestVehicleBrandsPostgreSQL(t *testing.T) {
	dsn := os.Getenv("INTEGRATION_DATABASE_URL")
	if dsn == "" {
		t.Skip("Requiere PostgreSQL aislado")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_test" {
		t.Fatal("Solo PostgreSQL de pruebas")
	}
	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	store := &postgres.Store{Pool: pool}
	handler, err := httpapi.New(store, httpapi.Options{Origin: "http://localhost:8080", Key: []byte(rand.Text() + rand.Text())})
	if err != nil {
		t.Fatal(err)
	}
	api := httptest.NewServer(handler)
	defer api.Close()
	h := workshopHarness{pool: pool, api: api, ctx: ctx, password: rand.Text(), suffix: rand.Text()[:8]}
	_, admin := h.user(t, "brands", "Brands_"+h.suffix, []string{"clientes.ver", "clientes.crear", "vehiculos.ver", "vehiculos.crear", "vehiculos.editar"})
	_, limited := h.user(t, "brandslimited", "BrandsLimited_"+h.suffix, []string{"ordenes.ver"})
	h.call(t, limited, "GET", "/vehiculos/marcas", nil, 403)
	fingerprint := `SELECT md5(coalesce(string_agg(row_to_json(v)::text,',' ORDER BY id_vehiculo),'')) FROM vehiculos v`
	before := h.scalar(t, fingerprint)
	if err = store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if after := h.scalar(t, fingerprint); after != before {
		t.Fatal("Migration modified vehicles")
	}
	list := h.call(t, admin, "GET", "/vehiculos/marcas", nil, 200)["data"].([]any)
	actual := map[string]bool{}
	for _, brand := range list {
		key := strings.ToLower(brand.(string))
		if actual[key] {
			t.Fatal("Duplicated brand", key)
		}
		actual[key] = true
	}
	expected := []string{"JAC", "Toyota", "Nissan", "Hyundai", "Kia", "Chevrolet", "Ford", "Volkswagen", "Suzuki", "Mitsubishi", "Mazda", "Renault", "Chery", "Changan", "Haval", "BYD", "MG", "Great Wall", "BMW", "Mercedes-Benz", "Audi", "Volvo", "Peugeot", "Citroën", "BAIC", "Jinbei", "DFM", "Brilliance", "Changhe", "Mahindra", "SsangYong", "Daewoo", "Land Rover", "Honda"}
	for _, name := range expected {
		if !actual[strings.ToLower(name)] {
			t.Fatal("Missing approved brand", name)
		}
	}
	client := responseID(t, h.call(t, admin, "POST", "/clientes", map[string]string{"name": "Marcas prueba " + h.suffix, "documentType": "Pasaporte", "documentNumber": "B" + h.suffix, "phone": "+51987654321", "address": "Prueba aislada"}, 201))
	for i, brand := range []string{"Toyota", "Hyundai", "BMW", "Honda", "Otra " + h.suffix} {
		body := map[string]any{"clientId": client, "plate": "B" + h.suffix[:4] + string(rune('A'+i)), "brand": brand, "model": "Prueba", "year": 2025, "mileage": "0.00"}
		id := responseID(t, h.call(t, admin, "POST", "/vehiculos", body, 201))
		if dataObject(t, h.call(t, admin, "GET", "/vehiculos/"+id, nil, 200))["brand"] != brand {
			t.Fatal("Brand not persisted", brand)
		}
		got := dataObject(t, h.call(t, admin, "PATCH", "/vehiculos/"+id, map[string]string{"brand": "  " + strings.ToLower(brand) + "  "}, 200))
		if got["brand"] != brand {
			t.Fatal("Brand not canonicalized", got)
		}
		for _, bad := range []string{"", "Otra marca", "__other__", strings.Repeat("x", 41), "<script>", "Marca\nFalsa"} {
			h.call(t, admin, "PATCH", "/vehiculos/"+id, map[string]string{"brand": bad}, 422)
		}
	}
	if n := h.scalar(t, `SELECT count(*)::text FROM marcas_vehiculos WHERE jac_normalize(nombre)=jac_normalize($1)`, "Otra "+h.suffix); n != "1" {
		t.Fatal("Custom brand not in unique catalogue", n)
	}
	if err = store.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if n := h.scalar(t, `SELECT count(*)::text FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'`); n != "26" {
		t.Fatal(n)
	}
	if n := h.scalar(t, `SELECT count(*)::text FROM pg_constraint WHERE connamespace='public'::regnamespace AND contype='f'`); n != "44" {
		t.Fatal(n)
	}
}
