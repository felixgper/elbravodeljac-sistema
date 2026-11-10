package sqlserver

import (
	"context"
	"encoding/json"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"strings"
	"time"
)

const clientCols = `CONVERT(varchar(32),id_cliente),nombre,tipo_documento,numero_documento,coalesce(telefono,''),coalesce(direccion,''),coalesce(correo,'')`

func scanClient(row interface{ Scan(...any) error }) (domain.Client, error) {
	var v domain.Client
	err := row.Scan(&v.ID, &v.Name, &v.DocumentType, &v.DocumentNumber, &v.Phone, &v.Address, &v.Email)
	return v, translate(err)
}
func (s *Store) Client(ctx context.Context, id string) (domain.Client, error) {
	return scanClient(s.db.QueryRowContext(ctx, `SELECT `+clientCols+` FROM dbo.clientes WHERE id_cliente=@p1`, id))
}
func (s *Store) Clients(ctx context.Context, p domain.Page) (domain.List[domain.Client], error) {
	out := domain.List[domain.Client]{Data: []domain.Client{}, Meta: domain.Meta{Page: p.Number, PageSize: p.Size}}
	where := ` WHERE CONCAT(nombre,N' ',numero_documento) COLLATE Latin1_General_100_CI_AI_SC LIKE @p1 OR EXISTS(SELECT 1 FROM dbo.vehiculos v WHERE v.id_cliente=clientes.id_cliente AND v.placa COLLATE Latin1_General_100_CI_AS_SC LIKE @p1)`
	err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM dbo.clientes`+where, "%"+p.Query+"%").Scan(&out.Meta.Total)
	if err != nil {
		return out, translate(err)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+clientCols+`,(SELECT CONVERT(varchar(20),v.id_vehiculo) AS id, CONVERT(varchar(20),v.id_cliente) AS clientId,
    v.placa AS plate,v.marca AS brand,v.modelo AS model FROM dbo.vehiculos v WHERE v.id_cliente=clientes.id_cliente
    ORDER BY v.id_vehiculo FOR JSON PATH, INCLUDE_NULL_VALUES) FROM dbo.clientes`+where+` ORDER BY id_cliente OFFSET @p3 ROWS FETCH NEXT @p2 ROWS ONLY`, "%"+p.Query+"%", p.Size, (p.Number-1)*p.Size)
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
	return scanClient(s.db.QueryRowContext(ctx, `
DECLARE @id bigint = NEXT VALUE FOR dbo.clientes_id_cliente_seq;
DECLARE @tipo bigint = (SELECT id_tipo_documento FROM dbo.tipos_documento WHERE codigo=dbo.jac_documento_codigo(@p2) AND activo=1);
IF @tipo IS NULL THROW 51002, N'Tipo de documento inexistente o inactivo.', 1;
INSERT dbo.clientes(id_cliente,nombre,tipo_documento,numero_documento,telefono,direccion,correo,id_tipo_documento)
VALUES(@id,@p1,@p2,@p3,@p4,@p5,NULLIF(@p6,N''),@tipo);
SELECT `+clientCols+` FROM dbo.clientes WHERE id_cliente=@id;`, v.Name, v.DocumentType, v.DocumentNumber, v.Phone, v.Address, v.Email))
}
func (s *Store) PatchClient(ctx context.Context, id string, patch map[string]string) (domain.Client, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Client{}, err
	}
	defer tx.Rollback()
	v, err := scanClient(tx.QueryRowContext(ctx, `SELECT `+clientCols+` FROM dbo.clientes WITH (UPDLOCK,HOLDLOCK) WHERE id_cliente=@p1`, id))
	if err != nil {
		return v, err
	}
	fields := map[string]*string{"name": &v.Name, "documentType": &v.DocumentType, "documentNumber": &v.DocumentNumber, "phone": &v.Phone, "address": &v.Address, "email": &v.Email}
	originalKind, originalNumber := v.DocumentType, v.DocumentNumber
	for k, value := range patch {
		ptr, ok := fields[k]
		if !ok {
			return v, domain.Invalid(k, "Campo no editable.")
		}
		*ptr = value
	}
	validate := v.ClientInput.Validate
	if v.DocumentType == originalKind && v.DocumentNumber == originalNumber {
		validate = v.ClientInput.ValidateRetainedIdentity
	}
	if err := validate(); err != nil {
		return v, err
	}
	v, err = scanClient(tx.QueryRowContext(ctx, `UPDATE dbo.clientes SET nombre=@p2,tipo_documento=@p3,numero_documento=@p4,telefono=@p5,direccion=@p6,correo=nullif(@p7,'') WHERE id_cliente=@p1;
SELECT `+clientCols+` FROM dbo.clientes WHERE id_cliente=@p1;`, id, v.Name, v.DocumentType, v.DocumentNumber, v.Phone, v.Address, v.Email))
	if err != nil {
		return v, err
	}
	return v, translate(tx.Commit())
}
func (s *Store) DeleteClient(ctx context.Context, id string) error {
	tag, err := s.db.ExecContext(ctx, `DELETE FROM dbo.clientes WHERE id_cliente=@p1`, id)
	if err != nil {
		return translate(err)
	}
	n, err := tag.RowsAffected()
	if err != nil {
		return translate(err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

const contactCols = `CONVERT(varchar(32),id_contacto),CONVERT(varchar(32),id_cliente),activo,nombre,tipo_documento,numero_documento,coalesce(telefono,''),coalesce(relacion,'')`

func scanContact(row interface{ Scan(...any) error }) (domain.Contact, error) {
	var c domain.Contact
	err := row.Scan(&c.ID, &c.ClientID, &c.Active, &c.Name, &c.DocumentType, &c.DocumentNumber, &c.Phone, &c.Relation)
	return c, translate(err)
}
func (s *Store) Contacts(ctx context.Context, id string) ([]domain.Contact, error) {
	if _, err := s.Client(ctx, id); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+contactCols+` FROM dbo.contactos_cliente WHERE id_cliente=@p1 ORDER BY id_contacto`, id)
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
	return scanContact(s.db.QueryRowContext(ctx, `
DECLARE @id bigint = NEXT VALUE FOR dbo.contactos_cliente_id_contacto_seq;
DECLARE @tipo bigint = (SELECT id_tipo_documento FROM dbo.tipos_documento WHERE codigo=dbo.jac_documento_codigo(@p3) AND activo=1);
IF @tipo IS NULL THROW 51002, N'Tipo de documento inexistente o inactivo.', 1;
INSERT dbo.contactos_cliente(id_contacto,id_cliente,nombre,tipo_documento,numero_documento,telefono,relacion,id_tipo_documento)
VALUES(@id,@p1,@p2,@p3,@p4,@p5,@p6,@tipo);
SELECT `+contactCols+` FROM dbo.contactos_cliente WHERE id_contacto=@id;`, id, v.Name, v.DocumentType, v.DocumentNumber, v.Phone, v.Relation))
}

const vehicleCols = `CONVERT(varchar(32),id_vehiculo),CONVERT(varchar(32),id_cliente),placa,coalesce(marca,''),coalesce(modelo,''),coalesce(anio,0),CONVERT(varchar(32),coalesce(kilometraje_actual,0)),coalesce(observaciones,'')`

func scanVehicle(row interface{ Scan(...any) error }) (domain.Vehicle, error) {
	var v domain.Vehicle
	err := row.Scan(&v.ID, &v.ClientID, &v.Plate, &v.Brand, &v.Model, &v.Year, &v.Mileage, &v.Notes)
	return v, translate(err)
}
func (s *Store) Vehicle(ctx context.Context, id string) (domain.Vehicle, error) {
	return scanVehicle(s.db.QueryRowContext(ctx, `SELECT `+vehicleCols+` FROM dbo.vehiculos WHERE id_vehiculo=@p1`, id))
}
func (s *Store) Vehicles(ctx context.Context, p domain.Page) (domain.List[domain.Vehicle], error) {
	out := domain.List[domain.Vehicle]{Data: []domain.Vehicle{}, Meta: domain.Meta{Page: p.Number, PageSize: p.Size}}
	where := ` WHERE (placa COLLATE Latin1_General_100_CI_AS_SC LIKE @p1 OR modelo COLLATE Latin1_General_100_CI_AS_SC LIKE @p1 OR EXISTS(SELECT 1 FROM dbo.clientes c WHERE c.id_cliente=vehiculos.id_cliente AND c.nombre COLLATE Latin1_General_100_CI_AS_SC LIKE @p1)) AND (@p2='' OR CONVERT(varchar(32),id_cliente)=@p2)`
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM dbo.vehiculos`+where, "%"+p.Query+"%", p.ClientID).Scan(&out.Meta.Total); err != nil {
		return out, translate(err)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+vehicleCols+` FROM dbo.vehiculos`+where+` ORDER BY id_vehiculo OFFSET @p4 ROWS FETCH NEXT @p3 ROWS ONLY`, "%"+p.Query+"%", p.ClientID, p.Size, (p.Number-1)*p.Size)
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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Vehicle{}, err
	}
	defer tx.Rollback()
	// Preserve the UI's equivalence ABC-123 == ABC123 without changing schema.
	canonical := strings.ReplaceAll(v.Plate, "-", "")
	if err = acquireLock(ctx, tx, "jac:plate:"+canonical); err != nil {
		return domain.Vehicle{}, translate(err)
	}
	var exists bool
	if err = tx.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN EXISTS(SELECT 1 FROM dbo.vehiculos WHERE REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(UPPER(placa),N'-',N''),N' ',N''),NCHAR(9),N''),NCHAR(10),N''),NCHAR(13),N'')=@p1) THEN 1 ELSE 0 END)`, canonical).Scan(&exists); err != nil {
		return domain.Vehicle{}, translate(err)
	}
	if exists {
		return domain.Vehicle{}, domain.ErrConflict
	}
	if err = tx.QueryRowContext(ctx, `SELECT coalesce((SELECT nombre FROM dbo.marcas_vehiculos WHERE nombre COLLATE Latin1_General_100_CI_AI_SC=@p1), @p1)`, v.Brand).Scan(&v.Brand); err != nil {
		return domain.Vehicle{}, translate(err)
	}
	out, err := scanVehicle(tx.QueryRowContext(ctx, `
