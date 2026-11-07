# Comparación conceptual, lógico y físico

## Referencias oficiales

[Conceptual V24](../V24_Chen_Resumido_Exposicion/El_Bravo_del_JAC_V24_Chen_Resumido.drawio), [lógico sin tipos](../Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.drawio) y referencia física con tipos de esta carpeta. La [comparación real](../COMPARACION_BASE_ACTUAL.md) documenta el contraste con la instancia del 5/11/2026. Antecedentes V23/V20 permanecen en Git como procedencia, no como dependencia para abrir los modelos vigentes.

## Niveles y diferencias

| Asunto | Conceptual V24 resumido | Lógico V24 y físico versionado |
| --- | --- | --- |
| Alcance | Negocio y proyecciones; V24 omite detalles | 28 tablas completas, 208 columnas y 50 FK; ninguna tabla omitida |
| PK/FK | No visibles por decisión didáctica | PK simples/compuestas, FK, NULL y UQ individual identificadas; índices conjuntos/parciales documentados |
| Detalle de cotización | V24 conecta servicios/repuestos como abstracción | detalle_cotizacion mantiene referencias opcionales/excluyentes, conceptos libres, cantidades/precios/descuentos |
| Asignaciones y ejecución | V24 resume trabajador–orden | asignaciones_ot conserva id_usuario, función, vigencia y reasignaciones; avances/pruebas/estados conservan sus actores |
| EMPLEADOS–USUARIOS, P1 | Objetivo cuenta interna con un empleado | usuarios.id_empleado nullable y UNIQUE: opcional en ambos sentidos. No se impone NOT NULL |
| Responsables, P2 | Persona empleada y autor registrador separados conceptualmente | FK operativas siguen hacia USUARIOS; empleado se resuelve mediante vínculo opcional. No se intercambian IDs ni se inventa FK directa a empleados |
| Auxilio–recepción, P3 | V23 proponía máximo uno en ambos sentidos | auxilios.id_recepcion nullable sin UNIQUE: una recepción admite varios auxilios; un auxilio tiene cero o una recepción |
| Recepcionista | Función de administración; no entidad | recepciones.id_recepcionista permanece NOT NULL y FK a usuarios; cambio a id_usuario_registro pendiente |
| TIPOS_DOCUMENTO | Omitida en V24 didáctico | Tabla explícita, PK y codigo UNIQUE; clientes/contactos con FK NOT NULL. No se elimina |
| Roles y permisos | Omitidos en V24 para exponer negocio | usuarios, roles, permisos, usuario_roles, rol_permisos íntegros; cargo no equivale a rol |
| MongoDB | V23 muestra referencias lógicas, V24 las omite | Fuera de las tablas PostgreSQL. No se dibuja una FK hacia resultados_escaner ni evidencias_tecnicas |
| Titularidad | Titular registrado y propiedad histórica | vehiculos.id_cliente + historial_propietarios, cliente de atención preservado en recepción; no prueba legal externa |
| Pagos/entrega | Flujo final del negocio | Ambas dependen de orden; no existe FK entrega→pago ni pago→entrega inventada |

## Correspondencia Go y formularios

| Funcionalidad | Tablas / FK | Validaciones y fuente |
| --- | --- | --- |
| Clientes, vehículos, historial | clientes, tipos_documento, vehiculos, historial_propietarios, contactos_cliente | domain/models.go, postgres/store.go, documents.go; views/clients-vehicles.js y payments.js. Catálogo documental, longitudes, traspaso verificado y autorización contextual |
| Recepción y diagnóstico | recepciones y diagnosticos; referencias técnicas y de captura a usuarios | CreateReception, SaveDiagnosis, technicalUser; views/operations.js. Modalidad explícita, técnico asignado, diagnóstico final de misma recepción |
| Cotización y detalles | cotizaciones, detalle_cotizacion, servicios y repuestos | SaveQuote y QuoteInput.Validate. Revisión/aprobación, conceptos libres tipificados, referencias excluyentes y totales calculados por servidor |
| Órdenes y ejecución | ordenes_trabajo, asignaciones_ot, avances_ot, pruebas_ot, historial_estados_ot | CreateOrder/UpdateOrder, transacciones y alcance según asignación. Sin nueva entidad Mecánicos |
| Pagos y entrega | pagos, entregas, contactos_cliente y usuarios | Saldo conciliado, rectificaciones, idempotencia de pagos, identidad/conformidad, snapshot y ficha; views/payments.js |
| Auxilios y catálogos | auxilios, unidades_auxilio, proveedores, repuestos, servicios, repuesto_proveedor | SaveAuxilio/LinkProviderPart; referencias existentes, costo/stock/precio válidos; views/administration.js |
| Empleados y acceso | empleados, usuarios, roles, permisos y sus puentes | SaveEmployee/SaveUser, identity.go, JWT/cookies/CSRF y protect; views/administration.js. Vínculo opcional, sin autoescalamiento, matriz transaccional |

Las rutas están en docs/backend/ENDPOINTS.md y router/handlers; no se reejecutaron ni se afirma que esta revisión pruebe HTTP real. El lógico no modela cada endpoint como una tabla.

## Normalización y desnormalización existente

Las N:M mantienen puentes y claves compuestas/conjuntas, evitando listas de roles o proveedores en columnas de texto. El detalle separa conceptos y sus importes de la cabecera. El historial de titulares mantiene períodos y no reemplaza al titular actual ni al cliente de atención. No se crea identidad duplicada PROPIETARIOS.

Persisten duplicaciones deliberadas: usuarios.nombre compatible con empleados.nombre mediante triggers; etiquetas documentales junto al catálogo; total/subtotal persistidos y validados; snapshots de entrega/ficha e historiales JSONB para preservar evidencia. No se declaran tablas en 3FN/BCNF por inspección superficial: una prueba formal exige todas las dependencias funcionales del dominio. Se describen estas decisiones sin “normalizarlas” alterando el aplicativo.

## Verificación visual

Un solo lienzo horizontal, seis áreas, cada tabla dibujada una vez. Recorrido superior clientes/vehículos → recepción/diagnóstico/cotización → órdenes; pagos/entrega debajo a la derecha. Auxilio/recursos y personal/seguridad ocupan la segunda fila. El orden visual no añade FK consecutivas.

Conectores ortogonales nativos con extremos Crow’s Foot y tooltip por referencia. Se mantuvieron las 50 FK, incluidas las de auditoría. No se declara un grafo sin cruces ni un óptimo global. El PDF es vectorial A0 horizontal, de una página, para conservar legibilidad. El PNG completo permite ampliar bloques sin capturas redundantes.

No se modificaron lógico/físico previos, aplicación, configuración ni bases. La limpieza conserva los editables y exportaciones oficiales sin regenerarlos.
