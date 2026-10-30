# Chen simplificado V22 — El Bravo del JAC

Versión didáctica sin atributos. Una sola página y un único grafo conectado.
Contiene 24 entidades relacionales, las tres N:M como rombos y dos conceptos
documentales MongoDB. No se muestran IDs, campos, tipos de datos ni claves.

## Qué se simplificó

Se conservan las entidades del negocio, sus asociaciones principales, el servicio
directo sin diagnóstico, las cotizaciones previas a la orden, los pagos y la
entrega asociados a la orden, y los responsables técnicos de diagnóstico,
asignaciones y auxilio. EMPLEADOS sigue separado de la cuenta y del acceso RBAC.
El cargo laboral permanece en el sistema, aunque aquí no se dibuja como atributo.

Se omiten conexiones de autoría y auditoría repetitivas: quién registró cada
ingreso, pago, entrega, prueba o cambio. Los eventos e historiales sí permanecen.
También se abstraen la autoría documental, la revisión de cotizaciones y las
dos conexiones adicionales de asignación de diagnóstico en recepción. No se
elimina ninguna relación del sistema: esta es una vista conceptual simplificada,
no un reemplazo del esquema físico ni una promesa de cobertura de todas sus FK.

## Participaciones y lectura

Cada etiqueta está junto a la entidad a la que pertenece y expresa mínimo:máximo:
`0:N` admite cero o varios, `1:1` exige exactamente uno y `0:1` admite uno opcional.
Una N:M se representa con `0:N` en ambos extremos. Las reglas condicionales no
se convierten en mínimos universales que los datos históricos no garantizan.
Las líneas discontinuas representan vínculos documentales MongoDB, sin FK
entre motores. Los pequeños sobrepasos distinguen cruces de uniones.

Servicio directo omite diagnóstico, pero conserva recepción y cotización.
Crear una orden exige cotización aprobada. La entrega corresponde a una orden
finalizada y con saldo neto cero; no se inventa una relación pago–entrega.
Las funciones de mecánico y auxiliar se mantienen dentro de las asignaciones;
no se confunden con los cargos laborales o con los roles de acceso.

## Relaciones visibles

| Entidad | Relación | Entidad | Participaciones respectivas |
|---|---|---|---|
| CLIENTES | POSEE | VEHÍCULOS | (0,N) / (1,1) |
| VEHÍCULOS | INGRESA | RECEPCIONES | (0,N) / (1,1) |
| RECEPCIONES | RECIBE DIAGNÓSTICO | DIAGNÓSTICOS | (0,1) / (1,1) |
| DIAGNÓSTICOS | SUSTENTA | COTIZACIONES | (0,N) / (0,1) |
| COTIZACIONES | AUTORIZA ORDEN | ÓRDENES DE TRABAJO | (0,1) / (1,1) |
| ÓRDENES DE TRABAJO | RECIBE PAGO | PAGOS | (0,N) / (1,1) |
| ÓRDENES DE TRABAJO | CULMINA EN ENTREGA | ENTREGAS | (0,1) / (1,1) |
| CLIENTES | DESIGNA CONTACTO | CONTACTOS DEL CLIENTE | (0,N) / (1,1) |
| EMPLEADOS | TIENE CUENTA | USUARIOS | (0,1) / (0,1) |
| USUARIOS | ASUME ROL | ROLES | (0,N) / (0,N) |
| ROLES | CONCEDE PERMISO | PERMISOS | (0,N) / (0,N) |
| PROVEEDORES | ABASTECE | REPUESTOS | (0,N) / (0,N) |
| VEHÍCULOS | CONSERVA TITULARIDAD | HISTORIAL DE PROPIETARIOS | (0,N) / (1,1) |
| RECEPCIONES | PROCEDE DE AUXILIO | AUXILIOS MECÁNICOS | (0,N) / (0,1) |
| ÓRDENES DE TRABAJO | TIENE ASIGNACIÓN | ASIGNACIONES DE PERSONAL | (0,N) / (1,1) |
| SERVICIOS | REFERENCIA SERVICIO | DETALLES DE COTIZACIÓN | (0,N) / (0,1) |
| UNIDADES AUXILIO | SE DESPLIEGA EN | AUXILIOS MECÁNICOS | (0,N) / (0,1) |
| ÓRDENES DE TRABAJO | CONSERVA ESTADOS | HISTORIAL DE ESTADOS | (0,N) / (1,1) |
| COTIZACIONES | CONTIENE CONCEPTO | DETALLES DE COTIZACIÓN | (0,N) / (1,1) |
| USUARIOS | DESEMPEÑA FUNCIÓN | ASIGNACIONES DE PERSONAL | (0,N) / (1,1) |
| DIAGNÓSTICOS | DOCUMENTA ESCÁNER | RESULTADOS DE ESCÁNER | (0,N) / (1,1) |
| CLIENTES | FIGURA COMO TITULAR | HISTORIAL DE PROPIETARIOS | (0,N) / (1,1) |
| ÓRDENES DE TRABAJO | REGISTRA AVANCE | AVANCES DE REPARACIÓN | (0,N) / (1,1) |
| VEHÍCULOS | REQUIERE AUXILIO | AUXILIOS MECÁNICOS | (0,N) / (0,1) |
| RECEPCIONES | ORIGINA COTIZACIÓN | COTIZACIONES | (0,N) / (1,1) |
| REPUESTOS | REFERENCIA REPUESTO | DETALLES DE COTIZACIÓN | (0,N) / (0,1) |
| CLIENTES | SOLICITA ATENCIÓN | RECEPCIONES | (0,N) / (1,1) |
| USUARIOS | DIAGNOSTICA | DIAGNÓSTICOS | (0,N) / (1,1) |
| ÓRDENES DE TRABAJO | INCLUYE PRUEBA | PRUEBAS DE REPARACIÓN | (0,N) / (1,1) |
| DIAGNÓSTICOS | REÚNE EVIDENCIA | EVIDENCIAS TÉCNICAS | (0,N) / (1,1) |
| CLIENTES | SOLICITA AUXILIO | AUXILIOS MECÁNICOS | (0,N) / (0,1) |
| USUARIOS | ATIENDE AUXILIO | AUXILIOS MECÁNICOS | (0,N) / (0,1) |
| CONTACTOS DEL CLIENTE | RECIBE ENTREGA | ENTREGAS | (0,N) / (0,1) |

## Validación

Se validaron el XML nativo, una sola página, extremos conectados, ausencia de
óvalos/atributos, cobertura de todas las entidades, participaciones y geometría.
Los resultados exactos figuran en VALIDACION.json. Se revisa la imagen completa
y ampliaciones antes de entregar. No se afirma haber abierto Draw.io Desktop:
la validación del archivo editable es local. Abrir desde Archivo → Abrir desde
→ Dispositivo en Draw.io y seleccionar el archivo .drawio.

Fuentes: inventario y relaciones contrastados del V21 con las migraciones
001–007 y la rama funcional que incluye EMPLEADOS. V21 queda conservado.
Solo se agregan documentos V22; no se modifican bases, backend, frontend ni
migraciones. Sin commit, push o merge.
