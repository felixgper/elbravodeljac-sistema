# El Bravo del JAC — Modelo conceptual Chen V21

**29/10/2026. Una página, un diagrama, completamente editable.**

Rama documental: `docs/modelo-conceptual-chen-v21`, creada desde
`backend/arquitectura-go-mongodb`, commit `7dd2738`, que ya incluye el merge
de EMPLEADOS (`ac0b12a`). El árbol estaba limpio antes de crear la rama.
Los V19/V20 permanecen sin cambios en `docs/propuesta-modelo-v20`, commit
`086f166`. Se leyeron desde Git sin mezclar ramas ni copiar sus cambios.

## Archivos y apertura

- `El_Bravo_del_JAC_V21_Chen_Completo.drawio`: abrir en Draw.io mediante
  **Archivo → Abrir desde → Dispositivo**. Tiene exactamente una página,
  «Chen completo · lienzo único». Cada figura, texto y conector es editable.
- `El_Bravo_del_JAC_V21_Chen_Completo.png`: vista íntegra de 12600 × 6150 píxeles.
  Ampliar para leer los atributos; no está concebido para reducirse a una hoja A4.
- `vista_general.png` y los PNG `detalle_*`: recortes de revisión visual del
  MISMO lienzo. No son páginas adicionales ni partes necesarias para completarlo.
- `generar_chen.py`, `inventario_actual.json`, `relaciones_chen.json` y
  `VALIDACION.json`: reproducción, cobertura y comprobaciones locales.

El generador utiliza Pillow y solo lee el inventario guardado: no conecta a
bases ni ejecuta migraciones. No incorpora un PNG dentro del `.drawio`.

## Fuentes contrastadas

Se revisaron el respaldo estructural `database/postgresql/esquema_actual.sql`,
las siete migraciones 001–007, V19/V20 y su inventario, y los archivos Go
`workshop_writes.go`, `diagnosis_workflow.go`, `technical_sheet.go`,
`employees.go`, `identity.go`, `technical.go` y el dominio MongoDB.
Las reglas documentadas se contrastaron con la implementación, no solo con
los nombres de las FK. Se consultó únicamente el catálogo de estructura de
`elbravodeljac_dev`, PostgreSQL aislado: **27 tablas, 201 columnas y 48 FK**.
No se leyeron registros personales. Los hashes del SQL consultado están en
`VALIDACION.json`; no se modificó ninguno.

## Cómo explicar el dibujo

Rectángulos azules: entidades. Óvalos individuales: atributos. Identificadores
subrayados: claves de identidad. Rombos: asociaciones verbales. No contiene
tipos SQL, listas de columnas dentro de rectángulos, patas de cuervo ni UML.

Cada par `(mínimo,máximo)` está junto al extremo del conector que llega al rombo
y se refiere a la **entidad conectada por esa línea**: cuántas participaciones
puede tener una instancia de esa entidad en esa relación. Los códigos R01…R49
permiten contrastar los dos extremos con la tabla de este informe. Las líneas
no tienen flechas; sus sobrepasos indican cruces sin unión.

Ejemplo: en POSEE, CLIENTES `(0,N)` significa que un cliente puede no tener
vehículos registrados o tener varios; VEHÍCULOS `(1,1)` exige exactamente
un titular actual. En TIENE CUENTA hay `(0,1)` en ambos extremos. En las
relaciones N:M ambos extremos admiten `(0,N)`.

Los mínimos globales conservadores evitan inventar obligatoriedad para datos
históricos. Una FK obligatoria exige un padre al hijo; no exige hijos a cada
padre. Una FK única limita al padre a un hijo. Los mínimos adicionales
condicionales del negocio se explican más abajo. No se presenta `(1,N)` como
restricción universal si la base y sus registros legados no lo garantizan.

## Cobertura de las 27 tablas físicas

Las tres tablas puente se transforman en relaciones N:M. Las otras 24 conservan
entidad propia. Los detalles, asignaciones y eventos no se colapsan en N:M:
tienen identidad, fechas, estados o información que necesita conservarse.

