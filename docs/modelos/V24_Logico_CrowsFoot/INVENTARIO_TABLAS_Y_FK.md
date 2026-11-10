# Inventario de tablas y claves foráneas

Fecha: 2026-11-02. Rama `backend/arquitectura-go-mongodb`. **28 tablas, 208 columnas y 50 FK**, reconstruidas desde la exportación y las migraciones 001–009. No se presupusieron los recuentos ni se ejecutó SQL. Alcance: estructura versionada esperada; no certifica el catálogo de una instancia viva. No se encontró migración posterior a 009.

## Método

`inventariar_esquema.py` inspecciona CREATE TABLE, PK, FK, UNIQUE, índices parciales/completos y ALTER COLUMN/COLUMN. Conserva definiciones, origen y SHA-256 en ESQUEMA_VERIFICADO.json. No es un parser SQL universal ni ejecuta bloques DO: reconstruye las declaraciones de esta cadena concreta, cuya aplicación puede abortar ante inconsistencias históricas. El diagrama no acredita un despliegue.

Los nombres de FK inline se derivan de la convención PostgreSQL y llevan generatedName=true en el JSON; no se afirma haberlos consultado en un catálogo vivo. F01–F50 son identificadores documentales, no columnas nuevas.

## Evolución reconstruida

| Fuente | Tablas | Columnas | FK |
| --- | ---: | ---: | ---: |
| 000_base_legacy.sql | 26 | 160 | 42 |
| 001_revision_funcional.sql | 26 | 163 | 42 |
| 002_marcas_vehiculos.sql | 26 | 163 | 42 |
| 003_atencion_ficha.sql | 26 | 171 | 44 |
| 004_flujo_diagnosticos.sql | 26 | 184 | 46 |
| 005_auditoria_flujo.sql | 26 | 191 | 47 |
| 006_pagos_entrega.sql | 26 | 196 | 47 |
| 007_empleados.sql | 27 | 201 | 48 |
| 008_tipos_documento.sql | 28 | 207 | 50 |
| 009_tres_roles.sql | 28 | 208 | 50 |

## Todas las FK

Cada fila se dibuja exactamente una vez. Una FK NOT NULL exige padre a la hija, no hijos al padre. Una UNIQUE compuesta/parcial no implica unicidad individual. El CSV complementario permite filtrar estas 50 referencias.

