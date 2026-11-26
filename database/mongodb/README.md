# MongoDB técnico

Colecciones `resultados_escaner` y `evidencias_tecnicas`; bucket privado GridFS `archivos_tecnicos.files/chunks`. No duplica clientes, vehículos, órdenes, pagos ni credenciales de SQL Server.

El contrato ejecutable es [validators.json](../../backend/internal/platform/mongodb/validators.json), embebido en el adaptador Go; incluye ObjectID y fechas BSON. Se retira la propuesta JSON anterior, que no se utilizaba para inicializar ni validar las colecciones.

Cada documento referencia `diagnosticoId` y `capturadoPor` SQL. Antes de leer/escribir, Go verifica existencia, permisos y alcance en SQL Server. Índices únicos `(diagnosticoId, requestId)` y `gridfsFileId`, índices por captura y los de GridFS. Reintentos usan digest y reservas `uploading`/`ready`; hay compensación y conciliación explícita.

Las rutas de escáner/evidencias y límites están en [ENDPOINTS.md](../../docs/backend/ENDPOINTS.md) y en los DTO/validaciones de `backend/internal/domain/technical.go`. El adaptador usa el driver oficial MongoDB Go v2. `cmd/mongo-init` aplica explícitamente validadores/índices solo en las bases separadas admitidas; arrancar la API no modifica validadores existentes.

[Configuración y ejecución](../../backend/README.md) · [Pruebas aisladas](../../checks/README.md). No ejecutar inicialización, conciliación o pruebas sobre datos reales sin autorización específica.

## Acceso opcional desde el diagnóstico

El modal «Evidencias y escáner» usa las colecciones y GridFS existentes. No se
requiere otro inicializador, esquema o colección. `PutScan` y `PutEvidence`
conservan requestId, digest, índices únicos y reservas uploading/ready; `File`
verifica SHA-256 y descarga mediante autorización SQL. El frontend no almacena
archivos en SQL, Base64 ni carpetas del servidor web.

Primero guardar el diagnóstico como borrador; después pueden registrarse archivos
o códigos. Al cerrar el diagnóstico se impiden nuevas escrituras también en Go,
con coordinación de bloqueo en el adaptador relacional. Los archivos guardados
siguen disponibles con permiso de consulta. Reintentar conserva la misma solicitud;
«Preparar otra carga» exige confirmación si el resultado previo es incierto.