DECLARE @id bigint = NEXT VALUE FOR dbo.vehiculos_id_vehiculo_seq;
INSERT dbo.vehiculos(id_vehiculo,id_cliente,placa,marca,modelo,anio,kilometraje_actual,observaciones)
VALUES(@id,@p1,@p2,@p3,@p4,@p5,@p6,@p7);
SELECT `+vehicleCols+` FROM dbo.vehiculos WHERE id_vehiculo=@id;`, v.ClientID, v.Plate, v.Brand, v.Model, v.Year, v.Mileage, v.Notes))
	if err != nil {
		return out, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO dbo.historial_propietarios(id_vehiculo,id_cliente,id_usuario_registro,verificado,observacion) VALUES(@p1,@p2,@p3,1,'Registro inicial')`, out.ID, v.ClientID, actor)
	if err != nil {
		return out, translate(err)
	}
	return out, translate(tx.Commit())
}
func (s *Store) PatchVehicle(ctx context.Context, id string, patch domain.VehiclePatch) (domain.Vehicle, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Vehicle{}, err
	}
	defer tx.Rollback()
	old, err := scanVehicle(tx.QueryRowContext(ctx, `SELECT `+vehicleCols+` FROM dbo.vehiculos WITH (UPDLOCK,HOLDLOCK) WHERE id_vehiculo=@p1`, id))
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
	if patch.Brand != nil {
		if err = tx.QueryRowContext(ctx, `SELECT coalesce((SELECT nombre FROM dbo.marcas_vehiculos WHERE nombre COLLATE Latin1_General_100_CI_AI_SC=@p1), @p1)`, v.Brand).Scan(&v.Brand); err != nil {
			return old, translate(err)
		}
	}
	out, err := scanVehicle(tx.QueryRowContext(ctx, `UPDATE dbo.vehiculos SET marca=@p2,modelo=@p3,anio=@p4,kilometraje_actual=@p5,observaciones=@p6 WHERE id_vehiculo=@p1;
SELECT `+vehicleCols+` FROM dbo.vehiculos WHERE id_vehiculo=@p1;`, id, v.Brand, v.Model, v.Year, v.Mileage, v.Notes))
	if err != nil {
		return out, err
	}
	return out, translate(tx.Commit())
}
func (s *Store) TransferVehicle(ctx context.Context, id string, v domain.Transfer, actor string) (domain.Vehicle, error) {
	if err := v.Validate(); err != nil {
		return domain.Vehicle{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Vehicle{}, err
	}
	defer tx.Rollback()
	old, err := scanVehicle(tx.QueryRowContext(ctx, `SELECT `+vehicleCols+` FROM dbo.vehiculos WITH (UPDLOCK,HOLDLOCK) WHERE id_vehiculo=@p1`, id))
	if err != nil {
		return old, err
	}
	if old.ClientID == v.NewClientID {
		return old, domain.Invalid("newClientId", "El cliente ya es propietario.")
	}
	// clock_timestamp after the lock avoids a waiting transaction's earlier start time.
	var at time.Time
	if err = tx.QueryRowContext(ctx, `SELECT TODATETIMEOFFSET(SYSUTCDATETIME(), '+00:00')`).Scan(&at); err != nil {
		return old, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE dbo.historial_propietarios SET fecha_fin=@p2 WHERE id_vehiculo=@p1 AND fecha_fin IS NULL`, id, at)
	if err != nil {
		return old, translate(err)
	}
	out, err := scanVehicle(tx.QueryRowContext(ctx, `UPDATE dbo.vehiculos SET id_cliente=@p2 WHERE id_vehiculo=@p1;
SELECT `+vehicleCols+` FROM dbo.vehiculos WHERE id_vehiculo=@p1;`, id, v.NewClientID))
	if err != nil {
		return out, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO dbo.historial_propietarios(id_vehiculo,id_cliente,id_usuario_registro,verificado,observacion,fecha_inicio) VALUES(@p1,@p2,@p3,1,@p4,@p5)`, id, v.NewClientID, actor, v.Observation, at)
	if err != nil {
		return out, translate(err)
	}
	return out, translate(tx.Commit())
}
func (s *Store) DeleteVehicle(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = scanVehicle(tx.QueryRowContext(ctx, `SELECT `+vehicleCols+` FROM dbo.vehiculos WITH (UPDLOCK,HOLDLOCK) WHERE id_vehiculo=@p1`, id)); err != nil {
		return err
	}
	var histories int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM dbo.historial_propietarios WHERE id_vehiculo=@p1`, id).Scan(&histories); err != nil {
		return err
	}
	if histories > 1 {
		return domain.ErrConflict
	}
	// Only the initial ownership record is removed. Any operational FK rolls back both deletions.
	if _, err = tx.ExecContext(ctx, `DELETE FROM dbo.historial_propietarios WHERE id_vehiculo=@p1`, id); err != nil {
		return translate(err)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM dbo.vehiculos WHERE id_vehiculo=@p1`, id); err != nil {
		return translate(err)
	}
	return translate(tx.Commit())
}
