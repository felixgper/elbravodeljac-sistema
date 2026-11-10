# Esquema físico de El Bravo del JAC en SQL Server

Esta etapa prepara una instalación vacía en SQL Server 2025 Standard o Standard Developer. Usa T-SQL disponible también en Express, sin características exclusivas de Enterprise. El backend oficial sigue apuntando a PostgreSQL.

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
- En la etapa Go quedan la asignación y autorización por rol, protección del último administrador, el control de versiones y la idempotencia transaccional de pagos y rectificaciones. El esquema conserva sus datos y restricciones de base. El adaptador deberá mantener la misma adquisición de bloqueos sobre la orden antes de operar sobre pagos o entregas y probar la concurrencia real. La futura carga histórica debe preservar IDs, secuencias, PDF e historiales anteriores a las reglas vigentes; por ejemplo, no tratar asignaciones auxiliares históricas como altas nuevas.
