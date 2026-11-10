package sqlserver

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
)

const providerJSON = `JSON_OBJECT('id':CONVERT(varchar(20),p.id_proveedor),'name':p.nombre,'document':p.documento,
'ruc':p.documento,'phone':p.telefono,'contact':p.contacto)`
const auxilioJSON = `JSON_OBJECT('id':CONVERT(varchar(20),a.id_auxilio),'clientId':CONVERT(varchar(20),a.id_cliente_solicitante),
'clientName':(SELECT nombre FROM dbo.clientes WHERE id_cliente=a.id_cliente_solicitante),
'vehicleId':CONVERT(varchar(20),a.id_vehiculo),'vehiclePlate':(SELECT placa FROM dbo.vehiculos WHERE id_vehiculo=a.id_vehiculo),
'technicianId':CONVERT(varchar(20),a.id_tecnico),'technicianName':(SELECT nombre FROM dbo.usuarios WHERE id_usuario=a.id_tecnico),
'unitId':CONVERT(varchar(20),a.id_unidad_auxilio),'receptionId':CONVERT(varchar(20),a.id_recepcion),'location':a.ubicacion,
'incident':a.tipo_incidente,'description':a.descripcion,'state':a.estado,'createdAt':a.fecha,'createdBy':CONVERT(varchar(20),a.id_usuario_registro))`
const unitJSON = `JSON_OBJECT('id':CONVERT(varchar(20),c.id_unidad_auxilio),'code':c.codigo,'description':c.descripcion,'active':c.activo)`
const linkJSON = `JSON_OBJECT('id':CONVERT(varchar(20),l.id),'providerId':CONVERT(varchar(20),l.id_proveedor),
'sparePartId':CONVERT(varchar(20),l.id_repuesto),'name':r.nombre,'referenceCost':CONVERT(varchar(32),l.costo_referencia),
'referencePrice':CONVERT(varchar(32),r.precio_referencia),'stock':CONVERT(varchar(32),r.stock),'active':r.activo)`