| ID | Tabla origen | Columna FK | Tabla destino | Columna referenciada | Permite NULL | Unicidad de FK | Cardinalidad lógica | Regla Go adicional | Regla SQL adicional | Restricción | Fuente | ON DELETE |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| F01 | asignaciones_ot | id_orden | ordenes_trabajo | id_orden | No | No individual | destino por origen 1; orígenes por destino 0..N | CreateOrder inserta asignación mecánica en la misma transacción; SQL no exige hijos por orden. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_asignaciones_ot_id_orden | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F02 | asignaciones_ot | id_usuario | usuarios | id_usuario | No | No individual | destino por origen 1; orígenes por destino 0..N | technicalUser exige cuenta activa y rol Mecánico activo; rechaza empleado vinculado inactivo. No exige vincular cuentas legadas sin empleado. 009 impide nuevas asignaciones AUXILIAR. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_asignaciones_ot_id_usuario | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F03 | auxilios | id_cliente_solicitante | clientes | id_cliente | Sí | No individual | destino por origen 0..1; orígenes por destino 0..N | SaveAuxilio exige cliente en altas y coincidencia con vehículo; SQL nullable sigue vigente para legado. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_auxilios_id_cliente_solicitante | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F04 | auxilios | id_recepcion | recepciones | id_recepcion | Sí | No individual | destino por origen 0..1; orígenes por destino 0..N | SaveAuxilio comprueba cliente y vehículo de la recepción, si se informa. Sin UNIQUE inversa; no aplicar P3. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_auxilios_id_recepcion | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F05 | auxilios | id_tecnico | usuarios | id_usuario | Sí | No individual | destino por origen 0..1; orígenes por destino 0..N | SaveAuxilio exige técnico activo en altas; la FK apunta a usuario, no a empleado. NULL histórico sigue permitido. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_auxilios_id_tecnico | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F06 | auxilios | id_unidad_auxilio | unidades_auxilio | id_unidad_auxilio | Sí | No individual | destino por origen 0..1; orígenes por destino 0..N | SaveAuxilio exige unidad activa en altas; SQL permite NULL. No certifica disponibilidad simultánea de la unidad. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_auxilios_id_unidad_auxilio | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F07 | auxilios | id_usuario_registro | usuarios | id_usuario | No | No individual | destino por origen 1; orígenes por destino 0..N | Usuario autenticado registrador, asignado por servidor; puede diferir del técnico. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_auxilios_id_usuario_registro | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F08 | auxilios | id_vehiculo | vehiculos | id_vehiculo | Sí | No individual | destino por origen 0..1; orígenes por destino 0..N | SaveAuxilio exige vehículo del cliente en altas. Muchos incidentes históricos admitidos. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_auxilios_id_vehiculo | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F09 | avances_ot | id_orden | ordenes_trabajo | id_orden | No | No individual | destino por origen 1; orígenes por destino 0..N | Avance dentro de orden abierta y autorización según asignación/permiso. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_avances_ot_id_orden | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F10 | avances_ot | id_usuario | usuarios | id_usuario | No | No individual | destino por origen 1; orígenes por destino 0..N | Cuenta autora del avance; no prueba por sí sola ejecución física del trabajo. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_avances_ot_id_usuario | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F11 | clientes | id_tipo_documento | tipos_documento | id_tipo_documento | No | No individual | destino por origen 1; orígenes por destino 0..N | Catálogo documental activo y normalización de identidad; trigger 008 conserva compatibilidad histórica. | 008: FK NN y trigger documental. | clientes_id_tipo_documento_fkey | backend/internal/platform/postgres/migrations/008_tipos_documento.sql | RESTRICT |
| F12 | contactos_cliente | id_cliente | clientes | id_cliente | No | No individual | destino por origen 1; orígenes por destino 0..N | Contacto del cliente; registrarlo no autoriza automáticamente a retirar el vehículo. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_contactos_cliente_id_cliente | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F13 | contactos_cliente | id_tipo_documento | tipos_documento | id_tipo_documento | No | No individual | destino por origen 1; orígenes por destino 0..N | Catálogo documental activo, con compatibilidad de etiquetas históricas al editar datos ordinarios. | 008: FK NN y trigger documental. | contactos_cliente_id_tipo_documento_fkey | backend/internal/platform/postgres/migrations/008_tipos_documento.sql | RESTRICT |
| F14 | cotizaciones | id_cotizacion_base | cotizaciones | id_cotizacion | Sí | No individual | destino por origen 0..1; orígenes por destino 0..N | Ampliación de misma recepción/base y orden base abierta; Go limita creación indebida. La FK por sí sola no impide ciclos arbitrarios. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | cotizaciones_id_cotizacion_base_fkey | backend/internal/platform/postgres/migrations/005_auditoria_flujo.sql | NO ACTION |
| F15 | cotizaciones | id_diagnostico | diagnosticos | id_diagnostico | Sí | No individual | destino por origen 0..1; orígenes por destino 0..N | SaveQuote y trigger SQL 004 exigen diagnóstico finalizado y de la misma recepción; directo omite diagnóstico. | 003/004: trigger valida modalidad, diagnóstico final y misma recepción. | fk_cotizaciones_id_diagnostico | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F16 | cotizaciones | id_recepcion | recepciones | id_recepcion | No | No individual | destino por origen 1; orígenes por destino 0..N | SaveQuote controla modalidad, cotización base, ampliaciones y revisión. Técnica pendiente no equivale a directa. | 003/004: trigger determina diagnóstico según modalidad. | cotizaciones_id_recepcion_fkey | backend/internal/platform/postgres/migrations/003_atencion_ficha.sql | NO ACTION |
| F17 | detalle_cotizacion | id_cotizacion | cotizaciones | id_cotizacion | No | No individual | destino por origen 1; orígenes por destino 0..N | SaveQuote exige 1–100 líneas y guarda cabecera/detalles atómicamente; FK sola no obliga a tener hijos. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_detalle_cotizacion_id_cotizacion | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F18 | detalle_cotizacion | id_repuesto | repuestos | id_repuesto | Sí | No individual | destino por origen 0..1; orígenes por destino 0..N | CHECK SQL 001 y Go excluyen dos referencias simultáneas; repuesto libre tipificado permitido; suministro CLIENTE sin cobro del repuesto. | 001: CHECK de referencias, cantidades, precios y descuentos. | fk_detalle_cotizacion_id_repuesto | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F19 | detalle_cotizacion | id_servicio | servicios | id_servicio | Sí | No individual | destino por origen 0..1; orígenes por destino 0..N | CHECK SQL 001 y Go permiten servicio de catálogo o libre tipificado, nunca referencia simultánea de repuesto. | 001: CHECK admite concepto libre tipificado y referencias excluyentes. | fk_detalle_cotizacion_id_servicio | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F20 | diagnosticos | id_recepcion | recepciones | id_recepcion | No | Individual | destino por origen 1; orígenes por destino 0..1 | UNIQUE individual; SaveDiagnosis y trigger 004 exigen recepción técnica, asignación previa y finalizado inmutable. | UNIQUE id_recepcion y trigger 004 de diagnóstico. | fk_diagnosticos_id_recepcion | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F21 | diagnosticos | id_tecnico | usuarios | id_usuario | No | No individual | destino por origen 1; orígenes por destino 0..N | Responsable coincide con técnico asignado en recepción. Autor del historial puede ser otro supervisor autorizado; sigue siendo usuario técnico. | 004: trigger verifica técnico asignado. | fk_diagnosticos_id_tecnico | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F22 | entregas | id_contacto_receptor | contactos_cliente | id_contacto | Sí | No individual | destino por origen 0..1; orígenes por destino 0..N | Contacto receptor debe corresponder al cliente de la atención; identidad/autorización y snapshot aparte. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_entregas_id_contacto_receptor | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F23 | entregas | id_orden | ordenes_trabajo | id_orden | No | Individual | destino por origen 1; orígenes por destino 0..1 | UNIQUE individual; Go y trigger 006 exigen orden finalizada, saldo neto cero y verificación; Go añade conciliación/conformidad. | UNIQUE id_orden; 006: trigger estado/saldo/autorización. | fk_entregas_id_orden | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F24 | entregas | id_usuario_entrega | usuarios | id_usuario | No | No individual | destino por origen 1; orígenes por destino 0..N | Cuenta autenticada que confirma entrega; puede ser distinta del mecánico. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_entregas_id_usuario_entrega | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F25 | entregas | id_usuario_verificador | usuarios | id_usuario | No | No individual | destino por origen 1; orígenes por destino 0..N | Cuenta verificadora de identidad/autorización, distinta de la persona que recibe el vehículo. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_entregas_id_usuario_verificador | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F26 | historial_estados_ot | id_orden | ordenes_trabajo | id_orden | No | No individual | destino por origen 1; orígenes por destino 0..N | CreateOrder crea evento inicial en la misma transacción; Go valida transiciones. FK sola no exige historial no vacío. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_historial_estados_ot_id_orden | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F27 | historial_estados_ot | id_usuario | usuarios | id_usuario | No | No individual | destino por origen 1; orígenes por destino 0..N | Autor autenticado de transición; no se convierte en empleado ejecutor por registrar el evento. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_historial_estados_ot_id_usuario | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F28 | historial_propietarios | id_cliente | clientes | id_cliente | No | No individual | destino por origen 1; orígenes por destino 0..N | Cliente titular del período; reutiliza CLIENTES sin entidad Propietario duplicada. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_historial_cliente | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F29 | historial_propietarios | id_usuario_registro | usuarios | id_usuario | No | No individual | destino por origen 1; orígenes por destino 0..N | Usuario autenticado que registra alta o transferencia, no el cliente propietario. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_historial_usuario | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F30 | historial_propietarios | id_vehiculo | vehiculos | id_vehiculo | No | No individual | destino por origen 1; orígenes por destino 0..N | Alta y traspaso con historial atómico. Índice parcial limita un período vigente, no un solo período histórico. | UNIQUE parcial WHERE fecha_fin IS NULL no limita historia completa. | fk_historial_vehiculo | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F31 | ordenes_trabajo | control_autor | usuarios | id_usuario | Sí | No individual | destino por origen 0..1; orígenes por destino 0..N | Autor del control técnico, opcional antes de iniciarlo; no asignación laboral directa. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | ordenes_trabajo_control_autor_fkey | backend/internal/platform/postgres/migrations/003_atencion_ficha.sql | NO ACTION |
| F32 | ordenes_trabajo | id_cotizacion | cotizaciones | id_cotizacion | No | Individual | destino por origen 1; orígenes por destino 0..1 | UNIQUE individual y trigger de estado aprobado; Go exige aprobación explícita y controla creación de OT. | UNIQUE id_cotizacion; trigger fn_bravo_ot_solo_aprobada. | fk_ordenes_trabajo_id_cotizacion | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F33 | ordenes_trabajo | id_usuario_creacion | usuarios | id_usuario | No | No individual | destino por origen 1; orígenes por destino 0..N | Usuario creador de la orden, distinto del responsable de asignaciones. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_ordenes_trabajo_id_usuario_creacion | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F34 | pagos | id_orden | ordenes_trabajo | id_orden | No | No individual | destino por origen 1; orígenes por destino 0..N | Pagos parciales y saldo conciliado; request_id para reintentos. No unicidad individual por orden. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_pagos_id_orden | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F35 | pagos | id_usuario_registro | usuarios | id_usuario | No | No individual | destino por origen 1; orígenes por destino 0..N | Usuario autenticado que registra el pago; no se deduce del técnico. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_pagos_id_usuario_registro | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F36 | pruebas_ot | id_orden | ordenes_trabajo | id_orden | No | No individual | destino por origen 1; orígenes por destino 0..N | Prueba en orden abierta, con permisos; pruebas satisfactorias para finalizar son regla por estado, no mínimo SQL global. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_pruebas_ot_id_orden | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F37 | pruebas_ot | id_usuario | usuarios | id_usuario | No | No individual | destino por origen 1; orígenes por destino 0..N | Cuenta autora de prueba; distinguir de ejecutor físico cuando Administración captura datos. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_pruebas_ot_id_usuario | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F38 | recepciones | diagnostico_asignado_por | usuarios | id_usuario | Sí | No individual | destino por origen 0..1; orígenes por destino 0..N | Usuario autorizado que asigna diagnóstico; opcional si no hay asignación o es directo. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | recepciones_diagnostico_asignado_por_fkey | backend/internal/platform/postgres/migrations/004_flujo_diagnosticos.sql | NO ACTION |
| F39 | recepciones | id_cliente_atencion | clientes | id_cliente | No | No individual | destino por origen 1; orígenes por destino 0..N | CreateReception exige titular registrado al alta; conserva cliente histórico aunque cambie propietario después. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_recepciones_id_cliente_atencion | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F40 | recepciones | id_mecanico_diagnostico | usuarios | id_usuario | Sí | No individual | destino por origen 0..1; orígenes por destino 0..N | Asignación diagnóstica opcional, solo técnica; requiere permiso y cuenta mecánica habilitada. No es empleados.id_empleado. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | recepciones_id_mecanico_diagnostico_fkey | backend/internal/platform/postgres/migrations/004_flujo_diagnosticos.sql | NO ACTION |
| F41 | recepciones | id_recepcionista | usuarios | id_usuario | No | No individual | destino por origen 1; orígenes por destino 0..N | Servidor registra u.ID; id_recepcionista sigue vigente. Renombrado aprobado a id_usuario_registro permanece pendiente. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_recepciones_id_recepcionista | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F42 | recepciones | id_vehiculo | vehiculos | id_vehiculo | No | No individual | destino por origen 1; orígenes por destino 0..N | Vehículo del cliente y kilometraje no decreciente; actualización transaccional. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_recepciones_id_vehiculo | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F43 | repuesto_proveedor | id_proveedor | proveedores | id_proveedor | No | No individual | destino por origen 1; orígenes por destino 0..N | Proveedor existente y costo de referencia válido; no implica compra registrada. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_repuesto_proveedor_id_proveedor | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F44 | repuesto_proveedor | id_repuesto | repuestos | id_repuesto | No | No individual | destino por origen 1; orígenes por destino 0..N | Repuesto existente y pareja proveedor/repuesto única; no unicidad de cada miembro por separado. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_repuesto_proveedor_id_repuesto | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F45 | rol_permisos | id_permiso | permisos | id_permiso | No | No individual | destino por origen 1; orígenes por destino 0..N | Matriz de permisos autorizada, sin autoescalamiento ni pérdida del último acceso administrativo. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_rol_permisos_permiso | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F46 | rol_permisos | id_rol | roles | id_rol | No | No individual | destino por origen 1; orígenes por destino 0..N | Matriz por rol transaccional; no GRANT PostgreSQL ni cambio de cargo laboral. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_rol_permisos_rol | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F47 | usuario_roles | id_rol | roles | id_rol | No | No individual | destino por origen 1; orígenes por destino 0..N | Roles activos asignables; usuario admite varios roles, sin borrar silenciosamente los adicionales. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_usuario_roles_rol | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F48 | usuario_roles | id_usuario | usuarios | id_usuario | No | No individual | destino por origen 1; orígenes por destino 0..N | Gestión de cuentas/roles autorizada y protección del último acceso administrativo. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_usuario_roles_usuario | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |
| F49 | usuarios | id_empleado | empleados | id_empleado | Sí | Individual | destino por origen 0..1; orígenes por destino 0..1 | Vinculación laboral verificada con permiso específico; no se infiere por nombre o coincidencia numérica. UNIQUE nullable: 0..1 en ambos sentidos. | 007: UNIQUE nullable y triggers de nombre compatible; sin NOT NULL. | usuarios_id_empleado_fkey | backend/internal/platform/postgres/migrations/007_empleados.sql | RESTRICT |
| F50 | vehiculos | id_cliente | clientes | id_cliente | No | No individual | destino por origen 1; orígenes por destino 0..N | Titular registrado actual; alta/historial y transferencia verificada atómicos. No acredita propiedad legal externa. | FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla. | fk_vehiculos_cliente | backend/internal/platform/postgres/migrations/000_base_legacy.sql | NO ACTION |

