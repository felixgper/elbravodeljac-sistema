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

El modal puede abrirse antes de guardar el diagnóstico. Al confirmar el primer
anexo se guarda automáticamente un borrador SQL con los textos actuales y se usa
su ID real para MongoDB. Si falla el anexo, se conserva el borrador y se reutilizan
su ID y la solicitud para reintentar sin duplicar. Finalizar es siempre explícito. Al cerrar el diagnóstico se impiden nuevas escrituras también en Go,
con coordinación de bloqueo en el adaptador relacional. Los archivos guardados
siguen disponibles con permiso de consulta. Reintentar conserva la misma solicitud;
«Preparar otra carga» exige confirmación si el resultado previo es incierto.

## Encuestas de satisfacción

El módulo usa una única colección `encuestas_invitaciones`, con la respuesta
embebida: consumir el enlace y guardar sus cinco notas/comentario es una operación
atómica de un documento, también en MongoDB standalone. No usa GridFS ni duplica
nombres, teléfonos o documentos. El contrato es
[surveys-validator.json](../../backend/internal/platform/mongodb/surveys-validator.json).

- Una invitación por entrega (índice único), token aleatorio de 256 bits con
  SHA-256 e índice único de hash para validación, y copia cifrada para recuperación.
  Vencimiento de 14 días; nunca se guarda el token en texto plano.
- Generar requiere `pagos.entregar` y una entrega SQL existente, verificada y de
  orden finalizada. Consultar estado/seguimiento requiere `pagos.ver`;
  estadísticas agregadas requieren `reportes.ver`. Se comprueban permisos efectivos.
- Regenerar requiere confirmación explícita, invalida el enlace anterior y está
  limitado a una vez por minuto. No se regenera una encuesta respondida. Generar con `regenerate:false` reutiliza
  el enlace vigente (HTTP 200), incluso después de reiniciar la API e iniciar
  sesión nuevamente; no renueva su vencimiento. La clave privada debe preservarse
  junto con los respaldos. Rotarla impide recuperar las copias antiguas; los enlaces
  ya compartidos siguen validándose por hash hasta su uso o vencimiento.
- El enlace usa fragmento (`/encuesta#t=…`), nunca query ni almacenamiento
  persistente del navegador. MongoDB conserva además una copia del secreto cifrada
  con AES-256-GCM, una clave derivada con propósito separado de la clave privada
  de la aplicación y autenticación del vínculo con la entrega y su hash. Solo una
  cuenta con permiso efectivo `pagos.entregar` puede recuperarla; el seguimiento
  con solo `pagos.ver` no devuelve enlaces, hashes ni secretos cifrados. Abrir WhatsApp es una acción local, no evidencia
  de envío, entrega o lectura. El envío es manual y voluntario.
- El velocímetro usa exclusivamente la quinta respuesta: rojo `[1,3)`, amarillo
  `[3,4)` y verde `[4,5]`. No incluye invitaciones sin respuesta. Sus filtros
  Semana/Mes/Año son independientes del filtro Día/Semana/Mes de ingresos, que
  reutiliza `/reportes`, sus permisos y `jac_movimientos_pagos`, sin cambiar SQL.
  Ambos usan intervalos de calendario de Lima. Las consultas filtran por fecha de respuesta en el intervalo semiabierto
  seleccionado. La quinta nota da el promedio general/distribución; no es NPS.
- No hay TTL: los resultados no se eliminan al vencer el enlace. La provisión y
  las escrituras rechazan cualquier índice TTL en esta colección, sin retirarlo
  automáticamente. Los comentarios
  no se devuelven en estadísticas ni seguimiento; no incluir datos personales.

### Incorporación posterior, previa autorización

1. En QA nuevo y aislado, ejecutar `go test ./internal/platform/mongodb -run TestSurvey`
   con `SURVEY_QA_URI` temporal y `SURVEY_QA_ISOLATED=yes`. Nunca usar la conexión
   oficial. El test crea su propia base con nombre aleatorio y retira su colección.
2. Ejecutar las pruebas SQL Server de flujo existentes y validar la consulta de
   entrega en QA, así como PDF/GridFS y navegador móvil. No escribir en producción.
3. Conservar respaldo y versión de API anterior. Construir la imagen que incluye
   `surveys-init`; ejecutar **explícitamente**, con aprobación y credencial de
   provisión existente, `/app/surveys-init` en la base documental correcta. Este
   comando solo crea/verifica la nueva colección y sus índices; no aplica `collMod`
   ni cambia colecciones técnicas, SQL, GridFS o volúmenes. Rechaza un validador
   existente distinto. No ampliar permisos automáticamente.
4. Verificar la nueva colección e índices y recrear solo la API con la configuración
   existente. Su arranque no inicializa encuestas. Conservar dominio HTTPS para
   que los enlaces emitidos usen el origen público correcto. No agregar recursos Azure.
5. Reversión: restaurar únicamente la imagen API anterior; conservar la colección
   de encuestas. No restaurar ni eliminar bases/volúmenes para revertir el aplicativo.

La API limita peticiones públicas a 30 por minuto, con mapa acotado e independiente
del login. Por defecto utiliza la IP de conexión e ignora las cabeceras reenviadas.
Detrás del Caddy verificado, configurar explícitamente `SURVEY_TRUSTED_PROXY_IPS`
con su IP privada exacta (varias IP separadas por comas, sin CIDR ni comodines).
Solo desde esos pares acepta un único `X-Forwarded-For` válido. Listas, cabeceras
duplicadas o direcciones inválidas conservan el límite de la IP de conexión.
`X-Real-IP` no interviene. No se cambia el limitador de autenticación.

