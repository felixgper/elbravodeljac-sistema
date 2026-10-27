package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ Pool *pgxpool.Pool }

func (s *Store) Ping(ctx context.Context) error { return s.Pool.Ping(ctx) }
func translate(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		switch pg.Code {
		case "23505", "23503", "40001", "40P01":
			return domain.ErrConflict
		case "23514", "22001", "22003", "23502":
			return domain.Invalid("request", "Los valores no cumplen las restricciones de PostgreSQL.")
		}
	}
	return err
}
func (s *Store) User(ctx context.Context, id string) (domain.User, error) {
	return s.user(ctx, "u.id_usuario::text=$1", id)
}
func (s *Store) FindLogin(ctx context.Context, login string) (domain.User, error) {
	// Ambiguous user/email matches are rejected instead of picking another account.
	return s.user(ctx, "(u.usuario=$1 OR u.correo=$1)", login)
}
func (s *Store) user(ctx context.Context, where, value string) (domain.User, error) {
	rows, err := s.Pool.Query(ctx, `SELECT u.id_usuario::text,u.nombre,u.usuario,coalesce(u.correo,''),u.activo,u.password_hash,
 ARRAY(SELECT r.nombre_rol FROM roles r JOIN usuario_roles ur USING(id_rol) WHERE ur.id_usuario=u.id_usuario ORDER BY r.nombre_rol),
 ARRAY(SELECT DISTINCT p.codigo FROM permisos p JOIN rol_permisos rp USING(id_permiso) JOIN usuario_roles ur USING(id_rol) WHERE ur.id_usuario=u.id_usuario ORDER BY p.codigo)
 FROM usuarios u WHERE `+where, value)
	if err != nil {
		return domain.User{}, translate(err)
	}
	defer rows.Close()
	users, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.User, error) {
		var u domain.User
		err := row.Scan(&u.ID, &u.Name, &u.Username, &u.Email, &u.Active, &u.PasswordHash, &u.Roles, &u.Permissions)
		return u, err
	})
	if err != nil {
		return domain.User{}, err
	}
	if len(users) != 1 {
		return domain.User{}, domain.ErrNotFound
	}
	return users[0], nil
}

const clientCols = `id_cliente::text,nombre,tipo_documento,numero_documento,coalesce(telefono,''),coalesce(direccion,''),coalesce(correo,'')`

