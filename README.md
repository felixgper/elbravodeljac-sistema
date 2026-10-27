# El Bravo del JAC · Frontend y backend Go · Fase E

[Atenciones directas/técnicas, gráfico de ingresos y ficha original de dos páginas](docs/backend/ATENCIONES_REPORTES_FICHA.md): implementación, migración 003, permisos y pruebas.

[Revisión funcional implementada y probada — 27/10/2026](docs/backend/REVISION_12_OBSERVACIONES.md): búsqueda remota, cotizaciones manuales/descuentos, trabajadores, auxilio y períodos. Incluye migración segura para desarrollo/pruebas y un campo pendiente de identificar.

Sistema web universitario de gestión del taller automotriz, realizado en **HTML, CSS y JavaScript**. Conserva la referencia funcional **Balsamiq V14.1 reparado** (23 pantallas y ventanas, notas internas V14.1.1) y adopta el estilo **Borgoña contemporáneo**: navegación color vino, superficies claras y títulos sans serif.

Las **13 opciones del menú consumen las APIs Go** conservando el diseño aprobado.
Go sirve el modo real en `/` y la demostración separada en `/demo/`. La sesión usa
cookies HttpOnly y CSRF; PostgreSQL mantiene los datos transaccionales y MongoDB
los resultados de escáner/evidencias. No se mezclan registros demo con persistidos.

**Modo real en Windows con Docker Desktop:** desde la raíz ejecutar
`./backend/scripts/start-dev.ps1` y abrir **http://localhost:8080/**.
Usuario inicial: `admin_pruebas`; contraseña aleatoria en el archivo local
`backend/.cache/acceso-desarrollo.txt`, excluido de Git. El script mantiene un
entorno separado y crea la cuenta con bcrypt solo si la nueva base no tiene usuarios.

[Ejecución Docker y Chrome realmente verificada](docs/backend/VERIFICACION_DOCKER.md):
recorrido completo, permisos, MongoDB y persistencia tras reiniciar los contenedores.
Contratos y límites de las pantallas en [el informe de Fase E](docs/backend/FASE_E.md).

La interfaz aprobada todavía no incluye controles para cargar archivos técnicos:
el servicio JavaScript de MongoDB ya está implementado y probado; esos controles
quedan pendientes de acordar. Especialidad de proveedores no se agrega al modelo.

Los pasos de Python que siguen abren **solo la demostración**.

## Abrir el proyecto

1. Descomprime el ZIP completo y abre su carpeta en Visual Studio Code.
2. Abre una terminal en esa carpeta.
3. Ejecuta, si tienes Python instalado:

   ```bash
   python -m http.server 8000 --directory dist
   ```

   En macOS/Linux puede llamarse `python3`.

4. Abre `http://localhost:8000`.
5. Inicia sesión con uno de los accesos de demostración:

| Usuario | Contraseña de ejemplo | Uso |
|---|---|---|
| `admin` | `Bravo2026!` | Recorrer todos los módulos |
| `kevin` | `Auxiliar2026!` | Entrar directamente a Mis tareas |

También puedes servir `dist/index.html` con Live Server de VS Code. No abras el HTML mediante doble clic: los módulos JavaScript requieren HTTP. No necesitas instalar paquetes npm ni compilar. Las fuentes remotas son opcionales y tienen alternativas locales.

**En el enlace alojado:** el acceso privado de ChatGPT aparece primero; después verás el Login de demostración del taller.

## Qué puedes hacer

| Módulo | Operaciones implementadas |
|---|---|
| Inicio | Cuatro indicadores y cinco gráficos calculados desde los registros locales |
| Clientes | Buscar, registrar y editar; documento por tipo y número; correo opcional |
| Vehículos | Buscar, registrar, actualizar kilometraje y datos; transferencia verificada; historial técnico |
| Recepciones | Guardar cliente, vehículo, kilometraje y síntomas; continuar al diagnóstico |
| Diagnósticos | Seleccionar recepción pendiente, registrar evaluación y continuar a cotización |
| Cotizaciones | Crear o abrir cotizaciones; agregar/quitar conceptos, calcular subtotales y total; aprobar o rechazar |
| Órdenes de trabajo | Buscar; crear desde cotización aprobada; actualizar personal, estado, avances y pruebas; eliminar con confirmación |
| Auxilio mecánico | Registrar atención externa con unidad y técnico; consultar solicitudes; abrir Mis tareas |
| Pagos y entrega | Registrar pagos; consultar saldo; identificar receptor; registrar contactos en una ventana; confirmar entrega |
| Mecánicos | Consultar personal técnico y sus órdenes activas; ver asignaciones |
| Proveedores | Buscar y registrar/editar proveedores en la misma ventana |
| Reportes | Filtrar Semana, Mes o Año; consultar servicios e ingresos |
| Usuarios y roles | Crear y editar usuarios; activar/desactivar; asignar roles; iniciar sesión con usuarios de ejemplo creados |

La vista **Mis tareas** muestra las órdenes activas asignadas a Kevin o al auxiliar que inició sesión. Permite registrar avances y cambiar el estado operativo; la finalización con pruebas se realiza desde Actualizar orden.

