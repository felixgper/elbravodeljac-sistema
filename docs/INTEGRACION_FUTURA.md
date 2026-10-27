# Preparación para el backend y las bases de datos

## Responsabilidades

| Capa | Tecnología | Responsabilidad |
|---|---|---|
| Presentación | HTML, CSS, JavaScript | Formularios, validación de ayuda y visualización |
| API REST | Go | Autenticación, permisos, reglas de negocio y transacciones |
| Base relacional | PostgreSQL | Clientes, vehículos, recepciones, diagnósticos estructurados, cotizaciones, órdenes, pagos y usuarios |
| Documentos técnicos | MongoDB | Resultados detallados de escáner, documentos flexibles y metadatos/evidencias de diagnóstico |

Las contraseñas y las credenciales de bases de datos nunca deben llegar al frontend. La ubicación de los binarios de fotos y archivos se definirá con el backend; no se impone un almacenamiento nuevo en esta entrega.

## Sustituir proveedores, conservar pantallas

`services/auth-service.js`, `services/workshop-service.js` y `services/dashboard-service.js` delimitan las llamadas que luego se sustituirán por peticiones HTTP. `store-service.js` es exclusivamente el proveedor local de demostración. El contrato definitivo de endpoints y los nombres exactos de los campos se alinearán con el modelo de datos de la otra conversación antes de conectar Go. Esta entrega conserva los campos y relaciones funcionales del prototipo. No sustituye el esquema PostgreSQL acordado, no incluye migraciones y no crea bases alternativas.

El archivo de controles Balsamiq se conserva como fuente de interfaz. Su inventario no reemplaza el diccionario de datos definitivo.

## Reglas que deben preservarse al construir formularios

- Cliente identificado por tipo y número de documento; correo opcional; nombre o razón social según corresponda.
- Kilometraje decimal compatible con `NUMERIC(10,2)`, sin VIN en el MVP. Acordar transporte exacto de decimales en la API.
- La placa y el historial identifican el mismo vehículo aunque se registre un cambio de propietario verificado.
- Guardar recepción antes del diagnóstico; guardar diagnóstico antes de cotizar.
- Crear OT únicamente desde cotización aprobada. Crear y actualizar serán rutas, formularios y operaciones diferentes.
- Actualizaciones de OT deben agregar avances con usuario/fecha, sin sobrescribir la historia previa. Finalizar solo tras las pruebas exigidas satisfactorias.
- El pago y la entrega son acciones distintas. La autorización para retirar requiere verificación específica.
- Agregar contacto es una ventana dentro de pago/entrega; no se crea un módulo independiente.
- Usuarios y roles administra al personal; Mecánicos consulta sus asignaciones.

El backend verificará todas estas reglas. La validación del navegador mejora la interacción, pero no sustituye el control del servidor.

## Acceso y permisos

El rol se conserva en el usuario y la vista del auxiliar filtra sus tareas. Los controles de acceso y permisos definitivos deben verificarse en Go en cada operación. La demo permite recorrer todos los módulos y no pretende proteger datos reales. Los resúmenes locales de claves ficticias no sustituyen una estrategia de autenticación del servidor.
