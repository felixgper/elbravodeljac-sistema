package postgres

import (
	"context"
	"encoding/json"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/jackc/pgx/v5"
)

const providerJSON = `jsonb_build_object('id',p.id_proveedor::text,'name',p.nombre,'document',p.documento,'ruc',p.documento,'phone',p.telefono,'contact',p.contacto)`
const auxilioJSON = `jsonb_build_object('id',a.id_auxilio::text,'clientId',a.id_cliente_solicitante::text,'clientName',(SELECT c.nombre FROM clientes c WHERE c.id_cliente=a.id_cliente_solicitante),'vehicleId',a.id_vehiculo::text,'vehiclePlate',(SELECT v.placa FROM vehiculos v WHERE v.id_vehiculo=a.id_vehiculo),'technicianId',a.id_tecnico::text,'technicianName',(SELECT u.nombre FROM usuarios u WHERE u.id_usuario=a.id_tecnico),'unitId',a.id_unidad_auxilio::text,'receptionId',a.id_recepcion::text,'location',a.ubicacion,'incident',a.tipo_incidente,'description',a.descripcion,'state',a.estado,'createdAt',a.fecha,'createdBy',a.id_usuario_registro::text)`
const unitJSON = `jsonb_build_object('id',c.id_unidad_auxilio::text,'code',c.codigo,'description',c.descripcion,'active',c.activo)`
const linkJSON = `jsonb_build_object('id',l.id::text,'providerId',l.id_proveedor::text,'sparePartId',l.id_repuesto::text,'name',r.nombre,'referenceCost',l.costo_referencia::text,'referencePrice',r.precio_referencia::text,'stock',r.stock::text,'active',r.activo)`

