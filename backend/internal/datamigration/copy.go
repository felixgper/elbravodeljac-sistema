// Package datamigration copies one frozen, isolated PostgreSQL snapshot to a new SQL Server database.
package datamigration

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	mssql "github.com/microsoft/go-mssqldb"
)

const (
	SourceName = "elbravodeljac_pg_migration"
	TargetName = "elbravodeljac_sqlserver_migration"
	MongoName  = "elbravodeljac_tecnico_migration"
)

var tables = strings.Fields("asignaciones_ot auxilios avances_ot clientes contactos_cliente cotizaciones detalle_cotizacion diagnosticos empleados entregas historial_estados_ot historial_propietarios ordenes_trabajo pagos permisos proveedores pruebas_ot recepciones repuesto_proveedor repuestos rol_permisos roles servicios tipos_documento unidades_auxilio usuario_roles usuarios vehiculos")
var safeName = regexp.MustCompile("^[a-z][a-z0-9_]*$")
var markerPattern = regexp.MustCompile("^[0-9a-f]{64}$")

const session = `SET ANSI_NULLS ON; SET ANSI_PADDING ON; SET ANSI_WARNINGS ON; SET ARITHABORT ON;
SET CONCAT_NULL_YIELDS_NULL ON; SET QUOTED_IDENTIFIER ON; SET NUMERIC_ROUNDABORT OFF; SET XACT_ABORT ON; SET LOCK_TIMEOUT 10000;`

type Config struct{ Source, Target, Mongo, Marker string }
type Report struct {
	Rows      map[string]int `json:"rows"`
	Sequences int            `json:"sequences"`
	Scans     int            `json:"scans"`
	Evidence  int            `json:"evidence"`
	Balances  int            `json:"balances"`
}

type column struct{ name, kind string }
type table struct {
	name string
	cols []column
	keys []string
}
type toggle struct{ table, name string }

func (c Config) validate() error {
	if !markerPattern.MatchString(c.Marker) {
		return fmt.Errorf("se requiere un marcador QA/ensayo de 64 dígitos hexadecimales")
	}
	for _, v := range []struct{ raw, scheme, name string }{{c.Source, "postgres", SourceName}, {c.Target, "sqlserver", TargetName}, {c.Mongo, "mongodb", MongoName}} {
		u, err := url.Parse(v.raw)
		if err != nil || u.Scheme != v.scheme || u.Hostname() != "127.0.0.1" || u.Port() == "" || u.Fragment != "" {
			return fmt.Errorf("conexión inválida para %s: se requiere 127.0.0.1 con puerto explícito", v.name)
		}
		q, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			return fmt.Errorf("parámetros inválidos para %s", v.name)
		}
		switch v.scheme {
		case "postgres":
			if u.Path != "/"+v.name {
				return fmt.Errorf("origen incorrecto")
			}
			for k, vals := range q {
				if k != "sslmode" || len(vals) != 1 {
					return fmt.Errorf("opción de origen no admitida")
				}
			}
		case "sqlserver":
			if u.Path != "" || q.Get("database") != v.name || len(q["database"]) != 1 {
				return fmt.Errorf("destino incorrecto")
			}
			for k, vals := range q {
				if !slices.Contains([]string{"database", "encrypt", "TrustServerCertificate"}, k) || len(vals) != 1 {
					return fmt.Errorf("opción de destino no admitida")
				}
			}
		case "mongodb":
			if u.Path != "/"+v.name || q.Get("directConnection") != "true" {
				return fmt.Errorf("copia MongoDB incorrecta")
			}
			for k, vals := range q {
				if !slices.Contains([]string{"directConnection", "authSource", "tls"}, k) || len(vals) != 1 {
					return fmt.Errorf("opción MongoDB no admitida")
				}
			}
		}
	}
	return nil
}

