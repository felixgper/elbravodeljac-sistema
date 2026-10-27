# Backend Go: arquitectura inicial

Estado: esqueleto ejecutable y contratos de rutas. Las APIs de negocio todavía
devuelven 501; no hay controladores de PostgreSQL/MongoDB ni autenticación real.

Desde backend/, con Go 1.22 o superior:

```powershell
go test ./...
go vet ./...
go run ./cmd/api
```

Escucha por defecto en http://127.0.0.1:8080. GET /healthz informa que el proceso
vive; GET /readyz devuelve 503 porque la persistencia no está implementada.
No sirve ni altera dist/. No requiere secretos ni instala paquetes externos.
HTTP_ADDR puede cambiarse mediante una variable de entorno del proceso.
.env.example documenta variables futuras; no se carga automáticamente.

## Organización

- cmd/api: composición y arranque del servidor.
- internal/httpapi: transporte, catálogo de rutas y pruebas del esqueleto.
- internal/application: futuros casos de uso y transacciones.
- internal/domain: futuras reglas, entidades e interfaces.
- internal/platform/postgres: futura persistencia transaccional.
- internal/platform/mongodb: futura persistencia técnica documental.
- ../docs/backend: análisis del esquema, correspondencia y plan de implementación.
- ../database/mongodb: modelo declarativo propuesto; no es un script de instalación.

El catálogo completo de rutas está en docs/backend/ENDPOINTS.md. La integración
del navegador se hará posteriormente por adaptadores de servicios asíncronos,
conservando diseño, rutas visuales, formularios y funcionalidad existentes.
