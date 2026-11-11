package datamigration

import (
	"context"
	"database/sql"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"
)

func inventory(ctx context.Context, pg pgx.Tx, dst *sql.Tx) ([]table, []toggle, []toggle, error) {
	source, err := pg.Query(ctx, `SELECT cl.relname, a.attname, typ.typname
 FROM pg_class cl JOIN pg_namespace ns ON ns.oid=cl.relnamespace
 JOIN pg_attribute a ON a.attrelid=cl.oid AND a.attnum>0 AND NOT a.attisdropped
 JOIN pg_type typ ON typ.oid=a.atttypid
 WHERE ns.nspname='public' AND cl.relkind='r' ORDER BY cl.relname,a.attname COLLATE "C"`)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("inventario PostgreSQL ilegible")
	}
	defer source.Close()
	src := map[string][]column{}
	cols := 0
	types := map[string]string{"int8": "bigint", "int4": "int", "int2": "smallint", "varchar": "nvarchar", "text": "nvarchar", "numeric": "decimal", "bool": "bit", "jsonb": "nvarchar", "timestamptz": "datetimeoffset", "bytea": "varbinary"}
	for source.Next() {
		var t string
		var c column
		if source.Scan(&t, &c.name, &c.kind) != nil || !slices.Contains(tables, t) || !safeName.MatchString(c.name) || types[c.kind] == "" {
			return nil, nil, nil, fmt.Errorf("tabla, columna o tipo PostgreSQL no admitido")
		}
		src[t] = append(src[t], c)
		cols++
	}
	if source.Err() != nil || len(src) != 28 || cols != 208 {
		return nil, nil, nil, fmt.Errorf("se esperan 28 tablas y 208 columnas en PostgreSQL")
	}

	target, err := dst.QueryContext(ctx, `SELECT t.name,c.name,TYPE_NAME(c.user_type_id)
 FROM sys.tables t JOIN sys.columns c ON c.object_id=t.object_id
 WHERE t.schema_id=SCHEMA_ID(N'dbo') ORDER BY t.name COLLATE Latin1_General_100_BIN2,c.name COLLATE Latin1_General_100_BIN2`)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("inventario SQL Server ilegible")
	}
	defer target.Close()
	seen := map[string]int{}
	for target.Next() {
		var t, c, k string
		if target.Scan(&t, &c, &k) != nil || seen[t] >= len(src[t]) {
			return nil, nil, nil, fmt.Errorf("tabla o columna inesperada en SQL Server")
		}
		expected := src[t][seen[t]]
		if c != expected.name || k != types[expected.kind] {
			return nil, nil, nil, fmt.Errorf("columna o tipo incompatible: %s.%s", t, c)
		}
		seen[t]++
	}
	if target.Err() != nil || len(seen) != len(src) {
		return nil, nil, nil, fmt.Errorf("inventario de destino incompleto")
	}
	for t, cs := range src {
		if seen[t] != len(cs) {
			return nil, nil, nil, fmt.Errorf("faltan columnas en %s", t)
		}
	}

	keys, err := pg.Query(ctx, `SELECT cl.relname,a.attname
 FROM pg_constraint k JOIN pg_class cl ON cl.oid=k.conrelid JOIN pg_namespace n ON n.oid=cl.relnamespace
 CROSS JOIN LATERAL unnest(k.conkey) WITH ORDINALITY p(attnum,pos)
 JOIN pg_attribute a ON a.attrelid=cl.oid AND a.attnum=p.attnum
 WHERE k.contype='p' AND n.nspname='public' ORDER BY cl.relname,p.pos`)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("PK de origen ilegibles")
	}
	defer keys.Close()
	pk := map[string][]string{}
	for keys.Next() {
		var t, c string
		if keys.Scan(&t, &c) != nil || !slices.Contains(tables, t) || !safeName.MatchString(c) {
			return nil, nil, nil, fmt.Errorf("PK de origen inesperada")
		}
		pk[t] = append(pk[t], c)
	}
	if keys.Err() != nil || len(pk) != 28 {
		return nil, nil, nil, fmt.Errorf("PK de origen incompletas")
	}
	target, err = dst.QueryContext(ctx, `SELECT t.name,c.name FROM sys.indexes i
 JOIN sys.tables t ON i.object_id=t.object_id JOIN sys.index_columns ic ON ic.object_id=i.object_id AND ic.index_id=i.index_id
 JOIN sys.columns c ON c.object_id=ic.object_id AND c.column_id=ic.column_id
 WHERE i.is_primary_key=1 AND t.schema_id=SCHEMA_ID(N'dbo') ORDER BY t.name,ic.key_ordinal`)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("PK de destino ilegibles")
	}
	defer target.Close()
	seen = map[string]int{}
	for target.Next() {
		var t, c string
		if target.Scan(&t, &c) != nil || seen[t] >= len(pk[t]) || c != pk[t][seen[t]] {
			return nil, nil, nil, fmt.Errorf("PK de destino diferente")
		}
		seen[t]++
	}
	if target.Err() != nil || len(seen) != 28 {
		return nil, nil, nil, fmt.Errorf("PK de destino incompletas")
	}
	catalog := make([]table, 0, len(tables))
	for _, t := range tables {
		if seen[t] != len(pk[t]) {
			return nil, nil, nil, fmt.Errorf("PK incompleta: %s", t)
		}
		catalog = append(catalog, table{t, src[t], pk[t]})
	}

	fk, err := pg.Query(ctx, `SELECT cl.relname,a.attname,ref.relname,b.attname,k.convalidated
 FROM pg_constraint k JOIN pg_class cl ON cl.oid=k.conrelid JOIN pg_namespace n ON n.oid=cl.relnamespace
 JOIN pg_class ref ON ref.oid=k.confrelid
 CROSS JOIN LATERAL unnest(k.conkey,k.confkey) p(c,r)
 JOIN pg_attribute a ON a.attrelid=cl.oid AND a.attnum=p.c
 JOIN pg_attribute b ON b.attrelid=ref.oid AND b.attnum=p.r
 WHERE k.contype='f' AND n.nspname='public'`)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("FK de origen ilegibles")
	}
	defer fk.Close()
	edges := map[[4]string]int{}
	for fk.Next() {
		var e [4]string
		var trusted bool
		if fk.Scan(&e[0], &e[1], &e[2], &e[3], &trusted) != nil || !trusted {
			return nil, nil, nil, fmt.Errorf("FK de origen sin validar")
		}
		edges[e]++
	}
	if fk.Err() != nil {
		return nil, nil, nil, fmt.Errorf("FK de origen ilegibles")
	}
	target, err = dst.QueryContext(ctx, `SELECT f.name,t.name,c.name,r.name,d.name
 FROM sys.foreign_keys f JOIN sys.foreign_key_columns x ON f.object_id=x.constraint_object_id
 JOIN sys.tables t ON t.object_id=x.parent_object_id JOIN sys.tables r ON r.object_id=x.referenced_object_id
 JOIN sys.columns c ON c.object_id=t.object_id AND c.column_id=x.parent_column_id
 JOIN sys.columns d ON d.object_id=r.object_id AND d.column_id=x.referenced_column_id
 WHERE f.schema_id=SCHEMA_ID(N'dbo') ORDER BY f.name`)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("FK de destino ilegibles")
	}
	defer target.Close()
	var fks []toggle
	for target.Next() {
		var n string
		var e [4]string
		if target.Scan(&n, &e[0], &e[1], &e[2], &e[3]) != nil || !safeName.MatchString(n) || !slices.Contains(tables, e[0]) || edges[e] != 1 {
			return nil, nil, nil, fmt.Errorf("relación de destino inesperada")
		}
		delete(edges, e)
		fks = append(fks, toggle{e[0], n})
	}
	if target.Err() != nil || len(fks) != 50 || len(edges) != 0 {
		return nil, nil, nil, fmt.Errorf("se esperan las mismas 50 relaciones")
	}
	if err = verifyIntegrity(ctx, dst); err != nil {
		return nil, nil, nil, err
	}
	target, err = dst.QueryContext(ctx, `SELECT t.name,g.name FROM sys.triggers g JOIN sys.tables t ON t.object_id=g.parent_id
 WHERE t.schema_id=SCHEMA_ID(N'dbo') ORDER BY t.name,g.name`)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("triggers de destino ilegibles")
	}
	defer target.Close()
	var triggers []toggle
	for target.Next() {
		var t toggle
		if target.Scan(&t.table, &t.name) != nil || !slices.Contains(tables, t.table) || !safeName.MatchString(t.name) {
			return nil, nil, nil, fmt.Errorf("trigger inesperado")
		}
		triggers = append(triggers, t)
	}
	if target.Err() != nil || len(triggers) != 10 {
		return nil, nil, nil, fmt.Errorf("se esperan 10 triggers")
	}
	return catalog, fks, triggers, nil
}

