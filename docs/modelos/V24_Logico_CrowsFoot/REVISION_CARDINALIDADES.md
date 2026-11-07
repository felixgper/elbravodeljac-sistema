# Revisión de cardinalidades y conexiones

## Crow’s Foot

Cada conector parte de la fila FK hija y termina en la tabla padre. Tooltip y metadata identifican la columna destino exacta. Los puertos de llegada se distribuyen por el borde para separar símbolos: su altura no asigna la FK a la fila dibujada a ese nivel. F01–F50 identifica inequívocamente origen/destino. Esto es trazado, no una referencia a campos diferentes.

- `||` junto al padre: exactamente un padre por hija, FK NOT NULL.
- `o|` junto al padre: cero o un padre por hija, FK nullable.
- `o<` junto a hija: cero o muchas hijas por padre, FK no única individualmente.
- `o|` junto a hija: cero o una hija por padre, FK única individualmente.

No hay rombos ni pares Chen. La dirección de dibujo no es secuencia de proceso. [PostgreSQL 18: restricciones](https://www.postgresql.org/docs/18/ddl-constraints.html) distingue PK, UNIQUE, FK y NULL: unicidad conjunta no implica individual y, por defecto, UNIQUE admite varios NULL. Un índice parcial afecta solo sus filas seleccionadas.

## Revisión individual

### F01: asignaciones_ot.id_orden → ordenes_trabajo.id_orden

Cada `asignaciones_ot` referencia exactamente un `ordenes_trabajo`. Cada `ordenes_trabajo` admite cero o muchos `asignaciones_ot` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: CreateOrder inserta asignación mecánica en la misma transacción; SQL no exige hijos por orden.

### F02: asignaciones_ot.id_usuario → usuarios.id_usuario

Cada `asignaciones_ot` referencia exactamente un `usuarios`. Cada `usuarios` admite cero o muchos `asignaciones_ot` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: technicalUser exige cuenta activa y rol Mecánico activo; rechaza empleado vinculado inactivo. No exige vincular cuentas legadas sin empleado. 009 impide nuevas asignaciones AUXILIAR.

### F03: auxilios.id_cliente_solicitante → clientes.id_cliente

Cada `auxilios` referencia cero o un `clientes`. Cada `clientes` admite cero o muchos `auxilios` por esta referencia.

SQL: nullable; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: SaveAuxilio exige cliente en altas y coincidencia con vehículo; SQL nullable sigue vigente para legado.

### F04: auxilios.id_recepcion → recepciones.id_recepcion

Cada `auxilios` referencia cero o un `recepciones`. Cada `recepciones` admite cero o muchos `auxilios` por esta referencia.

SQL: nullable; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: SaveAuxilio comprueba cliente y vehículo de la recepción, si se informa. Sin UNIQUE inversa; no aplicar P3.

### F05: auxilios.id_tecnico → usuarios.id_usuario

Cada `auxilios` referencia cero o un `usuarios`. Cada `usuarios` admite cero o muchos `auxilios` por esta referencia.

SQL: nullable; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: SaveAuxilio exige técnico activo en altas; la FK apunta a usuario, no a empleado. NULL histórico sigue permitido.

### F06: auxilios.id_unidad_auxilio → unidades_auxilio.id_unidad_auxilio

Cada `auxilios` referencia cero o un `unidades_auxilio`. Cada `unidades_auxilio` admite cero o muchos `auxilios` por esta referencia.

SQL: nullable; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: SaveAuxilio exige unidad activa en altas; SQL permite NULL. No certifica disponibilidad simultánea de la unidad.

### F07: auxilios.id_usuario_registro → usuarios.id_usuario

Cada `auxilios` referencia exactamente un `usuarios`. Cada `usuarios` admite cero o muchos `auxilios` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Usuario autenticado registrador, asignado por servidor; puede diferir del técnico.

### F08: auxilios.id_vehiculo → vehiculos.id_vehiculo

Cada `auxilios` referencia cero o un `vehiculos`. Cada `vehiculos` admite cero o muchos `auxilios` por esta referencia.

SQL: nullable; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: SaveAuxilio exige vehículo del cliente en altas. Muchos incidentes históricos admitidos.

### F09: avances_ot.id_orden → ordenes_trabajo.id_orden

Cada `avances_ot` referencia exactamente un `ordenes_trabajo`. Cada `ordenes_trabajo` admite cero o muchos `avances_ot` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Avance dentro de orden abierta y autorización según asignación/permiso.

### F10: avances_ot.id_usuario → usuarios.id_usuario

Cada `avances_ot` referencia exactamente un `usuarios`. Cada `usuarios` admite cero o muchos `avances_ot` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Cuenta autora del avance; no prueba por sí sola ejecución física del trabajo.

### F11: clientes.id_tipo_documento → tipos_documento.id_tipo_documento

Cada `clientes` referencia exactamente un `tipos_documento`. Cada `tipos_documento` admite cero o muchos `clientes` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE RESTRICT. 008: FK NN y trigger documental.

Go: Catálogo documental activo y normalización de identidad; trigger 008 conserva compatibilidad histórica.

### F12: contactos_cliente.id_cliente → clientes.id_cliente

Cada `contactos_cliente` referencia exactamente un `clientes`. Cada `clientes` admite cero o muchos `contactos_cliente` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Contacto del cliente; registrarlo no autoriza automáticamente a retirar el vehículo.

### F13: contactos_cliente.id_tipo_documento → tipos_documento.id_tipo_documento

Cada `contactos_cliente` referencia exactamente un `tipos_documento`. Cada `tipos_documento` admite cero o muchos `contactos_cliente` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE RESTRICT. 008: FK NN y trigger documental.

Go: Catálogo documental activo, con compatibilidad de etiquetas históricas al editar datos ordinarios.

### F14: cotizaciones.id_cotizacion_base → cotizaciones.id_cotizacion

Cada `cotizaciones` referencia cero o un `cotizaciones`. Cada `cotizaciones` admite cero o muchos `cotizaciones` por esta referencia.

SQL: nullable; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Ampliación de misma recepción/base y orden base abierta; Go limita creación indebida. La FK por sí sola no impide ciclos arbitrarios.

### F15: cotizaciones.id_diagnostico → diagnosticos.id_diagnostico

Cada `cotizaciones` referencia cero o un `diagnosticos`. Cada `diagnosticos` admite cero o muchos `cotizaciones` por esta referencia.

SQL: nullable; sin UNIQUE individual; ON DELETE NO ACTION. 003/004: trigger valida modalidad, diagnóstico final y misma recepción.

Go: SaveQuote y trigger SQL 004 exigen diagnóstico finalizado y de la misma recepción; directo omite diagnóstico.

### F16: cotizaciones.id_recepcion → recepciones.id_recepcion

Cada `cotizaciones` referencia exactamente un `recepciones`. Cada `recepciones` admite cero o muchos `cotizaciones` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. 003/004: trigger determina diagnóstico según modalidad.

Go: SaveQuote controla modalidad, cotización base, ampliaciones y revisión. Técnica pendiente no equivale a directa.

### F17: detalle_cotizacion.id_cotizacion → cotizaciones.id_cotizacion

Cada `detalle_cotizacion` referencia exactamente un `cotizaciones`. Cada `cotizaciones` admite cero o muchos `detalle_cotizacion` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: SaveQuote exige 1–100 líneas y guarda cabecera/detalles atómicamente; FK sola no obliga a tener hijos.

### F18: detalle_cotizacion.id_repuesto → repuestos.id_repuesto

Cada `detalle_cotizacion` referencia cero o un `repuestos`. Cada `repuestos` admite cero o muchos `detalle_cotizacion` por esta referencia.

SQL: nullable; sin UNIQUE individual; ON DELETE NO ACTION. 001: CHECK de referencias, cantidades, precios y descuentos.

Go: CHECK SQL 001 y Go excluyen dos referencias simultáneas; repuesto libre tipificado permitido; suministro CLIENTE sin cobro del repuesto.

### F19: detalle_cotizacion.id_servicio → servicios.id_servicio

Cada `detalle_cotizacion` referencia cero o un `servicios`. Cada `servicios` admite cero o muchos `detalle_cotizacion` por esta referencia.

SQL: nullable; sin UNIQUE individual; ON DELETE NO ACTION. 001: CHECK admite concepto libre tipificado y referencias excluyentes.

Go: CHECK SQL 001 y Go permiten servicio de catálogo o libre tipificado, nunca referencia simultánea de repuesto.

### F20: diagnosticos.id_recepcion → recepciones.id_recepcion

Cada `diagnosticos` referencia exactamente un `recepciones`. Cada `recepciones` admite cero o un `diagnosticos` por esta referencia.

SQL: NOT NULL; UNIQUE individual; ON DELETE NO ACTION. UNIQUE id_recepcion y trigger 004 de diagnóstico.

Go: UNIQUE individual; SaveDiagnosis y trigger 004 exigen recepción técnica, asignación previa y finalizado inmutable.

### F21: diagnosticos.id_tecnico → usuarios.id_usuario

Cada `diagnosticos` referencia exactamente un `usuarios`. Cada `usuarios` admite cero o muchos `diagnosticos` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. 004: trigger verifica técnico asignado.

Go: Responsable coincide con técnico asignado en recepción. Autor del historial puede ser otro supervisor autorizado; sigue siendo usuario técnico.

### F22: entregas.id_contacto_receptor → contactos_cliente.id_contacto

Cada `entregas` referencia cero o un `contactos_cliente`. Cada `contactos_cliente` admite cero o muchos `entregas` por esta referencia.

SQL: nullable; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Contacto receptor debe corresponder al cliente de la atención; identidad/autorización y snapshot aparte.

### F23: entregas.id_orden → ordenes_trabajo.id_orden

Cada `entregas` referencia exactamente un `ordenes_trabajo`. Cada `ordenes_trabajo` admite cero o un `entregas` por esta referencia.

SQL: NOT NULL; UNIQUE individual; ON DELETE NO ACTION. UNIQUE id_orden; 006: trigger estado/saldo/autorización.

Go: UNIQUE individual; Go y trigger 006 exigen orden finalizada, saldo neto cero y verificación; Go añade conciliación/conformidad.

### F24: entregas.id_usuario_entrega → usuarios.id_usuario

Cada `entregas` referencia exactamente un `usuarios`. Cada `usuarios` admite cero o muchos `entregas` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Cuenta autenticada que confirma entrega; puede ser distinta del mecánico.

### F25: entregas.id_usuario_verificador → usuarios.id_usuario

Cada `entregas` referencia exactamente un `usuarios`. Cada `usuarios` admite cero o muchos `entregas` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Cuenta verificadora de identidad/autorización, distinta de la persona que recibe el vehículo.

### F26: historial_estados_ot.id_orden → ordenes_trabajo.id_orden

Cada `historial_estados_ot` referencia exactamente un `ordenes_trabajo`. Cada `ordenes_trabajo` admite cero o muchos `historial_estados_ot` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: CreateOrder crea evento inicial en la misma transacción; Go valida transiciones. FK sola no exige historial no vacío.

### F27: historial_estados_ot.id_usuario → usuarios.id_usuario

Cada `historial_estados_ot` referencia exactamente un `usuarios`. Cada `usuarios` admite cero o muchos `historial_estados_ot` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Autor autenticado de transición; no se convierte en empleado ejecutor por registrar el evento.

### F28: historial_propietarios.id_cliente → clientes.id_cliente

Cada `historial_propietarios` referencia exactamente un `clientes`. Cada `clientes` admite cero o muchos `historial_propietarios` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Cliente titular del período; reutiliza CLIENTES sin entidad Propietario duplicada.

### F29: historial_propietarios.id_usuario_registro → usuarios.id_usuario

Cada `historial_propietarios` referencia exactamente un `usuarios`. Cada `usuarios` admite cero o muchos `historial_propietarios` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Usuario autenticado que registra alta o transferencia, no el cliente propietario.

### F30: historial_propietarios.id_vehiculo → vehiculos.id_vehiculo

Cada `historial_propietarios` referencia exactamente un `vehiculos`. Cada `vehiculos` admite cero o muchos `historial_propietarios` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. UNIQUE parcial WHERE fecha_fin IS NULL no limita historia completa.

Go: Alta y traspaso con historial atómico. Índice parcial limita un período vigente, no un solo período histórico.

### F31: ordenes_trabajo.control_autor → usuarios.id_usuario

Cada `ordenes_trabajo` referencia cero o un `usuarios`. Cada `usuarios` admite cero o muchos `ordenes_trabajo` por esta referencia.

SQL: nullable; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Autor del control técnico, opcional antes de iniciarlo; no asignación laboral directa.

### F32: ordenes_trabajo.id_cotizacion → cotizaciones.id_cotizacion

Cada `ordenes_trabajo` referencia exactamente un `cotizaciones`. Cada `cotizaciones` admite cero o un `ordenes_trabajo` por esta referencia.

SQL: NOT NULL; UNIQUE individual; ON DELETE NO ACTION. UNIQUE id_cotizacion; trigger fn_bravo_ot_solo_aprobada.

Go: UNIQUE individual y trigger de estado aprobado; Go exige aprobación explícita y controla creación de OT.

### F33: ordenes_trabajo.id_usuario_creacion → usuarios.id_usuario

Cada `ordenes_trabajo` referencia exactamente un `usuarios`. Cada `usuarios` admite cero o muchos `ordenes_trabajo` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Usuario creador de la orden, distinto del responsable de asignaciones.

### F34: pagos.id_orden → ordenes_trabajo.id_orden

Cada `pagos` referencia exactamente un `ordenes_trabajo`. Cada `ordenes_trabajo` admite cero o muchos `pagos` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Pagos parciales y saldo conciliado; request_id para reintentos. No unicidad individual por orden.

### F35: pagos.id_usuario_registro → usuarios.id_usuario

Cada `pagos` referencia exactamente un `usuarios`. Cada `usuarios` admite cero o muchos `pagos` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Usuario autenticado que registra el pago; no se deduce del técnico.

### F36: pruebas_ot.id_orden → ordenes_trabajo.id_orden

Cada `pruebas_ot` referencia exactamente un `ordenes_trabajo`. Cada `ordenes_trabajo` admite cero o muchos `pruebas_ot` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Prueba en orden abierta, con permisos; pruebas satisfactorias para finalizar son regla por estado, no mínimo SQL global.

### F37: pruebas_ot.id_usuario → usuarios.id_usuario

Cada `pruebas_ot` referencia exactamente un `usuarios`. Cada `usuarios` admite cero o muchos `pruebas_ot` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Cuenta autora de prueba; distinguir de ejecutor físico cuando Administración captura datos.

### F38: recepciones.diagnostico_asignado_por → usuarios.id_usuario

Cada `recepciones` referencia cero o un `usuarios`. Cada `usuarios` admite cero o muchos `recepciones` por esta referencia.

SQL: nullable; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Usuario autorizado que asigna diagnóstico; opcional si no hay asignación o es directo.

### F39: recepciones.id_cliente_atencion → clientes.id_cliente

Cada `recepciones` referencia exactamente un `clientes`. Cada `clientes` admite cero o muchos `recepciones` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: CreateReception exige titular registrado al alta; conserva cliente histórico aunque cambie propietario después.

### F40: recepciones.id_mecanico_diagnostico → usuarios.id_usuario

Cada `recepciones` referencia cero o un `usuarios`. Cada `usuarios` admite cero o muchos `recepciones` por esta referencia.

SQL: nullable; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Asignación diagnóstica opcional, solo técnica; requiere permiso y cuenta mecánica habilitada. No es empleados.id_empleado.

### F41: recepciones.id_recepcionista → usuarios.id_usuario

Cada `recepciones` referencia exactamente un `usuarios`. Cada `usuarios` admite cero o muchos `recepciones` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Servidor registra u.ID; id_recepcionista sigue vigente. Renombrado aprobado a id_usuario_registro permanece pendiente.

### F42: recepciones.id_vehiculo → vehiculos.id_vehiculo

Cada `recepciones` referencia exactamente un `vehiculos`. Cada `vehiculos` admite cero o muchos `recepciones` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Vehículo del cliente y kilometraje no decreciente; actualización transaccional.

### F43: repuesto_proveedor.id_proveedor → proveedores.id_proveedor

Cada `repuesto_proveedor` referencia exactamente un `proveedores`. Cada `proveedores` admite cero o muchos `repuesto_proveedor` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Proveedor existente y costo de referencia válido; no implica compra registrada.

### F44: repuesto_proveedor.id_repuesto → repuestos.id_repuesto

Cada `repuesto_proveedor` referencia exactamente un `repuestos`. Cada `repuestos` admite cero o muchos `repuesto_proveedor` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Repuesto existente y pareja proveedor/repuesto única; no unicidad de cada miembro por separado.

### F45: rol_permisos.id_permiso → permisos.id_permiso

Cada `rol_permisos` referencia exactamente un `permisos`. Cada `permisos` admite cero o muchos `rol_permisos` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Matriz de permisos autorizada, sin autoescalamiento ni pérdida del último acceso administrativo.

### F46: rol_permisos.id_rol → roles.id_rol

Cada `rol_permisos` referencia exactamente un `roles`. Cada `roles` admite cero o muchos `rol_permisos` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Matriz por rol transaccional; no GRANT PostgreSQL ni cambio de cargo laboral.

### F47: usuario_roles.id_rol → roles.id_rol

Cada `usuario_roles` referencia exactamente un `roles`. Cada `roles` admite cero o muchos `usuario_roles` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Roles activos asignables; usuario admite varios roles, sin borrar silenciosamente los adicionales.

### F48: usuario_roles.id_usuario → usuarios.id_usuario

Cada `usuario_roles` referencia exactamente un `usuarios`. Cada `usuarios` admite cero o muchos `usuario_roles` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Gestión de cuentas/roles autorizada y protección del último acceso administrativo.

### F49: usuarios.id_empleado → empleados.id_empleado

Cada `usuarios` referencia cero o un `empleados`. Cada `empleados` admite cero o un `usuarios` por esta referencia.

SQL: nullable; UNIQUE individual; ON DELETE RESTRICT. 007: UNIQUE nullable y triggers de nombre compatible; sin NOT NULL.

Go: Vinculación laboral verificada con permiso específico; no se infiere por nombre o coincidencia numérica. UNIQUE nullable: 0..1 en ambos sentidos.

### F50: vehiculos.id_cliente → clientes.id_cliente

Cada `vehiculos` referencia exactamente un `clientes`. Cada `clientes` admite cero o muchos `vehiculos` por esta referencia.

SQL: NOT NULL; sin UNIQUE individual; ON DELETE NO ACTION. FK/PK/NULL/UQ según DDL; ver definiciones y restricciones por tabla.

Go: Titular registrado actual; alta/historial y transferencia verificada atómicos. No acredita propiedad legal externa.

## N:M, autorrelación y casos particulares

USUARIOS–ROLES usa usuario_roles con PK compuesta; ROLES–PERMISOS usa rol_permisos con PK compuesta. Cada miembro puede repetirse, la pareja no. REPUESTOS–PROVEEDORES usa repuesto_proveedor, PK propia y UNIQUE conjunta. ÓRDENES–USUARIOS usa asignaciones_ot con función/vigencia: no se crea orden-empleado directa ni se exige simultaneidad de varios asignados.

F14 es autorrelación de cotizaciones para base/ampliaciones. FK sola no prohíbe ciclos; Go valida las ampliaciones. No se encontró ciclo de FK entre tablas distintas. El historial JSON no es otra tabla/puente inventada.

F20, F23 y F32 tienen máximo uno por UNIQUE individual. F49 es opcional en ambos extremos y única por empleado cuando está informado. F30 sigue siendo muchos históricamente, pese al índice único del período vigente. Pagos usa UNIQUE parcial de (orden,request_id); F34 sigue permitiendo pagos parciales múltiples.

Los mínimos de hijos SQL son cero. Go crea al menos un detalle/asignación/estado inicial mediante transacciones, sin convertir esos mínimos en restricciones SQL inexistentes. Modalidad directa y diagnóstico final de misma recepción también están en triggers 003/004, no solo en Go.

## Verificación

La revisión original comprobó tablas/columnas, PK, 50 FK y sus columnas, metadata y símbolos NULL/UQ, puentes, autorrelación, geometría, protección de archivos y PDF. El motor oficial draw.io local abre XML nativo y conserva una edición temporal de una tabla; no se afirma sesión manual del editor web.

Principal único y completo; vistas por bloque son recortes complementarios. PDF A0 horizontal de una sola página, sin compresión A4. Ampliar en pantalla o imprimir en formato amplio.