## Diccionario completo por tabla

Se incluyen todas las columnas. Los tipos permiten contrastar con el almacenamiento; no son modificaciones propuestas. PK compuesta se interpreta como conjunto.

### asignaciones_ot

PK: `id_asignacion`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_asignacion | BIGINT | No | PK | 000_base_legacy.sql |
| id_orden | BIGINT | No | F01 FK | 000_base_legacy.sql |
| id_usuario | BIGINT | No | F02 FK | 000_base_legacy.sql |
| rol_en_trabajo | VARCHAR(30) | No | - | 000_base_legacy.sql |
| fecha_asignacion | TIMESTAMPTZ | No | - | 000_base_legacy.sql |
| activo | BOOLEAN | No | - | 000_base_legacy.sql |

UNIQUE adicionales:

- Ninguna aparte de PK.

FK salientes: F01, F02. Entrantes: ninguna.

### auxilios

PK: `id_auxilio`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_auxilio | BIGINT | No | PK | 000_base_legacy.sql |
| id_vehiculo | BIGINT | Sí | F08 FK | 000_base_legacy.sql |
| id_cliente_solicitante | BIGINT | Sí | F03 FK | 000_base_legacy.sql |
| id_usuario_registro | BIGINT | No | F07 FK | 000_base_legacy.sql |
| id_tecnico | BIGINT | Sí | F05 FK | 000_base_legacy.sql |
| id_unidad_auxilio | BIGINT | Sí | F06 FK | 000_base_legacy.sql |
| id_recepcion | BIGINT | Sí | F04 FK | 000_base_legacy.sql |
| ubicacion | TEXT | No | - | 000_base_legacy.sql |
| tipo_incidente | VARCHAR(60) | Sí | - | 000_base_legacy.sql |
| descripcion | TEXT | Sí | - | 000_base_legacy.sql |
| fecha | TIMESTAMPTZ | No | - | 000_base_legacy.sql |
| estado | VARCHAR(30) | No | - | 000_base_legacy.sql |

