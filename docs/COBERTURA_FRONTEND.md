# Cobertura del Balsamiq V14.1 · Frontend v1.0

Se conservan las 13 opciones de navegación y se implementan las 23 pantallas/ventanas del archivo de referencia. Las rutas indicadas son internas y se escriben después de `#/`.

| Pantalla original | Ruta o acceso | Comportamiento |
|---|---|---|
| 01 · Inicio de sesión | `login` | Usuario/correo, contraseña, Recordarme e Iniciar sesión; sin menú |
| 02 · Dashboard | `inicio` | Cuatro indicadores y cinco gráficos con registros locales |
| 03 · Clientes | `clientes` | Búsqueda, listado y Editar |
| 04 · Nuevo cliente | `clientes/nuevo` y `clientes/editar/:id` | Un formulario para registro y edición |
| 05 · Vehículos | `vehiculos` | Búsqueda, Nuevo, Actualizar e Historial |
| 06 · Nuevo vehículo | `vehiculos/nuevo` | Propietario, placa, marca, modelo, año, kilometraje, observaciones |
| 06B · Actualizar vehículo | `vehiculos/editar/:id` | Actualización y transferencia verificada, sin cambiar placa/identidad |
| 07 · Nueva recepción | `recepciones` | Cliente y vehículo, kilometraje, síntomas, guardado y continuación |
| 08 · Diagnóstico técnico | `diagnosticos?recepcion=:id` | Recepción pendiente, datos de lectura, evaluación y guardado |
| 09 · Cotización | `cotizaciones` | Detalle de servicios/repuestos, cantidades, origen, precios, total y autorización |
| 10 · Órdenes de trabajo | `ordenes` | Listado, búsqueda, creación, lápiz y papelera |
| 10B · Confirmar eliminación | Papelera de una OT | Ventana con Cancelar y Eliminar |
| 11 · Nueva orden de trabajo | `ordenes/nueva` | Cotización aprobada, datos derivados, personal e inicio Pendiente |
| 11B · Actualizar orden | `ordenes/editar/:id` | Personal, estado, avance nuevo, prueba, resultado e historial |
| 12 · Auxilio mecánico | `auxilio` | Cliente, vehículo, ubicación, incidente, unidad, descripción y técnico |
| 13 · Pagos y entrega | `pagos?orden=:id` | Pago y entrega independientes, receptor y autorización específica |
| 13A · Nuevo contacto | + Nuevo contacto en Pagos | Nombre, documento, teléfono internacional y relación |
| 13B · Mecánicos | `mecanicos` | Consulta del personal y ventana de asignaciones |
| 13C · Proveedores | `proveedores` | Listado y ventana de alta/edición compartida |
| 14 · Reportes | `reportes` | Periodo, Aplicar filtros, servicios e ingresos |
| 15 · Usuarios y roles | `usuarios` | Listado, búsqueda y accesos al formulario compartido |
| 15A · Agregar / editar usuario | `usuarios/nuevo` y `usuarios/editar/:id` | Datos de acceso, rol, estado y clave ficticia inicial solo en alta |
| 16 · Vista móvil del auxiliar | `auxilio/tareas` | Órdenes activas asignadas y registro de avances/estado |

## Adaptaciones necesarias, sin módulos adicionales

- Los listados de recepciones, diagnósticos y cotizaciones guardadas permiten volver a abrir lo registrado y continuar el flujo. Permanecen dentro del módulo original.
- Los botones para quitar conceptos permiten corregir una cotización antes de guardarla.
- La paginación y búsqueda de tablas mantienen manejable el conjunto de demostración.
- Descargar y reiniciar los datos son utilidades del perfil para revisar la demo; no constituyen módulos de negocio.
- Propietario y Administrativo se incluyen entre los roles porque aparecen en el listado original de usuarios; no se descartan al abrir la edición.
- El Dashboard mantiene los indicadores y tipos de gráfico, pero calcula sus valores desde los registros. Los meses avanzan con la fecha del navegador para que los ejemplos sigan siendo utilizables.
- Se impide borrar una orden con pagos/entrega o modificar su cotización una vez creada la OT, para conservar las relaciones y los trabajos autorizados.
- El auxiliar puede registrar avances y pasar a reparación/pruebas; finalizar requiere el formulario completo y prueba satisfactoria.

## Qué significa completo en esta entrega

Las pantallas y su recorrido funcionan con datos locales de demostración. La operación real multiusuario, los permisos seguros, la autenticación del servidor y la persistencia compartida dependen del backend Go y las bases acordadas. La pantalla de diagnóstico conserva los campos del Balsamiq: no se añadió una carga de archivos que no figura allí.
