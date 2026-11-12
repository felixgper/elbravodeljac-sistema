# El Bravo del JAC

El Bravo del JAC es un sistema web de gestión automotriz. Administra clientes, vehículos, recepciones, diagnósticos, cotizaciones, órdenes de trabajo, auxilios, pagos, entregas, empleados y permisos. Go sirve el frontend en HTML, CSS y JavaScript. SQL Server 2025 almacena los datos transaccionales y MongoDB con GridFS conserva los resultados de escáner y las evidencias técnicas.

## Entregables oficiales

| Entregable | Archivo oficial |
|---|---|
| Prototipo Balsamiq V14.1 | [Editable BMPR](docs/Balsamiq_V14_1_REPARADO.bmpr) |
| Conceptual Chen V24 resumido | [Draw.io](docs/modelos/V24_Chen_Resumido_Exposicion/El_Bravo_del_JAC_V24_Chen_Resumido.drawio) · [PDF](docs/modelos/V24_Chen_Resumido_Exposicion/El_Bravo_del_JAC_V24_Chen_Resumido.pdf) · [PNG](docs/modelos/V24_Chen_Resumido_Exposicion/El_Bravo_del_JAC_V24_Chen_Resumido.png) |
| Lógico Crow’s Foot, sin tipos | [Draw.io](docs/modelos/Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.drawio) · [PDF](docs/modelos/Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.pdf) · [PNG](docs/modelos/Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.png) |
| Referencia física detallada V24, con tipos | [Draw.io](docs/modelos/V24_Logico_CrowsFoot/El_Bravo_del_JAC_V24_Logico_Completo.drawio) · [PDF](docs/modelos/V24_Logico_CrowsFoot/El_Bravo_del_JAC_V24_Logico_Completo.pdf) · [PNG](docs/modelos/V24_Logico_CrowsFoot/El_Bravo_del_JAC_V24_Logico_Completo.png) |
| SQL Server 2025: esquema oficial | [esquema_entrega.sql](database/sqlserver/esquema_entrega.sql) |

El esquema de SQL Server contiene 28 tablas, 208 columnas y 50 claves foráneas. Los diagramas conservan la referencia de diseño original del proyecto.

## Instalación y ejecución en Windows

Usa PowerShell 7, Git y Docker Desktop con motor Linux activo. Para una instalación desde cero, sigue la [guía de SQL Server y MongoDB](database/sqlserver/README.md#instalación-local-con-docker-compose).

Para iniciar o reiniciar una instalación ya provisionada, ejecuta desde la raíz:

```powershell
./backend/scripts/start-dev.ps1
./backend/scripts/start-dev.ps1 -Restart
```

Abre **http://localhost:8080/** con las credenciales de tu instalación. El script verifica los volúmenes y reanuda los contenedores existentes de SQL Server, MongoDB y Go. Si falta alguno, consulta la guía antes de continuar.

Go sirve UI y API bajo el mismo origen con JWT HttpOnly, CSRF y permisos en servidor. `/demo/` es demostración ficticia separada; Live Server solo sirve esa demostración. No abrir `index.html` con doble clic.

## Pruebas y mantenimiento

Node.js y Go 1.26 o superior para las herramientas locales:

```powershell
node checks/run.mjs
./backend/scripts/test-isolated.ps1
```

El primer comando ejecuta ocho suites de demo/unidad/prototipo. El segundo crea PostgreSQL/MongoDB temporales con UUID y ejecuta Go y `go vet`. Para integración Chrome, requisitos y `-Browser`: [guía de pruebas](checks/README.md). **PASS, FAIL y SKIP se reportan por separado; SKIP no significa aprobado.** Nunca dirigir pruebas de escritura a localhost:8080 ni a bases existentes.

[Backend y configuración](backend/README.md) · [API REST](docs/backend/ENDPOINTS.md).
