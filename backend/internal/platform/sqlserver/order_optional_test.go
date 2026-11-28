package sqlserver

// These transaction tests use only database/sql's driver contract, never a server.
import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"io"
	"strings"
	"testing"
	"time"
)

type optionalOrderDriver struct{ c *optionalOrderConn }

func (d optionalOrderDriver) Open(string) (driver.Conn, error)             { return d.c, nil }
func (d optionalOrderDriver) Connect(context.Context) (driver.Conn, error) { return d.c, nil }
func (d optionalOrderDriver) Driver() driver.Driver                        { return d }

type optionalOrderConn struct {
	assigned, closed, committed, rolled bool
	state, control, paid                string
	execs                               []string
	testValues                          []driver.NamedValue
	failUpdate                          bool
}

func (c *optionalOrderConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected Prepare")
}
func (c *optionalOrderConn) Close() error              { return nil }
func (c *optionalOrderConn) Begin() (driver.Tx, error) { return c, nil }
func (c *optionalOrderConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return c, nil
}
func (c *optionalOrderConn) Commit() error   { c.committed = true; return nil }
func (c *optionalOrderConn) Rollback() error { c.rolled = true; return nil }
func (c *optionalOrderConn) ExecContext(_ context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
	if strings.HasPrefix(q, "INSERT INTO pruebas_ot") {
		c.testValues = append([]driver.NamedValue(nil), args...)
	}
	if c.failUpdate && strings.HasPrefix(q, "UPDATE ordenes_trabajo") {
		return nil, errors.New("controlled update failure")
	}
	c.execs = append(c.execs, q)
	return driver.RowsAffected(1), nil
}

type optionalRows struct {
	data    [][]driver.Value
	columns []string
}

func (r *optionalRows) Columns() []string { return r.columns }
func (r *optionalRows) Close() error      { return nil }
func (r *optionalRows) Next(dest []driver.Value) error {
	if len(r.data) == 0 {
		return io.EOF
	}
	copy(dest, r.data[0])
	r.data = r.data[1:]
	return nil
}
func one(values ...driver.Value) driver.Rows {
	cols := make([]string, len(values))
	for i := range cols {
		cols[i] = fmt.Sprint(i)
	}
	return &optionalRows{[][]driver.Value{values}, cols}
}
func (c *optionalOrderConn) QueryContext(_ context.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
	switch {
	case strings.HasPrefix(q, "SELECT estado FROM ordenes_trabajo"):
		return one(c.state), nil
	case strings.Contains(q, "SELECT 1 FROM entregas"):
		return one(c.closed), nil
	case strings.Contains(q, "SELECT 1 FROM asignaciones_ot"):
		return one(c.assigned), nil
	case strings.Contains(q, "coalesce(max(CONVERT(varchar(20),id_usuario))"):
		return one("7"), nil
	case strings.HasPrefix(q, "SELECT greatest"):
		return one(time.Date(2026, 11, 26, 15, 0, 0, 0, time.UTC)), nil
	case strings.HasPrefix(q, "SELECT flujo_version,control_tecnico"):
		return one(int64(1), []byte(c.control)), nil
	case strings.HasPrefix(q, "SELECT CONVERT(varchar(20),d.id_detalle)"):
		return one("9"), nil
	case strings.HasPrefix(q, "SELECT CONVERT(varchar(42),dbo.jac_importe_orden"):
		return one("100.00", c.paid), nil
	case strings.Contains(q, "AS id") || strings.Contains(q, "FOR JSON PATH"):
		return one([]byte(`{"id":"1","state":"Finalizada"}`)), nil
	default:
		return nil, fmt.Errorf("unexpected query: %.100s", q)
	}
}
func TestOrderOptionalTestTransactions(t *testing.T) {
	for _, tc := range []struct {
		name, state, test, result                              string
		denied, closed, badExecution, overpaid, fail, conflict bool
	}{
		{name: "advance without test", state: "En reparación"},
		{name: "finalize without test", state: "Finalizada"},
		{name: "finalize with explicit test", state: "Finalizada", test: "Ruta", result: "Satisfactoria"},
		{name: "failed test remains recorded", state: "En pruebas", test: "Frenado", result: "Fallida"},
		{name: "unassigned mechanic denied", state: "Finalizada", denied: true},
		{name: "delivered order remains closed", state: "Finalizada", closed: true},
		{name: "finalized order remains closed", state: "Finalizada", conflict: true},
		{name: "execution still required", state: "Finalizada", badExecution: true},
		{name: "overpayment still blocked", state: "Finalizada", overpaid: true},
		{name: "write error rolls back test and history", state: "Finalizada", test: "Ruta", result: "Satisfactoria", fail: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &optionalOrderConn{assigned: !tc.denied, closed: tc.closed, state: "en_reparacion", control: `{"execution":{"9":{"status":"Realizado"}}}`, paid: "0.00", failUpdate: tc.fail}
			if tc.conflict {
				c.state = "finalizada"
			}
			if tc.badExecution {
				c.control = `{}`
			}
			if tc.overpaid {
				c.paid = "110.00"
			}
			db := sql.OpenDB(optionalOrderDriver{c})
			defer db.Close()
			store := &Store{db: db}
			_, err := store.UpdateOrder(context.Background(), "1", domain.UpdateOrderInput{State: tc.state, Note: "Avance sintético en memoria", Test: tc.test, Result: tc.result}, domain.User{ID: "7", Permissions: []string{"ordenes.editar"}})
			rejected := tc.denied || tc.closed || tc.conflict || tc.badExecution || tc.overpaid || tc.fail
			if rejected {
				if err == nil || c.committed || !c.rolled {
					t.Fatalf("expected rollback, err=%v commit=%v rollback=%v", err, c.committed, c.rolled)
				}
				return
			}
			if err != nil || !c.committed || c.rolled {
				t.Fatalf("expected commit: %v", err)
			}
			if tc.test == "" {
				if len(c.testValues) != 0 {
					t.Fatal("fabricated test")
				}
			} else {
				if len(c.testValues) != 6 || c.testValues[2].Value != tc.test || c.testValues[3].Value != strings.ToLower(tc.result) {
					t.Fatal("explicit test lost")
				}
			}
		})
	}
}
