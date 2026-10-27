# Correspondencia entre formularios y PostgreSQL

Inspección de `dist/assets/js/views/*.js`, `services/*.js`, `data/permissions.js`
y las definiciones reales de `database/postgresql/esquema_actual.sql`.
No se presupone que los datos demo existan en PostgreSQL.

## Mapa de campos y operaciones

| Formulario / fuente | Campos de la interfaz → almacenamiento | Observaciones |
|---|---|---|
| Clientes / clients-vehicles.js | name→clientes.nombre; documentType→tipo_documento; documentNumber→numero_documento; phone→telefono; email→correo; address→direccion | Documento compuesto único. Correo vacío se convierte en NULL. |
| Vehículos / clients-vehicles.js | clientId→vehiculos.id_cliente; plate→placa; brand→marca; model→modelo; year→anio; mileage→kilometraje_actual; notes→observaciones | Placa normalizada; no cambiar identidad al transferir. |
| Transferencia / clients-vehicles.js | newClientId→vehiculos.id_cliente y historial_propietarios.id_cliente; transferVerified→verificado | Cerrar historial vigente; anterior propietario se deriva, usuario/fecha los fija el servidor. |
| Recepción / operations.js | clientId→recepciones.id_cliente_atencion; vehicleId→id_vehiculo; mileage→km_ingreso; symptoms→sintomas | id_recepcionista y fecha_ingreso del servidor. Actualizar km del vehículo en la misma transacción. |
| Diagnóstico / operations.js | receptionId→diagnosticos.id_recepcion; findings→resumen | id_tecnico y fecha del servidor; una recepción tiene como máximo un diagnóstico por índice único. |
| Cotización / operations.js | diagnosisId→cotizaciones.id_diagnostico; status→estado; items→detalle_cotizacion | concept→descripcion, quantity→cantidad, unitPrice→precio_unitario; referencias faltantes descritas abajo. Total/subtotales calculados por Go. |
| Nueva OT / operations.js | quoteId→ordenes_trabajo.id_cotizacion; mechanicId/assistantId→asignaciones_ot.id_usuario + rol_en_trabajo | Código/fechas/creador del servidor; nombres y trabajos se derivan de la cotización. |
| Actualizar OT y Mis tareas / operations.js | state→ordenes_trabajo.estado; note→avances_ot.observacion; test→pruebas_ot.tipo_prueba; result→resultado | Registrar también historial_estados_ot. Updates demo mezcla avance, prueba y estado: separar persistencia, conservar una operación atómica. |
| Pago / payments.js | orderId→pagos.id_orden; amount→monto; method→metodo_pago | monto decimal positivo; comprobante opcional sin control actual; usuario/fecha del servidor. |
| Nuevo contacto / payments.js | clientId→contactos_cliente.id_cliente; name/documentType/documentNumber/phone/relation→nombre/tipo_documento/numero_documento/telefono/relacion | Alta no autoriza entrega; vincular al cliente de la atención. |
| Entrega / payments.js | orderId→entregas.id_orden; kind→tipo_receptor; contactId→id_contacto_receptor; verified→autorizacion_verificada; method→medio_verificacion | Snapshots de nombre/documento/teléfono desde cliente/contacto; verificador y fecha desde servidor. Entrega única por OT. |
| Auxilio / operations.js | clientId→auxilios.id_cliente_solicitante; vehicleId→id_vehiculo; location→ubicacion; incident→tipo_incidente; description→descripcion; technicianId→id_tecnico | unit textual requiere resolver id_unidad_auxilio. id_recepcion opcional no tiene control actual. |
| Proveedor / administration.js | name→proveedores.nombre; ruc→documento; phone→telefono | parts no tiene equivalente directo; contacto SQL no tiene control en la interfaz. |
| Usuario / administration.js | name→usuarios.nombre; username→usuario; email→correo; active→activo; password→hash del servidor | role→usuario_roles→roles; nunca enviar password_hash. |
| Roles y permisos / administration.js | role→roles; matriz modulo/acción→permisos.codigo y rol_permisos | SQL permite varios roles por usuario; acordar proyección de un único rol para la UI. |
| Mecánicos / administration.js | consulta usuarios + usuario_roles + roles + asignaciones_ot + ordenes_trabajo | No hay tabla mecánicos ni hay que crear una duplicada. |
| Inicio y Reportes / app.js y administration.js | agregaciones de pagos, OT, detalles, servicios, recepciones y auxilios | Calcular desde SQL, no usar seed/demo ni sumar pagos con joins que dupliquen filas. |

