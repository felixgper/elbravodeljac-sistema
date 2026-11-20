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

- **PASS:** 28 comprobaciones repetidas en el motor SQL Server existente mediante SELECT sin consultar tablas oficiales. Incluyen pagos revertidos, precisión decimal, corte de mes en Lima, revisiones, cinco estados, cantidades realmente ejecutadas, JSON legado sin duplicar conceptos, fechas de finalización, privacidad y seis fuentes vacías. No requieren instalar vistas ni crear registros; los SELECT de las seis vistas no cambian.
- **PASS actual:** nueve pruebas estáticas sobre rechazo antes de cambiar opciones, envoltura transaccional, SQL generado sin DDL/tablas oficiales, expresiones del backend, privacidad, medidas/relaciones, parámetro deshabilitado y preparación de casos QA. Validar el texto generado no demuestra su comportamiento en el motor.
- **PASS:** 33 definiciones PBIR validadas contra esquemas públicos de Microsoft, sin errores.
- **PASS:** las seis expresiones ejecutadas como CTE de solo lectura contra las fuentes de desarrollo concilian conteos y sumas, sin pagos ni servicios duplicados. No se han instalado vistas; los resultados vacíos no demuestran escenarios inexistentes en los datos actuales. Frontend, `/healthz` y `/readyz` responden HTTP 200.
- **SKIP:** fallo DDL provocado y preservación de transacción externa en el motor; se conservan pruebas preparadas pero no se ejecutan en la base de desarrollo. La protección se revisa estáticamente y no se presenta como rollback DDL probado.
- **SKIP:** ejecución de `verificar_vistas.sql` sobre vistas instaladas, comprobación efectiva de permisos y cruces de gráficos con analítica habilitada. Requieren la instalación y autorización posteriores.

Estas pruebas comprueban las expresiones SELECT, no reemplazan la validación del DDL ni del refresco completo de Power BI después de habilitarlo. No consultan ni exportan registros reales del taller.

### GitHub Actions: alcance de las omisiones