| Tabla física | Representación Chen | Motivo |
|---|---|---|
| `asignaciones_ot` | Entidad **ASIGNACIONES OT** | Entidad asociativa con atributos propios. |
| `auxilios` | Entidad **AUXILIOS** | Entidad de negocio o evento con identidad propia. |
| `avances_ot` | Entidad **AVANCES OT** | Entidad de negocio o evento con identidad propia. |
| `clientes` | Entidad **CLIENTES** | Entidad de negocio o evento con identidad propia. |
| `contactos_cliente` | Entidad **CONTACTOS CLIENTE** | Entidad de negocio o evento con identidad propia. |
| `cotizaciones` | Entidad **COTIZACIONES** | Entidad de negocio o evento con identidad propia. |
| `detalle_cotizacion` | Entidad **DETALLE COTIZACION** | Entidad asociativa con atributos propios. |
| `diagnosticos` | Entidad **DIAGNOSTICOS** | Entidad de negocio o evento con identidad propia. |
| `empleados` | Entidad **EMPLEADOS** | Entidad de negocio o evento con identidad propia. |
| `entregas` | Entidad **ENTREGAS** | Entidad de negocio o evento con identidad propia. |
| `historial_estados_ot` | Entidad **HISTORIAL ESTADOS OT** | Entidad de negocio o evento con identidad propia. |
| `historial_propietarios` | Entidad **HISTORIAL PROPIETARIOS** | Entidad asociativa con atributos propios. |
| `ordenes_trabajo` | Entidad **ORDENES TRABAJO** | Entidad de negocio o evento con identidad propia. |
| `pagos` | Entidad **PAGOS** | Entidad de negocio o evento con identidad propia. |
| `permisos` | Entidad **PERMISOS** | Entidad de negocio o evento con identidad propia. |
| `proveedores` | Entidad **PROVEEDORES** | Entidad de negocio o evento con identidad propia. |
| `pruebas_ot` | Entidad **PRUEBAS OT** | Entidad de negocio o evento con identidad propia. |
| `recepciones` | Entidad **RECEPCIONES** | Entidad de negocio o evento con identidad propia. |
| `repuesto_proveedor` | Relación N:M **ABASTECE** (R12) | Asociación con costo de referencia; su ID físico es una clave técnica. |
| `repuestos` | Entidad **REPUESTOS** | Entidad de negocio o evento con identidad propia. |
| `rol_permisos` | Relación N:M **CONCEDE PERMISO** (R11) | Puente sin atributos propios; su identidad es el par de participantes. |
| `roles` | Entidad **ROLES** | Entidad de negocio o evento con identidad propia. |
| `servicios` | Entidad **SERVICIOS** | Entidad de negocio o evento con identidad propia. |
| `unidades_auxilio` | Entidad **UNIDADES AUXILIO** | Entidad de negocio o evento con identidad propia. |
| `usuario_roles` | Relación N:M **ASUME ROL** (R10) | Puente sin atributos propios; su identidad es el par de participantes. |
| `usuarios` | Entidad **USUARIOS** | Entidad de negocio o evento con identidad propia. |
| `vehiculos` | Entidad **VEHICULOS** | Entidad de negocio o evento con identidad propia. |

`repuesto_proveedor.costo_referencia` es un óvalo unido a ABASTECE. No se
pierde el costo al sustituir la tabla por la asociación. El ID sustituto `id`
es exclusivamente físico; la pareja proveedor–repuesto es única.

## Relaciones y cardinalidades completas

En esta tabla, el par de cada columna pertenece a la entidad indicada en ella.
Las 45 asociaciones relacionales cubren las 48 FK sin omisiones ni duplicaciones;
cuatro asociaciones documentales son lógicas y se diferencian por línea discontinua.
La numeración se ordena por claridad visual, no por motor.