// Run never changes the source or MongoDB. All destination changes use a single transaction.
func Run(ctx context.Context, c Config) (Report, error) {
	report := Report{Rows: map[string]int{}}
	if err := c.validate(); err != nil {
		return Report{}, err
	}
	source, err := pgx.Connect(ctx, c.Source)
	if err != nil {
		return Report{}, fmt.Errorf("no se pudo conectar al origen aislado")
	}
	defer source.Close(context.Background())
	pg, err := source.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return Report{}, fmt.Errorf("no se pudo abrir la instantánea de solo lectura")
	}
	defer pg.Rollback(context.Background())
	var name, marker string
	if err = pg.QueryRow(ctx, "SELECT current_database(),coalesce(shobj_description(oid,'pg_database'),'') FROM pg_database WHERE datname=current_database()").Scan(&name, &marker); err != nil || name != SourceName || marker != c.Marker {
		return Report{}, fmt.Errorf("origen sin marcador de la copia aislada")
	}
	connector, err := mssql.NewConnector(c.Target)
	if err != nil {
		return Report{}, fmt.Errorf("configuración inválida del destino")
	}
	connector.SessionInitSQL = session
	db := sql.OpenDB(connector)
	defer db.Close()
	dst, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return Report{}, fmt.Errorf("no se pudo abrir la transacción de destino")
	}
	defer dst.Rollback()
	if err = dst.QueryRowContext(ctx, `SELECT DB_NAME(),COALESCE((SELECT CONVERT(varchar(64),value) FROM sys.extended_properties WHERE class=0 AND name=N'jac.migration_run'),'' )`).Scan(&name, &marker); err != nil || name != TargetName || marker != c.Marker {
		return Report{}, fmt.Errorf("destino sin marcador de esta ejecución")
	}
	var lock int
	if err = dst.QueryRowContext(ctx, `DECLARE @r int; EXEC @r=sys.sp_getapplock @Resource=N'jac:migration',@LockMode=N'Exclusive',@LockOwner=N'Transaction',@LockTimeout=0; SELECT @r;`).Scan(&lock); err != nil || lock < 0 {
		return Report{}, fmt.Errorf("otra migración tiene el bloqueo del destino")
	}

	catalog, fks, triggers, err := inventory(ctx, pg, dst)
	if err != nil {
		return Report{}, err
	}
	for _, t := range catalog {
		var count int64
		if err = dst.QueryRowContext(ctx, "SELECT COUNT_BIG(*) FROM dbo."+t.name+" WITH (TABLOCKX,HOLDLOCK)").Scan(&count); err != nil || count != 0 {
			return Report{}, fmt.Errorf("destino no está vacío o no se puede bloquear: %s", t.name)
		}
	}
	for _, t := range triggers {
		if _, err = dst.ExecContext(ctx, "DISABLE TRIGGER dbo."+t.name+" ON dbo."+t.table); err != nil {
			return Report{}, fmt.Errorf("no se pudo suspender trigger %s", t.name)
		}
	}
	for _, f := range fks {
		if _, err = dst.ExecContext(ctx, "ALTER TABLE dbo."+f.table+" NOCHECK CONSTRAINT "+f.name); err != nil {
			return Report{}, fmt.Errorf("no se pudo diferir FK %s", f.name)
		}
	}

	for _, t := range catalog {
		count, err := copyTable(ctx, pg, dst, t)
		if err != nil {
			return Report{}, err
		}
		report.Rows[t.name] = count
	}
	for _, f := range fks {
		if _, err = dst.ExecContext(ctx, "ALTER TABLE dbo."+f.table+" WITH CHECK CHECK CONSTRAINT "+f.name); err != nil {
			return Report{}, fmt.Errorf("FK no validada: %s", f.name)
		}
	}
	for _, t := range triggers {
		if _, err = dst.ExecContext(ctx, "ENABLE TRIGGER dbo."+t.name+" ON dbo."+t.table); err != nil {
			return Report{}, fmt.Errorf("no se pudo reactivar trigger %s", t.name)
		}
	}
	if err = verifyIntegrity(ctx, dst); err != nil {
		return Report{}, err
	}
	report.Balances, err = verifyBalances(ctx, pg, dst)
	if err != nil {
		return Report{}, err
	}
	report.Scans, report.Evidence, err = verifyMongo(ctx, c, pg, dst)
	if err != nil {
		return Report{}, err
	}
	report.Sequences, err = restartSequences(ctx, pg, dst, catalog)
	if err != nil {
		return Report{}, err
	}
	if err = dst.Commit(); err != nil {
		return Report{}, fmt.Errorf("confirmación del destino no verificable: descarta la copia y vuelve a ensayar")
	}
	return report, nil
}

