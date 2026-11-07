# MongoDB técnico

Colecciones `resultados_escaner` y `evidencias_tecnicas`; bucket privado GridFS `archivos_tecnicos.files/chunks`. No duplica clientes, vehículos, órdenes, pagos ni credenciales de PostgreSQL.

El contrato ejecutable es [validators.json](../../backend/internal/platform/mongodb/validators.json), embebido en el adaptador Go; incluye ObjectID y fechas BSON. Se retira la propuesta JSON anterior, que no se utilizaba para inicializar ni validar las colecciones.

Cada documento referencia `diagnosticoId` y `capturadoPor` SQL. Antes de leer/escribir, Go verifica existencia, permisos y alcance en PostgreSQL. Índices únicos `(diagnosticoId, requestId)` y `gridfsFileId`, índices por captura y los de GridFS. Reintentos usan digest y reservas `uploading`/`ready`; hay compensación y conciliación explícita.

Las rutas de escáner/evidencias y límites están en [ENDPOINTS.md](../../docs/backend/ENDPOINTS.md) y en los DTO/validaciones de `backend/internal/domain/technical.go`. El adaptador usa el driver oficial MongoDB Go v2. `cmd/mongo-init` aplica explícitamente validadores/índices solo en las bases separadas admitidas; arrancar la API no modifica validadores existentes.

[Configuración y ejecución](../../backend/README.md) · [Pruebas aisladas](../../checks/README.md). No ejecutar inicialización, conciliación o pruebas sobre datos reales sin autorización específica.
