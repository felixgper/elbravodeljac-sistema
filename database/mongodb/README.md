# Modelo técnico MongoDB (propuesto, no aplicado)

Base propuesta: elbravodeljac_tecnico. El archivo collections.schema.json es una
especificación JSON Schema para el contrato; no es un comando mongosh ni aplica
validadores automáticamente. Al implementar, traducir a validator/$jsonSchema con
bsonType (fechas Date e identificadores BSON) y probar en otra base.

## Colecciones

| Colección | Contenido | Índices propuestos |
|---|---|---|
| resultados_escaner | diagnosticoId SQL, requestId, capturadoPor SQL, capturadoEn, equipo, códigos DTC, mediciones flexibles, versión de esquema | unique(diagnosticoId, requestId); (diagnosticoId, capturadoEn descendente) |
| evidencias_tecnicas | diagnosticoId SQL, requestId, autor SQL, fecha, descripción, MIME, tamaño, SHA-256, gridfsFileId | unique(diagnosticoId, requestId); (diagnosticoId, capturadoEn descendente); unique(gridfsFileId) |
| archivos_tecnicos.files / .chunks | Binarios mediante GridFS, sin URLs públicas | Índices estándar de GridFS creados por el adaptador en entorno autorizado |

ID PostgreSQL como cadena decimal evita perder precisión y permite validación
antes de convertir a int64 en Go. El _id MongoDB se devuelve como cadena hexadecimal
sin usarlo como ID transaccional. requestId es clave de reintento por diagnóstico
y colección: el mismo requestId con payload diferente debe producir 409.

Los documentos no guardan contraseñas, tokens, pagos ni copias completas del cliente.
Verificar permisos sobre diagnóstico/OT en PostgreSQL en cada acceso. No hay FK
entre motores. Mantener diagnóstico aunque MongoDB falle, reintentar explícitamente
y reconciliar archivos sin metadatos. No borrar evidencia como efecto de eliminar
un cliente. No se ofrecen endpoints de borrado de evidencia en la primera etapa.

No hay interfaz de subida en el frontend actual. Endpoints documentales reservados
para la integración posterior. No inicializar estas colecciones contra datos reales
al arrancar el servidor.