Para ecfd628, [SQL Server 2025 QA](https://github.com/felixgper/elbravodeljac-sistema/actions/runs/44649161815) terminó `skipped` en el job `schema`, sin pasos: su condición exige el PR #12 y `codex/migracion-sql-server`.
[Backend SQL Server QA](https://github.com/felixgper/elbravodeljac-sistema/actions/runs/44649161809) terminó `skipped` en `integration`, sin pasos: para eventos de PR exige el #20 y `codex/permisos-financieros`.
El evento `synchronize` de este PR coincide con los disparadores generales, pero no con sus condiciones de job. No son fallos de SQL ni pruebas aprobadas del PR #22. No se modifican estos workflows.
Una comprobación adicional apropiada sería un workflow independiente, limitado al PR #22 y su rama, con `contents: read`, checkout sin persistir credenciales, timeout y `python -B -m unittest discover -s database/sqlserver/bi -p test_consultas.py -v`. No requeriría Docker, secretos ni acceso SQL. Es una propuesta, no se añade ni habilita aquí. La integración de vistas requiere una ejecución QA separada y autorizada; Go/Chrome completos no sustituyen las pruebas del instalador BI.

### Validación en la base de desarrollo existente

Se utiliza exclusivamente `elbravodeljac_sqlserver_dev`, en el SQL Server del proyecto oficial, accesible desde Windows por `127.0.0.1,1435`. La conexión TCP de SSMS y la configuración de Go se contrastan con la misma instancia. No se crean bases, contenedores ni respaldos adicionales ni se importan registros.
El catálogo conserva las 28 tablas/208 columnas/50 FK del aplicativo; `dbo.sysdiagrams` es una tabla auxiliar de SSMS adicional y queda fuera de los indicadores. No se altera ni se confunde con una entidad del sistema.

Antes del DDL, revisar las seis fuentes y sus columnas, propietarios, vistas ya existentes y permisos actuales. La revisión confirmó `dbo` como propietario de las fuentes y de `bi`, y únicamente la vista de resumen ya instalada. La cuenta de aplicación no tiene visibilidad completa de permisos ajenos: la revisión del lector debe realizarse desde una sesión administrativa ya autenticada, sin copiar sus credenciales. No concluir que un usuario o permiso no existe porque esté oculto por visibilidad de metadatos.

1. **Autorización DDL.** Solicitar aprobación expresa para ejecutar el archivo completo `vistas_analiticas.sql` en esa base, conservando transacción única y guarda inicial. Antes y después, comparar las definiciones de tablas, conteos de registros y estado de la API mediante lecturas. El instalador solo crea/altera las seis vistas, no toca tablas, triggers, registros ni la vista de resumen.
2. **Conciliación.** Tras la instalación autorizada, ejecutar `verificar_vistas.sql`, revisar fechas/importes/conteos y confrontar cada vista con su fuente. Con tablas operativas vacías, documentar resultados cero; no crear operaciones para rellenar los gráficos. Los escenarios con pagos revertidos, revisiones y servicios realizados siguen cubiertos por los SELECT sintéticos reproducibles.
3. **Autorización SELECT separada.** Conservar CONNECT y SELECT del resumen ya concedidos. Solicitar aprobación para los cinco GRANT por objeto indicados más abajo; no ejecutarlos junto con el instalador. El modelo usa la tabla de cobros para agregar meses, por lo que `vw_ingresos_mensuales` no necesita un GRANT adicional salvo que se apruebe su consulta directa. No conceder SELECT sobre tablas, escritura ni roles administrativos.
4. **Comprobar acceso.** Tras aprobar los GRANT, verificar lectura de las cinco vistas con `powerbi_reader`, ausencia de permisos directos sobre tablas y roles elevados, y funcionamiento normal de Go. No cambiar contraseña ni crear otro usuario. Una comprobación por impersonación no reemplaza la actualización real desde Desktop.
5. **Power BI.** Conservar el PBIP versionado con `HabilitarAnalitica=false`. En una copia local excluida de Git, activar temporalmente el parámetro, actualizar desde el mismo origen autorizado y verificar cinco tarjetas, cinco gráficos, filtro de fechas y cruces. Conciliar los valores con SQL respetando los criterios financieros anteriores. Con datos vacíos, comprobar estados vacíos sin sustituirlos por ejemplos. No publicar ni guardar cachés o datos importados en el PR.
6. **Cierre.** Repetir pruebas estáticas, comprobar enlaces/secretos y actualizar este mismo Draft PR con PASS/FAIL/SKIP reales. La instalación transaccional, los GRANT y la activación no se consideran aprobados por una prueba estática. Solicitar aprobación final antes del merge y de habilitar el parámetro oficial.

La prueba de fallo DDL intencional y los casos SQL de transacción externa quedan **SKIP: no se ejecutan en la base de desarrollo**. `preparar_qa.py` se conserva como prueba opcional aislada, sin ejecución y sin formar parte del procedimiento actual. No ejecutar su salida en desarrollo ni crear QA para este cierre. Las comprobaciones estáticas de ese generador siguen siendo válidas como revisión de código, no como pruebas ejecutadas en SQL Server.

## Habilitación pendiente de analítica

La inspección real confirmó que `powerbi_reader` tiene SELECT solamente sobre `bi.vw_resumen_taller`.
Las nuevas vistas todavía no están instaladas y este trabajo no modifica usuarios ni permisos.
Por ello **HabilitarAnalitica** se entrega en `false`. La segunda página muestra un aviso y medidas en blanco, sin datos de ejemplo.
La actualización de la página de resumen sigue consultando la vista real.

Antes de habilitarla, un administrador debe aprobar y aplicar por separado los siguientes pasos.

1. Solicitar autorización expresa y ejecutar [vistas_analiticas.sql](../database/sqlserver/bi/vistas_analiticas.sql). El script crea seis vistas SELECT y conserva el resumen existente. No modifica datos ni concede permisos.
2. Solicitar autorización separada para el acceso SELECT de `powerbi_reader` a las cinco vistas utilizadas por el modelo. Este paso requiere aprobación del propietario y queda fuera de los cambios realizados.
3. Ejecutar [verificar_vistas.sql](../database/sqlserver/bi/verificar_vistas.sql) con una cuenta que pueda consultar las fuentes.
4. En **Transformar datos > Administrar parámetros**, cambiar **HabilitarAnalitica** a `true`, aplicar y actualizar.
5. Revisar selección de fechas, cruces entre gráficos, conciliación y actualización con pagos revertidos. Los gráficos ya están construidos.

No active el parámetro mientras falten vistas o permisos. Los errores de fuente se mostrarán; no se sustituyen por datos ficticios.

### Permisos propuestos para revisión

La cuenta que instale las vistas necesita `CREATE VIEW` en la base, `ALTER` sobre el esquema `bi` y acceso a los objetos referenciados.
La cuenta lectora conserva su permiso existente sobre el resumen y requiere SELECT por objeto sobre `bi.vw_pagos_vigentes`, `bi.vw_ordenes_analiticas`, `bi.vw_cotizaciones_analiticas`, `bi.vw_vehiculos_atendidos` y `bi.vw_servicios_realizados`.
`bi.vw_ingresos_mensuales` es una agregación adicional; requiere SELECT solamente si se consulta directamente.
La cadena de propiedad permite consultar estas vistas sin conceder acceso directo a las tablas, siempre que se mantenga su propietario común.
El instalador ahora rechaza propietarios distintos entre las fuentes y las vistas antes de ejecutar DDL, en lugar de solicitar permisos adicionales. La prueba de acceso con el lector real queda pendiente hasta instalar y aprobar los SELECT.
La conciliación con tablas de origen debe ejecutarla una cuenta administrativa autorizada.
No se requieren permisos de escritura, `db_datareader` ni SELECT sobre todo el esquema para el lector.
Las concesiones propuestas se muestran solo para revisión y requieren autorización expresa antes de ejecutarse:

```sql
GRANT SELECT ON OBJECT::bi.vw_pagos_vigentes TO powerbi_reader;
GRANT SELECT ON OBJECT::bi.vw_ordenes_analiticas TO powerbi_reader;
GRANT SELECT ON OBJECT::bi.vw_cotizaciones_analiticas TO powerbi_reader;
GRANT SELECT ON OBJECT::bi.vw_vehiculos_atendidos TO powerbi_reader;
GRANT SELECT ON OBJECT::bi.vw_servicios_realizados TO powerbi_reader;
```

No revocar ni reemplazar permisos existentes. No ejecutar estos GRANT durante la revisión previa.

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
