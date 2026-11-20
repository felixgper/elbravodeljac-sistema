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