| Código | Relación | Entidad A y participación | Entidad B y participación | Origen |
|---|---|---|---|---|
| R01 | POSEE | CLIENTES **(0,N)** | VEHICULOS **(1,1)** | `vehiculos.id_cliente` |
| R02 | INGRESA | VEHICULOS **(0,N)** | RECEPCIONES **(1,1)** | `recepciones.id_vehiculo` |
| R03 | RECIBE DIAGNÓSTICO | RECEPCIONES **(0,1)** | DIAGNOSTICOS **(1,1)** | `diagnosticos.id_recepcion` |
| R04 | SUSTENTA | DIAGNOSTICOS **(0,N)** | COTIZACIONES **(0,1)** | `cotizaciones.id_diagnostico` |
| R05 | AUTORIZA ORDEN | COTIZACIONES **(0,1)** | ORDENES TRABAJO **(1,1)** | `ordenes_trabajo.id_cotizacion` |
| R06 | RECIBE PAGO | ORDENES TRABAJO **(0,N)** | PAGOS **(1,1)** | `pagos.id_orden` |
| R07 | CULMINA EN ENTREGA | ORDENES TRABAJO **(0,1)** | ENTREGAS **(1,1)** | `entregas.id_orden` |
| R08 | DESIGNA CONTACTO | CLIENTES **(0,N)** | CONTACTOS CLIENTE **(1,1)** | `contactos_cliente.id_cliente` |
| R09 | TIENE CUENTA | EMPLEADOS **(0,1)** | USUARIOS **(0,1)** | `usuarios.id_empleado` |
| R10 | ASUME ROL | USUARIOS **(0,N)** | ROLES **(0,N)** | Puente `usuario_roles` |
| R11 | CONCEDE PERMISO | ROLES **(0,N)** | PERMISOS **(0,N)** | Puente `rol_permisos` |
| R12 | ABASTECE | PROVEEDORES **(0,N)** | REPUESTOS **(0,N)** | Puente `repuesto_proveedor` |
| R13 | TIENE AMPLIACIÓN | COTIZACIONES **(0,N)** | COTIZACIONES **(0,1)** | `cotizaciones.id_cotizacion_base` |
| R14 | CAMBIA ESTADO | USUARIOS **(0,N)** | HISTORIAL ESTADOS OT **(1,1)** | `historial_estados_ot.id_usuario` |
| R15 | CONSERVA TITULARIDAD | VEHICULOS **(0,N)** | HISTORIAL PROPIETARIOS **(1,1)** | `historial_propietarios.id_vehiculo` |
| R16 | PROCEDE DE AUXILIO | RECEPCIONES **(0,N)** | AUXILIOS **(0,1)** | `auxilios.id_recepcion` |
| R17 | TIENE ASIGNACIÓN | ORDENES TRABAJO **(0,N)** | ASIGNACIONES OT **(1,1)** | `asignaciones_ot.id_orden` |
| R18 | REFERENCIA SERVICIO | SERVICIOS **(0,N)** | DETALLE COTIZACION **(0,1)** | `detalle_cotizacion.id_servicio` |
| R19 | SE DESPLIEGA EN | UNIDADES AUXILIO **(0,N)** | AUXILIOS **(0,1)** | `auxilios.id_unidad_auxilio` |
| R20 | CONSERVA ESTADOS | ORDENES TRABAJO **(0,N)** | HISTORIAL ESTADOS OT **(1,1)** | `historial_estados_ot.id_orden` |
| R21 | CONTIENE CONCEPTO | COTIZACIONES **(0,N)** | DETALLE COTIZACION **(1,1)** | `detalle_cotizacion.id_cotizacion` |
| R22 | DESEMPEÑA FUNCIÓN | USUARIOS **(0,N)** | ASIGNACIONES OT **(1,1)** | `asignaciones_ot.id_usuario` |
| R23 | DOCUMENTA ESCÁNER | DIAGNOSTICOS **(0,N)** | RESULTADOS ESCÁNER **(1,1)** | Referencia lógica MongoDB, sin FK |
| R24 | FIGURA COMO TITULAR | CLIENTES **(0,N)** | HISTORIAL PROPIETARIOS **(1,1)** | `historial_propietarios.id_cliente` |
| R25 | REGISTRA AVANCE | ORDENES TRABAJO **(0,N)** | AVANCES OT **(1,1)** | `avances_ot.id_orden` |
| R26 | REQUIERE AUXILIO | VEHICULOS **(0,N)** | AUXILIOS **(0,1)** | `auxilios.id_vehiculo` |
| R27 | ORIGINA COTIZACIÓN | RECEPCIONES **(0,N)** | COTIZACIONES **(1,1)** | `cotizaciones.id_recepcion` |
| R28 | REFERENCIA REPUESTO | REPUESTOS **(0,N)** | DETALLE COTIZACION **(0,1)** | `detalle_cotizacion.id_repuesto` |
| R29 | SOLICITA ATENCIÓN | CLIENTES **(0,N)** | RECEPCIONES **(1,1)** | `recepciones.id_cliente_atencion` |
| R30 | CREA ORDEN | USUARIOS **(0,N)** | ORDENES TRABAJO **(1,1)** | `ordenes_trabajo.id_usuario_creacion` |
| R31 | DIAGNOSTICA | USUARIOS **(0,N)** | DIAGNOSTICOS **(1,1)** | `diagnosticos.id_tecnico` |
| R32 | REGISTRA CONTROL | USUARIOS **(0,N)** | ORDENES TRABAJO **(0,1)** | `ordenes_trabajo.control_autor` |
| R33 | INCLUYE PRUEBA | ORDENES TRABAJO **(0,N)** | PRUEBAS OT **(1,1)** | `pruebas_ot.id_orden` |
| R34 | INFORMA AVANCE | USUARIOS **(0,N)** | AVANCES OT **(1,1)** | `avances_ot.id_usuario` |
| R35 | REÚNE EVIDENCIA | DIAGNOSTICOS **(0,N)** | EVIDENCIAS TÉCNICAS **(1,1)** | Referencia lógica MongoDB, sin FK |
| R36 | SOLICITA AUXILIO | CLIENTES **(0,N)** | AUXILIOS **(0,1)** | `auxilios.id_cliente_solicitante` |
| R37 | CAPTURA RESULTADO | USUARIOS **(0,N)** | RESULTADOS ESCÁNER **(1,1)** | Referencia lógica MongoDB, sin FK |
| R38 | ASIGNA DIAGNÓSTICO | USUARIOS **(0,N)** | RECEPCIONES **(0,1)** | `recepciones.diagnostico_asignado_por` |
| R39 | ASUME DIAGNÓSTICO | USUARIOS **(0,N)** | RECEPCIONES **(0,1)** | `recepciones.id_mecanico_diagnostico` |
| R40 | REGISTRA INGRESO | USUARIOS **(0,N)** | RECEPCIONES **(1,1)** | `recepciones.id_recepcionista` |
| R41 | REGISTRA PAGO | USUARIOS **(0,N)** | PAGOS **(1,1)** | `pagos.id_usuario_registro` |
| R42 | INCORPORA EVIDENCIA | USUARIOS **(0,N)** | EVIDENCIAS TÉCNICAS **(1,1)** | Referencia lógica MongoDB, sin FK |
| R43 | REALIZA PRUEBA | USUARIOS **(0,N)** | PRUEBAS OT **(1,1)** | `pruebas_ot.id_usuario` |
| R44 | ATIENDE AUXILIO | USUARIOS **(0,N)** | AUXILIOS **(0,1)** | `auxilios.id_tecnico` |
| R45 | REGISTRA AUXILIO | USUARIOS **(0,N)** | AUXILIOS **(1,1)** | `auxilios.id_usuario_registro` |
| R46 | EFECTÚA ENTREGA | USUARIOS **(0,N)** | ENTREGAS **(1,1)** | `entregas.id_usuario_entrega` |
| R47 | VERIFICA RECEPTOR | USUARIOS **(0,N)** | ENTREGAS **(1,1)** | `entregas.id_usuario_verificador` |
| R48 | REGISTRA TITULARIDAD | USUARIOS **(0,N)** | HISTORIAL PROPIETARIOS **(1,1)** | `historial_propietarios.id_usuario_registro` |
| R49 | RECIBE ENTREGA | CONTACTOS CLIENTE **(0,N)** | ENTREGAS **(0,1)** | `entregas.id_contacto_receptor` |