## Recorrido recomendado

Abre `docs/RECORRIDO_DE_REVISION.md` para seguir una prueba guiada completa. Incluye un ejemplo con importes y la ruta para revisar cada pantalla. La cobertura de las 23 pantallas está en `docs/COBERTURA_FRONTEND.md`.

## Datos de demostración

- Los registros se conservan en `localStorage` de este navegador y este origen. Sobreviven a recargas y al cierre de sesión.
- Cada navegador, dispositivo o dirección (`localhost`, enlace alojado, otro puerto) mantiene su propia copia.
- La sesión de acceso se conserva en `sessionStorage`. «Recordarme» conserva únicamente el usuario.
- El perfil superior permite **Descargar datos demo** en JSON y **Reiniciar datos demo** con confirmación. El JSON no exporta resúmenes de contraseñas; no es una función de importación o restauración.
- Los ejemplos iniciales incluyen clientes, vehículos, recepciones, diagnósticos, cotizaciones, 35 órdenes, pagos, personal, proveedores y auxilios. Los cambios del usuario actualizan Dashboard, reportes, saldos e historiales.
- Utiliza datos ficticios. El acceso local y los resúmenes de claves solo simulan el flujo; no son autenticación segura ni controles de permisos del servidor.

## Reglas principales comprobadas

- No se duplican clientes con el mismo tipo y número de documento ni vehículos con la misma placa.
- El kilometraje admite dos decimales exactos y no disminuye; no hay campo VIN.
- Cambiar de propietario conserva la identidad y el historial del vehículo.
- Recepción y diagnóstico se guardan por separado; cotizar requiere diagnóstico guardado.
- **Crear orden y actualizar orden son operaciones diferentes**, con rutas y formularios separados.
- Una cotización debe estar aprobada y no tener una orden existente para crear OT.
- Los avances se agregan al historial. Finalizar exige una prueba satisfactoria, posterior a las reparaciones registradas.
- El pago no puede exceder el saldo. La entrega requiere orden finalizada, pago completo e identidad/autorización verificadas.
- Registrar un contacto no lo autoriza automáticamente para retirar un vehículo.
- Una orden con pagos o entrega se conserva por trazabilidad, aunque el resto pueda eliminarse con confirmación.

## Estructura

| Ruta | Contenido |
|---|---|
| `dist/index.html` | Entrada del frontend |
| `dist/assets/css/styles.css` | Estilos estructurales y componentes base |
| `dist/assets/css/atelier.css` | Tema Borgoña contemporáneo para acceso, navegación, módulos, formularios y móvil |
| `dist/assets/js/app.js` | Login, contenedor general, rutas y Dashboard |
| `dist/assets/js/views/` | Formularios, listados y eventos de cada grupo de módulos |
| `dist/assets/js/components/` | Menú, iconos, gráficos y elementos reutilizables |
| `dist/assets/js/services/store-service.js` | Lectura y escritura local de demostración |
| `dist/assets/js/services/workshop-service.js` | Reglas y operaciones simuladas del taller |
| `dist/assets/js/services/auth-service.js` | Acceso de demostración |
| `dist/assets/js/services/dashboard-service.js` | Consultas de Dashboard y reportes |
| `dist/assets/js/data/seed.js` | Datos ficticios iniciales |
| `dist/assets/js/data/demo.js` | Muestra histórica de la primera etapa, conservada como referencia |
| `dist/assets/images/` | Referencia visual aprobada para marca y banner |
| `docs/` | Balsamiq original, inventario, cobertura, guías y validación |
| `checks/` | Verificación del inventario y pruebas del flujo completo |

El ZIP conserva todos los archivos de la primera etapa y añade los módulos completos. No contiene dependencias instaladas, credenciales de servicios ni información del navegador de pruebas.

## Comprobar el código

Con Node.js instalado:

```bash
node checks/verify.mjs
node checks/workflow.mjs
node checks/edge-cases.mjs
node checks/real-api.mjs
# O todas las pruebas sin npm: node checks/run.mjs
```

Las pruebas usan almacenamiento en memoria y no alteran los datos de tu navegador.

## Integración real

La integración está implementada en `api-service.js`, `real-workshop.js` y `technical-service.js`. Go publica el frontend en el mismo origen y aplica permisos en cada operación. Consulta [FASE_E.md](docs/backend/FASE_E.md) para ejecutar y probar la aplicación completa.

El Balsamiq no incluye un control de carga de evidencias en la pantalla de diagnóstico. Se conservó su formulario; la interfaz de archivos técnicos se acordará al conectar MongoDB, sin añadir un módulo por iniciativa propia.

## Diseño Borgoña contemporáneo

La capa visual se define en `dist/assets/css/atelier.css`. Clientes muestra las placas vinculadas a cada persona y permite buscar también por placa. Sus contadores se calculan desde los datos del modo seleccionado. Se mantienen las operaciones de los 13 módulos y sus permisos. Títulos y texto usan DM Sans con fuentes del sistema como alternativa. El logo vectorial reutilizable está en `dist/assets/images/bravo-borgona.svg`.
