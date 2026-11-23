# Power BI · El Bravo del JAC

Abra [ElBravoDelJAC.pbip](ElBravoDelJAC.pbip) con Power BI Desktop y seleccione **Inicio > Actualizar**.
El proyecto utiliza PBIR y TMDL editables. Los gráficos, filtros, relaciones y medidas están definidos en los archivos del proyecto.

## Fuente y seguridad

Se conserva la fuente local existente, SQL Server en `127.0.0.1,1435`, base `elbravodeljac_sqlserver_dev`.
Use la autenticación de base de datos de `powerbi_reader` ya configurada en Desktop.
El proyecto no contiene contraseñas ni credenciales. Las credenciales se administran en **Configuración de origen de datos** de Desktop.
Es un informe financiero para personal autorizado. Los roles del aplicativo Go no se trasladan automáticamente a Power BI.

Los archivos `.pbi`, `cache.abf`, PBIX, ajustes locales y copias de datos quedan excluidos mediante el [.gitignore local](.gitignore).
El original exportado desde la ventana “Sin título” permanece en `backend/.cache/powerbi-original/`, fuera del seguimiento de Git.

## Páginas y medidas

**Resumen del taller** conserva las cinco tarjetas y la vista `bi.vw_resumen_taller`.
Las medidas DAX usan MAX sobre la única fila de resumen, con formatos enteros y moneda peruana.
Los cinco indicadores son acumulados históricos. Clientes y vehículos registrados no tienen fecha de alta en el esquema actual.

**Actividad por período** contiene cinco indicadores, un filtro de fecha y cinco gráficos nativos editables.
La relación entre Calendario y cada hecho es de uno a muchos, con filtro en una sola dirección.
Los estados son los actuales de las órdenes creadas en el período seleccionado.
Cada hecho usa su fecha correspondiente, en hora de Lima.

| Indicador o gráfico | Grano y fecha |
|---|---|
| Ingresos | Un registro por pago vigente, fecha del cobro |
| Órdenes | Una fila por orden, fecha de creación |
| Cotizaciones | Una fila por cotización emitida, incluye revisiones |
| Clientes y vehículos atendidos | DISTINCTCOUNT sobre recepciones, fecha de ingreso |
| Servicios realizados | Un concepto confirmado de una orden finalizada, fecha de finalización |

La suma de cobros excluye pagos con `revertido_en` informado. No se une el hecho de pagos a conceptos ni asignaciones.
El total histórico coincide con la vista actual. La distribución mensual puede diferir del dashboard web, que utiliza movimientos positivos y negativos en la fecha de rectificación.
El importe comercial de una cotización no se considera ingreso cobrado.
La vista de servicios publica el nombre del catálogo o «Servicio manual», y omite las descripciones libres.
Los identificadores técnicos de atención se conservan para contar clientes y vehículos distintos; las vistas omiten nombres de personas, documentos de identidad, teléfonos, correos, contraseñas y comprobantes.

Los SELECT de las seis vistas son de lectura. El script de instalación contiene DDL (`CREATE OR ALTER VIEW`) y debe ejecutarse únicamente tras su aprobación.

### Comparación con el backend Go

Se contrastan las consultas de `backend/internal/platform/sqlserver/statistics.go` y la vista `dbo.jac_movimientos_pagos` del esquema oficial. No se cambia el criterio del backend.

| Métrica | Power BI | Backend Go |
|---|---|---|
| Cobros históricos | Suma de pagos no revertidos | Suma de movimientos de pago y reversión: mismo resultado histórico |
| Cobros por período | Excluye el pago revertido de su mes original | Mantiene el cobro original y resta la reversión en su fecha; puede haber meses negativos |
| Órdenes por fecha | Estado actual de las órdenes creadas en el intervalo | Inicio muestra activas sin limitar su fecha de creación; finalizadas/canceladas por cierre. Análisis cuenta finalizadas por cierre |
| Cotizaciones | Cada ID emitido, incluidas revisiones; fecha de emisión | `counts.quotes` cuenta cada emisión en el mismo intervalo; una revisión no es ingreso adicional |
| Clientes y vehículos | Registrados históricos en Resumen; distintos atendidos por ingreso en Actividad | `counts.clients` y `counts.vehicles` cuentan distintos atendidos por ingreso |
| Servicios realizados | Cantidades ejecutadas de órdenes finalizadas, por fecha de cierre | Mismo filtro técnico y cantidades; el dashboard muestra solo los cuatro grupos principales |

