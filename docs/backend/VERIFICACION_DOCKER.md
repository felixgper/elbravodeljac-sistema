# Ejecución real verificada con Docker y Chrome

Fecha: 27 de octubre de 2026. Rama: `backend/arquitectura-go-mongodb`.

## Abrir la aplicación

**Dirección exacta:** http://localhost:8080/

Es el modo REAL del frontend aprobado, servido por Go en el mismo origen que
`/api/v1`. No abrir `/demo/` para trabajar con la base de desarrollo.

- Usuario inicial: `admin_pruebas`.
- Contraseña aleatoria: consultar localmente `backend/.cache/acceso-desarrollo.txt`.
  No se publica en este informe ni se incluye en Git.
- Configuración privada: `.env.fase-e.local`, también excluida de Git.
- La cuenta fue creada con bcrypt exclusivamente en la nueva base aislada.
  El comando de inicialización se niega a reemplazar cuentas si ya existen usuarios.

Desde PowerShell en la raíz de este repositorio, con Docker Desktop activo:

```powershell
./backend/scripts/start-dev.ps1
```

El script conserva la configuración privada existente, construye las imágenes,
espera ambas bases, verifica/inicializa MongoDB, provisiona la cuenta solo en una
base sin usuarios y deja Go en segundo plano. Para arrancar con las imágenes ya
construidas puede usarse `-SkipBuild`. No detiene proyectos ajenos ni elimina
volúmenes. Reutilizar el mismo archivo privado al reiniciar.

## Inventario y aislamiento

`docker version` y `docker info` confirmaron motor Docker Desktop activo:
Docker Engine 29.7.2 y Docker Desktop 4.89.0. La consulta inicial sin acceso fuera
del aislamiento devolvió permission denied; con acceso al motor se comprobó
su disponibilidad. El motor no estaba ausente en esta ejecución.

Proyecto previo encontrado, mantenido intacto:

| Contenedor | Estado inicial y final | Puerto |
|---|---|---|
| `el-bravo-del-jac-postgres-1` | En ejecución, saludable, mismo ID y fecha de arranque | 127.0.0.1:5433 |
| `el-bravo-del-jac-mongo-1` | En ejecución, mismo ID y fecha de arranque | Sin puerto publicado |
| `sql22dev01` | Detenido desde antes de la tarea | Sin cambios |
| `oracle-free-lite` | Detenido desde antes de la tarea | Sin cambios |

El proyecto anterior usa otro archivo Compose, en
`C:\Users\Usuario\Documents\Codex\el-bravo-del-jac\compose.yaml`.

Recursos nuevos de este repositorio:

| Proyecto | Finalidad | Almacenamiento / acceso |
|---|---|---|
| `elbravodeljac-fase-e-dev` | Aplicación que queda funcionando | Volúmenes propios persistentes; solo Go publica 127.0.0.1:8080 |
| `elbravodeljac-fase-e-tests` | Integración automatizada | Bases tmpfs separadas, sin puertos ni volúmenes persistentes; contenedores de pruebas terminados con código 0 |

Desarrollo usa PostgreSQL 18.6, base `elbravodeljac_dev`, y MongoDB 9.0,
base `elbravodeljac_tecnico_dev`. Las pruebas usan únicamente
`elbravodeljac_test` y `elbravodeljac_tecnico_test`.

No se accedió al PostgreSQL original del equipo en el puerto 5432 ni a la base
`ElBravoDelJac`. El puerto 5432 indicado dentro de Compose es el puerto privado
del contenedor nuevo `postgres-dev`, no el puerto de Windows. El SQL aprobado
solo inicializó el volumen nuevo vacío de desarrollo y las bases nuevas tmpfs.
No se aplicó un respaldo sobre bases existentes, ni se cambió el modelo físico.

## Corrección necesaria para publicar Go

Se detectó un problema de la configuración Compose: Go arrancaba y conectaba
ambas bases, pero `NetworkSettings.Ports` permanecía vacío al tener únicamente
una red `internal: true`, a pesar del binding solicitado para 8080.

Se agregó `web-entry` solo a la API. Las bases siguen exclusivamente en
`dev-only`, interna y sin puertos publicados. Después del cambio, `docker port`
confirmó `8080/tcp -> 127.0.0.1:8080` y `/readyz` respondió:

```json
{"status":"ready","technicalConfigured":true}
```

También se priorizó IPv4 en los ensayos Node para que el cliente HTTP no intente
solo `::1` cuando el binding local está en 127.0.0.1. Chrome puede abrir la
dirección localhost indicada. No se modificaron colores, estilos, componentes
visuales, formularios ni el SQL del frontend aprobado.

## Pruebas realmente ejecutadas

