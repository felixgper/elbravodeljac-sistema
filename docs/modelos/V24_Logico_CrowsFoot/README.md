# Referencia física detallada V24 — con tipos

Referencia física oficial de PostgreSQL con tipos de datos. El [modelo lógico oficial sin tipos](../Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.drawio) está disponible por separado.

28 tablas, 208 atributos y 50 FK; PK, FK, unicidad y nulabilidad. Una página, seis bloques, Crow’s Foot ortogonal. El diagnóstico sigue siendo opcional según modalidad. La vinculación empleado/cuenta continúa opcional y única.

## Entregables

[Draw.io editable](El_Bravo_del_JAC_V24_Logico_Completo.drawio) · [PDF vectorial A0 horizontal](El_Bravo_del_JAC_V24_Logico_Completo.pdf) · [PNG 7993 × 5400](El_Bravo_del_JAC_V24_Logico_Completo.png).

Abre el archivo en draw.io Desktop o desde Archivo → Abrir desde → Dispositivo. Al exportar, conserva el área completa en una sola página.

## Alcance del modelo

El [esquema oficial](../../../database/postgresql/esquema_entrega.sql) define la estructura para instalaciones nuevas. La comparación del 5 de noviembre de 2026 verificó las tablas, columnas y claves del modelo lógico contra el catálogo de PostgreSQL de desarrollo de esa fecha.

El modelo representa las restricciones declarativas de PostgreSQL. Las referencias a responsables operativos siguen apuntando a usuarios. El vínculo entre empleado y cuenta es opcional y único. Una recepción puede estar referenciada por varios auxilios. Reglas como exigir un diagnóstico finalizado para cotizar una atención técnica, validar el alcance de un usuario o exigir asignaciones en nuevas órdenes deben leerse junto con el [contrato de la API](../../backend/ENDPOINTS.md).
