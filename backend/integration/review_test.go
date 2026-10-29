package integration

import (
	"context"
	"crypto/rand"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
)

func TestFunctionalReviewPostgreSQL(t *testing.T) {
	dsn := os.Getenv("INTEGRATION_DATABASE_URL")
	if dsn == "" {
		t.Skip("Requiere PostgreSQL aislado")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_test" {
		t.Fatal("Solo base de prueba")
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
	all := []string{}
	for m, actions := range domain.PermissionPolicy().Modules {
		for _, a := range actions {
			all = append(all, m+"."+a)
		}
	}
	_, admin := h.user(t, "review", "Review_"+h.suffix, all)
	mechanic, _ := h.user(t, "reviewmech", "Mecánico", []string{"ordenes.ver", "ordenes.editar"})
	assistant, _ := h.user(t, "reviewaux", "Auxiliar", []string{"auxilio.ver"})
	h.exec(t, `INSERT INTO clientes(nombre,tipo_documento,numero_documento,telefono,direccion) SELECT 'Félix González '||$1||' '||g,'Pasaporte','R'||$1||g::text,'+51987654321','Prueba aislada' FROM generate_series(1,1500) g`, h.suffix)
	client := h.scalar(t, `SELECT id_cliente::text FROM clientes WHERE numero_documento=$1`, "R"+h.suffix+"1")
	t.Run("búsqueda paginada normalizada e índice", func(t *testing.T) {
		r := h.call(t, admin, "GET", "/clientes?q=F&pageSize=10", nil, 200)
		if len(r["data"].([]any)) != 10 {
			t.Fatal("unbounded client page")
		}
		r = h.call(t, admin, "GET", "/clientes?q="+url.QueryEscape("FELIX GONZALEZ "+h.suffix)+"&pageSize=20", nil, 200)
		if r["meta"].(map[string]any)["total"] != float64(1500) {
			t.Fatal("accent/case search", r["meta"])
		}
		if n := h.scalar(t, `SELECT count(*)::text FROM pg_indexes WHERE indexname='ix_clientes_busqueda'`); n != "1" {
			t.Fatal("index missing")
		}
	})
	h.exec(t, `INSERT INTO vehiculos(id_cliente,placa,marca,modelo,anio,kilometraje_actual) SELECT $1,'R'||$2||'-'||g::text,CASE WHEN g%2=0 THEN 'Ford' ELSE 'Toyota' END,'Prueba',2024,10 FROM generate_series(1,105)g`, client, h.suffix[:4])
	vehicle := h.scalar(t, `SELECT id_vehiculo::text FROM vehiculos WHERE id_cliente=$1 ORDER BY id_vehiculo LIMIT 1`, client)
	t.Run("todos los vehículos del propietario y marcas", func(t *testing.T) {
		r := h.call(t, admin, "GET", "/vehiculos?clientId="+client+"&pageSize=100", nil, 200)
		if len(r["data"].([]any)) != 100 || r["meta"].(map[string]any)["total"] != float64(105) {
			t.Fatal(r["meta"])
		}
		r = h.call(t, admin, "GET", "/vehiculos?clientId="+client+"&pageSize=100&page=2", nil, 200)
		if len(r["data"].([]any)) != 5 {
			t.Fatal("second vehicle page")
		}
		h.call(t, admin, "GET", "/vehiculos/marcas", nil, 200)
		h.call(t, admin, "POST", "/recepciones", map[string]string{"clientId": "9223372036854775807", "vehicleId": vehicle, "mileage": "10", "symptoms": "Inválido"}, 422)
	})
	reception := responseID(t, h.call(t, admin, "POST", "/recepciones", map[string]string{"clientId": client, "vehicleId": vehicle, "mileage": "10", "symptoms": "Revisión"}, 201))
	h.call(t, admin, "PATCH", "/recepciones/"+reception+"/mecanico", map[string]any{"technicianId": mechanic, "version": 0}, 200)
	diagnosis := responseID(t, h.call(t, admin, "POST", "/diagnosticos", map[string]string{"status": "Finalizado", "evaluation": "Evaluación técnica de prueba", "recommendations": "Servicios recomendados de prueba", "receptionId": reception, "findings": "Revisión manual"}, 201))
	q := h.call(t, admin, "POST", "/cotizaciones", map[string]any{"diagnosisId": diagnosis, "status": "Aprobada", "items": []map[string]string{
		{"concept": "Mano de obra negociada", "type": "Servicio", "quantity": "1.25", "unitPrice": "100.00", "discount": "10.00"},
		{"concept": "Repuesto fuera de catálogo", "type": "Repuesto", "partKind": "Alternativo", "origin": "Taller", "quantity": "2", "unitPrice": "50.00", "discount": "5.00"},
	}}, 201)
	quoteID := responseID(t, q)
	if dataObject(t, q)["total"] != "202.50" {
		t.Fatal(q)
	}
	t.Run("manuales descuentos persistidos y migración repetible", func(t *testing.T) {
		if err = store.Migrate(ctx); err != nil {
			t.Fatal(err)
		}
		r := dataObject(t, h.call(t, admin, "GET", "/cotizaciones/"+quoteID, nil, 200))
		if r["total"] != "202.50" {
			t.Fatal(r)
		}
		items := r["items"].([]any)
		if items[1].(map[string]any)["partKind"] != "Alternativo" || items[0].(map[string]any)["discount"] != "10.00" {
			t.Fatal(items)
		}
		if n := h.scalar(t, `SELECT count(*)::text FROM detalle_cotizacion WHERE id_cotizacion=$1 AND id_servicio IS NULL AND id_repuesto IS NULL`, quoteID); n != "2" {
			t.Fatal(n)
		}
	})
	t.Run("snapshot independiente del precio del catálogo", func(t *testing.T) {
		service := responseID(t, h.call(t, admin, "POST", "/servicios", map[string]string{"name": "Precio snapshot " + h.suffix, "referencePrice": "100.00"}, 201))
		rx := responseID(t, h.call(t, admin, "POST", "/recepciones", map[string]string{"clientId": client, "vehicleId": vehicle, "mileage": "10", "symptoms": "Snapshot"}, 201))
		h.call(t, admin, "PATCH", "/recepciones/"+rx+"/mecanico", map[string]any{"technicianId": mechanic, "version": 0}, 200)
		dx := responseID(t, h.call(t, admin, "POST", "/diagnosticos", map[string]string{"status": "Finalizado", "evaluation": "Evaluación técnica de prueba", "recommendations": "Servicios recomendados de prueba", "receptionId": rx, "findings": "Snapshot"}, 201))
		qx := responseID(t, h.call(t, admin, "POST", "/cotizaciones", map[string]any{"diagnosisId": dx, "status": "Pendiente", "items": []map[string]string{{"serviceId": service, "quantity": "2", "unitPrice": "70.00", "discount": "10.00"}}}, 201))
		h.exec(t, `UPDATE servicios SET precio_referencia=999 WHERE id_servicio=$1`, service)
		saved := dataObject(t, h.call(t, admin, "GET", "/cotizaciones/"+qx, nil, 200))
		if saved["total"] != "120.00" || saved["items"].([]any)[0].(map[string]any)["unitPrice"] != "70.00" {
			t.Fatal("catalog changed snapshot", saved)
		}
		_, err := pool.Exec(ctx, `INSERT INTO detalle_cotizacion(id_cotizacion,descripcion,cantidad,precio_unitario,subtotal) VALUES($1,'Sin tipo ni referencia',1,1,1)`, qx)
		if err == nil {
			t.Fatal("SQL accepted anonymous untyped detail")
		}
	})
	order := responseID(t, h.call(t, admin, "POST", "/ordenes", map[string]string{"quoteId": quoteID, "mechanicId": mechanic, "assistantId": assistant}, 201))
	t.Run("avance sin prueba y asignación sin duplicados", func(t *testing.T) {
		h.call(t, admin, "PATCH", "/ordenes/"+order, map[string]string{"state": "En reparación", "note": "Avance sin prueba", "mechanicId": mechanic, "assistantId": assistant}, 200)
		if n := h.scalar(t, `SELECT count(*)::text FROM asignaciones_ot WHERE id_orden=$1`, order); n != "2" {
			t.Fatal("redundant assignments", n)
		}
		h.call(t, admin, "PATCH", "/ordenes/"+order, map[string]string{"state": "Finalizada", "note": "Falta prueba"}, 422)
		h.call(t, admin, "PATCH", "/ordenes/"+order, map[string]string{"state": "En pruebas", "note": "Prueba guardada", "test": "Ruta", "result": "Satisfactoria"}, 200)
		h.confirmExecution(t, admin, order)
		h.call(t, admin, "PATCH", "/ordenes/"+order, map[string]string{"state": "Finalizada", "note": "Usa prueba vigente"}, 200)
		h.call(t, admin, "DELETE", "/ordenes/"+order, nil, 409)
	})
	t.Run("períodos filtran ingresos sin multiplicar detalles", func(t *testing.T) {
		h.call(t, admin, "POST", "/ordenes/"+order+"/pagos", map[string]string{"amount": "202.50", "method": "Efectivo"}, 201)
		for _, period := range []string{"Semana", "Mes", "Año"} {
			r := dataObject(t, h.call(t, admin, "GET", "/dashboard?period="+url.QueryEscape(period), nil, 200))
			a, b := r["from"].(string), r["until"].(string)
			expected := h.scalar(t, `SELECT coalesce(sum(monto),0)::numeric(18,2)::text FROM jac_movimientos_pagos WHERE fecha >= $1::timestamptz AND fecha < $2::timestamptz`, a, b)
			if r["monthlyIncome"] != expected {
				t.Fatal(period, r["monthlyIncome"], expected)
			}
			statesExpected := h.scalar(t, `SELECT count(*)::text FROM ordenes_trabajo WHERE estado IN ('creada','en_reparacion','en_pruebas') OR (fecha_fin >= $1::timestamptz AND fecha_fin < $2::timestamptz)`, a, b)
			total := 0
			for _, v := range r["states"].([]any) {
				total += int(v.(map[string]any)["count"].(float64))
			}
			if strconv.Itoa(total) != statesExpected {
				t.Fatal("states outside range", period, total, statesExpected)
			}
		}
	})
	if n := h.scalar(t, `SELECT count(*)::text FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'`); n != "27" {
		t.Fatal(n)
	}
	if n := h.scalar(t, `SELECT count(*)::text FROM pg_constraint WHERE contype='f' AND connamespace='public'::regnamespace`); n != "48" {
		t.Fatal(n)
	}
}