func (s *Store) RelationalItem(ctx context.Context, resource, id string) (any, error) {
	switch resource {
	case "empleados":
		return readJSON(ctx, s.Pool, "SELECT "+employeeJSON+" FROM empleados e WHERE id_empleado=$1", id)
	case "usuarios":
		return readJSON(ctx, s.Pool, "SELECT "+userJSON+" FROM usuarios u WHERE u.id_usuario=$1", id)
	case "rolePermissions":
		return readJSON(ctx, s.Pool, "SELECT "+roleJSON+" FROM roles r WHERE r.id_rol=$1", id)
	case "proveedores":
		return readJSON(ctx, s.Pool, "SELECT "+providerJSON+" FROM proveedores p WHERE p.id_proveedor=$1", id)
	case "auxilios":
		return readJSON(ctx, s.Pool, "SELECT "+auxilioJSON+" FROM auxilios a WHERE a.id_auxilio=$1", id)
	}
	return nil, domain.ErrNotFound
}
func (s *Store) RelationalList(ctx context.Context, resource, parent string, p domain.Page) (any, error) {
	var expr, from, where, sort string
	args := []any{"%" + p.Query + "%"}
	switch resource {
	case "empleados":
		expr, from, where, sort = employeeJSON, "empleados e", "(e.nombre ILIKE $1 OR e.tipo_empleado ILIKE $1)", "e.id_empleado"
	case "usuarios":
		expr, from, where, sort = userJSON, "usuarios u", "(u.nombre ILIKE $1 OR u.usuario ILIKE $1 OR coalesce(u.correo,'') ILIKE $1)", "u.id_usuario"
	case "roles":
		expr, from, where, sort = `jsonb_build_object('id',r.id_rol::text,'name',r.nombre_rol,'active',r.activo)`, "roles r", "r.nombre_rol ILIKE $1", "r.id_rol"
	case "permisos":
		expr, from, where, sort = `jsonb_build_object('id',p.id_permiso::text,'code',p.codigo,'description',p.descripcion)`, "permisos p", "p.codigo ILIKE $1", "p.id_permiso"
	case "proveedores":
		expr, from, where, sort = providerJSON, "proveedores p", "(p.nombre ILIKE $1 OR p.documento ILIKE $1)", "p.id_proveedor"
	case "auxilios":
		expr, from, where, sort = auxilioJSON, "auxilios a", "(a.ubicacion ILIKE $1 OR coalesce(a.descripcion,'') ILIKE $1)", "a.id_auxilio"
	case "unidades-auxilio":
		expr, from, where, sort = unitJSON, "unidades_auxilio c", "c.activo AND (c.codigo ILIKE $1 OR coalesce(c.descripcion,'') ILIKE $1)", "c.id_unidad_auxilio"
	case "providerParts":
		if _, err := s.RelationalItem(ctx, "proveedores", parent); err != nil {
			return nil, err
		}
		expr, from, where, sort = linkJSON, "repuesto_proveedor l JOIN repuestos r USING(id_repuesto)", "l.id_proveedor::text=$1", "l.id"
		args = []any{parent}
	default:
		return nil, domain.ErrNotFound
	}
	out := domain.List[json.RawMessage]{Data: []json.RawMessage{}, Meta: domain.Meta{Page: p.Number, PageSize: p.Size}}
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if err = tx.QueryRow(ctx, "SELECT count(*) FROM "+from+" WHERE "+where, args...).Scan(&out.Meta.Total); err != nil {
		return nil, translate(err)
	}
	query := "SELECT " + expr + " FROM " + from + " WHERE " + where + " ORDER BY " + sort + " LIMIT $2 OFFSET $3"
	args = append(args, p.Size, (p.Number-1)*p.Size)
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, translate(err)
	}
	for rows.Next() {
		var raw []byte
		if err = rows.Scan(&raw); err != nil {
			rows.Close()
			return nil, err
		}
		out.Data = append(out.Data, json.RawMessage(raw))
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
func (s *Store) SaveProvider(ctx context.Context, id string, v domain.ProviderInput) (any, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	name, document, phone, contact := "", "", "", ""
	if id != "" {
		if err = tx.QueryRow(ctx, `SELECT nombre,documento,coalesce(telefono,''),coalesce(contacto,'') FROM proveedores WHERE id_proveedor=$1 FOR UPDATE`, id).Scan(&name, &document, &phone, &contact); err != nil {
			return nil, translate(err)
		}
	}
	if v.Name != nil {
		name = *v.Name
	}
	if v.Document != nil {
		document = *v.Document
	}
	if v.RUC != nil {
		document = *v.RUC
	}
	if v.Phone != nil {
		phone = *v.Phone
	}
	if v.Contact != nil {
		contact = *v.Contact
	}
	merged := domain.ProviderInput{Name: &name, Document: &document, Phone: &phone, Contact: &contact}
	if err = merged.Validate(); err != nil {
		return nil, err
	}
	if id == "" {
		err = tx.QueryRow(ctx, `INSERT INTO proveedores(nombre,documento,telefono,contacto) VALUES($1,$2,nullif($3,''),nullif($4,'')) RETURNING id_proveedor::text`, name, document, phone, contact).Scan(&id)
	} else {
		_, err = tx.Exec(ctx, `UPDATE proveedores SET nombre=$2,documento=$3,telefono=nullif($4,''),contacto=nullif($5,'') WHERE id_proveedor=$1`, id, name, document, phone, contact)
	}
	if err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+providerJSON+" FROM proveedores p WHERE p.id_proveedor=$1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
func (s *Store) DeleteProvider(ctx context.Context, id string) error {
	tag, err := s.Pool.Exec(ctx, `DELETE FROM proveedores WHERE id_proveedor=$1`, id)
	if err != nil {
		return translate(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
func (s *Store) SaveCatalog(ctx context.Context, resource string, v domain.CatalogInput) (any, error) {
	if err := v.Validate(resource); err != nil {
		return nil, err
	}
	active := true
	if v.Active != nil {
		active = *v.Active
	}
	switch resource {
	case "unidades-auxilio":
		return readJSON(ctx, s.Pool, `WITH inserted AS (INSERT INTO unidades_auxilio(codigo,descripcion,activo) VALUES($1,$2,$3) RETURNING *) SELECT `+unitJSON+` FROM inserted c`, v.Code, v.Description, active)
	case "servicios":
		return readJSON(ctx, s.Pool, `WITH inserted AS (INSERT INTO servicios(nombre,precio_referencia,activo) VALUES($1,$2,$3) RETURNING *) SELECT jsonb_build_object('id',id_servicio::text,'name',nombre,'referencePrice',precio_referencia::text,'active',activo) FROM inserted`, v.Name, v.ReferencePrice, active)
	case "repuestos":
		return readJSON(ctx, s.Pool, `WITH inserted AS (INSERT INTO repuestos(nombre,precio_referencia,stock,activo) VALUES($1,$2,$3,$4) RETURNING *) SELECT jsonb_build_object('id',id_repuesto::text,'name',nombre,'referencePrice',precio_referencia::text,'stock',stock::text,'active',activo) FROM inserted`, v.Name, v.ReferencePrice, v.Stock, active)
	}
	return nil, domain.ErrNotFound
}
func (s *Store) SaveAuxilio(ctx context.Context, id string, v domain.AuxilioInput, u domain.User) (any, error) {
	create := id == ""
	if err := v.Validate(create); err != nil {
		return nil, err
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if !create {
		var technician, unit, state string
		err = tx.QueryRow(ctx, `SELECT coalesce(id_tecnico::text,''),coalesce(id_unidad_auxilio::text,''),estado FROM auxilios WHERE id_auxilio=$1 FOR UPDATE`, id).Scan(&technician, &unit, &state)
		if err != nil {
			return nil, translate(err)
		}
		if state == "finalizado" || state == "cancelado" {
			return nil, domain.ErrConflict
		}
		if v.TechnicianID == "" {
			v.TechnicianID = technician
		}
		if v.UnitID == "" {
			v.UnitID = unit
		}
		if v.State == "" {
			v.State = state
		}
	}
	if err = technicalUser(ctx, tx, v.TechnicianID); err != nil {
		return nil, err
	}
	var unit string
	if err = tx.QueryRow(ctx, `SELECT id_unidad_auxilio::text FROM unidades_auxilio WHERE id_unidad_auxilio=$1 AND activo FOR SHARE`, v.UnitID).Scan(&unit); err != nil {
		return nil, translate(err)
	}
	if create {
		var owner string
		if err = tx.QueryRow(ctx, `SELECT id_cliente::text FROM vehiculos WHERE id_vehiculo=$1 FOR SHARE`, v.VehicleID).Scan(&owner); err != nil {
			return nil, translate(err)
		}
		if owner != v.ClientID {
			return nil, domain.Invalid("clientId", "Vehículo ajeno al cliente.")
		}
		if v.ReceptionID != "" {
			var matches bool
			err = tx.QueryRow(ctx, `SELECT id_vehiculo::text=$2 AND id_cliente_atencion::text=$3 FROM recepciones WHERE id_recepcion=$1`, v.ReceptionID, v.VehicleID, v.ClientID).Scan(&matches)
			if err != nil {
				return nil, translate(err)
			}
			if !matches {
				return nil, domain.Invalid("receptionId", "Recepción incompatible con cliente/vehículo.")
			}
		}
		if v.State != "" && v.State != "pendiente" {
			return nil, domain.Invalid("state", "El auxilio se registra pendiente.")
		}
		err = tx.QueryRow(ctx, `INSERT INTO auxilios(id_vehiculo,id_cliente_solicitante,id_usuario_registro,id_tecnico,id_unidad_auxilio,id_recepcion,ubicacion,tipo_incidente,descripcion) VALUES($1,$2,$3,$4,$5,nullif($6,'')::bigint,$7,$8,$9) RETURNING id_auxilio::text`, v.VehicleID, v.ClientID, u.ID, v.TechnicianID, v.UnitID, v.ReceptionID, v.Location, v.Incident, v.Description).Scan(&id)
	} else {
		_, err = tx.Exec(ctx, `UPDATE auxilios SET id_tecnico=$2,id_unidad_auxilio=$3,estado=$4 WHERE id_auxilio=$1`, id, v.TechnicianID, v.UnitID, v.State)
	}
	if err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+auxilioJSON+" FROM auxilios a WHERE a.id_auxilio=$1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
func (s *Store) LinkProviderPart(ctx context.Context, provider, part string, v domain.PartLinkInput, remove bool) (any, error) {
	if remove {
		tag, err := s.Pool.Exec(ctx, `DELETE FROM repuesto_proveedor WHERE id_proveedor=$1 AND id_repuesto=$2`, provider, part)
		if err != nil {
			return nil, translate(err)
		}
		if tag.RowsAffected() == 0 {
			return nil, domain.ErrNotFound
		}
		return nil, nil
	}
	n, err := domain.Money(v.ReferenceCost, "referenceCost")
	if err != nil {
		return nil, err
	}
	v.ReferenceCost = domain.FormatMileage(n)
	return readJSON(ctx, s.Pool, `WITH changed AS (INSERT INTO repuesto_proveedor(id_proveedor,id_repuesto,costo_referencia) VALUES($1,$2,$3) ON CONFLICT(id_repuesto,id_proveedor) DO UPDATE SET costo_referencia=excluded.costo_referencia RETURNING *) SELECT `+linkJSON+` FROM changed l JOIN repuestos r USING(id_repuesto)`, provider, part, v.ReferenceCost)
}
