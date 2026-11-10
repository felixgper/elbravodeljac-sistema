# El Bravo del JAC

El Bravo del JAC es un sistema web de gestión automotriz. Administra clientes, vehículos, recepciones, diagnósticos, cotizaciones, órdenes de trabajo, auxilios, pagos, entregas, empleados y permisos. Go sirve el frontend en HTML, CSS y JavaScript. PostgreSQL almacena los datos transaccionales y MongoDB con GridFS conserva los resultados de escáner y las evidencias técnicas.

## Entregables oficiales

| Entregable | Archivo oficial |
|---|---|
| Prototipo Balsamiq V14.1 | [Editable BMPR](docs/Balsamiq_V14_1_REPARADO.bmpr) |
| Conceptual Chen V24 resumido | [Draw.io](docs/modelos/V24_Chen_Resumido_Exposicion/El_Bravo_del_JAC_V24_Chen_Resumido.drawio) · [PDF](docs/modelos/V24_Chen_Resumido_Exposicion/El_Bravo_del_JAC_V24_Chen_Resumido.pdf) · [PNG](docs/modelos/V24_Chen_Resumido_Exposicion/El_Bravo_del_JAC_V24_Chen_Resumido.png) |
| Lógico Crow’s Foot, sin tipos | [Draw.io](docs/modelos/Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.drawio) · [PDF](docs/modelos/Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.pdf) · [PNG](docs/modelos/Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.png) |
| Referencia física detallada V24, con tipos | [Draw.io](docs/modelos/V24_Logico_CrowsFoot/El_Bravo_del_JAC_V24_Logico_Completo.drawio) · [PDF](docs/modelos/V24_Logico_CrowsFoot/El_Bravo_del_JAC_V24_Logico_Completo.pdf) · [PNG](docs/modelos/V24_Logico_CrowsFoot/El_Bravo_del_JAC_V24_Logico_Completo.png) |
| PostgreSQL: SQL oficial de entrega | [esquema_entrega.sql](database/postgresql/esquema_entrega.sql) |

La comparación del modelo lógico con el catálogo de PostgreSQL del 5 de noviembre de 2026 verificó 28 tablas, 208 columnas y 50 claves foráneas. El modelo Chen resume el negocio. La referencia física incluye los tipos y el SQL oficial contiene la estructura sin datos.

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

[Backend y configuración](backend/README.md) · [API REST](docs/backend/ENDPOINTS.md).
