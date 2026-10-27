# Fase D: MongoDB implementado y probado

Las cinco operaciones técnicas pendientes están implementadas: **71/71 del catálogo
original**, más las tres ampliaciones relacionales de C (74 handlers de negocio).
No quedan rutas originales reservadas. Una configuración ausente, un motor caído
o un esquema no inicializado producen 503; nunca se simula una conexión correcta.

Rama backend/arquitectura-go-mongodb. PostgreSQL sigue siendo la autoridad del
negocio, identidad y permisos. No se modificaron sus 26 tablas, 42 FK ni el SQL
aprobado. Tampoco dist/ o el diseño. No se conectó a la base real ElBravoDelJac.

## Endpoints reales

Base `/api/v1`; autenticación JWT en cookie y CSRF en los POST.

| Métodos | Ruta | Operación |
|---|---|---|
| POST | /diagnosticos/{id}/escaneres | Registrar resultado; 201 nuevo, 200 reintento idéntico |
| GET | /diagnosticos/{id}/escaneres | Resultados completos paginados, ordenados por captura/ID |
| POST | /diagnosticos/{id}/evidencias | Multipart, archivo validado y metadatos; 201/200 |
| GET | /diagnosticos/{id}/evidencias | Metadatos paginados de cargas completas |
| GET | /diagnosticos/{id}/evidencias/{evidenciaId}/archivo | Descarga autorizada e íntegra desde GridFS |

Listas admiten page/pageSize (1–100). Retornan data/meta; altas retornan data.
El listado de escáneres devuelve también equipo, DTC y mediciones, no solo IDs.
Errores: 400/413 por carga inválida/grande, 401 sin sesión, 403 sin permiso/alcance,
404 diagnóstico/evidencia inexistente o evidencia de otro diagnóstico, 409 por
reintento incompatible/carga en curso, 422 por validación y 503 por persistencia.

Ejemplo de contrato de escáner (usar ID SQL real en la ruta):

```json
{
  "requestId": "clave-estable-del-intento",
  "equipo": {"nombre": "Escáner del taller", "version": "1"},
  "dtcs": [{"codigo": "P0300", "descripcion": "Resultado de lectura"}],
  "mediciones": {"rpm": 1200, "temperatura": {"valor": 85.5, "unidad": "C"}}
}
```

capturadoEn RFC3339 es opcional; si falta usa hora del servidor. RegistradoEn y
capturadoPor siempre se fijan en servidor. Capturas se normalizan a UTC/milisegundos
de BSON; no se permite fecha anterior a 2000 o más de cinco minutos en el futuro.
dtcs debe ser un arreglo (hasta 1000), equipo.nombre requerido. Mediciones: objeto
de hasta 64 KiB, profundidad 8 y claves sin `$`, puntos o NUL. Se admiten números
de medición BSON double, no son importes contables. Body JSON máximo 256 KiB.

Evidencia: multipart/form-data con exactamente un `archivo`, `requestId` y
`descripcion` opcional. No se admite diagnosticoId/autor/estado/hash arbitrario en
el body. El nombre original no se almacena ni controla el formato de descarga.

## Autorización y relación entre motores

Cada petición comprueba en PostgreSQL el diagnóstico y el actor activo. GET exige
diagnosticos.ver; POST diagnosticos.crear. Además debe ser el técnico creador,
estar asignado activamente a una OT relacionada, o tener permiso de gestión
ordenes.asignar/usuarios.roles. El permiso de gestión no sustituye el permiso de
diagnóstico. No se copian clientes, vehículos, órdenes, pagos ni contraseñas.

MongoDB guarda diagnosticoId y capturadoPor como strings decimales SQL; _id es
ObjectID. No hay FK ni transacción común entre motores. No se ofrece DELETE de
diagnóstico y cada acceso posterior vuelve a verificar SQL. Si un proceso externo
elimina un diagnóstico por fuera de la API, sus documentos técnicos no se exponen
por estas rutas; la conciliación de referencias SQL externas sigue siendo una
tarea de operación, no una cascada automática destructiva.

