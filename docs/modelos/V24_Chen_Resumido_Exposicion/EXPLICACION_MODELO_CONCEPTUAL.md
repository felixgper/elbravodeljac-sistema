# Modelo conceptual Chen V24 - explicación

## Propósito y alcance

Fecha: 2026-11-02. Lámina académica de alto nivel, derivada del V23 original editable. No sustituye el modelo lógico/físico ni certifica que propuestas anteriores hayan sido migradas. Se conservan 13 entidades principales y 15 relaciones en un lienzo. Rectángulos y rombos nativos editables; sin atributos, identificadores, columnas ni tipos. Se mantiene fondo claro y azul contrastado para proyección e impresión económica; no se dispuso de un archivo MI FINCA en esta solicitud y se toma únicamente su simplicidad descrita.

La lectura empieza arriba de izquierda a derecha y continúa desde COTIZACIONES hacia ENTREGAS de derecha a izquierda. Así se evita una franja horizontal excesiva. Los recursos y empleados quedan debajo. Los pares mínimo-máximo describen a la entidad junto a su etiqueta, no a la opuesta. El 0 permite registros iniciales y procesos que no llegan a la entrega; N no significa simultaneidad.

## Fuentes revisadas y alcance de la comprobación

- V23: `docs/modelos/V23_Chen_Corregido/El_Bravo_del_JAC_V23_Chen_Corregido.drawio`, su especificación y auditoría. Se contrastaron los 27 rectángulos y las 42 relaciones.
- V22 corregido: archivo editable leído con `git show` desde `docs/modelo-conceptual-chen-v21`, sin cambiar de rama. Referencia, huella y recuentos en FUENTES_REVISADAS.json.
- Modelo lógico V20 Crow's Foot: XML original leído desde `docs/propuesta-modelo-v20`; su informe lo identifica como propuesta. No se confunde ese archivo histórico con la implementación posterior. No se modificó ni se reconstruyó.
- Esquema base `database/postgresql/esquema_actual.sql`, 26 tablas originales; migraciones versionadas 001–009 y su registro en `backend/internal/platform/postgres/migrate.go`. No se encontró ninguna migración posterior a 009 en el repositorio.
- Go: domain/workshop.go y relational.go, postgres/store.go, workshop_writes.go, diagnosis_workflow.go, relational.go, identity.go y documents.go. Formularios y funcionalidades referenciados por la auditoría V23; no se modificaron.
- Informes `docs/backend/EMPLEADOS.md` y `CATALOGO_DOCUMENTOS_TRES_ROLES.md`: documentan ejecución previa de 007 y de la actualización documental/roles en entornos aislados. Son evidencia histórica disponible, **no una nueva comprobación de despliegue**. Esta tarea no consultó PostgreSQL ni MongoDB y no certifica qué migraciones están aplicadas hoy en una instancia concreta.

## Entidades conservadas

- CLIENTES: clientes.
- VEHÍCULOS: vehiculos.
- RECEPCIONES: recepciones.
- DIAGNÓSTICOS: diagnosticos.
- COTIZACIONES: cotizaciones.
- ÓRDENES DE TRABAJO: ordenes_trabajo.
- PAGOS: pagos.
- ENTREGAS: entregas.
- AUXILIOS MECÁNICOS: auxilios.
- EMPLEADOS: empleados.
- SERVICIOS: servicios.
- REPUESTOS: repuestos.
- PROVEEDORES: proveedores.

EMPLEADOS representa a la persona trabajadora y su cargo laboral, no su cuenta ni su permiso. No se crean entidades MECÁNICOS ni RECEPCIONISTAS. La administradora realiza la recepción como función del negocio; el software conserva el usuario autenticado que la registra. Un campo `id_recepcionista` no se convierte en entidad.

## Entidades de V23 omitidas solamente de esta vista

