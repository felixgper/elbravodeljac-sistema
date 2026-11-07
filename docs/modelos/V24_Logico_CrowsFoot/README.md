# Referencia física detallada V24 — con tipos

Única referencia detallada oficial **con tipos**. Su nombre histórico de carpeta/archivo dice “lógico”; el [lógico oficial sin tipos](../Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.drawio) está separado. No se renombra ni modifica el diagrama.

28 tablas, 208 atributos y 50 FK; PK, FK, unicidad y nulabilidad. Una página, seis bloques, Crow’s Foot ortogonal. El diagnóstico sigue siendo opcional según modalidad. La vinculación empleado/cuenta continúa opcional y única.

## Entregables

[Draw.io editable](El_Bravo_del_JAC_V24_Logico_Completo.drawio) · [PDF vectorial A0 horizontal](El_Bravo_del_JAC_V24_Logico_Completo.pdf) · [PNG 7993 × 5400](El_Bravo_del_JAC_V24_Logico_Completo.png) · [SVG](El_Bravo_del_JAC_V24_Logico_Completo.svg).

Abrir en draw.io Desktop o Archivo → Abrir desde → Dispositivo. Exportar desde el editor manteniendo una página y área completa. El editable nativo contiene todo; no requiere V23, capturas de bloques ni herramientas ligadas a una caché privada.

## Fuentes de mantenimiento

Se conservan ESQUEMA_VERIFICADO.json, INVENTARIO_FK.csv, REGLAS_POR_FK.json, [inventario](INVENTARIO_TABLAS_Y_FK.md), [cardinalidades](REVISION_CARDINALIDADES.md), [niveles](COMPARACION_CONCEPTUAL_LOGICO_FISICO.md) y [limitaciones](INCONSISTENCIAS_DETECTADAS.md).

`inventariar_esquema.py` reconstruye el inventario desde el esquema inicial y 001–009; `generar_logico.py` produce el editable desde ese inventario. Son herramientas opcionales que sobrescriben sus salidas; no ejecutarlas para consultar/exportar la entrega aprobada. Sus salidas auxiliares de geometría no son entradas necesarias. Los validadores/exportadores retirados dependían de rutas privadas, GraphViewer en caché y hashes de un estado histórico.

La [comparación de PostgreSQL real del 5/11/2026](../COMPARACION_BASE_ACTUAL.md) complementa la revisión original basada en DDL; no acredita otra instancia. La limpieza conserva editables y exportaciones byte por byte, sin ejecutar generadores, SQL ni migraciones.
