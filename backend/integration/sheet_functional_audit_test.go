package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/technicalsheet"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Explicit opt-in and a dedicated port keep this audit away from development.
// Only account fixtures use INSERT. Every workshop transition uses authenticated HTTP.
func TestSheetFunctionalAudit(t *testing.T) {
	if os.Getenv("SHEET_FUNCTIONAL_AUDIT") != "1" {
		t.Skip("Requiere el cluster exclusivo de auditoría de ficha")
	}
	cfg, err := pgxpool.ParseConfig(os.Getenv("INTEGRATION_DATABASE_URL"))
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_test" || cfg.ConnConfig.Host != "127.0.0.1" || cfg.ConnConfig.Port != 25439 {
		t.Fatal("La auditoría solo admite el PostgreSQL nuevo de QA en 127.0.0.1:25439")
	}
	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	origin := "http://" + listener.Addr().String()
	handler, err := httpapi.New(&postgres.Store{Pool: pool}, httpapi.Options{Origin: origin, Key: []byte(rand.Text() + rand.Text()), FrontendDir: filepath.Join("..", "..", "dist")})
	if err != nil {
		t.Fatal(err)
	}
	api := httptest.NewUnstartedServer(handler)
	api.Listener = listener
	api.Start()
	defer api.Close()
	h := workshopHarness{ctx: ctx, pool: pool, api: api, origin: origin, password: rand.Text(), suffix: rand.Text()[:8]}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(h.password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	seed := func(label, role, job string) (string, workshopSession) {
		t.Helper()
		name := label + " " + h.suffix
		employee := h.scalar(t, `INSERT INTO empleados(nombre,tipo_empleado,activo) VALUES($1,$2,true) RETURNING id_empleado::text`, name, job)
		id := h.scalar(t, `INSERT INTO usuarios(nombre,usuario,password_hash,id_empleado) VALUES($1,$2,$3,$4) RETURNING id_usuario::text`, name, label+"_"+h.suffix, string(passwordHash), employee)
		h.exec(t, `INSERT INTO usuario_roles(id_usuario,id_rol) SELECT $1,id_rol FROM roles WHERE nombre_rol=$2 AND activo`, id, role)
		res := h.send(workshopSession{}, "POST", "/auth/login", map[string]string{"username": label + "_" + h.suffix, "password": h.password})
		if res.err != nil || res.status != 200 {
			t.Fatalf("Login independiente %s: %d", role, res.status)
		}
		var login struct {
			CSRF string `json:"csrfToken"`
		}
		if json.Unmarshal(res.raw, &login) != nil || login.CSRF == "" || len(res.cookies) == 0 || !res.cookies[0].HttpOnly {
			t.Fatal("JWT HttpOnly/CSRF ausentes")
		}
		return id, workshopSession{res.cookies[0], login.CSRF}
	}
	adminID, admin := seed("Administradora QA", "Administrador", "Administración")
	mechanicID, mechanic := seed("Mecánico QA", "Mecánico", "Mecánico")
	_, other := seed("Mecánico ajeno QA", "Mecánico", "Mecánico")
	var checks []map[string]any
	check := func(name string) {
		t.Helper()
		checks = append(checks, map[string]any{"prueba": name, "resultado": "aprobada"})
		t.Log("PASS:", name)
	}
	check("Login de tres cuentas independientes con roles canónicos, empleados, JWT HttpOnly y CSRF")
	clientBody := map[string]string{"name": "Cliente Ficha QA " + h.suffix, "documentType": "Pasaporte", "documentNumber": "FQ" + h.suffix, "phone": "987654321", "email": "ficha@example.invalid", "address": "Dirección ficticia de auditoría"}
	client := responseID(t, h.call(t, admin, "POST", "/clientes", clientBody, 201))
	vehicleBody := map[string]any{"clientId": client, "plate": "FQ" + h.suffix[:5], "brand": "Toyota", "model": "Vehículo QA", "year": 2025, "mileage": "24000.50"}
	vehicle := responseID(t, h.call(t, admin, "POST", "/vehiculos", vehicleBody, 201))
	reception := responseID(t, h.call(t, admin, "POST", "/recepciones", map[string]any{"clientId": client, "vehicleId": vehicle, "mileage": "24000.50", "symptoms": "Ruido y vibración en el sistema de frenos", "attentionType": "Diagnostico"}, 201))
	h.call(t, admin, "PATCH", "/recepciones/"+reception+"/mecanico", map[string]any{"technicianId": mechanicID, "version": 0}, 200)
	diagnosisText := "Pastillas delanteras desgastadas. Discos aptos para servicio."
	diagnosis := responseID(t, h.call(t, mechanic, "POST", "/diagnosticos", map[string]any{"receptionId": reception, "technicalDiagnosis": diagnosisText, "recommendations": "Revisar alineación en próxima visita; recomendación no ejecutada.", "status": "Finalizado"}, 201))
	h.call(t, other, "GET", "/diagnosticos/"+diagnosis, nil, 404)
	if h.scalar(t, `SELECT (r.id_recepcionista=$2::bigint AND d.id_tecnico=$3::bigint)::text FROM recepciones r JOIN diagnosticos d USING(id_recepcion) WHERE d.id_diagnostico=$1`, diagnosis, adminID, mechanicID) != "true" {
		t.Fatal("Actores incorrectos")
	}
	check("Administradora registra cliente/vehículo/recepción, asigna y el mecánico responsable diagnostica por HTTP")
	items := []map[string]string{
		{"type": "Servicio", "concept": "Mantenimiento de frenos efectivamente realizado", "quantity": "1", "unitPrice": "100", "discount": "10"},
		{"type": "Repuesto", "partKind": "Alternativo", "origin": "Taller", "concept": "Pastillas delanteras semicerámicas utilizadas", "quantity": "1", "unitPrice": "170", "discount": "0"},
		{"type": "Repuesto", "partKind": "Alternativo", "origin": "Taller", "concept": "ACEITE_SOLO_COTIZADO_NO_UTILIZADO", "quantity": "1", "unitPrice": "50", "discount": "0"},
	}
	quoteBody := map[string]any{"diagnosisId": diagnosis, "receptionId": reception, "status": "Pendiente", "items": items}
	quote := responseID(t, h.call(t, admin, "POST", "/cotizaciones", quoteBody, 201))
	h.call(t, admin, "POST", "/ordenes", map[string]string{"quoteId": quote, "mechanicId": mechanicID}, 422)
	quoteBody["status"] = "Aprobada"
	quoteBody["approval"] = map[string]string{"name": clientBody["name"], "method": "Presencial"}
	approved := dataObject(t, h.call(t, admin, "PUT", "/cotizaciones/"+quote, quoteBody, 200))
	if approved["total"] != "310.00" {
		t.Fatal("Cálculo aprobado incorrecto", approved["total"])
	}
	approvedItems := approved["items"].([]any)
	ids := []string{}
	for _, item := range approvedItems {
		ids = append(ids, item.(map[string]any)["id"].(string))
	}
	order := responseID(t, h.call(t, admin, "POST", "/ordenes", map[string]string{"quoteId": quote, "mechanicId": mechanicID}, 201))
	path := "/ordenes/" + order
	h.call(t, admin, "GET", path+"/ficha-tecnica.pdf", nil, 404)
	h.call(t, other, "GET", path+"/control-tecnico", nil, 404)
	h.call(t, admin, "POST", path+"/entrega", map[string]any{"kind": "Titular / representante", "method": "Documento", "verified": true}, 422)
	h.call(t, mechanic, "POST", path+"/pagos", map[string]string{"amount": "1", "method": "Efectivo"}, 403)
	noCSRF := admin
	noCSRF.csrf = ""
	h.call(t, noCSRF, "POST", path+"/pagos", map[string]string{"amount": "1", "method": "Efectivo"}, 403)
	check("Cotización pendiente no genera orden; entrega prematura, pagos del mecánico y petición sin CSRF bloqueados")
	h.call(t, mechanic, "PATCH", path, map[string]string{"state": "En reparación", "note": "Desmontaje y revisión de frenos por el mecánico responsable"}, 200)
	control := map[string]any{
		"version": 0, "inspections": map[string]string{"fugas_0": "Bien", "liquidos_0": "Atención", "neumaticos_0": "Urgente", "fugas_1": "Sin evaluar"},
		"maintenance": map[string][]string{"pastillas": {"1"}}, "otherServices": []string{},
		"tires":       map[string]string{"frontLeft": "32", "frontRight": "32", "rearLeft": "30", "rearRight": "30"},
		"clientNotes": "Solicita comprobar el frenado.", "technicalNotes": "Se instalaron pastillas semicerámicas. Neumáticos requieren atención en otra visita.", "nextMileage": "29000", "nextDate": "2027-02-03",
		"execution": map[string]any{ids[0]: map[string]string{"status": "Realizado", "notes": "Limpieza y lubricación comprobadas"}, ids[1]: map[string]string{"status": "Realizado", "notes": "Pastillas instaladas"}, ids[2]: map[string]string{"status": "No realizado", "notes": "Cliente solicita posponer el cambio de aceite"}},
	}
	h.call(t, other, "PUT", path+"/control-tecnico", control, 403)
	h.call(t, mechanic, "PUT", path+"/control-tecnico", control, 200)
	h.call(t, mechanic, "PUT", path+"/control-tecnico", control, 409)
	readControl := func() map[string]any {
		return dataObject(t, h.call(t, mechanic, "GET", path+"/control-tecnico", nil, 200))["record"].(map[string]any)
	}
	read := readControl()
	rawRead, _ := json.Marshal(read["control"])
	var persistedControl technicalsheet.Control
	if err = json.Unmarshal(rawRead, &persistedControl); err != nil {
		t.Fatal(err)
	}
	var dbControl []byte
	if err = pool.QueryRow(ctx, `SELECT control_tecnico FROM ordenes_trabajo WHERE id_orden=$1`, order).Scan(&dbControl); err != nil {
		t.Fatal(err)
	}
	var dbParsed technicalsheet.Control
	if json.Unmarshal(dbControl, &dbParsed) != nil || !reflect.DeepEqual(dbParsed, persistedControl) {
		t.Fatal("DB/API difieren")
	}
	for id, state := range control["inspections"].(map[string]string) {
		if persistedControl.Inspections[id] != state {
			t.Fatal("Estado perdido", id)
		}
	}
	if len(persistedControl.Maintenance["pastillas"]) != 1 || len(persistedControl.Maintenance["aceite"]) != 0 || len(persistedControl.OtherServices) != 0 {
		t.Fatal("Materiales inferidos o perdidos")
	}
	check("Bien/Atención/Urgente/Sin evaluar persisten y recargan; DB/API iguales; aceite solo cotizado sin selección")
	if os.Getenv("SHEET_BROWSER_AUDIT") == "1" {
		browserCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
		defer cancel()
		fixture, _ := json.Marshal(map[string]string{"url": origin, "username": "Mecánico QA_" + h.suffix, "password": h.password, "orderId": order})
		cmd := exec.CommandContext(browserCtx, "node", filepath.Join("..", "..", "checks", "sheet-functional-browser.mjs"))
		cmd.Stdin = bytes.NewReader(fixture)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Casillas en Chrome: %v\n%s", err, output)
		}
		check("Chrome real: guardar control desde formulario, F5 y reabrir conserva estados, mantenimiento y ejecución")
		read = readControl()
	}
	// An ordinary update preserves every selection. An explicit authorised
	// reconciliation then accounts for the quoted item that was not performed.
	control["version"] = read["version"]
	control["technicalNotes"] = "Se instalaron pastillas semicerámicas y se verificó el frenado. Cambio de aceite pospuesto; revisar neumáticos."
	h.call(t, mechanic, "PUT", path+"/control-tecnico", control, 200)
	read = readControl()
	control["version"] = read["version"]
	control["adjustmentReason"] = "Cliente pospone el aceite cotizado; se cobran únicamente servicio y pastillas realizados."
	h.call(t, admin, "PUT", path+"/control-tecnico", control, 200)
	read = readControl()
	rawRead, _ = json.Marshal(read["control"])
	if json.Unmarshal(rawRead, &persistedControl) != nil || persistedControl.Maintenance["pastillas"][0] != "1" || persistedControl.Inspections["neumaticos_0"] != "Urgente" {
		t.Fatal("Selecciones perdidas al actualizar")
	}
	h.call(t, mechanic, "PATCH", path, map[string]string{"state": "En pruebas", "note": "Verificación funcional de frenos", "test": "Prueba de frenado", "result": "Satisfactoria"}, 200)
	h.call(t, mechanic, "PATCH", path, map[string]string{"state": "Finalizada", "note": "Reparación comprobada y finalizada", "test": "Prueba de ruta", "result": "Satisfactoria"}, 200)
	control["version"] = read["version"]
	h.call(t, mechanic, "PUT", path+"/control-tecnico", control, 409)
	h.call(t, admin, "POST", path+"/entrega", map[string]any{"kind": "Titular / representante", "method": "Documento", "verified": true}, 422)
	h.call(t, admin, "POST", path+"/pagos", map[string]string{"amount": "50", "method": "Efectivo"}, 201)
	h.call(t, admin, "POST", path+"/pagos", map[string]string{"amount": "211", "method": "Efectivo"}, 422)
	h.call(t, admin, "POST", path+"/entrega", map[string]any{"kind": "Titular / representante", "method": "Documento", "verified": true}, 422)
	h.call(t, admin, "GET", path+"/ficha-tecnica.pdf", nil, 404)
	h.call(t, admin, "POST", path+"/pagos", map[string]string{"amount": "210", "method": "Efectivo"}, 201)
	h.call(t, admin, "POST", path+"/entrega", map[string]any{"kind": "Titular / representante", "method": "Documento", "verified": true, "conformity": true, "recipientKind": "Propietario o titular"}, 201)
	check("Actualización conserva selección; avances/pruebas/finalización, conciliación autorizada y pago parcial/completo respetan reglas")
	pdf := h.send(admin, "GET", path+"/ficha-tecnica.pdf", nil)
	if pdf.status != 200 || !bytes.HasPrefix(pdf.raw, []byte("%PDF-1.7")) {
		t.Fatal("PDF no descargable", pdf.status)
	}
	h.call(t, mechanic, "GET", path+"/ficha-tecnica.pdf", nil, 403)
	h.call(t, workshopSession{}, "GET", path+"/ficha-tecnica.pdf", nil, 401)
	var snapshot, archived []byte
	if err = pool.QueryRow(ctx, `SELECT ficha_snapshot,ficha_pdf FROM entregas WHERE id_orden=$1`, order).Scan(&snapshot, &archived); err != nil {
		t.Fatal(err)
	}
	var sheet technicalsheet.Snapshot
	if json.Unmarshal(snapshot, &sheet) != nil {
		t.Fatal("Instantánea inválida")
	}
	if sheet.Client != clientBody["name"] || sheet.Document != "Pasaporte "+clientBody["documentNumber"] || sheet.Plate != vehicleBody["plate"] || sheet.Mileage != "24000.50" || sheet.Mechanic != "Mecánico QA "+h.suffix || sheet.Diagnosis != diagnosisText || sheet.Total != "260.00" || sheet.AuthorizedTotal != "310.00" || len(sheet.Lines) != 2 || len(sheet.Advances) < 1 || len(sheet.Tests) < 1 {
		t.Fatalf("Instantánea no corresponde a atención: %+v", sheet)
	}
	if !reflect.DeepEqual(sheet.Control.Maintenance, persistedControl.Maintenance) || !reflect.DeepEqual(sheet.Control.Inspections, persistedControl.Inspections) {
		t.Fatal("Instantánea de casillas difiere")
	}
	for _, line := range sheet.Lines {
		if line.Concept == items[2]["concept"] {
			t.Fatal("Trabajo no realizado en PDF")
		}
	}
	if !bytes.Equal(pdf.raw, archived) {
		t.Fatal("Primera descarga distinta del PDF archivado")
	}
	var liveDB []byte
	if err = pool.QueryRow(ctx, `SELECT jsonb_build_object(
 'client',(SELECT to_jsonb(x) FROM clientes x WHERE id_cliente=$1),
 'vehicle',(SELECT to_jsonb(x) FROM vehiculos x WHERE id_vehiculo=$2),
 'diagnosis',(SELECT to_jsonb(x) FROM diagnosticos x WHERE id_diagnostico=$3),
 'order',(SELECT to_jsonb(x) FROM ordenes_trabajo x WHERE id_orden=$4),
 'details',(SELECT jsonb_agg(to_jsonb(x) ORDER BY id_detalle) FROM detalle_cotizacion x WHERE id_cotizacion=$5),
 'payments',(SELECT jsonb_agg(to_jsonb(x) ORDER BY id_pago) FROM pagos x WHERE id_orden=$4),
 'advances',(SELECT jsonb_agg(to_jsonb(x) ORDER BY id_avance) FROM avances_ot x WHERE id_orden=$4),
 'tests',(SELECT jsonb_agg(to_jsonb(x) ORDER BY id_prueba) FROM pruebas_ot x WHERE id_orden=$4))`, client, vehicle, diagnosis, order, quote).Scan(&liveDB); err != nil {
		t.Fatal(err)
	}
	quoteAPI := dataObject(t, h.call(t, admin, "GET", "/cotizaciones/"+quote, nil, 200))
	receptionAPI := dataObject(t, h.call(t, admin, "GET", "/recepciones/"+reception, nil, 200))
	clientAPI := dataObject(t, h.call(t, admin, "GET", "/clientes/"+client, nil, 200))
	vehicleAPI := dataObject(t, h.call(t, admin, "GET", "/vehiculos/"+vehicle, nil, 200))
	diagnosisAPI := dataObject(t, h.call(t, admin, "GET", "/diagnosticos/"+diagnosis, nil, 200))
	orderAPI := dataObject(t, h.call(t, admin, "GET", path, nil, 200))
	paymentsAPI := h.call(t, admin, "GET", path+"/pagos", nil, 200)
	history := dataObject(t, h.call(t, admin, "GET", "/vehiculos/"+vehicle+"/historial", nil, 200))
	check("Entrega archiva PDF/instantánea; cliente/vehículo/km/mecánico/diagnóstico/líneas/importes/controles corresponden a DB y API")
	clientBody["name"] = "Cliente actualizado después de la entrega QA " + h.suffix
	h.call(t, admin, "PATCH", "/clientes/"+client, clientBody, 200)
	vehicleBody["model"] = "Modelo actualizado después de entrega QA"
	h.call(t, admin, "PATCH", "/vehiculos/"+vehicle, map[string]string{"model": "Modelo actualizado después de entrega QA"}, 200)
	again := h.send(admin, "GET", path+"/ficha-tecnica.pdf", nil)
	if again.status != 200 || !bytes.Equal(pdf.raw, again.raw) {
		t.Fatal("Documento histórico cambió al editar otros registros")
	}
	var frozenAfter, archivedAfter []byte
	if err = pool.QueryRow(ctx, `SELECT ficha_snapshot,ficha_pdf FROM entregas WHERE id_orden=$1`, order).Scan(&frozenAfter, &archivedAfter); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(snapshot, frozenAfter) || !bytes.Equal(archived, archivedAfter) {
		t.Fatal("Descarga alteró archivo histórico")
	}
	check("Descarga repetida tras editar cliente/vehículo de prueba conserva bytes, instantánea y archivo histórico; permisos bloquean terceros")
	dir := filepath.Join("..", ".cache", "auditoria-ficha-funcional")
	if err = os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	evidence := map[string]any{"checks": checks, "ids": map[string]string{"client": client, "vehicle": vehicle, "reception": reception, "diagnosis": diagnosis, "quote": quote, "order": order}, "databaseSnapshot": sheet, "databaseLive": json.RawMessage(liveDB), "api": map[string]any{"client": clientAPI, "vehicle": vehicleAPI, "quote": quoteAPI, "reception": receptionAPI, "diagnosis": diagnosisAPI, "order": orderAPI, "payments": paymentsAPI, "control": read, "history": history}, "pdfSHA256": fmt.Sprintf("%x", sha256.Sum256(pdf.raw)), "historicalBytesUnchanged": true}
	jsonEvidence, _ := json.MarshalIndent(evidence, "", "  ")
	for name, content := range map[string][]byte{"evidencia-db-api.json": jsonEvidence, "ficha-atencion-real-qa.pdf": pdf.raw, "instantanea-db.json": snapshot} {
		if err = os.WriteFile(filepath.Join(dir, name), content, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