func verifyIntegrity(ctx context.Context, tx *sql.Tx) error {
	var ok int
	err := tx.QueryRowContext(ctx, `SELECT CASE WHEN
  (SELECT COUNT(*) FROM sys.foreign_keys WHERE schema_id=SCHEMA_ID(N'dbo') AND is_disabled=0 AND is_not_trusted=0)=50
  AND NOT EXISTS(SELECT 1 FROM sys.check_constraints WHERE schema_id=SCHEMA_ID(N'dbo') AND (is_disabled=1 OR is_not_trusted=1))
  AND NOT EXISTS(SELECT 1 FROM sys.indexes i JOIN sys.tables t ON t.object_id=i.object_id
    WHERE t.schema_id=SCHEMA_ID(N'dbo') AND i.is_disabled=1)
  AND (SELECT COUNT(*) FROM sys.triggers g JOIN sys.tables t ON t.object_id=g.parent_id
    WHERE t.schema_id=SCHEMA_ID(N'dbo') AND g.is_disabled=0)=10
  AND (SELECT COUNT(*) FROM sys.sequences WHERE schema_id=SCHEMA_ID(N'dbo'))=26
 THEN 1 ELSE 0 END`).Scan(&ok)
	if err != nil || ok != 1 {
		return fmt.Errorf("restricciones, índices, triggers o secuencias no disponibles o sin validar")
	}
	return nil
}
