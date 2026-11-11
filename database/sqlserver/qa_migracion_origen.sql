-- Solo para la base PostgreSQL efímera de GitHub Actions.
\set ON_ERROR_STOP on
DO $$ BEGIN
    IF current_database() <> 'elbravodeljac_pg_migration' THEN RAISE EXCEPTION 'Base QA incorrecta'; END IF;
END $$;
BEGIN;
INSERT INTO tipos_documento(id_tipo_documento,codigo,nombre) VALUES (31,'DNI','Documento QA');
INSERT INTO empleados(id_empleado,nombre,tipo_empleado) VALUES (21,'Mecánica QA 🚘','Mecánica');
INSERT INTO usuarios(id_usuario,nombre,usuario,password_hash,id_empleado)
VALUES (101,'Mecánica QA 🚘','operador_qa','fixture-only',21),
       (102,'Cuenta histórica','historica_qa','fixture-only',NULL);
INSERT INTO roles(id_rol,nombre_rol) VALUES (33,'Gerente');
INSERT INTO permisos(id_permiso,codigo) VALUES (34,'pagos.ver'), (35,'cotizaciones.ver');
INSERT INTO usuario_roles VALUES (101,33);
INSERT INTO rol_permisos VALUES (33,34),(33,35);
INSERT INTO clientes(id_cliente,tipo_documento,numero_documento,nombre,direccion)
VALUES (41,'DNI','00000000','Cliente QA áé 🚘','Texto con espacios  '),
       (42,'DNI','00000001','Titular anterior',NULL);
INSERT INTO contactos_cliente(id_contacto,id_cliente,nombre,tipo_documento,numero_documento,relacion)
VALUES (43,41,'Contacto QA','DNI','00000002','');
INSERT INTO vehiculos(id_vehiculo,id_cliente,placa,marca,modelo,anio,kilometraje_actual)
VALUES (51,41,'QA-051','Citroën','Prueba',2024,100.25);
INSERT INTO historial_propietarios(id_historial,id_vehiculo,id_cliente,fecha_inicio,fecha_fin,id_usuario_registro)
VALUES (52,51,42,'2020-02-28 16:30:12.123456-05','2024-03-01 16:30:12.123456-05',102),
       (53,51,41,'2024-03-01 16:30:12.123456-05',NULL,101);
INSERT INTO recepciones(id_recepcion,id_vehiculo,id_cliente_atencion,id_recepcionista,tipo_atencion,
    id_mecanico_diagnostico,fecha_ingreso,historial_asignaciones)
VALUES (61,51,41,102,'Diagnostico',101,'2024-04-01 12:01:02.654321-05','[{"by":"102","technician":"101"}]'),
       (62,51,41,102,'Directo',NULL,'2024-04-02 12:00:00-05','[]');
INSERT INTO diagnosticos(id_diagnostico,id_recepcion,id_tecnico,resumen,estado,version,finalizado_en,historial)
VALUES (9007199254740993,61,101,'Diagnóstico QA','Finalizado',2,'2024-04-01 12:20:00-05',
 '[{"by":"101","number":9007199254740993123,"observation":"á 🚘"}]');
INSERT INTO servicios(id_servicio,nombre,precio_referencia) VALUES (71,'Servicio QA',100.25);
INSERT INTO repuestos(id_repuesto,nombre,stock,precio_referencia) VALUES (72,'Repuesto QA',2.25,5.75);
INSERT INTO proveedores(id_proveedor,nombre,documento) VALUES (73,'Proveedor QA','DOC-QA');
INSERT INTO repuesto_proveedor(id,id_repuesto,id_proveedor,costo_referencia) VALUES (74,72,73,2.15);
INSERT INTO cotizaciones(id_cotizacion,id_diagnostico,id_recepcion,estado,total,version,aprobacion,historial_revision)
VALUES (81,9007199254740993,61,'aprobada',100.25,3,'{"by":"102","version":3}','[{"version":2,"total":90.12}]'),
       (82,NULL,62,'pendiente',5.75,1,NULL,'[]');
