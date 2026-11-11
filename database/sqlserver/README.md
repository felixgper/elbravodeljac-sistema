# Esquema físico de El Bravo del JAC en SQL Server

Para preparar una copia histórica desde PostgreSQL, consulta el [procedimiento de migración de datos](MIGRACION_DATOS.md).

El esquema prepara una instalación vacía en SQL Server 2025 Standard o Standard Developer. No usa características exclusivas de Enterprise. La validación automatizada del adaptador Go utiliza Standard Developer. El entorno habitual puede seguir usando PostgreSQL durante la transición.

Ejecutar desde un cliente compatible con lotes `GO` y detenerse ante cualquier error. Para probar, crear previamente una base desechable llamada `elbravodeljac_sqlserver_qa` en una instancia aislada con nivel de compatibilidad 170. Con una sesión de cliente configurada para esa base:

```powershell
sqlcmd -S <instancia-qa> -d elbravodeljac_sqlserver_qa -b -f 65001 -i database/sqlserver/esquema_entrega.sql
sqlcmd -S <instancia-qa> -d elbravodeljac_sqlserver_qa -b -f 65001 -i database/sqlserver/pruebas_esquema.sql
```

El primer script crea el esquema en una transacción y no instala datos iniciales. El segundo carga datos sintéticos en la base desechable y verifica las reglas por lotes. Una instrucción con una fila inválida debe revertirse completa. Los casos negativos comprueban el error concreto y, cuando corresponde, que tampoco se escribió la fila válida.

- Se mantienen las 28 tablas, 208 columnas, 28 PK, 50 FK, 22 CHECK de origen, 26 secuencias, 3 vistas y 93 índices contando los de las claves. Las nueve funciones de trigger de PostgreSQL se integran en diez triggers T-SQL. Las otras tres funciones conservan sus nombres.
- Los importes mantienen precisión y escala. Las fechas usan `datetimeoffset(6)` con valor por defecto UTC. SQL Server genera ese valor al escribir, mientras que `CURRENT_TIMESTAMP` de PostgreSQL se fija al comenzar la transacción. Cuando varias filas deban compartir un mismo instante de negocio, el adaptador pasará el mismo valor explícito.
- Los textos de longitud fija de contrato usan `nvarchar` con capacidad para pares sustitutos y un CHECK de longitud en caracteres. El JSON se conserva como texto Unicode validado y las fichas archivadas como `varbinary(max)`. El adaptador debe usar una representación JSON estable para no confundir cambios de formato con cambios en datos históricos.
- La colación binaria conserva distinciones de mayúsculas y acentos en claves. SQL Server aún puede equiparar cadenas con espacios finales en índices únicos, así que la carga debe detectar esas colisiones y detenerse para revisión, sin descartar ni modificar filas automáticamente. Para búsquedas se usa una colación insensible a acentos. El índice de clientes cubre nombre y documento, pero no reemplaza el rendimiento del índice GIN para búsquedas por subcadena. El puerto del adaptador debe verificar ambas cosas con datos representativos.
- Las FK `RESTRICT` se expresan como `NO ACTION`. Correo y vínculo con empleado tienen índices únicos filtrados, con varios NULL permitidos. Las normalizaciones de escritura usan triggers `INSTEAD OF`; el futuro adaptador proporcionará los campos obligatorios, incluido el identificador de recepción al cotizar, antes de enviar el DML. Los triggers que correlacionan filas por PK rechazan actualizaciones de ese identificador.
- El adaptador Go protege el último acceso administrativo, controla versiones y toma un bloqueo sobre la orden antes de registrar pagos, rectificaciones o entregas. La futura carga histórica debe preservar IDs, secuencias, PDF e historiales anteriores a las reglas vigentes; por ejemplo, no tratar asignaciones auxiliares históricas como altas nuevas.

## Ejecutar la API con una base nueva y aislada

1. Crea `elbravodeljac_sqlserver_dev` con nivel de compatibilidad 170 y aplica únicamente `esquema_entrega.sql`. Reserva `elbravodeljac_sqlserver_qa` para pruebas automatizadas. `pruebas_esquema.sql` carga datos de prueba y se ejecuta en otra base desechable.
2. Configura `SQLSERVER_DSN` como URL `sqlserver://` que indique la base y los parámetros TLS del servidor. Quita `POSTGRES_DSN` del proceso de la API. Configura la URL y el secreto JWT como variables privadas; no guardes sus valores en el repositorio.
3. Para una instalación vacía, establece `SQLSERVER_BOOTSTRAP_CONFIRM` con el nombre exacto de la base, `DEV_ADMIN_USERNAME` y `DEV_ADMIN_PASSWORD`. Desde `backend`, ejecuta `go run ./cmd/sqlserver-bootstrap` para crear la cuenta Gerente, los roles, permisos y tipos documentales. El comando rechaza una base con datos; el arranque normal no lo invoca.
4. Configura `APP_ORIGIN`, `HTTP_ADDR`, `JWT_SECRET_BASE64` y `FRONTEND_DIR` como se indica en el [README del backend](../../backend/README.md). Para evidencias usa un MongoDB de desarrollo o pruebas aislado con `MONGODB_URI` y `MONGODB_DATABASE`; conserva su inicialización explícita y los nombres de base admitidos. Ejecuta `go run ./cmd/api` desde `backend`.

El workflow de la rama de migración levanta instancias desechables en GitHub Actions. La suite usa `SQLSERVER_QA_DSN` y `SQLSERVER_QA_MONGO_URI` y crea datos sintéticos. Este procedimiento no importa datos desde PostgreSQL.