UNIQUE adicionales:

- Ninguna aparte de PK.

FK salientes: F03, F04, F05, F06, F07, F08. Entrantes: ninguna.

### avances_ot

PK: `id_avance`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_avance | BIGINT | No | PK | 000_base_legacy.sql |
| id_orden | BIGINT | No | F09 FK | 000_base_legacy.sql |
| id_usuario | BIGINT | No | F10 FK | 000_base_legacy.sql |
| observacion | TEXT | No | - | 000_base_legacy.sql |
| fecha | TIMESTAMPTZ | No | - | 000_base_legacy.sql |

UNIQUE adicionales:

- Ninguna aparte de PK.

FK salientes: F09, F10. Entrantes: ninguna.

### clientes

PK: `id_cliente`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_cliente | BIGINT | No | PK | 000_base_legacy.sql |
| tipo_documento | VARCHAR(20) | No | - | 000_base_legacy.sql |
| numero_documento | VARCHAR(20) | No | - | 000_base_legacy.sql |
| nombre | VARCHAR(150) | No | - | 000_base_legacy.sql |
| telefono | VARCHAR(20) | Sí | - | 000_base_legacy.sql |
| direccion | TEXT | Sí | - | 000_base_legacy.sql |
| correo | VARCHAR(150) | Sí | - | 000_base_legacy.sql |
| id_tipo_documento | BIGINT | No | F11 FK | 008_tipos_documento.sql |

UNIQUE adicionales:

- `uq_cliente_documento`: (tipo_documento, numero_documento); todos los registros.
- `uq_clientes_tipo_catalogo_numero`: (id_tipo_documento, numero_documento); todos los registros.

FK salientes: F11. Entrantes: F03, F12, F28, F39, F50.

### contactos_cliente

PK: `id_contacto`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_contacto | BIGINT | No | PK | 000_base_legacy.sql |
| id_cliente | BIGINT | No | F12 FK | 000_base_legacy.sql |
| nombre | VARCHAR(150) | No | - | 000_base_legacy.sql |
| tipo_documento | VARCHAR(20) | No | - | 000_base_legacy.sql |
| numero_documento | VARCHAR(20) | No | - | 000_base_legacy.sql |
| telefono | VARCHAR(20) | Sí | - | 000_base_legacy.sql |
| relacion | VARCHAR(50) | Sí | - | 000_base_legacy.sql |
| activo | BOOLEAN | No | - | 000_base_legacy.sql |
| id_tipo_documento | BIGINT | No | F13 FK | 008_tipos_documento.sql |

UNIQUE adicionales:

- Ninguna aparte de PK.

FK salientes: F12, F13. Entrantes: F22.

### cotizaciones

