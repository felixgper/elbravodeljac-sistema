# Validación del modelo lógico Crow’s Foot sin tipos

Fecha: 2026-11-02. Rama: `backend/arquitectura-go-mongodb`.

## Referencia principal

Se utilizó el último diagrama detallado local disponible: `docs/modelos/V24_Logico_CrowsFoot/El_Bravo_del_JAC_V24_Logico_Completo.drawio`. Contiene las tablas, atributos, claves y tipos procedentes del SQL y migraciones 001–009. Aunque su nombre anterior dice “lógico”, sí mostraba los tipos; esta copia retira exclusivamente esa información visual. No se presupone la existencia de otro archivo detallado aprobado fuera del repositorio.

Se conserva el original y su carpeta, que tenían cambios locales sin seguimiento. No se descartó ni reemplazó ningún cambio previo.

## Resultado

| Elemento | Cantidad / estado |
| --- | --- |
| Tablas | 28, mismas tablas y nombres |
| Atributos | 208, mismos nombres reales |
| Claves primarias | 28, incluidas las compuestas |
| FK / conectores | 50, sin omisiones ni duplicados |
| Celdas visibles de tipos | 0; se retiraron las 208 celdas correspondientes |
| Distribución | Seis bloques, una sola página; cajas, filas y rutas de conectores originales |
| PDF | Una página vectorial A0 horizontal, completa y sin recortes |
| PNG | 7993 × 5400 píxeles |

## Comprobaciones realizadas

- Comparación de todos los nombres de atributos con el editable de referencia; ninguna columna añadida ni eliminada.
- PK simples/compuestas, FK y UQ individual conservadas. Las etiquetas FK F01–F50 identifican las mismas referencias del modelo detallado.
- Comparación exacta de los 50 conectores: origen, destino, columna referenciada, geometría y símbolos Crow’s Foot idénticos.
- Extremos revisados contra la nulabilidad y unicidad del inventario SQL existente, sin alterar las cardinalidades del original. Se mantienen puentes N:M y autorrelación de cotizaciones.
- Relación empleados–usuarios opcional y única en ambos sentidos; diagnóstico opcional según modalidad, sin convertir atención técnica pendiente en directa; auxilio–recepción conserva varios auxilios por recepción.
- Eliminación de tipos y longitudes de todas las etiquetas visibles. Los nombres ocupan ahora el espacio disponible, manteniendo las dimensiones de las tablas. Verificación de anchos de textos sin desbordamiento.
- Apertura y renderizado del XML nativo con el motor oficial draw.io GraphViewer/mxGraph en Chrome local. Edición temporal de una tabla y reexportación comprobadas sin reemplazar el archivo final. No se afirma una sesión manual del editor web.
- PDF reabierto y comprobado: contiene las 28 tablas, los 208 nombres de atributos y las 50 identificaciones FK, sin tipos visibles. Revisión visual mediante renderizado Poppler.

La posición vertical de un extremo sobre el borde del padre es un puerto distribuido del diseño original: la columna referenciada exacta permanece en el tooltip/metadata del conector y en el inventario del modelo detallado. No se interpreta como referencia al atributo situado a esa misma altura.

## Diferencias pendientes heredadas

No se detectaron cambios estructurales introducidos por esta conversión. Se mantienen las reservas documentadas en `docs/modelos/V24_Logico_CrowsFoot/INCONSISTENCIAS_DETECTADAS.md`: cuenta laboral obligatoria no aplicada, responsables operativos todavía referenciando usuarios, P3 sin unicidad inversa y renombrado de `id_recepcionista` pendiente. La copia no presenta esas propuestas como implementadas.

## Protección y uso

Se verificaron las huellas de los archivos previos, incluido el detallado original; sin modificaciones. No se accedió a PostgreSQL ni MongoDB, no se ejecutaron migraciones, pruebas de aplicación, limpieza ni operaciones Docker. No hubo commit, push ni merge.

Los cuatro entregables están en esta carpeta. `.validacion/` contiene únicamente evidencias y herramientas temporales de esta comprobación, ignoradas por Git. Abrir el `.drawio` en draw.io Desktop o Archivo → Abrir desde → Dispositivo. Ampliar el PNG o PDF para revisar atributos; imprimir en formato amplio para conservar legibilidad.
