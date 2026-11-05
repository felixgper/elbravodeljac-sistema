# Entrega del modelo lógico V24

**28 tablas / 208 columnas / 50 FK.** Diagrama único en seis bloques, sin omitir auditoría ni puentes. Fuente: SQL base y migraciones 001–009 del repositorio; no se consultaron bases vivas.

## Archivos principales

1. `El_Bravo_del_JAC_V24_Logico_Completo.drawio`: XML nativo editable; tablas, filas y conectores separados. Abrir en draw.io Desktop o Archivo → Abrir desde → Dispositivo en diagrams.net.
2. `El_Bravo_del_JAC_V24_Logico_Completo.png`: 7993×5400, ampliar para leer atributos.
3. `El_Bravo_del_JAC_V24_Logico_Completo.pdf`: vectorial, una página A0 horizontal (1189×841 mm). Mantener formato amplio; A4 reducirá legibilidad.
4. `INVENTARIO_TABLAS_Y_FK.md`: diccionario completo, nulabilidad, PK, índices únicos y 50 referencias.
5. `REVISION_CARDINALIDADES.md`: lectura de los dos extremos, reglas SQL/Go, puentes y autorrelación.
6. `COMPARACION_CONCEPTUAL_LOGICO_FISICO.md`: correspondencia entre V23/V24 conceptual y esquema versionado.
7. `INCONSISTENCIAS_DETECTADAS.md`: propuestas no aplicadas y diferencias documentadas.

Complementos: SVG vectorial, vista previa, seis PNG en `vistas/`, inventario CSV/JSON y comprobaciones finales. `F01`–`F50` significa FK; ver tooltip del conector y CSV para columna referenciada exacta. Los puertos distribuidos en el padre no representan referencias a filas distintas de la documentada.

## Resultados

- 28 PK y todas las columnas/FK verificadas; cero referencias faltantes, duplicadas o con símbolos incompatibles.
- Empleados–usuarios opcional y única en ambos sentidos; TIPOS_DOCUMENTO presente; autorrelación de cotizaciones y tres puentes N:M conservados.
- Cero tablas superpuestas, conectores atravesando tablas, tramos colineales compartidos o textos desbordados en las filas. Hay **121 cruces medidos**; no se afirma un óptimo global ni ausencia de cruces. Se separaron los tramos terminales para evitar fusionar símbolos opcionales.
- Apertura, exportación y edición temporal con motor oficial draw.io GraphViewer/mxGraph en Chrome local. PDF reabierto con pypdf y renderizado con Poppler para revisión visual. No se afirma una sesión manual del editor web.
- 254 archivos versionados previos idénticos a su línea base; cambios solo en esta carpeta. No se accedió a PostgreSQL/MongoDB ni se ejecutaron migraciones, Docker o pruebas de aplicación. No hubo commit/push/merge.

## Reproducibilidad

Desde la raíz del repositorio:

```powershell
python docs/modelos/V24_Logico_CrowsFoot/inventariar_esquema.py
python docs/modelos/V24_Logico_CrowsFoot/generar_logico.py
node docs/modelos/V24_Logico_CrowsFoot/verificar_exportar.mjs
python docs/modelos/V24_Logico_CrowsFoot/validar_entrega.py
```

Python requiere Pillow y pypdf; el validador usa Arial de Windows. Node requiere Playwright y Chrome; el exportador contiene las rutas disponibles en esta computadora. Usa el bundle oficial `https://viewer.diagrams.net/js/viewer-static.min.js`, leído de la caché local preexistente `backend/.cache/chen-v23/viewer-static.min.js`. No se incorpora esa dependencia de terceros al repositorio; otra máquina debe disponer del bundle y ajustar su ruta. El servidor temporal se limita a 127.0.0.1 y cierra al terminar.

PROTECCION_ANTES.json es la línea base de esta revisión. No debe regenerarse después para ocultar modificaciones. `ESQUEMA_VERIFICADO.json` registra huellas de fuentes y evolución; `MODELOS_PREVIOS_REVISADOS.json` acredita los archivos anteriores efectivamente disponibles. Las comprobaciones evalúan la estructura versionada esperada; no certifican que todas las migraciones estén hoy aplicadas en producción.

P1/P2/P3 y el renombrado de `id_recepcionista` se mantienen como pendientes, sin cambiar SQL ni Go. Los conceptos de MongoDB no reciben FK PostgreSQL. Se evita equiparar cargo laboral con permisos de acceso.