PK: `id_cotizacion`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_cotizacion | BIGINT | No | PK | 000_base_legacy.sql |
| id_diagnostico | BIGINT | Sí | F15 FK | 000_base_legacy.sql |
| estado | VARCHAR(20) | No | - | 000_base_legacy.sql |
| total | NUMERIC(12,2) | No | - | 000_base_legacy.sql |
| fecha_emision | TIMESTAMPTZ | No | - | 000_base_legacy.sql |
| aprobada_en | TIMESTAMPTZ | Sí | - | 000_base_legacy.sql |
| id_recepcion | BIGINT | No | F16 FK | 003_atencion_ficha.sql |
| version | BIGINT | No | - | 005_auditoria_flujo.sql |
| contenido_hash | TEXT | Sí | - | 005_auditoria_flujo.sql |
| aprobacion | JSONB | Sí | - | 005_auditoria_flujo.sql |
| historial_revision | JSONB | No | - | 005_auditoria_flujo.sql |
| id_cotizacion_base | BIGINT | Sí | F14 FK | 005_auditoria_flujo.sql |

UNIQUE adicionales:

- Ninguna aparte de PK.

CHECK declarados:

- `CONSTRAINT ck_cotizaciones_total CHECK ((total >= (0)::numeric))`

FK salientes: F14, F15, F16. Entrantes: F14, F17, F32.

### detalle_cotizacion

PK: `id_detalle`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_detalle | BIGINT | No | PK | 000_base_legacy.sql |
| id_cotizacion | BIGINT | No | F17 FK | 000_base_legacy.sql |
| id_servicio | BIGINT | Sí | F19 FK | 000_base_legacy.sql |
| id_repuesto | BIGINT | Sí | F18 FK | 000_base_legacy.sql |
| descripcion | TEXT | Sí | - | 000_base_legacy.sql |
| cantidad | NUMERIC(10,2) | No | - | 000_base_legacy.sql |
| precio_unitario | NUMERIC(12,2) | No | - | 000_base_legacy.sql |
| subtotal | NUMERIC(12,2) | No | - | 000_base_legacy.sql |
| origen_repuesto | VARCHAR(12) | Sí | - | 000_base_legacy.sql |
| tipo_concepto | VARCHAR(10) | Sí | - | 001_revision_funcional.sql |
| tipo_repuesto | VARCHAR(12) | Sí | - | 001_revision_funcional.sql |
| descuento_unitario | NUMERIC(12,2) | No | - | 001_revision_funcional.sql |

UNIQUE adicionales:

- Ninguna aparte de PK.

CHECK declarados:

- `CONSTRAINT ck_detalle_cantidad CHECK ((cantidad > (0)::numeric))`
- `CONSTRAINT ck_detalle_origen CHECK (((origen_repuesto IS NULL) OR ((origen_repuesto)::text = ANY ((ARRAY['TALLER'::character varying, 'CLIENTE'::character varying])::text[]))))`
- `CONSTRAINT ck_detalle_precio CHECK ((precio_unitario >= (0)::numeric))`
- `CONSTRAINT ck_detalle_subtotal CHECK ((subtotal >= (0)::numeric))`
- `ck_detalle_referencia:   (id_servicio IS NOT NULL AND id_repuesto IS NULL AND (tipo_concepto IS NULL OR tipo_concepto='Servicio')) OR  (id_repuesto IS NOT NULL AND id_servicio IS NULL AND (tipo_concepto IS NULL OR tipo_concepto='Repuesto')) OR  (id_servicio IS NULL AND id_repuesto IS NULL AND tipo_concepto IS NOT NULL AND tipo_concepto IN ('Servicio','Repuesto') AND descripcion IS NOT NULL AND btrim(descripcion)<>'') )`

FK salientes: F17, F18, F19. Entrantes: ninguna.

### diagnosticos

PK: `id_diagnostico`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_diagnostico | BIGINT | No | PK | 000_base_legacy.sql |
| id_recepcion | BIGINT | No | F20 FK, UQ individual | 000_base_legacy.sql |
| id_tecnico | BIGINT | No | F21 FK | 000_base_legacy.sql |
| fecha | TIMESTAMPTZ | No | - | 000_base_legacy.sql |
| resumen | TEXT | No | - | 000_base_legacy.sql |
| estado | VARCHAR(15) | No | - | 004_flujo_diagnosticos.sql |
| evaluacion | TEXT | No | - | 004_flujo_diagnosticos.sql |
| recomendaciones | TEXT | No | - | 004_flujo_diagnosticos.sql |
| observaciones | TEXT | No | - | 004_flujo_diagnosticos.sql |
| version | BIGINT | No | - | 004_flujo_diagnosticos.sql |
| finalizado_en | TIMESTAMPTZ | Sí | - | 004_flujo_diagnosticos.sql |
| actualizado_en | TIMESTAMPTZ | Sí | - | 004_flujo_diagnosticos.sql |
| historial | JSONB | No | - | 004_flujo_diagnosticos.sql |

UNIQUE adicionales:

- `uq_diagnosticos_id_recepcion`: (id_recepcion); todos los registros.

FK salientes: F20, F21. Entrantes: F15.

### empleados