## Diferencias que deben resolverse antes de conectar

1. **Cotizaciones:** UI permite concepto libre + tipo Servicio/Repuesto. SQL exige
   exactamente id_servicio o id_repuesto. Contrato propuesto transporta serviceId
   o sparePartId además de concept. El futuro adaptador resuelve/crea un catálogo
   con autorización explícita o solicita selección; no deducir IDs por nombre
   ambiguo ni desactivar el CHECK. Hoy no se altera el formulario. Los datos demo
   no contienen estas referencias y no se importarán automáticamente.
2. **Proveedor.parts:** es texto libre de especialidad/repuestos; repuesto_proveedor
   necesita repuesto concreto y costo_referencia. No equivale a proveedores.contacto.
   Propuesta pendiente: columna especialidad mediante migración revisada, o modelo
   explícito de catálogo si se acuerda. No perder ni guardar ese campo en contacto.
3. **Roles:** un select en la UI frente a relación N:M en SQL. Propuesta MVP: un rol
   administrado por usuario; no borrar asignaciones múltiples existentes. Si existen,
   detener edición incompatible y acordar rol principal/múltiple antes de integrar.
4. **Límites:** UI acepta nombres de cliente/proveedor/usuario/contacto de 180 y SQL
   150; teléfonos 25 frente a 20; correo 254 frente a 150; prueba 200 frente a 100.
   API debe validar sin truncar y devolver errores por campo. Acordar ampliar SQL o
   ajustar validación del frontend en una fase futura, manteniendo su diseño.
5. **Estados y etiquetas:** UI OT Pendiente/En reparación/En pruebas/Finalizada/Cancelada;
   SQL inicia en creada y no tiene catálogo cerrado de estados. Propuesta de DTO
   conserva etiquetas UI, traduce Pendiente→creada, En reparación→en_reparacion,
   En pruebas→en_pruebas, Finalizada→finalizada, Cancelada→cancelada. Revisar valores
   reales antes de fijar traducción. La entrega es registro separado, no sobrescribir
   estado arbitrariamente. Cotización Pendiente/Aprobada/Rechazada→minúsculas;
   origen Taller/Cliente→TALLER/CLIENTE; No aplica no genera prueba. Resultado y
   tipo_receptor/metodo_pago también necesitan traducción explícita, no capitalización
   automática. SQL solo restringe algunos valores, no confirma los catálogos vivos.
6. **Unidades de auxilio:** UI Unidad 01/Unidad 02 es texto fijo. Resolver desde
   unidades_auxilio (ID y código reales), sin asumir que esas filas ya existen.
7. **Identidad/precisión:** demo usa IDs como c1/u-admin/OT-060. SQL usa bigint y
   codigo_ot separado. API devuelve IDs como strings decimales y code visible para OT.
   Importes, cantidades y km como strings decimales de dos posiciones; no float64
   para dinero ni conversión directa de amountCents a monto. Go recalcula totales.
8. **Evidencias:** no hay carga de archivos/escáner en la UI actual ni referencia
   MongoDB en SQL. Vincular documentos por diagnosticoId SQL; no hace falta duplicar
   clientes/vehículos en MongoDB ni modificar SQL para esta relación de lectura.
9. **Fechas/auditoría:** en demo las pone el navegador. API ignora actor/fechas de
   auditoría del cliente, usa sesión y reloj del servidor. Transporta RFC3339 UTC y
   presenta America/Lima. Año UI 1950..año+1 frente a CHECK SQL 1900..2100: conservar
   regla funcional más restrictiva y documentar, no asumir que ambos límites son iguales.

## Resultado

Las ocho áreas principales tienen soporte relacional y formularios existentes.
La integración no es directa al 100 %: las diferencias anteriores requieren DTO,
traducciones y decisiones de datos. Esta preparación no cambia interfaz ni esquema.
