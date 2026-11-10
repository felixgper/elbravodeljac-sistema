# El Bravo del JAC

Sistema web de gestión automotriz: clientes, vehículos, recepciones, diagnósticos, cotizaciones, órdenes, auxilios, pagos/entrega, empleados y control de acceso. Frontend HTML/CSS/JavaScript servido por Go, PostgreSQL transaccional y MongoDB/GridFS para escáner y evidencias.

## Entregables oficiales

Una versión vigente por nivel. Las versiones anteriores se recuperan desde el historial de Git; no son alternativas para instalar.

| Entregable | Archivo oficial |
|---|---|
| Prototipo Balsamiq V14.1 | [Editable BMPR](docs/Balsamiq_V14_1_REPARADO.bmpr) · [Inventario de pantallas](docs/balsamiq-v14.1.json) |
| Conceptual Chen V24 resumido | [Draw.io](docs/modelos/V24_Chen_Resumido_Exposicion/El_Bravo_del_JAC_V24_Chen_Resumido.drawio) · [PDF](docs/modelos/V24_Chen_Resumido_Exposicion/El_Bravo_del_JAC_V24_Chen_Resumido.pdf) · [PNG](docs/modelos/V24_Chen_Resumido_Exposicion/El_Bravo_del_JAC_V24_Chen_Resumido.png) |
| Lógico Crow’s Foot, sin tipos | [Draw.io](docs/modelos/Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.drawio) · [PDF](docs/modelos/Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.pdf) · [PNG](docs/modelos/Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.png) |
| Referencia física detallada V24, con tipos | [Draw.io](docs/modelos/V24_Logico_CrowsFoot/El_Bravo_del_JAC_V24_Logico_Completo.drawio) · [PDF](docs/modelos/V24_Logico_CrowsFoot/El_Bravo_del_JAC_V24_Logico_Completo.pdf) · [PNG](docs/modelos/V24_Logico_CrowsFoot/El_Bravo_del_JAC_V24_Logico_Completo.png) |
| PostgreSQL: SQL oficial de entrega | [esquema_entrega.sql](database/postgresql/esquema_entrega.sql) · [Comparación y límites de verificación](docs/modelos/COMPARACION_BASE_ACTUAL.md) |

**PostgreSQL verificado el 5 de noviembre de 2026: 28 tablas, 208 columnas y 50 FK.** El SQL de entrega es estructura sin filas, contraseñas, propietarios ni ACL; no se debe cargar sobre una base existente. El nombre de carpeta `V24_Logico_CrowsFoot` se conserva para no romper rutas: su diagrama con tipos es la referencia física, no otro lógico oficial. El conceptual resume el negocio (13 entidades y 15 relaciones), no todas las tablas.

## Diseños actuales

- **Login:** oscuro de dos columnas, logo Diseño 19, campos «Correo o usuario» y «Contraseña».
- **Dashboard:** carbón y borgoña, datos reales de la API en modo real.
- **Cotización PDF:** borgoña, estados/revisiones e importes de la cotización guardada.
- **Ficha técnica:** borgoña, tres recuadros de inspección, casillas B/A/U y continuación de páginas. Los PDF ya archivados se descargan sin regenerar sus bytes.

[Login y Dashboard](docs/frontend/LOGIN_DASHBOARD_CARBON.md) · [PDF oficiales y capturas QA](docs/backend/CONSOLIDACION_PDF_DISENO19.md). El aplicativo está en `dist/`; no hay que crear ni compilar otro frontend.

## Instalación y ejecución en Windows

Requisitos: PowerShell 7, Git y Docker Desktop con motor Linux activo. Desde una copia del repositorio, para preparar **un entorno de desarrollo nuevo y separado**:

```powershell
./backend/scripts/start-dev.ps1
```

Abrir **http://localhost:8080/** en Chrome. El script genera configuración privada; solo provisiona la cuenta inicial cuando la base de desarrollo está vacía de usuarios. El acceso inicial queda en `backend/.cache/acceso-desarrollo.txt`, ignorado por Git. Una instalación existente conserva sus cuentas y datos; utiliza su acceso propio.

`start-dev.ps1` usa el **único esquema físico oficial**, `database/postgresql/esquema_entrega.sql`, al crear una base nueva. El migrador detecta su marca y no repite migraciones; únicamente prepara el catálogo inicial de documentos. Las migraciones **001–009** se mantienen para actualizar bases históricas; no son otros modelos físicos. **Nunca ejecutar el SQL final sobre una base existente.** No restaurar SQL ni reinicializar volúmenes para reanudar un entorno detenido. Descargar GitHub tampoco actualiza por sí solo un binario/contenedor que ya está ejecutándose; una actualización de la imagen habitual requiere revisión independiente y conserva los volúmenes.

Go sirve UI y API bajo el mismo origen con JWT HttpOnly, CSRF y permisos en servidor. `/demo/` es demostración ficticia separada; Live Server solo sirve esa demostración. No abrir `index.html` con doble clic.

## Pruebas y mantenimiento

Node.js y Go 1.26 o superior para las herramientas locales:

```powershell
node checks/run.mjs
./backend/scripts/test-isolated.ps1
```

El primer comando ejecuta ocho suites de demo/unidad/prototipo. El segundo crea PostgreSQL/MongoDB temporales con UUID y ejecuta Go y `go vet`. Para integración Chrome, requisitos y `-Browser`: [guía de pruebas](checks/README.md). **PASS, FAIL y SKIP se reportan por separado; SKIP no significa aprobado.** Nunca dirigir pruebas de escritura a localhost:8080 ni a bases existentes.

[Backend y configuración](backend/README.md) · [API REST](docs/backend/ENDPOINTS.md) · [Informe de limpieza y conservación](docs/LIMPIEZA_REPOSITORIO.md).