PK: `id_empleado`. Origen: `backend/internal/platform/postgres/migrations/007_empleados.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_empleado | BIGSERIAL | No | PK | 007_empleados.sql |
| nombre | VARCHAR(150) | No | - | 007_empleados.sql |
| tipo_empleado | VARCHAR(40) | No | - | 007_empleados.sql |
| activo | BOOLEAN | No | - | 007_empleados.sql |

UNIQUE adicionales:

- Ninguna aparte de PK.

FK salientes: ninguna. Entrantes: F49.

### entregas

PK: `id_entrega`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_entrega | BIGINT | No | PK | 000_base_legacy.sql |
| id_orden | BIGINT | No | F23 FK, UQ individual | 000_base_legacy.sql |
| id_usuario_entrega | BIGINT | No | F24 FK | 000_base_legacy.sql |
| id_contacto_receptor | BIGINT | Sí | F22 FK | 000_base_legacy.sql |
| fecha_entrega | TIMESTAMPTZ | No | - | 000_base_legacy.sql |
| tipo_receptor | VARCHAR(30) | No | - | 000_base_legacy.sql |
| nombre_receptor_snap | VARCHAR(150) | Sí | - | 000_base_legacy.sql |
| tipo_doc_receptor_snap | VARCHAR(20) | Sí | - | 000_base_legacy.sql |
| documento_receptor_snap | VARCHAR(30) | Sí | - | 000_base_legacy.sql |
| telefono_receptor_snap | VARCHAR(20) | Sí | - | 000_base_legacy.sql |
| autorizacion_verificada | BOOLEAN | No | - | 000_base_legacy.sql |
| medio_verificacion | VARCHAR(80) | Sí | - | 000_base_legacy.sql |
| fecha_verificacion | TIMESTAMPTZ | Sí | - | 000_base_legacy.sql |
| id_usuario_verificador | BIGINT | No | F25 FK | 000_base_legacy.sql |
| ficha_snapshot | JSONB | Sí | - | 003_atencion_ficha.sql |
| ficha_pdf | BYTEA | Sí | - | 003_atencion_ficha.sql |
| receptor_real | JSONB | Sí | - | 006_pagos_entrega.sql |
| conformidad | JSONB | Sí | - | 006_pagos_entrega.sql |

UNIQUE adicionales:

- `uq_entregas_id_orden`: (id_orden); todos los registros.

FK salientes: F22, F23, F24, F25. Entrantes: ninguna.

### historial_estados_ot

PK: `id_historial`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_historial | BIGINT | No | PK | 000_base_legacy.sql |
| id_orden | BIGINT | No | F26 FK | 000_base_legacy.sql |
| id_usuario | BIGINT | No | F27 FK | 000_base_legacy.sql |
| estado_anterior | VARCHAR(30) | Sí | - | 000_base_legacy.sql |
| estado_nuevo | VARCHAR(30) | No | - | 000_base_legacy.sql |
| fecha | TIMESTAMPTZ | No | - | 000_base_legacy.sql |

UNIQUE adicionales:

- Ninguna aparte de PK.

FK salientes: F26, F27. Entrantes: ninguna.

### historial_propietarios

PK: `id_historial`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_historial | BIGINT | No | PK | 000_base_legacy.sql |
| id_vehiculo | BIGINT | No | F30 FK | 000_base_legacy.sql |
| id_cliente | BIGINT | No | F28 FK | 000_base_legacy.sql |
| fecha_inicio | TIMESTAMPTZ | No | - | 000_base_legacy.sql |
| fecha_fin | TIMESTAMPTZ | Sí | - | 000_base_legacy.sql |
| id_usuario_registro | BIGINT | No | F29 FK | 000_base_legacy.sql |
| verificado | BOOLEAN | No | - | 000_base_legacy.sql |
| observacion | TEXT | Sí | - | 000_base_legacy.sql |

UNIQUE adicionales:

- `uq_historial_propietario_vigente`: (id_vehiculo); parcial: `WHERE (fecha_fin IS NULL)`.

CHECK declarados:

- `CONSTRAINT chk_fechas_propietario CHECK (((fecha_fin IS NULL) OR (fecha_fin >= fecha_inicio)))`
- `CONSTRAINT ck_historial_fechas CHECK (((fecha_fin IS NULL) OR (fecha_fin >= fecha_inicio)))`

FK salientes: F28, F29, F30. Entrantes: ninguna.

### ordenes_trabajo

PK: `id_orden`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_orden | BIGINT | No | PK | 000_base_legacy.sql |
| id_cotizacion | BIGINT | No | F32 FK, UQ individual | 000_base_legacy.sql |
| codigo_ot | VARCHAR(24) | No | UQ individual | 000_base_legacy.sql |
| estado | VARCHAR(30) | No | - | 000_base_legacy.sql |
| id_usuario_creacion | BIGINT | No | F33 FK | 000_base_legacy.sql |
| fecha_creacion | TIMESTAMPTZ | No | - | 000_base_legacy.sql |
| fecha_inicio | TIMESTAMPTZ | Sí | - | 000_base_legacy.sql |
| fecha_fin | TIMESTAMPTZ | Sí | - | 000_base_legacy.sql |
| control_tecnico | JSONB | No | - | 003_atencion_ficha.sql |
| control_version | BIGINT | No | - | 003_atencion_ficha.sql |
| control_autor | BIGINT | Sí | F31 FK | 003_atencion_ficha.sql |
| control_fecha | TIMESTAMPTZ | Sí | - | 003_atencion_ficha.sql |
| flujo_version | INTEGER | No | - | 005_auditoria_flujo.sql |
| control_historial | JSONB | No | - | 005_auditoria_flujo.sql |

UNIQUE adicionales:

- `uq_ordenes_trabajo_codigo_ot`: (codigo_ot); todos los registros.
- `uq_ordenes_trabajo_id_cotizacion`: (id_cotizacion); todos los registros.

FK salientes: F31, F32, F33. Entrantes: F01, F09, F23, F26, F34, F36.

### pagos

PK: `id_pago`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_pago | BIGINT | No | PK | 000_base_legacy.sql |
| id_orden | BIGINT | No | F34 FK | 000_base_legacy.sql |
| monto | NUMERIC(12,2) | No | - | 000_base_legacy.sql |
| metodo_pago | VARCHAR(30) | No | - | 000_base_legacy.sql |
| fecha_pago | TIMESTAMPTZ | No | - | 000_base_legacy.sql |
| comprobante | VARCHAR(50) | Sí | - | 000_base_legacy.sql |
| id_usuario_registro | BIGINT | No | F35 FK | 000_base_legacy.sql |
| request_id | VARCHAR(100) | Sí | - | 006_pagos_entrega.sql |
| revertido_en | TIMESTAMPTZ | Sí | - | 006_pagos_entrega.sql |
| rectificacion | JSONB | Sí | - | 006_pagos_entrega.sql |

UNIQUE adicionales:

- `ix_pago_request`: (id_orden, request_id); parcial: `WHERE request_id IS NOT NULL`.

CHECK declarados:

- `CONSTRAINT ck_pagos_monto CHECK ((monto > (0)::numeric))`

FK salientes: F34, F35. Entrantes: ninguna.

### permisos

PK: `id_permiso`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_permiso | BIGINT | No | PK | 000_base_legacy.sql |
| codigo | VARCHAR(80) | No | UQ individual | 000_base_legacy.sql |
| descripcion | VARCHAR(150) | Sí | - | 000_base_legacy.sql |

UNIQUE adicionales:

- `permisos_codigo_key`: (codigo); todos los registros.

FK salientes: ninguna. Entrantes: F45.

### proveedores

PK: `id_proveedor`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_proveedor | BIGINT | No | PK | 000_base_legacy.sql |
| nombre | VARCHAR(150) | No | - | 000_base_legacy.sql |
| documento | VARCHAR(20) | No | UQ individual | 000_base_legacy.sql |
| telefono | VARCHAR(20) | Sí | - | 000_base_legacy.sql |
| contacto | VARCHAR(120) | Sí | - | 000_base_legacy.sql |

UNIQUE adicionales:

- `uq_proveedores_documento`: (documento); todos los registros.

FK salientes: ninguna. Entrantes: F43.

### pruebas_ot

PK: `id_prueba`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_prueba | BIGINT | No | PK | 000_base_legacy.sql |
| id_orden | BIGINT | No | F36 FK | 000_base_legacy.sql |
| id_usuario | BIGINT | No | F37 FK | 000_base_legacy.sql |
| tipo_prueba | VARCHAR(100) | No | - | 000_base_legacy.sql |
| resultado | VARCHAR(20) | No | - | 000_base_legacy.sql |
| observaciones | TEXT | Sí | - | 000_base_legacy.sql |
| fecha | TIMESTAMPTZ | No | - | 000_base_legacy.sql |

UNIQUE adicionales:

- Ninguna aparte de PK.

FK salientes: F36, F37. Entrantes: ninguna.

### recepciones

PK: `id_recepcion`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_recepcion | BIGINT | No | PK | 000_base_legacy.sql |
| id_vehiculo | BIGINT | No | F42 FK | 000_base_legacy.sql |
| id_cliente_atencion | BIGINT | No | F39 FK | 000_base_legacy.sql |
| id_recepcionista | BIGINT | No | F41 FK | 000_base_legacy.sql |
| fecha_ingreso | TIMESTAMPTZ | No | - | 000_base_legacy.sql |
| km_ingreso | NUMERIC(10,2) | Sí | - | 000_base_legacy.sql |
| sintomas | TEXT | Sí | - | 000_base_legacy.sql |
| tipo_atencion | VARCHAR(15) | No | - | 003_atencion_ficha.sql |
| id_mecanico_diagnostico | BIGINT | Sí | F40 FK | 004_flujo_diagnosticos.sql |
| diagnostico_asignado_por | BIGINT | Sí | F38 FK | 004_flujo_diagnosticos.sql |
| diagnostico_asignado_en | TIMESTAMPTZ | Sí | - | 004_flujo_diagnosticos.sql |
| asignacion_version | BIGINT | No | - | 004_flujo_diagnosticos.sql |
| historial_asignaciones | JSONB | No | - | 004_flujo_diagnosticos.sql |

UNIQUE adicionales:

- Ninguna aparte de PK.

CHECK declarados:

- `CONSTRAINT ck_recepciones_km CHECK (((km_ingreso IS NULL) OR (km_ingreso >= (0)::numeric)))`

FK salientes: F38, F39, F40, F41, F42. Entrantes: F04, F16, F20.

### repuesto_proveedor

PK: `id`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id | BIGINT | No | PK | 000_base_legacy.sql |
| id_repuesto | BIGINT | No | F44 FK | 000_base_legacy.sql |
| id_proveedor | BIGINT | No | F43 FK | 000_base_legacy.sql |
| costo_referencia | NUMERIC(12,2) | No | - | 000_base_legacy.sql |

UNIQUE adicionales:

- `uq_repuesto_proveedor_id_repuesto_id_proveedor`: (id_repuesto, id_proveedor); todos los registros.

CHECK declarados:

- `CONSTRAINT ck_rp_costo CHECK ((costo_referencia >= (0)::numeric))`

FK salientes: F43, F44. Entrantes: ninguna.

### repuestos

PK: `id_repuesto`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_repuesto | BIGINT | No | PK | 000_base_legacy.sql |
| nombre | VARCHAR(120) | No | - | 000_base_legacy.sql |
| stock | NUMERIC(12,2) | No | - | 000_base_legacy.sql |
| precio_referencia | NUMERIC(12,2) | No | - | 000_base_legacy.sql |
| activo | BOOLEAN | No | - | 000_base_legacy.sql |

UNIQUE adicionales:

- Ninguna aparte de PK.

CHECK declarados:

- `CONSTRAINT ck_repuestos_precio CHECK ((precio_referencia >= (0)::numeric))`
- `CONSTRAINT ck_repuestos_stock CHECK ((stock >= (0)::numeric))`

FK salientes: ninguna. Entrantes: F18, F44.

### rol_permisos

PK: `id_rol, id_permiso`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_rol | BIGINT | No | PK compuesta, F46 FK | 000_base_legacy.sql |
| id_permiso | BIGINT | No | PK compuesta, F45 FK | 000_base_legacy.sql |

UNIQUE adicionales:

- Ninguna aparte de PK.

FK salientes: F45, F46. Entrantes: ninguna.

### roles

PK: `id_rol`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_rol | BIGINT | No | PK | 000_base_legacy.sql |
| nombre_rol | VARCHAR(60) | No | UQ individual | 000_base_legacy.sql |
| activo | BOOLEAN | No | - | 009_tres_roles.sql |

UNIQUE adicionales:

- `roles_nombre_rol_key`: (nombre_rol); todos los registros.

FK salientes: ninguna. Entrantes: F46, F47.

### servicios

PK: `id_servicio`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_servicio | BIGINT | No | PK | 000_base_legacy.sql |
| nombre | VARCHAR(100) | No | UQ individual | 000_base_legacy.sql |
| precio_referencia | NUMERIC(12,2) | No | - | 000_base_legacy.sql |
| activo | BOOLEAN | No | - | 000_base_legacy.sql |

UNIQUE adicionales:

- `uq_servicios_nombre`: (nombre); todos los registros.

CHECK declarados:

- `CONSTRAINT ck_servicios_precio CHECK ((precio_referencia >= (0)::numeric))`

FK salientes: ninguna. Entrantes: F19.

### tipos_documento

PK: `id_tipo_documento`. Origen: `backend/internal/platform/postgres/migrations/008_tipos_documento.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_tipo_documento | BIGSERIAL | No | PK | 008_tipos_documento.sql |
| codigo | VARCHAR(20) | No | UQ individual | 008_tipos_documento.sql |
| nombre | VARCHAR(150) | No | - | 008_tipos_documento.sql |
| activo | BOOLEAN | No | - | 008_tipos_documento.sql |