func (s *Store) RelationalItem(ctx context.Context, resource, id string) (any, error) {
	switch resource {
	case "empleados":
		return readJSON(ctx, s.db, "SELECT "+employeeJSON+" FROM empleados e WHERE id_empleado=@p1", id)
	case "usuarios":
		return readJSON(ctx, s.db, "SELECT "+userJSON+" FROM usuarios u WHERE u.id_usuario=@p1", id)
	case "rolePermissions":
		return readJSON(ctx, s.db, "SELECT "+roleJSON+" FROM roles r WHERE r.id_rol=@p1", id)
	case "proveedores":
		return readJSON(ctx, s.db, "SELECT "+providerJSON+" FROM proveedores p WHERE p.id_proveedor=@p1", id)
	case "auxilios":
		return readJSON(ctx, s.db, "SELECT "+auxilioJSON+" FROM auxilios a WHERE a.id_auxilio=@p1", id)
	}
	return nil, domain.ErrNotFound
}
func (s *Store) RelationalList(ctx context.Context, resource, parent string, p domain.Page) (any, error) {
	var expr, from, where, sort string
	args := []any{"%" + p.Query + "%"}
	switch resource {
	case "empleados":
		expr, from, where, sort = employeeJSON, "empleados e", "(e.nombre COLLATE Latin1_General_100_CI_AS_SC LIKE @p1 OR e.tipo_empleado COLLATE Latin1_General_100_CI_AS_SC LIKE @p1)", "e.id_empleado"
	case "usuarios":
		expr, from, where, sort = userJSON, "usuarios u", "(u.nombre COLLATE Latin1_General_100_CI_AS_SC LIKE @p1 OR u.usuario COLLATE Latin1_General_100_CI_AS_SC LIKE @p1 OR coalesce(u.correo,'') COLLATE Latin1_General_100_CI_AS_SC LIKE @p1)", "u.id_usuario"
	case "roles":
		expr, from, where, sort = `JSON_OBJECT('id':CONVERT(varchar(20),r.id_rol),'name':r.nombre_rol,'active':r.activo)`, "roles r", "r.nombre_rol COLLATE Latin1_General_100_CI_AS_SC LIKE @p1", "r.id_rol"
	case "permisos":
		expr, from, where, sort = `JSON_OBJECT('id':CONVERT(varchar(20),p.id_permiso),'code':p.codigo,'description':p.descripcion)`, "permisos p", "p.codigo COLLATE Latin1_General_100_CI_AS_SC LIKE @p1", "p.id_permiso"
	case "proveedores":
		expr, from, where, sort = providerJSON, "proveedores p", "(p.nombre COLLATE Latin1_General_100_CI_AS_SC LIKE @p1 OR p.documento COLLATE Latin1_General_100_CI_AS_SC LIKE @p1)", "p.id_proveedor"
	case "auxilios":
		expr, from, where, sort = auxilioJSON, "auxilios a", "(a.ubicacion COLLATE Latin1_General_100_CI_AS_SC LIKE @p1 OR coalesce(a.descripcion,'') COLLATE Latin1_General_100_CI_AS_SC LIKE @p1)", "a.id_auxilio"
	case "unidades-auxilio":
		expr, from, where, sort = unitJSON, "unidades_auxilio c", "c.activo=1 AND (c.codigo COLLATE Latin1_General_100_CI_AS_SC LIKE @p1 OR coalesce(c.descripcion,'') COLLATE Latin1_General_100_CI_AS_SC LIKE @p1)", "c.id_unidad_auxilio"
	case "providerParts":
		if _, err := s.RelationalItem(ctx, "proveedores", parent); err != nil {
			return nil, err
		}
		expr, from, where, sort = linkJSON, "repuesto_proveedor l JOIN dbo.repuestos r ON r.id_repuesto=l.id_repuesto", "CONVERT(varchar(32),l.id_proveedor)=@p1", "l.id"
		args = []any{parent}
	default:
		return nil, domain.ErrNotFound
	}
	return s.jsonPage(ctx, p, expr, from, where, sort, args...)
}
func (s *Store) SaveProvider(ctx context.Context, id string, v domain.ProviderInput) (any, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	name, document, phone, contact := "", "", "", ""
	if id != "" {
		if err = tx.QueryRowContext(ctx, `SELECT nombre,documento,coalesce(telefono,''),coalesce(contacto,'') FROM proveedores WITH (UPDLOCK,HOLDLOCK) WHERE id_proveedor=@p1`, id).Scan(&name, &document, &phone, &contact); err != nil {
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
		err = tx.QueryRowContext(ctx, `INSERT INTO proveedores(nombre,documento,telefono,contacto) OUTPUT CONVERT(varchar(20),inserted.id_proveedor) VALUES(@p1,@p2,nullif(@p3,''),nullif(@p4,''))`, name, document, phone, contact).Scan(&id)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE proveedores SET nombre=@p2,documento=@p3,telefono=nullif(@p4,''),contacto=nullif(@p5,'') WHERE id_proveedor=@p1`, id, name, document, phone, contact)
	}
	if err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+providerJSON+" FROM proveedores p WHERE p.id_proveedor=@p1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit())
}
func (s *Store) DeleteProvider(ctx context.Context, id string) error {
	tag, err := s.db.ExecContext(ctx, `DELETE FROM proveedores WHERE id_proveedor=@p1`, id)
	if err != nil {
		return translate(err)
	}
	affected, err := tag.RowsAffected()
	if err != nil {
		return translate(err)
	}
	if affected == 0 {
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
		return readJSON(ctx, s.db, `DECLARE @id bigint = NEXT VALUE FOR dbo.unidades_auxilio_id_unidad_auxilio_seq;
INSERT dbo.unidades_auxilio(id_unidad_auxilio,codigo,descripcion,activo) VALUES(@id,@p1,@p2,@p3);
SELECT `+unitJSON+` FROM dbo.unidades_auxilio c WHERE c.id_unidad_auxilio=@id;`, v.Code, v.Description, active)
	case "servicios":
		return readJSON(ctx, s.db, `DECLARE @id bigint = NEXT VALUE FOR dbo.servicios_id_servicio_seq;
INSERT dbo.servicios(id_servicio,nombre,precio_referencia,activo) VALUES(@id,@p1,@p2,@p3);
SELECT JSON_OBJECT('id':CONVERT(varchar(20),id_servicio),'name':nombre,'referencePrice':CONVERT(varchar(32),precio_referencia),'active':activo)
FROM dbo.servicios WHERE id_servicio=@id;`, v.Name, v.ReferencePrice, active)
	case "repuestos":
		return readJSON(ctx, s.db, `DECLARE @id bigint = NEXT VALUE FOR dbo.repuestos_id_repuesto_seq;
INSERT dbo.repuestos(id_repuesto,nombre,precio_referencia,stock,activo) VALUES(@id,@p1,@p2,@p3,@p4);
SELECT JSON_OBJECT('id':CONVERT(varchar(20),id_repuesto),'name':nombre,'referencePrice':CONVERT(varchar(32),precio_referencia),
'stock':CONVERT(varchar(32),stock),'active':activo) FROM dbo.repuestos WHERE id_repuesto=@id;`, v.Name, v.ReferencePrice, v.Stock, active)
	}
	return nil, domain.ErrNotFound
}
func (s *Store) SaveAuxilio(ctx context.Context, id string, v domain.AuxilioInput, u domain.User) (any, error) {
	create := id == ""
	if err := v.Validate(create); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if !create {
		var technician, unit, state string
		err = tx.QueryRowContext(ctx, `SELECT coalesce(CONVERT(varchar(32),id_tecnico),''),coalesce(CONVERT(varchar(32),id_unidad_auxilio),''),estado FROM auxilios WITH (UPDLOCK,HOLDLOCK) WHERE id_auxilio=@p1`, id).Scan(&technician, &unit, &state)
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
	if err = tx.QueryRowContext(ctx, `SELECT CONVERT(varchar(32),id_unidad_auxilio) FROM unidades_auxilio WITH (HOLDLOCK) WHERE id_unidad_auxilio=@p1 AND activo=1`, v.UnitID).Scan(&unit); err != nil {
		return nil, translate(err)
	}
	if create {
		var owner string
		if err = tx.QueryRowContext(ctx, `SELECT CONVERT(varchar(32),id_cliente) FROM vehiculos WITH (HOLDLOCK) WHERE id_vehiculo=@p1`, v.VehicleID).Scan(&owner); err != nil {
			return nil, translate(err)
		}
		if owner != v.ClientID {
			return nil, domain.Invalid("clientId", "Vehículo ajeno al cliente.")
		}
		if v.ReceptionID != "" {
			var matches bool
			err = tx.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN id_vehiculo=@p2 AND id_cliente_atencion=@p3 THEN 1 ELSE 0 END) FROM dbo.recepciones WHERE id_recepcion=@p1`, v.ReceptionID, v.VehicleID, v.ClientID).Scan(&matches)
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
		err = tx.QueryRowContext(ctx, `INSERT INTO auxilios(id_vehiculo,id_cliente_solicitante,id_usuario_registro,id_tecnico,id_unidad_auxilio,id_recepcion,ubicacion,tipo_incidente,descripcion) OUTPUT CONVERT(varchar(20),inserted.id_auxilio) VALUES(@p1,@p2,@p3,@p4,@p5,nullif(@p6,N''),@p7,@p8,@p9)`, v.VehicleID, v.ClientID, u.ID, v.TechnicianID, v.UnitID, v.ReceptionID, v.Location, v.Incident, v.Description).Scan(&id)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE auxilios SET id_tecnico=@p2,id_unidad_auxilio=@p3,estado=@p4 WHERE id_auxilio=@p1`, id, v.TechnicianID, v.UnitID, v.State)
	}
	if err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+auxilioJSON+" FROM auxilios a WHERE a.id_auxilio=@p1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit())
}
func (s *Store) LinkProviderPart(ctx context.Context, provider, part string, v domain.PartLinkInput, remove bool) (any, error) {
	if remove {
		tag, err := s.db.ExecContext(ctx, `DELETE FROM repuesto_proveedor WHERE id_proveedor=@p1 AND id_repuesto=@p2`, provider, part)
		if err != nil {
			return nil, translate(err)
		}
		affected, err := tag.RowsAffected()
		if err != nil {
			return nil, translate(err)
		}
		if affected == 0 {
			return nil, domain.ErrNotFound
		}
		return nil, nil
	}
	n, err := domain.Money(v.ReferenceCost, "referenceCost")
	if err != nil {
		return nil, err
	}
	v.ReferenceCost = domain.FormatMileage(n)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, translate(err)
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `UPDATE dbo.repuesto_proveedor WITH (UPDLOCK,HOLDLOCK) SET costo_referencia=@p3
WHERE id_proveedor=@p1 AND id_repuesto=@p2;
IF @@ROWCOUNT=0
    INSERT dbo.repuesto_proveedor(id_proveedor,id_repuesto,costo_referencia) VALUES(@p1,@p2,@p3);`, provider, part, v.ReferenceCost)
	if err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+linkJSON+` FROM dbo.repuesto_proveedor l JOIN dbo.repuestos r ON r.id_repuesto=l.id_repuesto
WHERE l.id_proveedor=@p1 AND l.id_repuesto=@p2`, provider, part)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit())
}