## Esquema, índices e inicialización explícita

Driver oficial fijado: go.mongodb.org/mongo-driver/v2 v2.9.2. Colecciones:
resultados_escaner y evidencias_tecnicas; binarios en archivos_tecnicos.files/chunks.
Solo se permiten bases elbravodeljac_tecnico_test y elbravodeljac_tecnico_dev.

`internal/platform/mongodb/validators.json` contiene los validadores BSON ejecutables,
derivados del contrato JSON de database/mongodb. Incluyen _id, fechas BSON Date,
versión, digest de reintento, tipos, límites y propiedades internas de carga.
Se instalan strict/error: MongoDB rechaza documentos que no cumplen el esquema.

Índices: unique(diagnosticoId,requestId) en ambas colecciones;
(diagnosticoId,capturadoEn descendente); unique(gridfsFileId) en evidencias; los
índices estándar GridFS de filename/uploadDate y unique(files_id,n).

El servidor API **no instala ni cambia validadores al arrancar**. Inicializar el
entorno separado con sus variables privadas, desde backend/:

```powershell
go run ./cmd/mongo-init
go run ./cmd/api
```

Si hay validadores diferentes, se rechaza la operación y se solicita revisión;
no se sobrescriben silenciosamente mediante collMod. Cada operación técnica
comprueba esquema e índices únicos, evitando escrituras sin deduplicación efectiva.
GET /readyz verifica SQL y, cuando está configurado, MongoDB/esquema. Sin URI,
readyz declara technicalConfigured=false y los endpoints técnicos devuelven 503;
la parte SQL sigue disponible. URI configurado no equivale a conexión verificada.

## GridFS y reintentos

Se eligió GridFS para mantener el binario fuera de los metadatos y usar un bucket
privado con recuperación por chunks; no porque un archivo de 10 MiB exceda BSON.
Límite 10 MiB. Formatos PDF/JPEG/PNG según contenido detectado, no extensión o MIME
del multipart. Imágenes: decodificación completa y máximo 40 megapíxeles. PDF:
firma y final estructural `%PDF-`/`%%EOF`; no se afirma validación semántica completa
ni análisis antivirus. Descarga como attachment, nombre generado, nosniff y CSP
sandbox; jamás una URL pública o ruta de disco controlada por el cliente.

SHA-256 y tamaño se comprueban antes de entregar bytes. La descarga se mantiene
acotada a 10 MiB y se valida antes de emitir 200; chunks faltantes/corrupción generan
error de persistencia, sin entregar un archivo parcial como si fuera válido.

requestId es obligatorio, estable y único dentro del diagnóstico y colección.
Se calcula un digest canónico del contenido de negocio. El mismo requestId y
contenido devuelve el registro anterior sin cambiar autor/fecha; contenido distinto
produce 409. El nombre original de archivo no forma parte de ese contenido.

Para evidencias se reserva primero un documento `uploading` con lease de cinco
minutos y candidato GridFS. Una reserva activa concurrente devuelve 409; el cliente
puede volver a consultar/reintentar la misma solicitud. Solo al terminar el archivo
se confirma `ready`. Listado y descarga excluyen cargas incompletas. Una reserva
vencida puede recuperarse con otro candidato, sin que el trabajador anterior pueda
confirmar metadatos sobre él. Un resultado ambiguo se consulta antes de eliminar
archivos potencialmente válidos. Fallos conocidos compensan candidato y reserva.

Para interrupciones donde no fue posible compensar por desconexión, hay una
herramienta explícita de conciliación en este bucket dedicado:

```powershell
go run ./cmd/mongo-reconcile          # inspección, no elimina
go run ./cmd/mongo-reconcile --repair # limpia candidatos antiguos no referenciados
```

Considera candidatos de más de diez minutos, chunks sin files y reservas vencidas.
Conserva evidencias ready y leases activos; no borra colecciones ni datos SQL.
No procesa archivos existentes que declaren otro managedBy. Debe programarse o
ejecutarse operacionalmente si se desea limpieza periódica; no se creó un scheduler.