UNIQUE adicionales:

- `tipos_documento_codigo_key`: (codigo); todos los registros.

FK salientes: ninguna. Entrantes: F11, F13.

### unidades_auxilio

PK: `id_unidad_auxilio`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_unidad_auxilio | BIGINT | No | PK | 000_base_legacy.sql |
| codigo | VARCHAR(30) | No | UQ individual | 000_base_legacy.sql |
| descripcion | VARCHAR(150) | Sí | - | 000_base_legacy.sql |
| activo | BOOLEAN | No | - | 000_base_legacy.sql |

UNIQUE adicionales:

- `uq_unidades_auxilio_codigo`: (codigo); todos los registros.

FK salientes: ninguna. Entrantes: F06.

### usuario_roles

PK: `id_usuario, id_rol`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_usuario | BIGINT | No | PK compuesta, F48 FK | 000_base_legacy.sql |
| id_rol | BIGINT | No | PK compuesta, F47 FK | 000_base_legacy.sql |

UNIQUE adicionales:

- Ninguna aparte de PK.

FK salientes: F47, F48. Entrantes: ninguna.

### usuarios

PK: `id_usuario`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_usuario | BIGINT | No | PK | 000_base_legacy.sql |
| nombre | VARCHAR(150) | No | - | 000_base_legacy.sql |
| usuario | VARCHAR(60) | No | UQ individual | 000_base_legacy.sql |
| correo | VARCHAR(150) | Sí | UQ individual | 000_base_legacy.sql |
| password_hash | TEXT | No | - | 000_base_legacy.sql |
| activo | BOOLEAN | No | - | 000_base_legacy.sql |
| id_empleado | BIGINT | Sí | F49 FK, UQ individual | 007_empleados.sql |