El mapa exacto de nombres de restricciones está en `relaciones_chen.json`.
Las tres N:M representan seis FK; las otras 42 asociaciones relacionales
representan una FK cada una: 6 + 42 = **48**. Sumadas las cuatro lógicas MongoDB,
el dibujo tiene **49 rombos y 98 etiquetas de participación**.

## Atributos y nivel conceptual

No se duplican como óvalos las columnas que ya representan una asociación.
Los atributos con identidad propia se subrayan. Los códigos de negocio únicos
adicionales (placa, código OT, documento, nombre de rol, etc.) siguen siendo
atributos; que no estén subrayados no elimina sus restricciones físicas.

El modelo muestra 154 óvalos: 140 corresponden a entidades relacionales,
13 al complemento MongoDB y uno al costo de abastecimiento. Los objetos
compuestos se resumen por su significado para mantener legible todo el lienzo:

- **Aprobación del cliente:** nombre, medio, fecha, versión e importes/conceptos
  aprobados; `aprobacion`. El cliente del vehículo llega a través de RECEPCIONES,
  pero el aprobante consignado no es una FK independiente a CLIENTES.
- **Control técnico:** inspecciones, estados Bien/Atención/Urgente/Sin evaluar,
  neumáticos, fluidos, observaciones, mantenimiento recomendado y ejecución de
  conceptos; `control_tecnico`. Se conserva su historial, autor y fecha.
