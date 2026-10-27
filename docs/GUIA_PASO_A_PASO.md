# Cómo entender el frontend completo

## 1. Entrada

`dist/index.html` carga la hoja de estilos y `app.js`. El elemento `#app` recibe la interfaz actual. Los módulos JavaScript se sirven por HTTP; por eso se usa Live Server o `python -m http.server`.

## 2. Rutas y menú compartido

Una ruta como `#/clientes/nuevo` tiene un módulo (`clientes`) y una acción (`nuevo`). `app.js` verifica la sesión demo, dibuja el menú y envía la petición a la vista correspondiente. Los parámetros después de `?` permiten continuar con una recepción o cotización ya guardada.

`components/navigation.js` define las 13 opciones una sola vez. El Login no tiene menú. En celular el menú se despliega desde la izquierda.

## 3. Vistas

| Archivo en `views/` | Pantallas |
|---|---|
| `clients-vehicles.js` | Clientes, vehículos, transferencia e historial |
| `operations.js` | Recepción, diagnóstico, cotización, órdenes, auxilio y tareas |
| `payments.js` | Pago, entrega y ventana de contacto |
| `administration.js` | Mecánicos, proveedores, reportes y usuarios |

Cada vista presenta controles, recoge datos y llama a un servicio. Los elementos reutilizados —formularios, tablas, ventanas, mensajes y búsqueda— están en `components/ui.js`.

## 4. Servicios y datos

`workshop-service.js` aplica las reglas de la simulación. Por ejemplo, `createOrder()` exige cotización aprobada. `updateOrder()` trabaja sobre una OT existente y agrega un avance.

`store-service.js` lee una copia de los datos locales, aplica una operación y la guarda. Si una validación falla o el navegador no permite escribir, el cambio no se presenta como guardado. `data/seed.js` contiene los ejemplos iniciales. No se escribe en PostgreSQL todavía.

## 5. Relaciones

Una recepción guarda los identificadores de cliente y vehículo. El diagnóstico apunta a la recepción; la cotización apunta al diagnóstico y la orden a la cotización. Un pago apunta a una orden. Al consultar una OT, el servicio reconstruye su contexto recorriendo esas relaciones.

Cambiar el propietario actual del vehículo no modifica los clientes guardados en las recepciones anteriores. Por eso se conserva la historia técnica y comercial.

## 6. Números exactos

Los precios se almacenan en céntimos enteros: `15500` significa S/ 155.00. El kilometraje se conserva como cadena decimal normalizada, por ejemplo `2023.50`, para adaptarlo después a `NUMERIC(10,2)`. Las cantidades de detalle se validan con dos decimales. Los subtotales se redondean a céntimos sin depender de sumas aproximadas de coma flotante.

## 7. Dashboard y reportes

`dashboard-service.js` consulta los mismos registros locales que los formularios. No incrementa contadores independientes: suma pagos para ingresos y cuenta órdenes por estado. Si finalizas una OT o registras un pago, el siguiente acceso al Dashboard refleja ese cambio.

La línea de ingresos usa SVG adaptable; el anillo y las barras se generan desde sus cantidades. No son capturas del diseño.

## 8. Acceso de demostración

`auth-service.js` permite entrar con usuarios ficticios o con los creados desde Usuarios y roles. Las claves nuevas se conservan como resúmenes de demostración, no como texto plano. Este mecanismo está en el navegador y no es una medida de seguridad de producción. Go implementará autenticación, hashes robustos, sesiones y permisos definitivos.

## 9. Diseño e imágenes

`styles.css` mantiene los colores y componentes compartidos. El logo y banner usan un encuadre CSS de la referencia visual seleccionada; pueden sustituirse por los originales sin cambiar la estructura. El resto de la interfaz son formularios, tablas y controles web reales.

## 10. Probar y continuar

Primero sigue `RECORRIDO_DE_REVISION.md`. Luego abre una vista, localiza la llamada al servicio y sigue cómo se valida y guarda. Cuando conectemos Go, los servicios se convertirán en adaptadores HTTP y mantendremos los formularios y el estilo.