func scanClient(row pgx.Row) (domain.Client, error) {
	var v domain.Client
	err := row.Scan(&v.ID, &v.Name, &v.DocumentType, &v.DocumentNumber, &v.Phone, &v.Address, &v.Email)
	return v, translate(err)
}
func (s *Store) Client(ctx context.Context, id string) (domain.Client, error) {
	return scanClient(s.Pool.QueryRow(ctx, `SELECT `+clientCols+` FROM clientes WHERE id_cliente=$1`, id))
}
func (s *Store) Clients(ctx context.Context, p domain.Page) (domain.List[domain.Client], error) {
	out := domain.List[domain.Client]{Data: []domain.Client{}, Meta: domain.Meta{Page: p.Number, PageSize: p.Size}}
	where := ` WHERE jac_normalize(nombre || ' ' || numero_documento) LIKE jac_normalize($1) OR EXISTS(SELECT 1 FROM vehiculos v WHERE v.id_cliente=clientes.id_cliente AND v.placa ILIKE $1)`
	err := s.Pool.QueryRow(ctx, `SELECT count(*) FROM clientes`+where, "%"+p.Query+"%").Scan(&out.Meta.Total)
	if err != nil {
		return out, translate(err)
	}
	rows, err := s.Pool.Query(ctx, `SELECT `+clientCols+`,(SELECT coalesce(jsonb_agg(jsonb_build_object('id',v.id_vehiculo::text,'clientId',v.id_cliente::text,'plate',v.placa,'brand',v.marca,'model',v.modelo) ORDER BY v.id_vehiculo),'[]'::jsonb) FROM vehiculos v WHERE v.id_cliente=clientes.id_cliente) FROM clientes`+where+` ORDER BY id_cliente LIMIT $2 OFFSET $3`, "%"+p.Query+"%", p.Size, (p.Number-1)*p.Size)
	if err != nil {
		return out, translate(err)
	}
	defer rows.Close()
	for rows.Next() {
		var v domain.Client
		var raw []byte
		err := rows.Scan(&v.ID, &v.Name, &v.DocumentType, &v.DocumentNumber, &v.Phone, &v.Address, &v.Email, &raw)
		if err == nil {
			err = json.Unmarshal(raw, &v.Vehicles)
		}
		if err != nil {
			return out, err
		}
		out.Data = append(out.Data, v)
	}
	return out, translate(rows.Err())
}
func (s *Store) CreateClient(ctx context.Context, v domain.ClientInput) (domain.Client, error) {
	if err := v.Validate(); err != nil {
		return domain.Client{}, err
	}
	return scanClient(s.Pool.QueryRow(ctx, `INSERT INTO clientes(nombre,tipo_documento,numero_documento,telefono,direccion,correo) VALUES($1,$2,$3,$4,$5,nullif($6,'')) RETURNING `+clientCols, v.Name, v.DocumentType, v.DocumentNumber, v.Phone, v.Address, v.Email))
}
func (s *Store) PatchClient(ctx context.Context, id string, patch map[string]string) (domain.Client, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return domain.Client{}, err
	}
	defer tx.Rollback(ctx)
	v, err := scanClient(tx.QueryRow(ctx, `SELECT `+clientCols+` FROM clientes WHERE id_cliente=$1 FOR UPDATE`, id))
	if err != nil {
		return v, err
	}
	fields := map[string]*string{"name": &v.Name, "documentType": &v.DocumentType, "documentNumber": &v.DocumentNumber, "phone": &v.Phone, "address": &v.Address, "email": &v.Email}
	for k, value := range patch {
		ptr, ok := fields[k]
		if !ok {
			return v, domain.Invalid(k, "Campo no editable.")
		}
		*ptr = value
	}
	if err := v.ClientInput.Validate(); err != nil {
		return v, err
	}
	v, err = scanClient(tx.QueryRow(ctx, `UPDATE clientes SET nombre=$2,tipo_documento=$3,numero_documento=$4,telefono=$5,direccion=$6,correo=nullif($7,'') WHERE id_cliente=$1 RETURNING `+clientCols, id, v.Name, v.DocumentType, v.DocumentNumber, v.Phone, v.Address, v.Email))
	if err != nil {
		return v, err
	}
	return v, translate(tx.Commit(ctx))
}
func (s *Store) DeleteClient(ctx context.Context, id string) error {
	tag, err := s.Pool.Exec(ctx, `DELETE FROM clientes WHERE id_cliente=$1`, id)
	if err != nil {
		return translate(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

const contactCols = `id_contacto::text,id_cliente::text,activo,nombre,tipo_documento,numero_documento,coalesce(telefono,''),coalesce(relacion,'')`

func scanContact(row pgx.Row) (domain.Contact, error) {
	var c domain.Contact
	err := row.Scan(&c.ID, &c.ClientID, &c.Active, &c.Name, &c.DocumentType, &c.DocumentNumber, &c.Phone, &c.Relation)
	return c, translate(err)
}
func (s *Store) Contacts(ctx context.Context, id string) ([]domain.Contact, error) {
	if _, err := s.Client(ctx, id); err != nil {
		return nil, err
	}
	rows, err := s.Pool.Query(ctx, `SELECT `+contactCols+` FROM contactos_cliente WHERE id_cliente=$1 ORDER BY id_contacto`, id)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	out := []domain.Contact{}
	for rows.Next() {
		c, err := scanContact(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, translate(rows.Err())
}
func (s *Store) CreateContact(ctx context.Context, id string, v domain.ContactInput) (domain.Contact, error) {
	if err := v.Validate(); err != nil {
		return domain.Contact{}, err
	}
	return scanContact(s.Pool.QueryRow(ctx, `INSERT INTO contactos_cliente(id_cliente,nombre,tipo_documento,numero_documento,telefono,relacion) VALUES($1,$2,$3,$4,$5,$6) RETURNING `+contactCols, id, v.Name, v.DocumentType, v.DocumentNumber, v.Phone, v.Relation))
}

const vehicleCols = `id_vehiculo::text,id_cliente::text,placa,coalesce(marca,''),coalesce(modelo,''),coalesce(anio,0),coalesce(kilometraje_actual,0)::text,coalesce(observaciones,'')`

func scanVehicle(row pgx.Row) (domain.Vehicle, error) {
	var v domain.Vehicle
	err := row.Scan(&v.ID, &v.ClientID, &v.Plate, &v.Brand, &v.Model, &v.Year, &v.Mileage, &v.Notes)
	return v, translate(err)
}
func (s *Store) Vehicle(ctx context.Context, id string) (domain.Vehicle, error) {
	return scanVehicle(s.Pool.QueryRow(ctx, `SELECT `+vehicleCols+` FROM vehiculos WHERE id_vehiculo=$1`, id))
}
func (s *Store) Vehicles(ctx context.Context, p domain.Page) (domain.List[domain.Vehicle], error) {
	out := domain.List[domain.Vehicle]{Data: []domain.Vehicle{}, Meta: domain.Meta{Page: p.Number, PageSize: p.Size}}
	where := ` WHERE (placa ILIKE $1 OR modelo ILIKE $1 OR EXISTS(SELECT 1 FROM clientes c WHERE c.id_cliente=vehiculos.id_cliente AND c.nombre ILIKE $1)) AND ($2='' OR id_cliente::text=$2)`
	if err := s.Pool.QueryRow(ctx, `SELECT count(*) FROM vehiculos`+where, "%"+p.Query+"%", p.ClientID).Scan(&out.Meta.Total); err != nil {
		return out, translate(err)
	}
	rows, err := s.Pool.Query(ctx, `SELECT `+vehicleCols+` FROM vehiculos`+where+` ORDER BY id_vehiculo LIMIT $3 OFFSET $4`, "%"+p.Query+"%", p.ClientID, p.Size, (p.Number-1)*p.Size)
	if err != nil {
		return out, translate(err)
	}
	defer rows.Close()
	for rows.Next() {
		v, err := scanVehicle(rows)
		if err != nil {
			return out, err
		}
		out.Data = append(out.Data, v)
	}
	return out, translate(rows.Err())
}
func (s *Store) CreateVehicle(ctx context.Context, v domain.VehicleInput, actor string) (domain.Vehicle, error) {
	if err := v.NormalizeValidate(); err != nil {
		return domain.Vehicle{}, err
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return domain.Vehicle{}, err
	}
	defer tx.Rollback(ctx)
	// Preserve the UI's equivalence ABC-123 == ABC123 without changing schema.
	canonical := strings.ReplaceAll(v.Plate, "-", "")
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, canonical); err != nil {
		return domain.Vehicle{}, translate(err)
	}
	var exists bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM vehiculos WHERE regexp_replace(upper(placa),'[[:space:]-]','','g')=$1)`, canonical).Scan(&exists); err != nil {
		return domain.Vehicle{}, translate(err)
	}
	if exists {
		return domain.Vehicle{}, domain.ErrConflict
	}
	out, err := scanVehicle(tx.QueryRow(ctx, `INSERT INTO vehiculos(id_cliente,placa,marca,modelo,anio,kilometraje_actual,observaciones) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING `+vehicleCols, v.ClientID, v.Plate, v.Brand, v.Model, v.Year, v.Mileage, v.Notes))
	if err != nil {
		return out, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO historial_propietarios(id_vehiculo,id_cliente,id_usuario_registro,verificado,observacion) VALUES($1,$2,$3,true,'Registro inicial')`, out.ID, v.ClientID, actor)
	if err != nil {
		return out, translate(err)
	}
	return out, translate(tx.Commit(ctx))
}
func (s *Store) PatchVehicle(ctx context.Context, id string, patch domain.VehiclePatch) (domain.Vehicle, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return domain.Vehicle{}, err
	}
	defer tx.Rollback(ctx)
	old, err := scanVehicle(tx.QueryRow(ctx, `SELECT `+vehicleCols+` FROM vehiculos WHERE id_vehiculo=$1 FOR UPDATE`, id))
	if err != nil {
		return old, err
	}
	v := patch.Apply(old.VehicleInput)
	if err := v.NormalizeValidate(); err != nil {
		return old, err
	}
	before, _ := domain.Mileage(old.Mileage)
	after, _ := domain.Mileage(v.Mileage)
	if after < before {
		return old, domain.Invalid("mileage", "El kilometraje no puede disminuir.")
	}
	out, err := scanVehicle(tx.QueryRow(ctx, `UPDATE vehiculos SET marca=$2,modelo=$3,anio=$4,kilometraje_actual=$5,observaciones=$6 WHERE id_vehiculo=$1 RETURNING `+vehicleCols, id, v.Brand, v.Model, v.Year, v.Mileage, v.Notes))
	if err != nil {
		return out, err
	}
	return out, translate(tx.Commit(ctx))
}
func (s *Store) TransferVehicle(ctx context.Context, id string, v domain.Transfer, actor string) (domain.Vehicle, error) {
	if err := v.Validate(); err != nil {
		return domain.Vehicle{}, err
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return domain.Vehicle{}, err
	}
	defer tx.Rollback(ctx)
	old, err := scanVehicle(tx.QueryRow(ctx, `SELECT `+vehicleCols+` FROM vehiculos WHERE id_vehiculo=$1 FOR UPDATE`, id))
	if err != nil {
		return old, err
	}
	if old.ClientID == v.NewClientID {
		return old, domain.Invalid("newClientId", "El cliente ya es propietario.")
	}
	// clock_timestamp after the lock avoids a waiting transaction's earlier start time.
	var at time.Time
	if err = tx.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&at); err != nil {
		return old, err
	}
	_, err = tx.Exec(ctx, `UPDATE historial_propietarios SET fecha_fin=$2 WHERE id_vehiculo=$1 AND fecha_fin IS NULL`, id, at)
	if err != nil {
		return old, translate(err)
	}
	out, err := scanVehicle(tx.QueryRow(ctx, `UPDATE vehiculos SET id_cliente=$2 WHERE id_vehiculo=$1 RETURNING `+vehicleCols, id, v.NewClientID))
	if err != nil {
		return out, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO historial_propietarios(id_vehiculo,id_cliente,id_usuario_registro,verificado,observacion,fecha_inicio) VALUES($1,$2,$3,true,$4,$5)`, id, v.NewClientID, actor, v.Observation, at)
	if err != nil {
		return out, translate(err)
	}
	return out, translate(tx.Commit(ctx))
}
func (s *Store) DeleteVehicle(ctx context.Context, id string) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = scanVehicle(tx.QueryRow(ctx, `SELECT `+vehicleCols+` FROM vehiculos WHERE id_vehiculo=$1 FOR UPDATE`, id)); err != nil {
		return err
	}
	var histories int
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM historial_propietarios WHERE id_vehiculo=$1`, id).Scan(&histories); err != nil {
		return err
	}
	if histories > 1 {
		return domain.ErrConflict
	}
	// Only the initial ownership record is removed. Any operational FK rolls back both deletions.
	if _, err = tx.Exec(ctx, `DELETE FROM historial_propietarios WHERE id_vehiculo=$1`, id); err != nil {
		return translate(err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM vehiculos WHERE id_vehiculo=$1`, id); err != nil {
		return translate(err)
	}
	return translate(tx.Commit(ctx))
}
func (s *Store) VehicleHistory(ctx context.Context, id string) (any, error) {
	var raw []byte
	err := s.Pool.QueryRow(ctx, `SELECT jsonb_build_object('vehicleId',v.id_vehiculo::text,
 'transfers',coalesce((SELECT jsonb_agg(jsonb_build_object('id',h.id_historial::text,'clientId',h.id_cliente::text,'from',h.fecha_inicio,'until',h.fecha_fin,'verified',h.verificado,'actorId',h.id_usuario_registro::text,'observation',h.observacion) ORDER BY h.fecha_inicio,h.id_historial) FROM historial_propietarios h WHERE h.id_vehiculo=v.id_vehiculo),'[]'::jsonb),
 'receptions',coalesce((SELECT jsonb_agg(jsonb_build_object('id',r.id_recepcion::text,'clientId',r.id_cliente_atencion::text,'createdAt',r.fecha_ingreso,'symptoms',r.sintomas,'diagnosisId',d.id_diagnostico::text,'findings',d.resumen,
 'quotes',coalesce((SELECT jsonb_agg(jsonb_build_object('id',c.id_cotizacion::text,'orderId',o.id_orden::text,'code',o.codigo_ot,'state',o.estado) ORDER BY c.id_cotizacion) FROM cotizaciones c LEFT JOIN ordenes_trabajo o USING(id_cotizacion) WHERE c.id_diagnostico=d.id_diagnostico),'[]'::jsonb)) ORDER BY r.fecha_ingreso DESC,r.id_recepcion DESC) FROM recepciones r LEFT JOIN diagnosticos d USING(id_recepcion) WHERE r.id_vehiculo=v.id_vehiculo),'[]'::jsonb))
 FROM vehiculos v WHERE v.id_vehiculo=$1`, id).Scan(&raw)
	if err != nil {
		return nil, translate(err)
	}
	return json.RawMessage(raw), nil
}