- **Identidad receptor:** nombre, tipo/número de documento y teléfono históricos
  (`*_receptor_snap`). **Receptor final** y **conformidad** representan sus
  objetos registrados. No se inventan firmas ni personas receptoras.
- **Ficha técnica emitida:** copia de los datos y documento conservado
  (`ficha_snapshot`, `ficha_pdf`). No se crea una nueva tabla conceptual de fichas.
- **Historiales:** versiones anteriores de asignación, diagnóstico, cotización
  y control almacenadas en sus agregados. El historial de estados de OT sí
  conserva su entidad porque es una tabla/evento independiente.
- **Escáner:** equipo (nombre/versión), conjunto de códigos DTC (código y
  descripción) y mediciones; se muestran como atributos de contenido documental.

Estos óvalos representan contenido agregado, no un desglose de todos los
subcampos de almacenamiento. No son rectángulos con columnas. Los anexos
documentan la correspondencia de **cada una de las 201 columnas** para que los
campos técnicos no se confundan con conceptos faltantes.

Se omiten deliberadamente hashes de credenciales/contenido, claves de reintento
y contadores de concurrencia. Son mecanismos de almacenamiento o seguridad,
no nuevas entidades. El nombre de USUARIOS se conserva por compatibilidad con
legado; cuando hay EMPLEADO vinculado, su nombre laboral es el principal y los
triggers 007 los sincronizan. No se incluyen datos personales ni contraseñas.

## Reglas del taller y límites de la cardinalidad general

1. **Servicio directo:** RECEPCIÓN de tipo Directo no tiene DIAGNÓSTICO. La
   relación ORIGINA COTIZACIÓN permite cotizar directamente desde la recepción;
   no se sintetizan diagnósticos. En una atención técnica, cotizar exige el
   diagnóstico finalizado y perteneciente a esa misma recepción. Así, SUSTENTA
   tiene `(0,1)` desde cotización globalmente, con obligatoriedad condicional
   para tipo Diagnostico (migraciones 003/004).
2. **Responsables del diagnóstico:** ASUME DIAGNÓSTICO es la cuenta asignada a
   RECEPCIÓN; ASIGNA DIAGNÓSTICO es la cuenta que hizo la asignación;
   DIAGNOSTICA es la cuenta responsable del diagnóstico. No son tres empleados
   nuevos. Hay como máximo un diagnóstico por recepción, por índice único.
   La API acepta recepciones técnicas todavía sin asignar por compatibilidad;
   no admite diagnosticar hasta que exista responsable válido. No se impone un
   mínimo uno incondicional a esa asignación.
3. **Cotizaciones:** toda cotización corresponde a exactamente una recepción.
   Su orden solo se crea con estado aprobado y evidencia de aprobación vigente
   (`CreateOrder`). Una cotización puede no producir OT y como máximo produce
   una, por unicidad. La ampliación es otra cotización relacionada a su base;
   las revisiones de contenido siguen en el atributo historial/versiones.