El Caddy frontal debe reemplazar las cabeceras externas, como hace `reverse_proxy`
sin `trusted_proxies`. Si se añade otro proxy/CDN, se debe revisar de nuevo la
cadena antes de habilitar confianza. La API no debe publicar su puerto directamente
a Internet. Revalidar la IP si Caddy se recrea: una IP desactualizada conserva la
protección pero vuelve a compartir el límite. Clientes bajo la misma IP pública/NAT
siguen compartiendo 30 peticiones/minuto. No registrar cuerpos públicos ni tokens.

### Publicación y reversión en el servidor existente

Estos pasos requieren autorización posterior; no se ejecutan al arrancar la API.

1. Crear el PR únicamente con los archivos del módulo y sus pruebas; excluir
   Power BI, cachés, archivos privados y el manual independiente. Fusionar solo
   después de revisar el diff, resultados y configuración del proxy.
   Antes del commit autorizado, comprobar que el índice Git esté vacío y añadir
   solo las rutas revisadas con `git add -- <rutas>`; verificar que
   `git diff --cached --name-only` coincida exactamente con esa lista. No usar
   `git add .`. Conservar y volver a comprobar los hashes de cambios ajenos.
2. Conservar la imagen API anterior por ID y etiqueta de reversión, la configuración
   privada y respaldos SQL/Mongo verificados. Preservar **el mismo**
   `JAC_JWT_SECRET_BASE64` en `/opt/elbravodeljac/private/azure.env`; Compose lo
   entrega como `JWT_SECRET_BASE64`. No generar otra clave al construir o desplegar.
   Guardar una copia cifrada fuera de la VM junto con los respaldos Mongo, con
   acceso restringido y manifiesto SHA-256 del archivo cifrado. Probar su recuperación
   en un entorno aislado antes de depender de ella.
3. Construir la imagen aprobada que contiene `/app/surveys-init`. Preparar únicamente
   la imagen del servicio `api` y añadir `SURVEY_TRUSTED_PROXY_IPS` a su entorno
   privado con la IP de Caddy recién inspeccionada. Validar Compose sin imprimir
   secretos y comparar que SQL Server, MongoDB, redes y volúmenes no cambien.
   La variable debe figurar tanto en `private/azure.env` como en el bloque
   `services.api.environment` del Compose privado mediante
   `SURVEY_TRUSTED_PROXY_IPS: ${SURVEY_TRUSTED_PROXY_IPS:?}`. No basta con añadirla
   al archivo de entorno sin entregarla al contenedor. Usar `docker compose ...
   config --quiet` para validar sin volcar conexiones ni secretos.
4. Antes de provisionar, consultar solo los metadatos de `encuestas_invitaciones`:
   si existe, revisar validador estricto e índices. No aceptar TTL ni sustituir
   validadores incompatibles. Usar la credencial de provisión existente, limitada
   a la base documental; no conceder privilegios automáticamente.
5. Con la configuración privada ya preparada y la imagen candidata seleccionada,
   ejecutar exclusivamente el provisionador de encuestas, sin publicar puertos ni
   levantar dependencias, y comprobar el resultado antes de continuar:

   ```sh
   cd /opt/elbravodeljac/deploy
   dc() { docker compose -p elbravodeljac-azure-demo --env-file ../private/azure.env -f azure.compose.yaml "$@"; }
   dc run --rm --no-deps --pull never --entrypoint /app/surveys-init api
   ```

   Deben existir `_id_`, `survey_delivery_unique` y `survey_token_unique` (únicos),
   `survey_answer_date` y `survey_tracking`; ninguno lleva `expireAfterSeconds`.
   El comando no inserta respuestas, no ejecuta `mongo-init` ni modifica SQL,
   colecciones técnicas o GridFS. Si falla, detenerse, sin reparar datos ni permisos.
6. Actualizar solo la API: `dc up -d --no-deps --no-build api`. Comparar los IDs,
   imágenes, montajes y recuentos protegidos; comprobar `/healthz`, `/readyz`, login,
   permisos, dashboard y rutas públicas sin registrar datos de prueba en producción.
   Una entrega real autorizada permite preparar/reutilizar el enlace y revisar
   WhatsApp sin enviar automáticamente. Comprobar una respuesta voluntaria real
   y su Q5 en el período de Lima; usar QA para pruebas sintéticas o destructivas.
7. Si falla una comprobación crítica, utilizar un overlay privado que cambie **solo**
   `services.api.image` a la etiqueta anterior y ejecutar
   `dc -f ../private/rollback-api.yaml up -d --no-deps --no-build api`.
   Comprobar salud e inicio de sesión. Conservar la clave y la colección de encuestas,
   incluidas respuestas recibidas; no restaurar bases ni eliminar índices/volúmenes.

La interfaz de entrega no muestra listas ni estados de personas que no responden.
Solo ofrece compartir/copiar y, cuando corresponde, confirmación de reemplazo.
El validador nuevo incluye `tokenCipher` opcional para reconocer documentos
anteriores. `surveys-init` sigue rechazando un validador instalado distinto: no
actualiza colecciones mediante `collMod` automáticamente. Si una instalación
tuviera ya el validador previo, hace falta revisar y autorizar explícitamente su
actualización antes de publicar. Un documento anterior sin copia cifrada requiere
reemplazo confirmado una sola vez para poder recuperar su enlace en el futuro.