// Compare the business-facing amount, net payments and outstanding balance as
// decimals on both engines. Row equality alone cannot catch divergent functions.
func verifyBalances(ctx context.Context, pg pgx.Tx, dst *sql.Tx) (int, error) {
	source, err := pg.Query(ctx, `SELECT o.id_orden::text,
  x.amount::numeric(38,2)::text,x.paid::numeric(38,2)::text,(x.amount-x.paid)::numeric(38,2)::text
 FROM public.ordenes_trabajo o
 CROSS JOIN LATERAL (SELECT public.jac_importe_orden(o.id_orden) amount,
   COALESCE((SELECT SUM(p.monto) FROM public.pagos p WHERE p.id_orden=o.id_orden AND p.revertido_en IS NULL),0) paid) x
 ORDER BY o.id_orden`)
	if err != nil {
		return 0, fmt.Errorf("no se pudo calcular saldos de origen")
	}
	defer source.Close()
	target, err := dst.QueryContext(ctx, `SELECT CONVERT(varchar(20),o.id_orden),
  CONVERT(varchar(42),CAST(x.amount AS decimal(38,2))),
  CONVERT(varchar(42),CAST(x.paid AS decimal(38,2))),
  CONVERT(varchar(42),CAST(x.amount AS decimal(37,2))-CAST(x.paid AS decimal(37,2)))
 FROM dbo.ordenes_trabajo o
 CROSS APPLY (SELECT dbo.jac_importe_orden(o.id_orden) amount,
   COALESCE((SELECT SUM(p.monto) FROM dbo.pagos p WHERE p.id_orden=o.id_orden AND p.revertido_en IS NULL),0) paid) x
 ORDER BY o.id_orden`)
	if err != nil {
		return 0, fmt.Errorf("no se pudo calcular saldos de destino")
	}
	defer target.Close()
	count := 0
	for source.Next() {
		var a, b [4]string
		if !target.Next() || source.Scan(&a[0], &a[1], &a[2], &a[3]) != nil ||
			target.Scan(&b[0], &b[1], &b[2], &b[3]) != nil || a != b {
			return 0, fmt.Errorf("los saldos difieren en la orden de posición %d", count+1)
		}
		count++
	}
	if target.Next() || source.Err() != nil || target.Err() != nil {
		return 0, fmt.Errorf("la verificación de saldos no fue completa")
	}
	return count, nil
}

func fields(t table, target bool) string {
	out := make([]string, len(t.cols))
	for i, c := range t.cols {
		switch c.kind {
		case "bool":
			if target {
				out[i] = "CASE " + c.name + " WHEN 1 THEN N'1' WHEN 0 THEN N'0' END"
			} else {
				out[i] = "CASE " + c.name + " WHEN true THEN '1' WHEN false THEN '0' END"
			}
		case "bytea", "timestamptz":
			out[i] = c.name
		default:
			if target {
				out[i] = "CONVERT(nvarchar(max)," + c.name + ")"
			} else {
				out[i] = c.name + "::text"
			}
		}
	}
	return strings.Join(out, ",")
}

func rowHash(h hash.Hash, row []any) error {
	for _, v := range row {
		var raw []byte
		switch x := v.(type) {
		case nil:
			h.Write([]byte{0})
			continue
		case []byte:
			raw = x
		case string:
			raw = []byte(x)
		case time.Time:
			if x.Year() < 1 || x.Year() > 9999 {
				return fmt.Errorf("fecha fuera del rango de SQL Server")
			}
			raw = []byte(x.UTC().Format(time.RFC3339Nano))
		default:
			return fmt.Errorf("tipo no representable: %T", v)
		}
		h.Write([]byte{1})
		h.Write(binary.BigEndian.AppendUint64(nil, uint64(len(raw))))
		h.Write(raw)
	}
	return nil
}