INSERT INTO cotizaciones(id_cotizacion,id_diagnostico,id_recepcion,estado,total,id_cotizacion_base)
VALUES (83,9007199254740993,61,'pendiente',5.75,81);
INSERT INTO detalle_cotizacion(id_detalle,id_cotizacion,id_servicio,id_repuesto,descripcion,cantidad,precio_unitario,subtotal,tipo_concepto)
VALUES (84,81,71,NULL,'Servicio autorizado QA',1,100.25,100.25,'Servicio'),
       (85,82,NULL,72,'Repuesto QA',1,5.75,5.75,'Repuesto');
INSERT INTO ordenes_trabajo(id_orden,id_cotizacion,codigo_ot,estado,id_usuario_creacion,control_tecnico,control_autor,control_historial)
VALUES (91,81,'OT-91','finalizada',102,'{"performedItems":["84"]}',101,'[{"version":1}]');
INSERT INTO asignaciones_ot(id_asignacion,id_orden,id_usuario,rol_en_trabajo) VALUES (92,91,101,'MECANICO');
-- Registro permitido en una versión anterior. El trigger actual debe quedar activo.
ALTER TABLE asignaciones_ot DISABLE TRIGGER jac_sin_auxiliar_nuevo;
INSERT INTO asignaciones_ot(id_asignacion,id_orden,id_usuario,rol_en_trabajo,activo) VALUES (93,91,102,'AUXILIAR',false);
ALTER TABLE asignaciones_ot ENABLE TRIGGER jac_sin_auxiliar_nuevo;
INSERT INTO avances_ot(id_avance,id_orden,id_usuario,observacion) VALUES (94,91,101,'Trabajo histórico');
INSERT INTO pruebas_ot(id_prueba,id_orden,id_usuario,tipo_prueba,resultado) VALUES (95,91,101,'Ruta','satisfactoria');
INSERT INTO historial_estados_ot(id_historial,id_orden,id_usuario,estado_anterior,estado_nuevo)
VALUES (96,91,101,'en_pruebas','finalizada');
INSERT INTO pagos(id_pago,id_orden,monto,metodo_pago,id_usuario_registro,request_id,revertido_en,rectificacion)
VALUES (97,91,20.00,'Efectivo',102,'qa-original','2024-04-03T12:00:00-05','{"reason":"QA","by":"102"}'),
       (98,91,20.00,'Tarjeta',102,'qa-replacement',NULL,NULL),
       (99,91,80.25,'Transferencia',102,'qa-balance',NULL,NULL);
INSERT INTO entregas(id_entrega,id_orden,id_usuario_entrega,tipo_receptor,autorizacion_verificada,
    medio_verificacion,id_usuario_verificador,ficha_snapshot,ficha_pdf,receptor_real,conformidad)
VALUES (111,91,102,'Titular / representante',true,'Documento',102,
    '{"historico":true}',decode('255044462d312e340a'||repeat('61',9000)||'0a2525454f46','hex'),
    '{"name":"Cliente QA"}','{"accepted":true}');
INSERT INTO unidades_auxilio(id_unidad_auxilio,codigo,descripcion) VALUES (113,'QA-UNIT','');
INSERT INTO auxilios(id_auxilio,id_vehiculo,id_cliente_solicitante,id_usuario_registro,id_tecnico,id_unidad_auxilio,id_recepcion,ubicacion)
VALUES (114,51,41,102,101,113,61,'Ubicación QA');
-- La secuencia puede estar por delante de la última fila debido a borrados, caché o rollback.
SELECT setval('public.pagos_id_pago_seq',1000,true);
SELECT setval('public.servicios_id_servicio_seq',400,false);
-- Un documento ya emitido debe conservar su catálogo aunque haya quedado inactivo.
UPDATE tipos_documento SET activo=false WHERE id_tipo_documento=31;
COMMIT;
