# Auditoría estructural V24 — SQL Server 2025

**Estado:** revisión de modelos V24 frente a `database/sqlserver/esquema_entrega.sql` de `main`; no se alteraron el esquema ni los datos.

## Comprobaciones

| Control | Script SQL Server | Diagrama lógico | Diagrama físico | Resultado |
|---|---:|---:|---:|---|
| Tablas | 28 | 28 | 28 | Coincide |
| Atributos/columnas | 208 | 208 | 208 | Coincide |
| Relaciones FK | 50 | 50 | 50 | Coincide |
| Tipos SQL Server | 208 | No muestra tipos | 208 | Los 208 coinciden |
| Claves primarias | 28 | Representadas | Representadas | Coinciden |
| Geometría y conexiones frente al V24 de `main` | — | Sin alteración | Sin alteración | Conservadas |

Se contrastaron nombres de tablas y columnas, tipos, nulabilidad, FK (origen y destino) y multiplicidades basadas en nulabilidad/unicidad. Esto es una auditoría documental: no sustituye pruebas de comportamiento o validación de restricciones condicionales.

## Auditoría en SSMS de la base activa

El propietario ejecutó consultas de solo lectura en `elbravodeljac_sqlserver_dev` y obtuvo:

- Recuento general de objetos de usuario: **29 tablas, 213 columnas, 50 FK**.
- La tabla adicional era **`dbo.sysdiagrams` (5 columnas)**, creada para guardar diagramas de SSMS, sin relación con el negocio del taller.
- Al excluir solo `dbo.sysdiagrams`: **28 tablas y 208 columnas**.

Por lo tanto, no se agrega `sysdiagrams` al modelo. Los recuentos por sí solos no comprueban cada índice, trigger o restricción del motor en ejecución.

## Aspectos de negocio documentados sin ejecutar cambios

- `recepciones.id_recepcionista` es el nombre **vigente en SQL Server y Go**. El nombre `id_usuario_registro` sigue como propuesta; cambiarlo requiere revisión coordinada del código, del esquema y de las pruebas.
- `diagnosticos` conserva columnas heredadas, como `evaluacion`, aunque el formulario unificado usa una presentación simplificada. Suprimir columnas requiere análisis de compatibilidad, no basta con editar el diagrama.
- `cotizaciones.id_cotizacion_base` es una **FK autorreferente real** para revisiones. Se conserva.
- `historial_propietarios` posee unicidad **filtrada** para el propietario vigente, que no convierte toda la relación histórica en 1:1.
- `usuarios.id_empleado` es nullable y tiene índice único filtrado para valores no nulos.
- El conceptual Chen de exposición es deliberadamente resumido y no representa una correspondencia 1:1 con las 28 tablas.

## Exportación y aprobación

Los `.drawio` de este PR son la referencia editable. Las cuatro exportaciones PDF/PNG anteriores de PostgreSQL fueron retiradas de la rama. Se prepararon versiones revisadas en la conversación, **pero no se han adjuntado a esta rama de GitHub**. Incorporar las exportaciones tras revisar su presentación antes de fusionar.

**PR #16 se mantiene en borrador.** No tocar SQL Server, MongoDB o Go para resolver diferencias documentales.