Ejemplo sintético probado: un pago de S/75 en octubre revertido en noviembre se excluye de ambos meses en Power BI; Go conserva +75 en octubre y -75 en noviembre. Con otros cobros vigentes, la prueba obtiene octubre/noviembre 150/50 en Power BI y 225/-25 en Go; ambos suman S/200 históricos.
La vista de servicios usa el nombre de catálogo, sin descripción libre. Go agrupa por ID y descripción comercial, por lo que las etiquetas y agrupaciones pueden diferir aunque la cantidad total coincida. Análisis agrega auxilios por separado; no forman parte de esta vista de servicios.
La compatibilidad técnica conserva el criterio existente de Go: ejecución `Realizado` **o** pertenencia a `performedItems`. No se redefine aquí el contrato de Control técnico.

### Instalación atómica y pruebas reproducibles

El instalador rechaza primero una transacción ajena mediante `RAISERROR` de severidad 16 y `RETURN`, antes de cualquier `SET`. No usa `THROW` en ese rechazo: `RAISERROR` no aplica `XACT_ABORT`, incluso si el llamador ya lo tiene activado. No cambia opciones, confirma ni revierte trabajo del llamador. Después comprueba base/esquema e instala las seis definiciones en una sola transacción con `XACT_ABORT`, `TRY/CATCH`, rollback si `XACT_STATE()` no es cero y propagación de errores propios mediante `THROW`.
Cada `CREATE OR ALTER VIEW` se compila mediante `sp_executesql` dentro del bloque protegido: así los errores de compilación de una vista también llegan al `CATCH`, sin continuar hasta un commit parcial. Se conserva la definición previa de las vistas si se revierte la transacción.
Ejecute el archivo completo en una sesión dedicada, preferentemente `sqlcmd -b`. No ejecute fragmentos. Si se cancela manualmente desde un editor, cierre la sesión o asegúrese de revertir cualquier transacción abierta; una cancelación no equivale a un error capturable por `CATCH`.
Referencia técnica: [TRY/CATCH de SQL Server](https://learn.microsoft.com/en-us/sql/t-sql/language-elements/try-catch-transact-sql?view=sql-server-ver17).
La excepción deliberada para rechazar una transacción ajena se fundamenta en [RAISERROR y XACT_ABORT](https://learn.microsoft.com/en-us/sql/t-sql/language-elements/raiserror-transact-sql?view=sql-server-ver17).

Las pruebas no necesitan instalar vistas ni acceder a tablas:

```powershell
python -B -m unittest discover -s database/sqlserver/bi -p test_consultas.py -v
python -B -X utf8 database/sqlserver/bi/probar_consultas.py > consultas.sql
```

Ejecute el SQL generado con una conexión autorizada y `sqlcmd -f 65001 -b -h -1 -w 65535`. Revise que las 28 filas indiquen PASS; un FAIL de comparación es una fila de resultado y debe detectarlo quien ejecute la prueba, aunque sqlcmd termine con código cero.
El generador extrae las seis consultas del propio instalador y sustituye sus fuentes por CTE de valores sintéticos. También extrae del código Go el filtro de servicios y del esquema los movimientos de pagos; no los reimplementa para hacerlos coincidir.

## Estado verificado de la integración

Las siguientes comprobaciones ya se completaron con autorizaciones separadas:

- **PASS — Instalación:** las seis vistas de [vistas_analiticas.sql](../database/sqlserver/bi/vistas_analiticas.sql) están instaladas y verificadas en `elbravodeljac_sqlserver_dev`. Se ejecutó [verificar_vistas.sql](../database/sqlserver/bi/verificar_vistas.sql) mediante lecturas; se conservaron las tablas, los registros y `bi.vw_resumen_taller`.
- **PASS — Acceso:** se concedieron los cinco SELECT por objeto indicados abajo. Se comprobó el acceso con `powerbi_reader`, la cadena de propiedad y la ausencia de permisos directos innecesarios sobre tablas operativas, escritura o administración. Se conservó su SELECT del resumen. No se cambiaron las credenciales.
- **PASS — Power BI Desktop:** una copia local excluida de Git, con analítica temporalmente habilitada, abrió y actualizó desde SQL Server usando la conexión configurada del lector. Se comprobaron cinco indicadores, cinco gráficos, las cinco relaciones, el filtro de fechas y el cruce desde vehículos atendidos. Las fuentes vacías mostraron estados adecuados; no se insertaron registros.
- **PASS — Diseño:** el usuario aprobó el diseño carbón y granate, las tarjetas legibles, los gráficos y los mensajes sin datos. Se trasladó al PBIP oficial sin modificar los cálculos. Se conservan las 13 medidas originales y se agregan cinco medidas de texto para los mensajes.
- **PASS — SELECT sintéticos, ejecución anterior:** 28 comprobaciones en SQL Server sin consultar tablas oficiales. Cubren pagos revertidos, precisión decimal, corte de mes en Lima, revisiones, cinco estados, servicios ejecutados, JSON legado sin duplicar conceptos, privacidad y seis fuentes vacías. Su código contiene ejemplos sintéticos, no datos extraídos del taller.
- **PASS — Conciliación anterior:** las seis expresiones, inicialmente probadas como CTE de lectura, conciliaron conteos y sumas con sus fuentes; posteriormente se verificaron las vistas instaladas. Los resultados sin operaciones no demuestran escenarios positivos ausentes. Go, frontend, `/healthz` y `/readyz` continuaron funcionando en aquellas comprobaciones; esta revisión documental no vuelve a intervenirlos.
- **PASS — Revisión estática:** nueve pruebas de `test_consultas.py`, definiciones PBIR publicadas, referencias TMDL, consultas, relaciones y conservación de medidas. La comprobación estática no reemplaza la ejecución del DDL ni del motor DAX.

### Pendientes y límites de la validación

- **SKIP — DDL provocado y transacción externa:** no se ejecutaron fallos DDL intencionales ni casos de transacción externa en la base de desarrollo. `preparar_qa.py` conserva esos casos para un QA aislado y autorizado, con marcador `synthetic-only`; su revisión estática no demuestra rollback real.
- **SKIP — Series positivas en Desktop:** faltan operaciones para comprobar valores positivos y todos los cruces de cobros, órdenes, cotizaciones y servicios. Los SELECT sintéticos previos cubren esos criterios SQL, pero no sustituyen la comprobación visual con datos autorizados. No se crearán datos para rellenar gráficos.
- **SKIP — Resolución exacta 1366×768:** se revisaron escritorio y ventana reducida; no se reprodujo esa resolución exacta. El selector conserva el formato regional que muestre Desktop.
- **SKIP — Esquema de metadatos locales 2.14.0:** la URL pública de Microsoft devolvió 404. La compatibilidad estructural con 2.7.0 pasó; no se presenta como validación exacta. El contenido visual publicado conserva 2.13.0 y sí pasó su esquema exacto. Los metadatos locales ajenos no se publicaron.
- **Pendiente de autorización:** habilitar la analítica en el PBIP oficial y fusionar el PR. Son decisiones separadas; ninguna se ejecuta durante esta revisión.

### GitHub Actions: alcance de las omisiones

Los workflows se disparan ante `synchronize`, pero sus condiciones de job excluyen el PR #22:

- **SQL Server 2025 QA** exige el PR #12 y `codex/migracion-sql-server`.
- **Backend SQL Server QA** exige, para eventos de PR, el #20 y `codex/permisos-financieros`; su disparador push se limita a `codex/backend-sql-server`.

Para `ecfd628` terminaron `skipped` sin ejecutar sus jobs: [esquema](https://github.com/felixgper/elbravodeljac-sistema/actions/runs/44649161815) y [backend](https://github.com/felixgper/elbravodeljac-sistema/actions/runs/44649161809). También se confirmó `skipped` para `b590c42`: [esquema](https://github.com/felixgper/elbravodeljac-sistema/actions/runs/45107719045) y [backend](https://github.com/felixgper/elbravodeljac-sistema/actions/runs/45107718982).
No son fallos SQL ni aprobaciones automáticas del PR #22. Las pruebas estáticas se ejecutan localmente sin Docker ni acceso a bases. No se modifican ni amplían estos workflows. Un workflow estático específico sería opcional y requeriría revisión independiente.

## Configuración oficial y activación posterior

El PBIP oficial conserva **HabilitarAnalitica=false** por decisión expresa, aunque las seis vistas y sus permisos ya están preparados y comprobados. El aviso indica «Analítica desactivada por configuración», sin atribuirlo a falta de permisos. Actividad muestra sus medidas deshabilitadas y el resumen sigue consultando su vista real.

Tras una autorización nueva para habilitar el proyecto oficial:

1. En **Transformar datos > Administrar parámetros**, cambiar **HabilitarAnalitica** a `true`, aplicar y actualizar mediante la conexión existente de `powerbi_reader`.
2. Verificar tarjetas, fechas, relaciones, cruces y estados sin datos, respetando los criterios financieros anteriores. Una respuesta vacía no debe sustituirse por datos ficticios.
3. Comprobar escenarios positivos cuando existan datos autorizados. Los errores de fuente deben mostrarse, no ocultarse con cifras de ejemplo.

No es necesario reinstalar las vistas ni repetir los GRANT para habilitar el parámetro. Una instalación nueva en otra base requiere autorización y comprobación independiente de destino, fuentes, propietarios y permisos.

### Permisos concedidos: referencia de instalación

Se conservó CONNECT y el SELECT existente sobre `bi.vw_resumen_taller`. Los cinco permisos siguientes ya fueron autorizados, ejecutados y comprobados; se documentan como referencia, no como tareas pendientes ni instrucciones para volver a ejecutarlos:

```sql
GRANT SELECT ON OBJECT::bi.vw_pagos_vigentes TO powerbi_reader;
GRANT SELECT ON OBJECT::bi.vw_ordenes_analiticas TO powerbi_reader;
GRANT SELECT ON OBJECT::bi.vw_cotizaciones_analiticas TO powerbi_reader;
GRANT SELECT ON OBJECT::bi.vw_vehiculos_atendidos TO powerbi_reader;
GRANT SELECT ON OBJECT::bi.vw_servicios_realizados TO powerbi_reader;
```

El modelo agrega meses desde Cobros, por lo que no necesita SELECT directo sobre `bi.vw_ingresos_mensuales`. La cadena de propiedad común permite consultar las vistas sin SELECT sobre tablas operativas. No se requieren escritura, `db_datareader`, SELECT global del esquema ni privilegios administrativos.
La conciliación con las tablas de origen requiere una cuenta administrativa autorizada; el lector del informe no debe recibir esos permisos.
Para una instalación nueva, la cuenta instaladora necesita `CREATE VIEW`, `ALTER` sobre `bi` y acceso a las fuentes. El instalador rechaza propietarios incompatibles sin ampliarlos automáticamente.
No se revocan ni reemplazan permisos existentes en este PR.

## Healthcheck de SQL Server

El contenedor ejecuta el healthcheck con el valor de `MSSQL_SA_PASSWORD` guardado al crearlo.
El registro actual devuelve 18456, State 8, contraseña incorrecta para `sa`. La cuenta `jac_app` sigue aceptando consultas.
Cambiar la contraseña dentro de SQL Server no actualiza el entorno del contenedor.

La reparación propuesta requiere una ventana separada y autorización. Se propone un login dedicado `jac_healthcheck` con CONNECT SQL, usuario asociado y CONNECT en la base `elbravodeljac_sqlserver_dev`, además de la pertenencia implícita a public.
La comprobación ejecutaría `SELECT 1` en esa base. No necesita SELECT sobre tablas, vistas, roles de lectura o escritura, sysadmin ni VIEW SERVER STATE.
Su contraseña se generaría y guardaría únicamente en configuración privada y se pasaría a sqlcmd mediante SQLCMDPASSWORD, sin imprimirla.
El login y su usuario no se han creado. Deben revisarse también los permisos heredados de public antes de autorizarlo.
El cambio requiere actualizar el healthcheck y recrear exclusivamente el contenedor SQL Server en una ventana aprobada, conservando imagen, volúmenes, montajes y redes.
Se conservarían los respaldos y una copia privada de la configuración anterior. La reversión reutilizaría la misma imagen y el mismo volumen, sin restaurar ni sobrescribir datos.
La puesta en marcha desde una base vacía también requiere revisar el orden de inicialización para que la comprobación no espere a un usuario aún inexistente.
Este trabajo no ejecuta esa reparación ni modifica Compose.