4. **Conceptos:** las nuevas cotizaciones de la API exigen al menos un detalle.
   La FK, por sí misma, no obliga a todas las cotizaciones históricas a tener
   hijos: el mínimo general no se endurece sin esa garantía. Cada detalle puede
   referenciar un servicio O un repuesto, nunca ambos; un concepto manual
   validado puede no referenciar ninguno (CHECK de migración 001). No se
   dibuja una dependencia obligatoria del catálogo que el sistema ya no exige.
   Cantidad, precio, descuento y subtotal conservan su significado actual.
5. **Asignaciones:** cada registro tiene una orden, una cuenta, una función,
   fecha y vigencia. `rol_en_trabajo` distingue MECANICO y AUXILIAR; no es el
   rol RBAC ni `tipo_empleado`. Go exige mecánico activo para crear OT y valida
   el auxiliar opcional. Se conservan todas las asignaciones históricas, por
   eso el máximo es N. La base no tiene unicidad universal de mecánico activo
   por orden ni obliga mediante FK a una asignación en todo registro legado.
6. **Estados, avances y pruebas:** se conservan como eventos ligados a OT y
   a la cuenta autora. La finalización exige en Go la última prueba satisfactoria,
   posterior a la última reparación. El control técnico tiene su autor/fecha;
   no se completa automáticamente al pagar.
7. **Pagos y entrega:** una orden puede tener varios pagos, anticipos o ninguno
   todavía. El importe neto descuenta reversiones y la entrega exige saldo cero,
   orden finalizada y verificación de identidad/autorización (006). Hay como
   máximo una entrega por orden. ENTREGAS depende de OT, no de cada pago.
   Un contacto receptor es opcional; también se conservan identidad histórica
   y receptor real. No hay FK independiente para cada nombre dentro de una copia.
8. **Propiedad:** POSEE identifica el titular actual. HISTORIAL PROPIETARIOS
   conserva varios períodos por cliente/vehículo. El índice parcial admite
   como máximo un período abierto por vehículo; no vuelve 1:1 todo el historial
   ni garantiza que siempre exista un período abierto para cada vehículo.
9. **Auxilio:** puede registrarse sin cliente, vehículo, técnico, unidad o
   recepción vinculados todavía. El usuario de registro es obligatorio. No se
   impone en el dibujo una asignación inmediata que el esquema permite posponer.
10. **Empleados y acceso:** TIENE CUENTA es opcional y única en ambos lados.
    El cargo no otorga permisos. Las N:M ASUME ROL y CONCEDE PERMISO conservan
    varias asignaciones. Una cuenta legada/técnica puede no tener empleado.
    Estado laboral y habilitación de cuenta son distintos. Las asignaciones
    históricas siguen relacionadas con USUARIOS, tal como está implementado.
    No se confunden filas de ROLES con GRANT del motor PostgreSQL.

## Discrepancias y decisiones para revisión

- **Auxiliar y permisos:** la política base de Auxiliar es de consulta. Sin
  embargo, `canEditOrder` comprueba asignación activa, sin filtrar su función
  MECANICO/AUXILIAR; el endpoint comprueba además `ordenes.editar`. Por tanto,
  una cuenta auxiliar a la que se conceda ese permiso u otro rol habilitante
  puede editar dentro del alcance permitido. La prohibición absoluta por ser
  auxiliar **no está garantizada** por esa función. El control técnico sí
  comprueba MECANICO o permiso de asignación. El dibujo no cambia la seguridad:
  esta diferencia se presenta para revisión y no se corrige en una rama documental.
- **V19:** refleja 26 tablas, 160 columnas, 42 FK; no incluye EMPLEADOS ni las
  ampliaciones posteriores de atención directa, control, asignación y pagos.
- **V20:** incluyó 27/201/48 como propuesta. EMPLEADOS ya está implementado
  por 007; dejó de ser solo una hipótesis para este V21. La tabla de cuentas
  mantiene el vínculo opcional, no una supuesta obligación global de cuenta humana.
- **Error de nulabilidad del inventario V20:** `cotizaciones.id_recepcion`
  aparecía nullable a pesar de que 003 ejecuta SET NOT NULL. V21 lo corrige a
  `(1,1)` desde COTIZACIONES en ORIGINA COTIZACIÓN. El SQL original no se altera.
