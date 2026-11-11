package datamigration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/sqlserver"
	"github.com/jackc/pgx/v5"
	mssql "github.com/microsoft/go-mssqldb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func TestMigrationQA(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "true" || os.Getenv("MIGRATION_RUN_TOKEN") == "" {
		t.Skip("requires temporary GitHub Actions databases")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	c := Config{Source: os.Getenv("MIGRATION_POSTGRES_DSN"), Target: os.Getenv("MIGRATION_SQLSERVER_DSN"), Mongo: os.Getenv("MIGRATION_MONGODB_URI"), Marker: os.Getenv("MIGRATION_RUN_TOKEN")}
	if err := c.validate(); err != nil {
		t.Fatal(err)
	}
	pg, err := pgx.Connect(ctx, c.Source)
	if err != nil {
		t.Fatal("QA PostgreSQL inaccessible")
	}
	defer pg.Close(ctx)
	var marker string
	if err := pg.QueryRow(ctx, "SELECT shobj_description(oid,'pg_database') FROM pg_database WHERE datname=current_database()").Scan(&marker); err != nil || marker != c.Marker {
		t.Fatal("QA source not marked")
	}
	connector, err := mssql.NewConnector(c.Target)
	if err != nil {
		t.Fatal("QA destination connection invalid")
	}
	connector.SessionInitSQL = session
	dst := sql.OpenDB(connector)
	defer dst.Close()
	mg, err := mongo.Connect(options.Client().ApplyURI(c.Mongo))
	if err != nil {
		t.Fatal("QA MongoDB inaccessible")
	}
	defer mg.Disconnect(ctx)
	md := mg.Database(MongoName)
	if md.Collection("migration_control").FindOne(ctx, bson.M{"_id": c.Marker}).Err() != nil {
		t.Fatal("QA MongoDB not marked")
	}

	for _, coll := range []string{"resultados_escaner", "evidencias_tecnicas"} {
		if err := md.CreateCollection(ctx, coll); err != nil {
			t.Fatal(err)
		}
	}
	const diagnostic = "9007199254740993"
	payload := bytes.Repeat([]byte("Synthetic GridFS.\n"), 20000)
	checksum := sha256.Sum256(payload)
	evidenceID, fileID := bson.NewObjectID(), bson.NewObjectID()
	bucket := md.GridFSBucket(options.GridFSBucket().SetName("archivos_tecnicos"))
	metadata := bson.M{"evidenceId": evidenceID, "diagnosticoId": diagnostic, "managedBy": "elbravodeljac"}
	if err := bucket.UploadFromStreamWithID(ctx, fileID, "qa", bytes.NewReader(payload), options.GridFSUpload().SetMetadata(metadata)); err != nil {
		t.Fatal(err)
	}
	if _, err := md.Collection("evidencias_tecnicas").InsertOne(ctx, bson.M{"_id": evidenceID, "diagnosticoId": diagnostic, "capturadoPor": "101", "gridfsFileId": fileID.Hex(), "state": "ready", "sha256": hex.EncodeToString(checksum[:]), "sizeBytes": len(payload)}); err != nil {
		t.Fatal(err)
	}
	scans := md.Collection("resultados_escaner")
	if _, err := scans.InsertOne(ctx, bson.M{"diagnosticoId": diagnostic, "capturadoPor": "101"}); err != nil {
		t.Fatal(err)
	}
	orphan, err := scans.InsertOne(ctx, bson.M{"diagnosticoId": "9999", "capturadoPor": "101"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Run(ctx, c); err == nil || !strings.Contains(err.Error(), "referencia técnica huérfana") {
		t.Fatalf("expected rejection of orphan: %v", err)
	}
	assertEmpty := func() {
		t.Helper()
		tx, err := dst.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		if err := verifyIntegrity(ctx, tx); err != nil {
			t.Fatal(err)
		}
		for _, table := range tables {
			var n int
			if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM dbo."+table).Scan(&n); err != nil || n != 0 {
				t.Fatalf("rollback incomplete on %s", table)
			}
		}
	}
	assertEmpty()
	t.Log("PASS: orphan MongoDB aborts and rolls back all 28 tables and trigger/FK state")
	if _, err := scans.DeleteOne(ctx, bson.M{"_id": orphan.InsertedID}); err != nil {
		t.Fatal(err)
	}
	if _, err := pg.Exec(ctx, "INSERT INTO public.proveedores(id_proveedor,nombre,documento) VALUES(9999,'Colisión QA','DOC-QA ')"); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(ctx, c); err == nil || !strings.Contains(err.Error(), "proveedores") {
		t.Fatalf("expected rejection of trailing-space collision: %v", err)
	}
	assertEmpty()
	if _, err := pg.Exec(ctx, "DELETE FROM public.proveedores WHERE id_proveedor=9999"); err != nil {
		t.Fatal(err)
	}
	t.Log("PASS: SQL Server collation collision is rejected without partial imports")

	const password = "Only synthetic QA migration"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pg.Exec(ctx, "UPDATE public.usuarios SET password_hash=$1 WHERE id_usuario=101", string(hash)); err != nil {
		t.Fatal(err)
	}
	report, err := Run(ctx, c)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Rows) != 28 || report.Sequences != 26 || report.Scans != 1 || report.Evidence != 1 {
		t.Fatalf("unexpected report: %+v", report)
	}
	for table, n := range report.Rows {
		if n == 0 {
			t.Fatalf("fixture missing table %s", table)
		}
	}
	if _, err := Run(ctx, c); err == nil || !strings.Contains(err.Error(), "destino no está vacío") {
		t.Fatalf("second run must reject occupied destination: %v", err)
	}
	t.Logf("PASS: 28 tables, 208 columns, 50 trusted FKs, 26 sequences, 1 scan and 1 GridFS file; counts %+v", report.Rows)

	for table, next := range map[string]int64{"pagos_id_pago_seq": 1001, "servicios_id_servicio_seq": 400, "usuarios_id_usuario_seq": 103, "diagnosticos_id_diagnostico_seq": 9007199254740994} {
		var actual int64
		if err := dst.QueryRowContext(ctx, "SELECT NEXT VALUE FOR dbo."+table).Scan(&actual); err != nil || actual != next {
			t.Fatalf("wrong next value for %s: %d", table, actual)
		}
	}
	store, err := sqlserver.Open(c.Target)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	user, err := store.FindLogin(ctx, "operador_qa")
	if err != nil || user.ID != "101" || !user.Can("pagos.ver") || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		t.Fatal("migrated credentials and permissions are unusable")
	}
	doc, err := store.SheetPDF(ctx, "91", user)
	var original []byte
	if e := pg.QueryRow(ctx, "SELECT ficha_pdf FROM public.entregas WHERE id_entrega=111").Scan(&original); e != nil || err != nil || !bytes.Equal(doc, original) {
		t.Fatal("archived PDF differs through the SQL Server adapter")
	}
	var historical int
	if err := dst.QueryRowContext(ctx, "SELECT COUNT(*) FROM dbo.asignaciones_ot WHERE id_asignacion=93 AND rol_en_trabajo=N'AUXILIAR' AND activo=0").Scan(&historical); err != nil || historical != 1 {
		t.Fatal("historical assistant assignment was altered")
	}
	_, err = dst.ExecContext(ctx, "INSERT dbo.asignaciones_ot(id_orden,id_usuario,rol_en_trabajo) VALUES(91,102,N'AUXILIAR')")
	var dbErr mssql.Error
	if !errors.As(err, &dbErr) || dbErr.Number != 51008 {
		t.Fatalf("expected active business trigger, got %v", err)
	}
	_, err = dst.ExecContext(ctx, "INSERT dbo.vehiculos(id_cliente,placa) VALUES(99999,N'QA-FK')")
	if !errors.As(err, &dbErr) || dbErr.Number != 547 {
		t.Fatalf("expected active foreign key, got %v", err)
	}
	t.Log("PASS: next IDs, migrated credentials, unchanged archived PDF and historical assignment, active trigger and FK")
}