| Entidad omitida | Motivo y conservación |
| --- | --- |
| CONTACTOS DEL CLIENTE | Detalle de personas autorizadas; permanece `contactos_cliente` y autorización específica de entrega. No se convierte al contacto en propietario. |
| HISTORIAL DE PROPIETARIOS | Detalle temporal de titularidad; permanece `historial_propietarios`, período vigente y verificación del traspaso. S01 muestra solo titular actual registrado. |
| TIPOS DE DOCUMENTO | Catálogo de validación que sobrecarga la exposición. **TIPOS_DOCUMENTO continúa implementada en PostgreSQL mediante 008, con sus relaciones a clientes y contactos; no se elimina ninguna.** No se afirma haber migrado producción. |
| UNIDADES DE AUXILIO | Recurso de asignación; permanece `unidades_auxilio` y la referencia del incidente. |
| RESULTADOS DE ESCÁNER | Documento técnico fuera del flujo general; permanece MongoDB con referencia al diagnóstico y autorización Go. |
| EVIDENCIAS TÉCNICAS | Detalle documental; permanece MongoDB/GridFS y su referencia al diagnóstico. No se traslada a PostgreSQL. |
| DETALLES DE COTIZACIÓN | Se resume en S13/S14; permanece `detalle_cotizacion` con descripción, cantidades, precios, descuentos y exclusión de referencias. |
| ASIGNACIONES A ÓRDENES | Se resume en S11; permanece `asignaciones_ot`, responsables y reasignaciones, sin alterar el historial. |
| AVANCES DE REPARACIÓN | Actividad interna de la orden; permanece `avances_ot` y sus autores. |
| PRUEBAS DE REPARACIÓN | Control de ejecución; permanece `pruebas_ot` y las precondiciones de finalización. |
| HISTORIAL DE ESTADOS | Trazabilidad interna; permanece `historial_estados_ot`. |
| USUARIOS | Cuenta de autenticación, omitida para concentrarse en negocio. Permanece `usuarios`, JWT, cookies/CSRF y actores auditados. |
| ROLES | Política de acceso; permanece `roles` y la relación N:M `usuario_roles`. Cargo laboral no equivale a rol. |
| PERMISOS | Detalle de seguridad; permanece `permisos`, `rol_permisos` y verificaciones en Go. |

Las tablas puente `repuesto_proveedor`, `usuario_roles` y `rol_permisos` también continúan. El inventario conceptual no es el recuento físico de tablas. Marcas de vehículos, datos del taller y movimientos de pagos son vistas/funcionalidades existentes, no módulos nuevos que deban dibujarse.

## Relaciones y lectura de ambos extremos

El asterisco de S10–S14 identifica una proyección conceptual resumida por relaciones existentes. No constituye una relación física directa creada para ahorrar líneas. Sus caminos están documentados debajo; el modelo lógico conserva todos los intermediarios.

### S01: CLIENTES — TIENE TITULAR REGISTRADO — VEHÍCULOS

- CLIENTES (0,N): Un cliente puede tener cero o muchos vehículos registrados a su nombre.
- VEHÍCULOS (1,1): Cada vehículo tiene exactamente un titular actual registrado.

No certifica propiedad legal externa. El historial de titulares se conserva fuera de la vista.

Evidencia: `vehiculos.id_cliente; historial_propietarios; postgres/store.go`.

### S02: VEHÍCULOS — INGRESA — RECEPCIONES

- VEHÍCULOS (0,N): Un vehículo puede tener cero o muchas recepciones durante su historial.
- RECEPCIONES (1,1): Cada recepción corresponde exactamente a un vehículo.

La recepción conserva cliente de atención y kilometraje.

Evidencia: `recepciones.id_vehiculo; CreateReception`.

### S03: RECEPCIONES — REQUIERE — DIAGNÓSTICOS

- RECEPCIONES (0,1): Una recepción puede requerir cero o un diagnóstico.
- DIAGNÓSTICOS (1,1): Cada diagnóstico pertenece exactamente a una recepción.

Índice uq_diagnosticos_id_recepcion fija máximo uno. Directo no admite diagnóstico; una atención técnica pendiente puede carecer todavía de él.

Evidencia: `esquema_actual.sql: uq_diagnosticos_id_recepcion; migraciones 003/004`.

### S04: RECEPCIONES — GENERA — COTIZACIONES

- RECEPCIONES (0,N): Una recepción puede generar cero o muchas cotizaciones.
- COTIZACIONES (1,1): Cada cotización procede exactamente de una recepción.

Incluye base y ampliaciones. La API controla duplicación de bases; no significa varias bases libres. Atención directa no exige diagnóstico.

Evidencia: `cotizaciones.id_recepcion; migraciones 003/005; SaveQuote`.

### S05: DIAGNÓSTICOS — SUSTENTA — COTIZACIONES

- DIAGNÓSTICOS (0,N): Un diagnóstico puede sustentar cero o muchas cotizaciones.
- COTIZACIONES (0,1): Una cotización puede tener cero o un diagnóstico.

Si es técnica, el diagnóstico debe estar finalizado y ser de la misma recepción; si es directa, no lo lleva. No es una segunda vía independiente de S04.

Evidencia: `SaveQuote; jac_validar_origen_cotizacion (003/004)`.

### S06: COTIZACIONES — AUTORIZA — ÓRDENES DE TRABAJO