- **Unicidades revalidadas:** se comprobaron también los índices únicos separados
  (diagnóstico por recepción, OT por cotización y entrega por OT), no solo
  constraints. Se conservan esos máximos del V20; no son restricciones nuevas.
- **Catálogos/vistas:** marcas, datos del taller y movimientos netos son vistas,
  no tres nuevas entidades persistentes. Especialidad de proveedor no se inventa.
- **MongoDB:** cada resultado/evidencia referencia un diagnóstico y una cuenta
  de captura por identificadores lógicos. Go comprueba existencia/autorización
  en PostgreSQL, pero MongoDB no impone una FK entre motores. Una escritura
  externa que eluda Go podría dejar referencias huérfanas: no se promete
  integridad referencial distribuida automática. GridFS es almacenamiento
  técnico del archivo, no otro cliente, vehículo, orden o pago duplicado.

## Validación y revisión visual

El archivo XML se volvió a analizar y se comprobaron IDs únicos, extremos
existentes, geometrías, una sola página, conectividad y cobertura exacta.
Los 98 pares se derivaron de nulabilidad y unicidad reales, más las N:M y
referencias documentales explícitas. La cobertura se puede repetir con el
inventario sin conectarse a PostgreSQL. No se ejecutaron pruebas de Go ni
migraciones: el sistema no cambió en esta tarea.

Se revisaron visualmente la vista íntegra y ampliaciones de recepción/diagnóstico/
cotización, empleados/cuentas/roles, MongoDB y el flujo de pago/entrega.
Se ajustaron el espacio de salida de los conectores y sus rutas antes de cerrar
la entrega. No hay figuras superpuestas, conectores atravesando figuras, textos
fuera de sus cajas ni objetos cortados por el lienzo, según las verificaciones
geométricas. Las relaciones de auditoría hacen que el grafo no sea una cadena:
existen cruces, representados con sobrepasos, sin unir conectores distintos.
No se afirma que el dibujo sea un grafo plano sin cruces.

**Validación Draw.io realizada: local, del XML nativo mxGraph.** No se encontró
Draw.io Desktop en las rutas habituales y no se subió el modelo a un servicio externo.
No se presenta esta validación como una apertura interactiva del editor. El
archivo se entrega sin comprimir, con figuras nativas y sin recursos remotos.

## Correspondencia de todas las columnas

Esta tabla permite revisar qué se visualiza y qué se abstrae como mecanismo
técnico. No se elimina ninguna columna del sistema.