## Entornos aislados

compose.test.yaml: MongoDB y PostgreSQL internos, sin puertos públicos y con tmpfs.
La suite provisiona la base MongoDB de prueba. Solo ese Mongo de pruebas habilita
enableTestCommands para inyectar fallos controlados; no el entorno de desarrollo.

compose.dev.yaml: bases _dev en volúmenes nuevos del proyecto aislado, sin publicar
sus puertos. Solo API queda en localhost:8080. Requiere variables privadas
DEV_POSTGRES_PASSWORD, DEV_MONGO_PASSWORD y JWT_SECRET_BASE64. Usar contraseñas
aleatorias hexadecimales evita caracteres reservados en las URI interpoladas.
Es un entorno local de desarrollo con cuentas de aprovisionamiento; no constituye
un despliegue de producción ni una configuración de privilegio mínimo para él.

```powershell
docker compose -f compose.dev.yaml up -d postgres-dev mongo-dev
docker compose -f compose.dev.yaml run --rm mongo-init
docker compose -f compose.dev.yaml up -d api
```

No monta datos del servicio local ni usa ElBravoDelJac. PostgreSQL se inicializa
solo en el volumen de desarrollo nuevo; no se siembran cuentas reales. Conservar
las variables en el entorno privado o un archivo excluido de Git; nunca publicar
la salida expandida de compose config con secretos.

## Pruebas realmente ejecutadas

`scripts/test-local-databases.ps1` arrancó MongoDB instalado y un PostgreSQL nuevo,
cada uno con carpeta temporal y puerto loopback aleatorio. Ejecutó go test ./...
-count=1 -v con ambas conexiones. PostgreSQL se detuvo al finalizar y el proceso
MongoDB temporal se cerró; ningún servicio real fue utilizado.

Pasaron A+B+C+D, incluyendo:

- Persistencia BSON, validadores reales, índices/reintentos concurrentes de escáner.
- Subida y descarga byte a byte de PNG, PDF permitido, falsos formatos y tamaño.
- Reintento idéntico, conflicto de contenido y concurrencia de evidencias.
- JWT, CSRF, autorización SQL por diagnóstico y descarga cruzada denegada.
- Fallo real inyectado en la actualización de metadatos después del upload; limpieza
  de la reserva y éxito del reintento posterior, sin modificar diagnóstico SQL.
- Recuperación de reserva vencida; conciliación dry-run y reparación real de
  archivo/chunks huérfanos conservando una evidencia lista.
- MongoDB inaccesible: 503 técnico, mientras login y diagnóstico SQL responden 200.
- Inventario registrado: 71/71 operaciones originales, 74 con ampliaciones.

go vet ./... y las tres verificaciones del frontend también pasaron. Los dos
Compose validaron con config --quiet. **No se ejecutaron contenedores/build Docker**:
su motor desktop-linux no está disponible. La conexión real sí se probó por el
camino nativo descrito, usando registros sintéticos solo en las bases de pruebas.

## Pendiente para conectar el frontend completo

Reemplazar progresivamente servicios locales por adaptadores HTTP asíncronos;
gestionar sesión/CSRF, carga, errores y reintentos estables; asociar conceptos a
IDs de catálogos; respetar roles N:M y el rechazo de especialidad no persistible.
Acordar la interfaz de escáner/evidencias (no existe en el diseño actual), integrar
upload/descarga y hacer pruebas de navegador del recorrido completo. La demo
permanece intacta; no se cambiaron pantallas por iniciativa de esta fase.

Fuentes del driver/esquema:
[GridFS oficial](https://www.mongodb.com/docs/drivers/go/current/crud/gridfs/),
[cliente Go](https://www.mongodb.com/docs/drivers/go/current/connect/mongoclient/),
[validación BSON](https://www.mongodb.com/docs/manual/core/schema-validation/specify-json-schema/).