- COTIZACIONES (0,1): Una cotización puede autorizar cero o una orden.
- ÓRDENES DE TRABAJO (1,1): Cada orden nace exactamente de una cotización aprobada.

Cotizaciones pendientes/rechazadas admiten cero órdenes. uq_ordenes_trabajo_id_cotizacion impide duplicación. Revisión y aprobación explícitas; una ampliación no autoriza duplicar la orden base.

Evidencia: `CreateOrder; uq_ordenes_trabajo_id_cotizacion; migración 005`.

### S07: ÓRDENES DE TRABAJO — RECIBE — PAGOS

- ÓRDENES DE TRABAJO (0,N): Una orden puede recibir cero o muchos pagos.
- PAGOS (1,1): Cada pago corresponde exactamente a una orden.

Anticipos, rectificaciones y saldo neto se controlan aparte. La cardinalidad no equivale a cancelación total ni registra ingresos desde cotizaciones pendientes.

Evidencia: `pagos.id_orden; migración 006; jac_movimientos_pagos`.

### S08: ÓRDENES DE TRABAJO — CULMINA EN — ENTREGAS

- ÓRDENES DE TRABAJO (0,1): Una orden puede culminar en cero o una entrega.
- ENTREGAS (1,1): Cada entrega corresponde exactamente a una orden.

Orden finalizada, trabajos conciliados, saldo neto cero, identidad/autorización y conformidad verificadas. No existe aquí un vínculo inventado PAGO–ENTREGA: ambos dependen de la orden.

Evidencia: `uq_entregas_id_orden; migración 006; CreateDelivery`.

### S09: CLIENTES — SOLICITA — AUXILIOS MECÁNICOS

- CLIENTES (0,N): Un cliente puede solicitar cero o muchos auxilios a lo largo del tiempo.
- AUXILIOS MECÁNICOS (1,1): Cada auxilio nuevo debe tener exactamente un cliente identificado.

La columna física id_cliente_solicitante es nullable. El mínimo uno es regla de altas por API, no una garantía sobre registros antiguos. N significa incidentes históricos; no simultaneidad.

Evidencia: `auxilios.id_cliente_solicitante; AuxilioInput.Validate; SaveAuxilio`.

### S10: EMPLEADOS — ATIENDE * — AUXILIOS MECÁNICOS

- EMPLEADOS (0,N): Un empleado puede atender cero o muchos auxilios en su permanencia.
- AUXILIOS MECÁNICOS (1,1): Cada auxilio nuevo tiene un responsable técnico.

Proyección laboral de auxilios.id_tecnico → usuarios → empleados. La referencia técnica física es nullable y usuario-empleado también; mínimo uno es responsabilidad de negocio, pendiente de asegurar para todo legado. No se habilitan técnicos sin cuenta por dibujar este vínculo.

Evidencia: `SaveAuxilio; technicalUser; migración 007`.

### S11: EMPLEADOS — REALIZA * — ÓRDENES DE TRABAJO

- EMPLEADOS (0,N): Un empleado puede participar en cero o muchas órdenes.
- ÓRDENES DE TRABAJO (1,N): Una orden nueva debe contar con uno o varios trabajadores responsables mediante sus asignaciones.

Relación resumida, no FK directa: orden → asignaciones_ot → usuario técnico → empleado. CreateOrder inserta un mecánico atómicamente. La vinculación laboral completa no está garantizada por SQL nullable; no afirmar regularización del legado. El N permite diferentes participantes históricos, no exige varios simultáneos.

Evidencia: `CreateOrder; insertAssignment; asignaciones_ot; migraciones 007/009`.

### S12: EMPLEADOS — REALIZA EVALUACIÓN * — DIAGNÓSTICOS

- EMPLEADOS (0,N): Un empleado puede realizar cero o muchas evaluaciones diagnósticas.
- DIAGNÓSTICOS (1,1): Cada diagnóstico tiene un responsable técnico.

Proyección de diagnosticos.id_tecnico → usuarios → empleados. El usuario autor del historial puede ser diferente. El mínimo laboral expresa negocio, no garantiza que todas las cuentas históricas tengan empleado.

Evidencia: `SaveDiagnosis; diagnosis_workflow.go; migración 007`.

### S13: COTIZACIONES — INCLUYE SERVICIO * — SERVICIOS

- COTIZACIONES (0,N): Una cotización puede incluir cero o muchos servicios del catálogo.
- SERVICIOS (0,N): Un servicio del catálogo puede aparecer en cero o muchas cotizaciones.

Derivada de detalle_cotizacion, no una nueva FK cotización-servicio. Puede no usar catálogo si todos sus conceptos son libres tipificados. No declara que un servicio cotizado ya fue realizado.