| Tabla | Óvalos directos | Relaciones (no repetir como óvalos) | Agrupación conceptual / campos técnicos |
|---|---|---|---|
| `asignaciones_ot` | `id_asignacion`, `rol_en_trabajo`, `fecha_asignacion`, `activo` | `id_orden`, `id_usuario` | — |
| `auxilios` | `id_auxilio`, `ubicacion`, `tipo_incidente`, `descripcion`, `fecha`, `estado` | `id_vehiculo`, `id_cliente_solicitante`, `id_usuario_registro`, `id_tecnico`, `id_unidad_auxilio`, `id_recepcion` | — |
| `avances_ot` | `id_avance`, `observacion`, `fecha` | `id_orden`, `id_usuario` | — |
| `clientes` | `id_cliente`, `tipo_documento`, `numero_documento`, `nombre`, `telefono`, `direccion`, `correo` |  | — |
| `contactos_cliente` | `id_contacto`, `nombre`, `tipo_documento`, `numero_documento`, `telefono`, `relacion`, `activo` | `id_cliente` | — |
| `cotizaciones` | `id_cotizacion`, `estado`, `total`, `fecha_emision`, `aprobada_en`, `version`, `aprobacion`, `historial_revision` | `id_diagnostico`, `id_recepcion`, `id_cotizacion_base` | `contenido_hash` |
| `detalle_cotizacion` | `id_detalle`, `descripcion`, `cantidad`, `precio_unitario`, `subtotal`, `origen_repuesto`, `tipo_concepto`, `tipo_repuesto`, `descuento_unitario` | `id_cotizacion`, `id_servicio`, `id_repuesto` | — |
| `diagnosticos` | `id_diagnostico`, `fecha`, `resumen`, `estado`, `evaluacion`, `recomendaciones`, `observaciones`, `finalizado_en`, `actualizado_en`, `historial` | `id_recepcion`, `id_tecnico` | `version` |
| `empleados` | `id_empleado`, `nombre`, `tipo_empleado`, `activo` |  | — |
| `entregas` | `id_entrega`, `fecha_entrega`, `tipo_receptor`, `autorizacion_verificada`, `medio_verificacion`, `fecha_verificacion`, `receptor_real`, `conformidad` | `id_orden`, `id_usuario_entrega`, `id_contacto_receptor`, `id_usuario_verificador` | `nombre_receptor_snap`, `tipo_doc_receptor_snap`, `documento_receptor_snap`, `telefono_receptor_snap`, `ficha_snapshot`, `ficha_pdf` |
| `historial_estados_ot` | `id_historial`, `estado_anterior`, `estado_nuevo`, `fecha` | `id_orden`, `id_usuario` | — |
| `historial_propietarios` | `id_historial`, `fecha_inicio`, `fecha_fin`, `verificado`, `observacion` | `id_vehiculo`, `id_cliente`, `id_usuario_registro` | — |
| `ordenes_trabajo` | `id_orden`, `codigo_ot`, `estado`, `fecha_creacion`, `fecha_inicio`, `fecha_fin`, `control_tecnico`, `control_fecha`, `control_historial` | `id_cotizacion`, `id_usuario_creacion`, `control_autor` | `control_version`, `flujo_version` |
| `pagos` | `id_pago`, `monto`, `metodo_pago`, `fecha_pago`, `comprobante`, `revertido_en`, `rectificacion` | `id_orden`, `id_usuario_registro` | `request_id` |
| `permisos` | `id_permiso`, `codigo`, `descripcion` |  | — |
| `proveedores` | `id_proveedor`, `nombre`, `documento`, `telefono`, `contacto` |  | — |
| `pruebas_ot` | `id_prueba`, `tipo_prueba`, `resultado`, `observaciones`, `fecha` | `id_orden`, `id_usuario` | — |
| `recepciones` | `id_recepcion`, `fecha_ingreso`, `km_ingreso`, `sintomas`, `tipo_atencion`, `diagnostico_asignado_en`, `historial_asignaciones` | `id_vehiculo`, `id_cliente_atencion`, `id_recepcionista`, `id_mecanico_diagnostico`, `diagnostico_asignado_por` | `asignacion_version` |
| `repuesto_proveedor` | `costo_referencia` | `id_repuesto`, `id_proveedor` | `id (identidad técnica de la asociación)` |
| `repuestos` | `id_repuesto`, `nombre`, `stock`, `precio_referencia`, `activo` |  | — |
| `rol_permisos` |  | `id_rol`, `id_permiso` | — |
| `roles` | `id_rol`, `nombre_rol` |  | — |
| `servicios` | `id_servicio`, `nombre`, `precio_referencia`, `activo` |  | — |
| `unidades_auxilio` | `id_unidad_auxilio`, `codigo`, `descripcion`, `activo` |  | — |
| `usuario_roles` |  | `id_usuario`, `id_rol` | — |
| `usuarios` | `id_usuario`, `nombre`, `usuario`, `correo`, `activo` | `id_empleado` | `password_hash` |
| `vehiculos` | `id_vehiculo`, `placa`, `marca`, `modelo`, `anio`, `kilometraje_actual`, `observaciones` | `id_cliente` | — |

Los cuatro campos `*_receptor_snap` se agrupan como identidad receptor; los
dos `ficha_*` como ficha técnica emitida. `password_hash` se mantiene en el
sistema y se omite como detalle criptográfico del modelo conceptual.
`request_id`, `contenido_hash`, `asignacion_version`, `control_version`,
`flujo_version` y la `version` técnica del diagnóstico son controles técnicos.
La versión de cotización sí se muestra porque identifica lo autorizado.

**Seguridad de la entrega:** solo archivos bajo `docs/modelos/V21_Chen/`.
Sin cambios a Go, frontend, PostgreSQL, MongoDB o migraciones. Sin borrados,
recreación de bases, commit, push o merge. Los diagramas V19/V20 conservan sus blobs Git.
