# Backend Go vigente

Go sirve `dist/` y `/api/v1` en el mismo origen. SQL Server 2025 mantiene transacciones, cuentas y permisos. MongoDB guarda resultados de escáner y evidencias en GridFS vinculados a diagnósticos SQL. Las dependencias están fijadas en `go.mod` y `go.sum`. Usa Go 1.26 o superior.

## Ejecutar

Para instalar desde cero, sigue la [guía de SQL Server y MongoDB](../database/sqlserver/README.md#instalación-local-con-docker-compose). Para reanudar una instalación existente, usa PowerShell 7 y Docker Desktop con motor Linux. Ejecuta desde la raíz:

```powershell
./backend/scripts/start-dev.ps1
./backend/scripts/start-dev.ps1 -Restart
```

Abre http://localhost:8080/. El arranque habitual usa los contenedores y volúmenes ya provisionados por `compose.sqlserver.dev.yaml`. Los Compose de PostgreSQL, las migraciones 001–009 y `000_base_legacy.sql` corresponden a pruebas y mantenimiento históricos.

## Configuración

`compose.sqlserver.dev.yaml` configura `SQLSERVER_DSN`, `JWT_SECRET_BASE64`, `APP_ORIGIN`, `MONGODB_URI` y `MONGODB_DATABASE` a partir de la configuración privada. `HTTP_ADDR`, `COOKIE_SECURE` y `FRONTEND_DIR` ajustan la escucha, las cookies y la ubicación del frontend. Las pruebas usan bases aisladas.

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
