# MongoDB técnico: implementado en Fase D

Colecciones: resultados_escaner y evidencias_tecnicas; bucket privado GridFS
archivos_tecnicos.files/chunks. No se copian datos transaccionales de PostgreSQL.
Los documentos se vinculan a diagnosticoId y capturadoPor SQL; cada acceso verifica
existencia, usuario, permisos y alcance del diagnóstico en PostgreSQL.

collections.schema.json conserva el contrato JSON de dominio. Los validadores BSON
aplicables, con _id ObjectID, fechas Date, payloadHash y estado/lease internos,
están en backend/internal/platform/mongodb/validators.json. El driver oficial Go
v2 los instala únicamente mediante cmd/mongo-init en una base _test o _dev separada.
El arranque de la API no cambia los validadores de una colección existente.

Índices únicos por (diagnosticoId,requestId) y gridfsFileId; índice de captura por
diagnóstico e índices estándar GridFS. Idempotencia con digest, reservas uploading,
confirmación ready, compensación y herramienta explícita de conciliación.

Configuración, límites, payloads, permisos, Compose y resultados de integración real:
[FASE_D.md](../../docs/backend/FASE_D.md).
