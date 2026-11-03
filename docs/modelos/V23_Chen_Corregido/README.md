# Entrega Chen V23

Revisión documental del 2026-11-02. No reemplaza V22 ni aplica sus propuestas a la aplicación.

- **El_Bravo_del_JAC_V23_Chen_Corregido.drawio**: archivo nativo editable, una página. Abrir con Archivo → Abrir desde → Dispositivo en draw.io, o con draw.io Desktop. Entidades, rombos, etiquetas y conectores son objetos individuales.
- **El_Bravo_del_JAC_V23_Chen_Corregido.png**: exportación de 9190×5572 del motor oficial; ampliar para leer.
- **El_Bravo_del_JAC_V23_Chen_Corregido.svg**: exportación vectorial complementaria.
- **vista_general.png**: previsualización reducida.
- **INFORME_AUDITORIA_V23.md** y **AUDITORIA_RELACIONES.csv**: auditoría de las 42 relaciones, anteriores/corregidas y evidencias.
- **DECISIONES_PENDIENTES.md**: P1–P3 y límites del negocio que necesitan revisión.
- **DIFERENCIAS_IMPLEMENTACION.md**: contraste con PostgreSQL, MongoDB, Go y formularios.
- **PROCEDENCIA_V22.json**: fuente Git y huella original de V22.
- **ESPECIFICACION_V23.json**, **RELACIONES_V23.json** y **generar_v23.py**: definición y generador documental. Requiere Python 3.12 o posterior, Pillow y Arial de Windows. La generación recrea una previsualización Pillow, que luego debe sustituirse por la exportación del motor draw.io usando el verificador local.
- **VALIDACION_GEOMETRIA.json**, **OPTIMIZACION_TRAZADO.json**, **VALIDACION_INDEPENDIENTE.json**: verificaciones y comparación de trazados.
- **VALIDACION_RENDER_LOCAL.json**: apertura, render y prueba de edición local aprobados.
- **VALIDACION_DRAWIO.json**: intento remoto bloqueado HTTP 403; se conserva como evidencia de la limitación, no como resultado exitoso.

## Reproducir la validación

Desde la raíz del repositorio:

```powershell
python docs/modelos/V23_Chen_Corregido/validar_v23.py
node docs/modelos/V23_Chen_Corregido/verificar_local_drawio.mjs
```

El validador independiente utiliza el registro local inicial de 170 hashes en `backend/.cache/chen-v23/proteccion-antes.json` y el blob original conservado en Git. Es una comprobación de esta sesión; en otra copia debe generarse una línea base nueva antes de trabajar, no reemplazarla para ocultar cambios.

El renderizador requiere Node, Playwright y Chrome; el script indica las rutas disponibles en esta computadora. Usa el bundle oficial `https://viewer.diagrams.net/js/viewer-static.min.js`, descargado en `backend/.cache/chen-v23/viewer-static.min.js`. Abre únicamente un servidor HTTP efímero en 127.0.0.1, carga el XML, exporta, comprueba una edición temporal y cierra navegador/servidor. No se incluye ese bundle de terceros en los entregables. La alternativa remota está en `verificar_render_drawio.mjs`; no es necesaria para utilizar el archivo editable.

Las pruebas de renderizado no se conectan a PostgreSQL ni MongoDB. Los hashes de los entregables están en MANIFIESTO_SHA256.json. Las referencias P1/P2/P3 en ámbar deben revisarse antes de considerar migraciones o endurecer restricciones.