func copyTable(ctx context.Context, pg pgx.Tx, dst *sql.Tx, t table) (int, error) {
	keys := strings.Join(t.keys, ",")
	names, params := make([]string, len(t.cols)), make([]string, len(t.cols))
	for i, c := range t.cols {
		names[i] = c.name
		params[i] = fmt.Sprintf("@p%d", i+1)
		if c.kind == "bytea" {
			params[i] = "CONVERT(varbinary(max)," + params[i] + ")"
		}
	}
	stmt, err := dst.PrepareContext(ctx, "INSERT dbo."+t.name+" ("+strings.Join(names, ",")+") VALUES ("+strings.Join(params, ",")+")")
	if err != nil {
		return 0, fmt.Errorf("no se pudo preparar carga: %s", t.name)
	}
	defer stmt.Close()
	rows, err := pg.Query(ctx, "SELECT "+fields(t, false)+" FROM public."+t.name+" ORDER BY "+keys)
	if err != nil {
		return 0, fmt.Errorf("no se pudo leer origen: %s", t.name)
	}
	defer rows.Close()
	left, right := sha256.New(), sha256.New()
	count := 0
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil || rowHash(left, vals) != nil {
			return 0, fmt.Errorf("dato no representable en %s, fila %d", t.name, count+1)
		}
		for i, v := range vals {
			if at, ok := v.(time.Time); ok {
				vals[i] = at.UTC()
			}
		}
		if _, err = stmt.ExecContext(ctx, vals...); err != nil {
			if dbErr, ok := err.(mssql.Error); ok {
				return 0, fmt.Errorf("SQL Server rechazó %s fila %d (error %d)", t.name, count+1, dbErr.Number)
			}
			return 0, fmt.Errorf("SQL Server rechazó %s fila %d", t.name, count+1)
		}
		count++
	}
	if rows.Err() != nil {
		return 0, fmt.Errorf("lectura de origen interrumpida: %s", t.name)
	}
	target, err := dst.QueryContext(ctx, "SELECT "+fields(t, true)+" FROM dbo."+t.name+" ORDER BY "+keys)
	if err != nil {
		return 0, fmt.Errorf("no se pudo releer el destino: %s", t.name)
	}
	defer target.Close()
	vals, pointers := make([]any, len(t.cols)), make([]any, len(t.cols))
	for i := range vals {
		pointers[i] = &vals[i]
	}
	n := 0
	for target.Next() {
		if err = target.Scan(pointers...); err != nil || rowHash(right, vals) != nil {
			return 0, fmt.Errorf("dato de destino ilegible: %s", t.name)
		}
		n++
	}
	if target.Err() != nil || n != count || !slices.Equal(left.Sum(nil), right.Sum(nil)) {
		return 0, fmt.Errorf("conteo o contenido no coincide: %s", t.name)
	}
	return count, nil
}

func restartSequences(ctx context.Context, pg pgx.Tx, dst *sql.Tx, catalog []table) (int, error) {
	n := 0
	for _, t := range catalog {
		if len(t.keys) != 1 {
			continue
		}
		key := t.keys[0]
		seq := t.name + "_" + key + "_seq"
		var owned string
		if err := pg.QueryRow(ctx, "SELECT COALESCE(pg_get_serial_sequence(@table,@col),'')", pgx.NamedArgs{"table": "public." + t.name, "col": key}).Scan(&owned); err != nil || owned != "public."+seq {
			return 0, fmt.Errorf("secuencia de origen inesperada: %s", t.name)
		}
		var last, maxID int64
		var called bool
		if err := pg.QueryRow(ctx, "SELECT last_value,is_called FROM public."+seq).Scan(&last, &called); err != nil {
			return 0, fmt.Errorf("secuencia de origen ilegible: %s", seq)
		}
		if err := dst.QueryRowContext(ctx, "SELECT COALESCE(MAX("+key+"),0) FROM dbo."+t.name).Scan(&maxID); err != nil {
			return 0, fmt.Errorf("clave máxima ilegible: %s", t.name)
		}
		if last < 1 || last == math.MaxInt64 || maxID == math.MaxInt64 {
			return 0, fmt.Errorf("secuencia fuera de rango: %s", seq)
		}
		if called {
			last++
		}
		next := max(last, maxID+1)
		if _, err := dst.ExecContext(ctx, fmt.Sprintf("ALTER SEQUENCE dbo.%s RESTART WITH %d", seq, next)); err != nil {
			return 0, fmt.Errorf("no se pudo reiniciar %s", seq)
		}
		n++
	}
	if n != 26 {
		return 0, fmt.Errorf("inventario de secuencias inesperado")
	}
	return n, nil
}
