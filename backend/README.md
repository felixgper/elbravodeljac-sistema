# Backend Go vigente

Monolito modular: Go sirve `dist/` y `/api/v1` en el mismo origen. PostgreSQL mantiene transacciones, cuentas y permisos; MongoDB guarda resultados de escáner y evidencias/GridFS vinculados a diagnósticos SQL. Dependencias fijadas en `go.mod`/`go.sum`; Go 1.26 o superior (verificado con 1.27.1).

## Ejecutar

Desde la raíz, con Docker Desktop Linux y PowerShell 7:

```powershell
./backend/scripts/start-dev.ps1
```

Abrir http://localhost:8080/. Las credenciales iniciales son privadas (`backend/.cache/acceso-desarrollo.txt`); una base ya provisionada conserva sus cuentas. El comando `dev-bootstrap` solo permite `elbravodeljac_dev`, confirmación explícita y ausencia de usuarios.

La ruta técnica de instalación nueva es `database/postgresql/esquema_entrega.sql` (único esquema físico consolidado) → catálogo documental por el migrador → inicialización MongoDB → provisión autorizada → API. Su marca evita reaplicar migraciones históricas. Las migraciones embebidas 001–009 y `000_base_legacy.sql` se conservan únicamente para actualizar/probar bases anteriores. Nunca ejecutar el esquema final sobre bases existentes ni recrear volúmenes.

Antes de reiniciar PostgreSQL de una instalación existente, revisa el origen del bind mount con `docker inspect`. Compose puede recrear el contenedor si cambia el montaje. Verifica el volumen y los respaldos antes de programar esa recreación. PostgreSQL solo ejecuta el SQL de inicialización con un directorio de datos vacío.

`compose.dev.yaml` conserva volúmenes de desarrollo; `compose.test.yaml` utiliza `tmpfs` y red aislada. El script manual `scripts/backup-audit.ps1` y sus dos Compose de auditoría se reservan para el procedimiento de respaldo y restauración aislada. El script exige identificar los proyectos y sus archivos Compose antes de leer la fuente o restaurar en QA.

## Configuración

`.env.example` es documentación; Go no la carga automáticamente. Configurar privadamente `POSTGRES_DSN`, `JWT_SECRET_BASE64`, `APP_ORIGIN` y, para documentos técnicos, `MONGODB_URI`/`MONGODB_DATABASE`. `HTTP_ADDR`, `COOKIE_SECURE` y `FRONTEND_DIR` permiten ajustar escucha, cookies y ubicación del mismo frontend. Las bases admitidas son las separadas `_dev`/`_test` previstas por `internal/config`; no usar `ElBravoDelJac`.

Desde `backend/`, con variables privadas configuradas: `go run ./cmd/api`. `cmd/mongo-init` instala explícitamente validadores e índices del entorno permitido; arrancar la API no los cambia. Sin MongoDB configurado, las operaciones técnicas informan indisponibilidad; no simulan persistencia.

En HTTP loopback de desarrollo, `COOKIE_SECURE=false`; en HTTPS usar Secure. Login exige Origin y JSON; devuelve cookie HttpOnly y token CSRF. Las escrituras exigen Origin y `X-CSRF-Token`. Logout revoca sesión. Los permisos efectivos unen todos los roles existentes, sin bypass por su nombre. Se protege el último acceso administrativo y se impide autoescalamiento.

## Contratos y organización

- [API REST y reglas operativas](../docs/backend/ENDPOINTS.md).
- [MongoDB y su contrato BSON](../database/mongodb/README.md).
- [Generación y conservación de PDF](internal/platform/technicalsheet/README.md).

`cmd/` contiene ejecutables; `internal/httpapi` transporte y autorización; `internal/auth` JWT/CSRF; `internal/domain` DTO y validaciones; `internal/application` interfaz Store; `internal/platform/postgres` SQL/transacciones/migraciones; `internal/platform/mongodb` driver oficial/índices/GridFS; `internal/platform/technicalsheet` generadores aprobados. `integration/` contiene pruebas HTTP con motores reales aislados.

## Pruebas

Desde la raíz: `node checks/run.mjs` (ocho suites), `./backend/scripts/test-isolated.ps1` (QA nuevo). Para Chrome, runners nativos y requisitos: [checks/README.md](../checks/README.md). No reutilizar bases ni volúmenes. El runner no reconstruye ni reemplaza el aplicativo habitual.

Ejecuta `go test ./...` y `go vet ./...` desde `backend/`. Las pruebas que requieren una conexión de integración o una opción explícita se omiten cuando falta esa configuración.

## Límites de despliegue

Las sesiones/revocaciones y el limitador son de una instancia; el reinicio invalida sesiones. No hay refresh token. Un despliegue empresarial requiere revisión de TLS/proxy, estado compartido y observabilidad. Esta documentación no autoriza cambios del entorno habitual ni publica secretos.
