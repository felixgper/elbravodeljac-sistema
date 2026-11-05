# El Bravo del JAC

Sistema web de gestión de un taller automotriz, desarrollado como proyecto universitario y aplicativo para una empresa. Incluye clientes, vehículos, recepciones, diagnósticos, cotizaciones, órdenes, auxilios, pagos/entrega, empleados, usuarios y permisos.

Tecnologías: HTML/CSS/JavaScript, Go, PostgreSQL transaccional y MongoDB/GridFS para escáner/evidencias. Go sirve el frontend y las APIs en un mismo origen, con JWT en cookies HttpOnly, protección CSRF y permisos del servidor.

## Entrega universitaria vigente

**La estructura REAL de PostgreSQL de desarrollo fue consultada el 5 de noviembre de 2026: 28 tablas, 208 columnas y 50 FK.** No es un conteo supuesto por la presencia de migraciones.

| Entregable | Archivo vigente |
| --- | --- |
| SQL oficial de entrega: estructura real sin datos | [database/postgresql/esquema_entrega.sql](database/postgresql/esquema_entrega.sql) |
| Modelo lógico Crow’s Foot sin tipos | [Draw.io editable](docs/modelos/Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.drawio) · [PDF](docs/modelos/Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.pdf) · [PNG](docs/modelos/Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.png) |
| Comparación con la instancia existente | [Resultado, inventario y alcance](docs/modelos/COMPARACION_BASE_ACTUAL.md) |
| Conceptual Chen V24: resumen del negocio | [Draw.io editable](docs/modelos/V24_Chen_Resumido_Exposicion/El_Bravo_del_JAC_V24_Chen_Resumido.drawio) · [PDF](docs/modelos/V24_Chen_Resumido_Exposicion/El_Bravo_del_JAC_V24_Chen_Resumido.pdf) |

El SQL de entrega incluye tablas, campos, tipos, restricciones, índices, secuencias, vistas, funciones, triggers y extensiones de la base consultada. Es una exportación de **estructura**, no un respaldo de registros ni de cuentas o asignaciones. Omite propietarios y ACL para no depender de cuentas PostgreSQL locales. Los inventarios y diagramas anteriores conservados en `docs/` documentan la evolución; no sustituyen estos entregables. El conceptual resume el negocio y no enumera todas las tablas físicas.

## Login y Dashboard actuales

La versión vigente utiliza **logo Diseño 19**, **Login oscuro de dos columnas** con **Correo o usuario** y **Contraseña**, y **Dashboard carbón/borgoña** con indicadores, filtros y datos de la API en modo real. No hay un frontend nuevo separado: se mantiene `dist/`.

- Entrada: `dist/index.html`; Login y navegación: `dist/assets/js/app.js`.
- Dashboard: `dist/assets/js/views/dashboard.js`.
- Tema actual: `dist/assets/css/gestion-carbon.css`, sobre los estilos base necesarios.
- Identidad visual: `dist/assets/images/diseno19/`.

**Descargar o actualizar GitHub no reemplaza un servidor o contenedor que ya estaba ejecutándose.** Si aparece un diseño anterior, comprueba qué versión está sirviendo la API y reconstruye solo su imagen cuando corresponda; no borres volúmenes ni recrees bases para actualizar la interfaz.

## Ejecutar el aplicativo real

Entorno documentado: Windows, PowerShell 7 y Docker Desktop con motor Linux. Para **preparar un entorno de desarrollo nuevo y separado**, desde la raíz:

```powershell
./backend/scripts/start-dev.ps1
```

Abrir **http://localhost:8080/**. El script genera configuración privada y crea la cuenta inicial de pruebas solo si la base nueva no tiene usuarios. El acceso se guarda en `backend/.cache/acceso-desarrollo.txt`, excluido de Git. No se publican credenciales en este README. Una instalación ya utilizada conserva sus cuentas: utiliza el acceso propio de ese entorno.

Si el entorno solo está detenido, inicia sus contenedores existentes de PostgreSQL/MongoDB, espera que estén saludables y después inicia la API. No hace falta restaurar SQL ni volver a provisionar cuentas para reanudarlo.

### Procedimiento técnico conservado

`compose.dev.yaml` y `start-dev.ps1` siguen usando **esquema inicial → migraciones oficiales 001–009 → provisión del entorno nuevo → API**. No se cambia ese procedimiento para esta entrega.

**`database/postgresql/esquema_actual.sql` mantiene su nombre histórico porque Go, Docker y las pruebas lo utilizan como esquema inicial técnico. No es el SQL vigente para presentar.** Las migraciones siguen siendo necesarias para su evolución.

**No sustituir ese archivo técnico por `esquema_entrega.sql`, ni cargar ambos y volver a aplicar las mismas modificaciones.** El SQL de entrega permite revisar la estructura actual o cargarla por separado en una base vacía independiente; no es una nueva ruta de arranque ni contiene los catálogos y cuentas del procedimiento técnico. No ejecutarlo sobre bases existentes.

## Demostración y pruebas

- **Real:** `http://localhost:8080/`, servido por Go con PostgreSQL/MongoDB y permisos del backend.
- **Demo:** `http://localhost:8080/demo/`, con datos ficticios y almacenamiento del navegador; no acredita seguridad ni persistencia del servidor.
- Live Server o `python -m http.server 8000 --directory dist` sirven la demostración, no el aplicativo conectado. No abrir el HTML mediante doble clic.

```powershell
node checks/run.mjs
cd backend
go test ./...
go vet ./...
```

Go y Node.js se necesitan para herramientas/pruebas ejecutadas fuera de Docker. Las integraciones sin configuración se omiten: no deben presentarse como aprobadas. El procedimiento aislado está en `backend/scripts/test-compose.ps1`; no apuntar pruebas a bases existentes.

Detalles de implementación en [backend/README.md](backend/README.md) y [endpoints](docs/backend/ENDPOINTS.md). Los informes de fases conservan su fecha y alcance histórico. Esta entrega documental no modifica el aplicativo ni los datos.