| Prueba | Resultado |
|---|---|
| `docker version`, `docker info`, inventarios de Compose, contenedores, redes y volúmenes | Aprobados |
| Construcción de imágenes Go y arranque de PostgreSQL/MongoDB propios | Aprobados |
| `go test ./... -count=1 -v` en Compose de pruebas | A/B/C/D, unidades y HTTP aprobados; 26 tablas y 42 FK verificadas |
| `go vet ./...` en la imagen Docker de pruebas | Aprobado, sin hallazgos |
| `TestDevelopmentTargetRejectsRealDatabase` | Aprobado: rechazo de base real y confirmación ausente |
| `node checks/run.mjs` | Cuatro suites aprobadas |
| `checks/live-docker.mjs` con Chrome instalado en Windows contra el Go de Docker | Aprobado |
| `checks/read-persisted-docker.mjs` después de reiniciar nuestros tres contenedores | Aprobado |
| Arranque repetido y bootstrap sobre desarrollo ya provisionado | Aprobado; no se reemplazaron cuentas ni permisos |

El recorrido real de Chrome registró mediante los formularios existentes:
cliente, vehículo, recepción, diagnóstico, cotización con servicio y repuesto
de catálogo, orden con mecánico, avance, prueba satisfactoria, pago completo y
entrega. Después de recargar se consultaron nuevamente los datos desde las APIs.

Se verificaron 14 vistas reales, cookies HttpOnly/SameSite, 401 sin sesión,
403 al omitir CSRF, rechazo de autoedición del rol administrativo y denegación
de Clientes para un auxiliar. No hubo errores JavaScript ni datos demo mezclados.

MongoDB se probó desde el servicio JavaScript del navegador: resultado de escáner,
archivo PNG técnico, descarga GridFS y reintentos que conservan el mismo documento.
No se inventaron controles visuales de evidencias que el diseño no contiene.

Se reiniciaron exclusivamente Go, PostgreSQL y MongoDB de **nuestro** proyecto
de desarrollo, sin eliminar contenedores ni volúmenes. Una sesión nueva de Chrome
consultó otra vez cliente, OT finalizada, pago de 120.00, saldo 0.00, entrega,
resultado de escáner y archivo GridFS: todos se conservaron.

Registros de la verificación, conservados para consulta en la aplicación:

- Cliente 1 y vehículo 1.
- Recepción 1, diagnóstico 1 y cotización 1.
- Orden `OT-1`, finalizada; pago 1 y entrega 1.
- Escáner y evidencia ligados al diagnóstico 1 en MongoDB.

Son datos de prueba identificados como verificación Docker, persistidos en los
motores reales aislados. No representan clientes reales del taller.
El reporte JSON sin credenciales y capturas se guardan en
`backend/.cache/live-docker/`, excluido de Git.

## Pruebas omitidas y pendientes

El ensayo Go opcional `TestPhaseEFrontendBrowserPostgreSQLMongoDB` muestra SKIP
dentro de la imagen de Go porque allí no se instalaron Node/Playwright/Chrome
ni se habilitó FRONTEND_E2E. **No se cuenta ese SKIP como aprobado.** La
verificación de navegador sí se ejecutó por separado desde Windows con Chrome
contra la aplicación persistente de Docker, como se detalla arriba.

No quedan bloqueadas las conexiones Docker, Go, PostgreSQL y MongoDB ni el
recorrido principal solicitado. Continúan pendientes los controles visuales de
escáner/evidencias y otros controles no incluidos en el diseño aprobado,
documentados en [FASE_E.md](FASE_E.md). Los servicios y endpoints técnicos sí
fueron probados.

## Repetir la verificación

El ensayo completo crea nuevos registros identificados como pruebas y los
conserva. El ensayo de lectura solo consulta el último reporte persistido.
Requiere Node, Playwright y Chrome en Windows:

```powershell
# Si Playwright está fuera de node_modules, indicar su ruta local privada:
# $env:PLAYWRIGHT_MODULE = 'RUTA_ABSOLUTA_AL_PAQUETE_PLAYWRIGHT'
node checks/live-docker.mjs
node checks/read-persisted-docker.mjs
```

Para repetir las suites Docker, solo en el proyecto temporal identificado:

```powershell
docker compose --env-file .env.fase-e.local -p elbravodeljac-fase-e-tests -f compose.test.yaml up --build --abort-on-container-exit --exit-code-from tests
docker compose --env-file .env.fase-e.local -p elbravodeljac-fase-e-tests -f compose.test.yaml run --rm --no-deps tests go vet ./...
```

Los cambios permanecen en la rama actual, sin merge con main. La configuración
privada y las contraseñas no forman parte de los archivos versionados.
