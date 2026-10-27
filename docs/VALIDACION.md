# Validación del frontend v1.0

## Correspondencia con la fuente

Se conserva el Balsamiq original y su extracción íntegra de 23 recursos de pantallas/ventanas. El menú coincide con las 13 opciones de esa fuente. `COBERTURA_FRONTEND.md` enlaza cada pantalla con su ruta o ventana.

## Pruebas automáticas ejecutadas

| Archivo | Comprobaciones |
|---|---|
| `checks/verify.mjs` | Acceso demo, sesión, recordatorio, datos y correspondencia del inventario |
| `checks/workflow.mjs` | Registro completo desde cliente hasta entrega; duplicados, relaciones, cotización aprobada, avances, pruebas, pagos parciales, receptor, transferencia, proveedores, usuarios, reportes y fallos de almacenamiento |
| `checks/edge-cases.mjs` | Códigos de OT no reutilizados, prueba posterior a nueva reparación, último administrador activo, límites decimales, gráficos sin cantidades y compatibilidad de claves ficticias en vistas HTTP locales |

Se comprobó también la sintaxis de todos los archivos JavaScript. Las pruebas usan almacenamiento en memoria y datos ficticios.

## Recorrido manual en navegador

Se realizó con una sesión sintética local, sin credenciales reales:

1. Se abrió el Dashboard y el listado de clientes.
2. Se registró Cliente QA Taller y se guardó su vehículo QAQ-123 con 2023.50 km.
3. Se creó la recepción RX-063, su diagnóstico DX-062 y una cotización con un servicio de S/ 120 y un repuesto de S/ 35.
4. Se aprobó la cotización y se creó OT-060 con Pedro y Kevin.
5. Se intentó finalizar sin prueba: la interfaz rechazó la operación con el mensaje esperado.
6. Se registró una prueba de ruta satisfactoria y se finalizó la OT.
7. Se registró un pago local ficticio de S/ 155.00, quedando saldo cero.
8. Se creó Receptor QA Taller desde la ventana de contacto, se verificó la autorización y se confirmó la entrega a tercero.
9. Se abrieron Auxilio, Mecánicos, Proveedores, Reportes y Usuarios; no aparecieron errores de aplicación ni desbordamiento horizontal en escritorio.
10. Se revisó el alta de usuarios, la ventana de proveedores y Mis tareas de Kevin.
11. Se revisó un formulario en un contenedor móvil de 390 px: sus campos y botones se distribuyen en una columna y el documento no desborda horizontalmente.

Se corrigió un problema de compatibilidad detectado al guardar desde el entorno HTTP de pruebas: los identificadores usan una generación aleatoria disponible también en ese entorno. Se conservaron los controles originales del frontend.

## Límites de la validación

- La prueba móvil utiliza un contenedor estrecho de navegador; no un teléfono físico.
- La lógica de alta de usuario y verificación de claves se probó automáticamente; no se introdujeron contraseñas reales en el navegador.
- Las operaciones de eliminación y reinicio tienen confirmación visible; sus reglas se comprobaron mediante pruebas de datos.
- No se validaron APIs ni bases de datos porque esta entrega trabaja completamente con datos locales.
- No se certifican seguridad, concurrencia multiusuario, cumplimiento tributario o reglas de producción; corresponden al backend y a su validación posterior.

Los archivos temporales que habilitaron sesiones sintéticas durante la revisión se eliminan de la entrega y de la publicación.