Evidencia: `detalle_cotizacion.id_servicio; QuoteInput.Validate; migración 001`.

### S14: COTIZACIONES — INCLUYE REPUESTO * — REPUESTOS

- COTIZACIONES (0,N): Una cotización puede incluir cero o muchos repuestos del catálogo.
- REPUESTOS (0,N): Un repuesto del catálogo puede aparecer en cero o muchas cotizaciones.

Derivada de detalle_cotizacion. Cero porque admite solo servicios o repuestos libres. Por línea: servicio o repuesto, nunca ambas referencias; concepto libre explícito. Cantidades y precios permanecen en el detalle lógico/físico.

Evidencia: `detalle_cotizacion.id_repuesto; QuoteInput.Validate; migración 001`.

### S15: PROVEEDORES — SUMINISTRA — REPUESTOS

- PROVEEDORES (0,N): Un proveedor puede suministrar cero o muchos repuestos.
- REPUESTOS (0,N): Un repuesto puede tener cero o muchos proveedores.

N:M mediante repuesto_proveedor, con costo de referencia. Representa la relación de abastecimiento, no prueba una compra o movimiento de stock efectuado.

Evidencia: `repuesto_proveedor; LinkProviderPart`.

## Reglas que permanecen fuera del dibujo

Una cotización válida tiene de 1 a 100 líneas aunque S13 y S14 admitan cero servicios/repuestos catalogados: las líneas libres tipificadas son válidas. Una línea puede referenciar un servicio, un repuesto o ninguno si es libre; nunca ambos. Las operaciones realizadas y los importes conciliados pertenecen a la ejecución de la orden, no al simple hecho de cotizar.

Se conservan revisiones, huella y aprobación; la autorrelación de ampliaciones se omite visualmente para no sobrecargar. Cada ampliación mantiene su recepción/base y autorización según Go; el máximo una orden por cotización permanece. No se afirma una nueva orden obligatoria por revisión.

Cliente solicitante y titular actual pueden diferir históricamente por traspaso: la recepción conserva su cliente. La API de alta requiere el titular registrado en ese momento; V24 no elimina esa identidad histórica por omitir su línea directa. Un auxilio mantiene cliente, vehículo, unidad y posible recepción. Se omite su vínculo con recepción porque V23 P3 aún requiere confirmar el máximo inverso: la base permite varios auxilios por recepción. No se adopta unicidad 1:1.

S10, S11 y S12 muestran responsabilidad humana de negocio. **No se presenta como implementada la obligatoriedad de vincular cada cuenta con EMPLEADOS.** La FK de usuarios es nullable y única, y las FK operativas siguen hacia usuarios técnicos. Las cuentas históricas sin empleado no quedan regularizadas mediante el dibujo. El registrador administrativo puede ser distinto del ejecutor; la administradora no se vuelve mecánica por capturar información. Las operaciones nuevas requieren cuenta técnica habilitada; no se permite asignar directamente a cualquier empleado sin ella.

No se inventa una entidad TIPO_EMPLEADO: el cargo está en empleados y se omite como atributo, coherente con esta presentación. N:M de roles permite permisos adicionales sin cambiar cargo laboral. No se representan GRANT de PostgreSQL.

## Formatos y verificación

Draw.io contiene 13 rectángulos, 15 rombos y 30 conectores, todos nativos. Una página; cero atributos/óvalos. Geometría revisada: cero solapamientos de formas y cero conectores atravesando formas/etiquetas; tres cruces computados, sin afirmar ausencia total ni óptimo global.

Se abrió el XML con el motor oficial GraphViewer/mxGraph de draw.io en Chrome local, se exportó SVG/PNG y se comprobó una modificación temporal de CLIENTES sin reemplazar el original. El PDF se exportó desde esa representación vectorial, no desde una captura: una página A3 horizontal con márgenes, verificable con pypdf y renderizado Poppler. Se inspeccionó la imagen de la página completa. No es una sesión manual del editor remoto.

Los hashes de 234 archivos versionados previos se conservan en PROTECCION_ORIGINALES.json; la validación final verifica que siguen intactos. Se añaden documentos solo dentro de esta carpeta. No se ejecutaron migraciones, SQL, limpieza, pruebas de aplicación ni operaciones Docker. No se hizo commit, push ni merge.

Para exponer: utilizar PDF a pantalla completa; para editar: abrir el .drawio en draw.io Desktop o Archivo → Abrir desde → Dispositivo. Para imprimir se recomienda A3 horizontal, ajustar a una página; A4 reduce legibilidad de las cardinalidades.
