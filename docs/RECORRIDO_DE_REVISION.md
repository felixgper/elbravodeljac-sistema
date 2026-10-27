# Recorrido completo para revisar el frontend

Inicia sesión como `admin` con `Bravo2026!`. Todos los ejemplos siguientes son ficticios. Puedes comenzar con los datos iniciales o registrar los tuyos. Si repites exactamente un documento o una placa que ya guardaste, el sistema te avisará del duplicado.

## 1. Cliente

En **Clientes → + Nuevo cliente** registra:

| Campo | Ejemplo |
|---|---|
| Nombre completo | Cliente de demostración |
| Tipo de documento | DNI |
| Número | 45678912 |
| Teléfono | 987654300 |
| Correo | Déjalo vacío: es opcional |
| Dirección | Dirección de prueba, Lima |

Guarda, busca el cliente por nombre y prueba Editar. Si eliges RUC, el nombre del campo pasa a Razón social y el documento requiere 11 dígitos. La pareja tipo + número es única.

## 2. Vehículo e historial

En **Vehículos → + Nuevo vehículo**, selecciona ese cliente, placa `QAQ-123`, marca JAC, modelo T8, año 2024 y kilometraje `2023.50`. Guarda.

Prueba **Actualizar vehículo**: puedes corregir datos y aumentar el kilometraje. Activa Cambio de propietario únicamente para una transferencia; requiere nuevo cliente y verificación. El historial previo conserva al cliente que participó en cada atención. Puedes consultar el historial desde Ver historial.

## 3. Recepción

En **Recepciones**, elige al cliente y su vehículo. Al elegir la placa se completa su kilometraje actual. Escribe: «Cambio de aceite y revisión de frenos». Guarda.

La pantalla muestra el código RX generado y habilita **Ir a diagnóstico**. Los botones + Cliente y + Nuevo vehículo permiten completar esos registros y volver a la recepción; el borrador se conserva en la pestaña.

## 4. Diagnóstico

Desde la recepción guardada abre Diagnóstico. El cliente, vehículo y síntomas se muestran solo para lectura. Escribe: «Se recomienda cambio de aceite y filtro; realizar prueba de ruta al concluir».

Guarda. Se genera un código DX y se habilita **Continuar a cotización**.

## 5. Cotización

Pulsa **+ Agregar servicio o repuesto** para registrar estas filas:

| Concepto | Tipo | Origen | Cantidad | Precio unitario |
|---|---|---|---:|---:|
| Cambio de aceite | Servicio | — | 1 | 120.00 |
| Filtro de aceite | Repuesto | Taller | 1 | 35.00 |

El total será **S/ 155.00**. Cambia la autorización a **Aprobada** y guarda. También puedes guardar Pendiente y abrir después la misma cotización desde el listado. Un repuesto aportado por el cliente se conserva como tal, con precio cero.

## 6. Crear una orden

Pulsa **Crear orden**, elige mecánico Pedro y auxiliar Kevin, y crea la OT. El código y fecha son automáticos. El estado inicial será Pendiente.

Esta operación tiene un formulario y ruta propios. No se permite crear desde una cotización rechazada/pendiente ni crear dos órdenes para la misma cotización.

## 7. Actualizar una orden

En Actualizar orden registra el avance y cambia el estado a En reparación. Guarda. El avance aparecerá en el historial.

Después registra una prueba fallida en En pruebas. Intenta finalizar: el sistema exigirá un resultado satisfactorio. Registra la reparación y, finalmente, una prueba de ruta satisfactoria con un avance nuevo. Pulsa **Finalizar orden**.

Cada actualización conserva el avance anterior y registra quién la hizo y cuándo. La lista de órdenes permite buscar por código, cliente o placa. La papelera abre una confirmación; no elimina automáticamente. Las órdenes con pagos o entrega se conservan por trazabilidad.

## 8. Pago

En **Pagos y entrega**, selecciona la OT finalizada. Verás el total, lo pagado y el saldo. Registra **155.00** con medio Transferencia.

También puedes probar pagos parciales; su suma no puede superar el total. El registro es solo una simulación local: no realiza transferencias ni cobra tarjetas.

## 9. Retiro por tercero

Selecciona **Tercero → + Nuevo contacto**. Registra un nombre ficticio, DNI de 8 dígitos, teléfono con prefijo internacional (ejemplo +51987654310) y relación Conductor.

El nuevo contacto queda seleccionado, pero todavía debes elegir el método de verificación y marcar **Verifiqué identidad y autorización específica antes de entregar**. Confirma la entrega. Se registra receptor, usuario, fecha y método.

Si no se ha completado el pago, si la OT no está finalizada o si falta verificar la autorización, la entrega se rechaza con un mensaje.

## 10. Revisar resultados

Vuelve al Dashboard y Reportes. Los importes, estados, servicios y asignaciones se calculan desde tus registros. Recarga la página: tus datos locales se mantienen.

## 11. Otros módulos

- **Auxilio mecánico:** selecciona cliente y vehículo; registra ubicación, incidente, unidad, descripción y técnico. El listado muestra la solicitud guardada.
- **Mecánicos:** busca por nombre/rol y abre Ver asignaciones.
- **Proveedores:** Nuevo y Editar abren la misma ventana con nombre, RUC, teléfono y repuestos.
- **Usuarios y roles:** crea un trabajador de prueba con usuario, correo y contraseña ficticios. Puedes activarlo/desactivarlo y cambiar su rol. Los usuarios activos creados pueden iniciar sesión localmente.
- **Auxiliar:** cierra sesión y entra como `kevin` / `Auxiliar2026!`. Verás Mis tareas con sus órdenes activas. Abre Actualizar estado y registra un avance. Un auxiliar sin órdenes asignadas muestra una lista vacía.

## 12. Descargar y reiniciar los datos

En el perfil superior puedes descargar los registros demo en JSON. Reiniciar solicita confirmación y sustituye los datos de este navegador por los iniciales. El JSON exportado no incluye resúmenes de contraseñas y no se importa desde esta versión.

Los registros de un dispositivo no aparecen automáticamente en otro: esa persistencia compartida llegará al conectar Go y las bases de datos.
