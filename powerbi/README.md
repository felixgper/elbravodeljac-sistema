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

El instalador exige una sesión sin transacciones abiertas y comprueba base/esquema antes de iniciar. Las seis definiciones se ejecutan dentro de una sola transacción con `XACT_ABORT`, `TRY/CATCH`, rollback si `XACT_STATE()` no es cero y propagación del error mediante `THROW`.
Cada `CREATE OR ALTER VIEW` se compila mediante `sp_executesql` dentro del bloque protegido: así los errores de compilación de una vista también llegan al `CATCH`, sin continuar hasta un commit parcial. Se conserva la definición previa de las vistas si se revierte la transacción.
Ejecute el archivo completo en una sesión dedicada, preferentemente `sqlcmd -b`. No ejecute fragmentos. Si se cancela manualmente desde un editor, cierre la sesión o asegúrese de revertir cualquier transacción abierta; una cancelación no equivale a un error capturable por `CATCH`.
Referencia técnica: [TRY/CATCH de SQL Server](https://learn.microsoft.com/en-us/sql/t-sql/language-elements/try-catch-transact-sql?view=sql-server-ver17).

Las pruebas no necesitan instalar vistas ni acceder a tablas:

```powershell
python -m unittest discover -s database/sqlserver/bi -p test_consultas.py -v
python -X utf8 database/sqlserver/bi/probar_consultas.py > consultas.sql
```

Ejecute el SQL generado con una conexión autorizada y `sqlcmd -f 65001 -b -h -1 -w 65535`. Revise que las 28 filas indiquen PASS; un FAIL de comparación es una fila de resultado y debe detectarlo quien ejecute la prueba, aunque sqlcmd termine con código cero.
El generador extrae las seis consultas del propio instalador y sustituye sus fuentes por CTE de valores sintéticos. También extrae del código Go el filtro de servicios y del esquema los movimientos de pagos; no los reimplementa para hacerlos coincidir.

- **PASS:** 28 comprobaciones en el motor SQL Server existente mediante SELECT sin consultar tablas oficiales. Incluyen pagos revertidos, precisión decimal, corte de mes en Lima, revisiones, cinco estados, cantidades realmente ejecutadas, JSON legado sin duplicar conceptos, fechas de finalización, privacidad y seis fuentes vacías.
- **PASS:** seis pruebas estáticas sobre la envoltura transaccional, SQL generado sin DDL/tablas oficiales, expresiones del backend, privacidad, medidas/relaciones y parámetro deshabilitado.
- **SKIP:** fallo DDL provocado después de crear/alterar vistas y restauración efectiva de sus definiciones; no se ejecuta DDL en la base oficial ni se crean contenedores. La atomicidad está revisada estructuralmente, pero esta prueba de integración sigue pendiente en una base QA autorizada.
- **SKIP:** ejecución de `verificar_vistas.sql` sobre vistas instaladas, comprobación efectiva de permisos y cruces de gráficos con analítica habilitada. Requieren la instalación y autorización posteriores.

Estas pruebas comprueban las expresiones SELECT, no reemplazan la validación del DDL ni del refresco completo de Power BI después de habilitarlo. No consultan ni exportan registros reales del taller.

## Habilitación pendiente de analítica

La inspección real confirmó que `powerbi_reader` tiene SELECT solamente sobre `bi.vw_resumen_taller`.
Las nuevas vistas todavía no están instaladas y este trabajo no modifica usuarios ni permisos.
Por ello **HabilitarAnalitica** se entrega en `false`. La segunda página muestra un aviso y medidas en blanco, sin datos de ejemplo.
La actualización de la página de resumen sigue consultando la vista real.

Antes de habilitarla, un administrador debe aprobar y aplicar por separado los siguientes pasos.

1. Revisar y ejecutar [vistas_analiticas.sql](../database/sqlserver/bi/vistas_analiticas.sql). El script crea seis vistas SELECT y conserva el resumen existente. No modifica datos ni concede permisos.
2. Autorizar específicamente el acceso SELECT de `powerbi_reader` a las cinco vistas utilizadas por el modelo. Este paso requiere aprobación del propietario y queda fuera de los cambios realizados.
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
Este proyecto no incluye ni ejecuta concesiones de permisos.

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
