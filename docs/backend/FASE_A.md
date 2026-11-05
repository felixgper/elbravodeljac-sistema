# Fase A: backend real

Rama: `backend/arquitectura-go-mongodb`. Se mantienen las 26 tablas y 42 claves
foráneas del esquema aprobado, sin migraciones ni cambios en dist/. No se accede
a la base real ElBravoDelJac. El proceso rechaza DSN cuyo nombre de base no sea
`elbravodeljac_test` o `elbravodeljac_dev` antes de abrir el pool.

## Endpoints implementados

Todos bajo `/api/v1`; estos endpoints ya no son placeholders 501.

| Métodos | Ruta | Comportamiento |
|---|---|---|
| POST | /auth/login | Verifica bcrypt, emite JWT en cookie HttpOnly y devuelve csrfToken |
| GET | /auth/me | Identidad, todos los roles, permisos efectivos y csrfToken |
| POST | /auth/logout | Revoca la sesión y elimina la cookie |
| GET, POST | /clientes | Listado con búsqueda/paginación y registro |
| GET, PATCH, DELETE | /clientes/{id} | Consulta, actualización parcial y borrado sujeto a FK |
| GET, POST | /clientes/{id}/contactos | Listado/alta vinculados; alta requiere pagos.entregar |
| GET, POST | /vehiculos | Búsqueda/paginación/filtro clientId y alta con historial inicial |
| GET, PATCH, DELETE | /vehiculos/{id} | Consulta, actualización parcial y borrado protegido |
| POST | /vehiculos/{id}/transferencias | Cambia propietario con verificación e historial atómico |
| GET | /vehiculos/{id}/historial | Propietarios y atenciones/diagnósticos/cotizaciones/OT existentes |

GET /healthz informa proceso vivo; GET /readyz comprueba PostgreSQL. Datos en
`{data: ...}`, listas con `meta: {page,pageSize,total}`. Errores 400/401/403/404/
409/415/422/429/503. No se devuelven hashes, DSN ni errores SQL al cliente.

## Seguridad implementada

- JWT HS256, clave de mínimo 32 bytes, issuer/audience/exp/iat/nbf/sub/jti y vida
  de 15 minutos. No acepta algoritmo none ni otro algoritmo de firma.
- Cookie bravo_access: HttpOnly, SameSite=Strict y Secure por defecto. Configuración
  explícita permite HTTP sin Secure únicamente en localhost para pruebas.
- Origin exacto obligatorio en login y escrituras. X-CSRF-Token ligado al JWT
  en cada escritura autenticada, incluido logout. El token se obtiene de login o
  me y no se guarda en localStorage. No hay CORS abierto ni tokens en URLs.
- Contraseñas bcrypt; no se aceptan SHA-256 demo ni contraseñas de más de 72 bytes
  (límite bcrypt). Rate limit de login: cinco intentos/IP/minuto, estado acotado.
  No confiar en X-Forwarded-For sin un proxy configurado deliberadamente.
- Permisos `modulo.accion` consultados desde PostgreSQL en cada petición, uniendo
  **todos** los roles asignados. Cambios de permisos/desactivación tienen efecto
  en la siguiente petición. No se modifica usuario_roles ni rol_permisos en login.
- Allowlist de sesiones en memoria: logout invalida el JWT inmediatamente; reiniciar
  servidor invalida todas las sesiones. Adecuado para una instancia de desarrollo.
  Compartir sesiones/limitador entre instancias y añadir observabilidad son trabajos
  pendientes antes de producción. No hay refresh token en esta fase.

## Integridad implementada

Clientes: longitudes SQL sin truncar (nombres 150, documentos 20, teléfono 20,
correo 150), tipo/número de documento y formato de correo/teléfono. Normalización
de número de documento a mayúsculas y correo a minúsculas; correo vacío→NULL.
PATCH bloquea la fila y combina solo campos permitidos.

Vehículos: placa en mayúsculas sin espacios, equivalencia con/sin guion verificada
con bloqueo advisory transaccional para altas concurrentes, marca 40/modelo 60,
kilometraje NUMERIC(10,2) como cadena decimal exacta; no se usa float para km.
Alta + propietario inicial atómicos; PATCH bloquea la fila e impide disminuir km;
transferencia cierra historial vigente y abre el nuevo en una sola transacción.
El actor se toma de la sesión. Placa y cliente no son editables con PATCH.

El borrado de cliente respeta FK. El de vehículo solo elimina su registro inicial
de propietario si no hay transferencias; cualquier FK operativa revierte ambos
borrados. Nunca usa CASCADE ni borra recepciones. Restricciones únicas/FK→409;
datos fuera de CHECK/longitudes→422. Transacciones SQL parametrizadas.

Campos SQL TEXT mantienen una protección adicional de tamaño de API (address y
notes: 10000 caracteres; observation de transferencia: 4000); el body JSON se
limita a 64 KiB. No hay recorte silencioso. Estos límites de API no son columnas
nuevas ni modificaciones de PostgreSQL.

## Decisiones conservadas para próximas fases

- Cotizaciones deberán usar IDs reales de servicios/repuestos; no se quitará el
  CHECK que exige una referencia por línea.
- Administración futura de un rol por trabajador: conservar asignaciones N:M
  adicionales y acordar cómo identificar el rol administrado; hoy no se editan roles.
- Proveedores queda sin implementar ni modificar. Propuesta pendiente: decidir
  si parts es especialidad textual o relación de repuestos; no guardar en contacto
  ni agregar columna sin aprobación de la propuesta.
- Frontend conserva demo, diseño y servicios locales; todavía no apunta al backend.
- Recepciones, Diagnósticos, Cotizaciones, OT, Pagos, Entregas y MongoDB quedan
  para las siguientes fases. Sus rutas reservadas requieren autenticación y devuelven 501.

## Referencias

- [pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool).
- [Validación JWT](https://pkg.go.dev/github.com/golang-jwt/jwt/v5).
- [bcrypt y su límite de entrada](https://pkg.go.dev/golang.org/x/crypto/bcrypt).
- [Protección CSRF OWASP](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html).