UNIQUE adicionales:

- `usuarios_correo_key`: (correo); todos los registros.
- `usuarios_usuario_key`: (usuario); todos los registros.
- `uq_usuarios_empleado`: (id_empleado); todos los registros.

FK salientes: F49. Entrantes: F02, F05, F07, F10, F21, F24, F25, F27, F29, F31, F33, F35, F37, F38, F40, F41, F48.

### vehiculos

PK: `id_vehiculo`. Origen: `backend/internal/platform/postgres/migrations/000_base_legacy.sql`.

| Columna | Tipo | NULL | Clave | Origen |
| --- | --- | --- | --- | --- |
| id_vehiculo | BIGINT | No | PK | 000_base_legacy.sql |
| id_cliente | BIGINT | No | F50 FK | 000_base_legacy.sql |
| placa | VARCHAR(12) | No | UQ individual | 000_base_legacy.sql |
| marca | VARCHAR(40) | Sí | - | 000_base_legacy.sql |
| modelo | VARCHAR(60) | Sí | - | 000_base_legacy.sql |
| anio | SMALLINT | Sí | - | 000_base_legacy.sql |
| kilometraje_actual | NUMERIC(10,2) | Sí | - | 000_base_legacy.sql |
| observaciones | TEXT | Sí | - | 000_base_legacy.sql |

UNIQUE adicionales:

- `vehiculos_placa_key`: (placa); todos los registros.

CHECK declarados:

- `CONSTRAINT ck_vehiculos_anio CHECK (((anio IS NULL) OR ((anio >= 1900) AND (anio <= 2100))))`
- `CONSTRAINT ck_vehiculos_km CHECK (((kilometraje_actual IS NULL) OR (kilometraje_actual >= (0)::numeric)))`

FK salientes: F50. Entrantes: F08, F30, F42.

## Restricciones fuera de las FK

Los CHECK inline, definiciones y DEFAULT están en columns[].definition/effectiveDefault. JSONB de historial/control y snapshots puede almacenar referencias; no se dibujan rutas JSON como FK. Funciones/disparadores añaden reglas que no caben en cardinalidad binaria. Go añade permisos, concurrencia, revisiones y transacciones.

No hay tabla MECÁNICOS/RECEPCIONISTAS/TÉCNICOS. Se conserva recepciones.id_recepcionista hacia usuarios.id_usuario; renombrado pendiente. TIPOS_DOCUMENTO aparece con sus dos referencias NOT NULL. Cargo laboral no equivale a rol.
