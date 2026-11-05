# Comparación del lógico vigente con PostgreSQL de desarrollo

Fecha de verificación: **5 de noviembre de 2026**. Consulta real de **elbravodeljac_dev**, PostgreSQL **18.6**, en el contenedor **elbravodeljac-fase-e-dev-postgres-dev-1** del aplicativo local. No se utilizó como fuente de esta entrega la instancia separada del puerto 5433 ni se supuso que las migraciones estuvieran aplicadas.

## Resultado

**La base existente y el modelo lógico vigente coinciden. No se requiere corregir ni regenerar el diagrama.**

- 28 tablas, 208 columnas, 28 claves primarias y 50 claves foráneas.
- 93 índices, 251 restricciones de catálogo (incluidas las NOT NULL registradas por PostgreSQL 18), 26 secuencias, tres vistas, 12 funciones propias y diez triggers de usuario.
- Extensiones instaladas: pg_trgm 1.6, plpgsql 1.0, unaccent 1.1. Las funciones pertenecientes a extensiones no se cuentan como funciones propias.

Se compararon nombres y orden de columnas, tipos, nulabilidad, PK simples/compuestas, pares de columnas de las 50 FK, nombres de sus restricciones, acciones ON DELETE y unicidad individual de las FK. Se normalizaron representaciones equivalentes como BIGSERIAL/BIGINT con secuencia y TIMESTAMPTZ/timestamp with time zone; no se alteraron tipos para forzar coincidencias.

También se contrastó el Draw.io sin tipos: mismas 28 tablas y 208 filas de atributos, y mismos 50 conectores, columnas origen/destino y símbolos Crow’s Foot. Sus cardinalidades concuerdan con NOT NULL e índices únicos no parciales del catálogo real. No se confundieron condiciones de negocio de Go con restricciones declarativas de PostgreSQL.

Se comprobaron además las marcas PK/FK de cada atributo del editable. Una segunda lectura del catálogo confirmó que las definiciones no cambiaron durante la exportación. Se contaron directamente en el SQL generado sus 28 tablas, 208 columnas, 50 FK, 28 PK, 93 índices, 26 secuencias, tres vistas, 12 funciones y diez triggers; también se revisaron los 13 enlaces locales del README y este informe.

**Diferencias detectadas: cero** en los elementos comparados, incluidos los nombres de las restricciones FK. No se añadió ninguna tabla, campo, relación o funcionalidad futura.

## Inventario comprobado en la instancia real

| Tabla | Columnas | PK | FK salientes |
| --- | ---: | --- | ---: |
| pagos | 10 | id_pago | 2 |
| roles | 3 | id_rol | 0 |
| auxilios | 12 | id_auxilio | 6 |
| clientes | 8 | id_cliente | 1 |
| entregas | 18 | id_entrega | 4 |
| permisos | 3 | id_permiso | 0 |
| usuarios | 7 | id_usuario | 1 |
| empleados | 4 | id_empleado | 0 |
| repuestos | 5 | id_repuesto | 0 |
| servicios | 4 | id_servicio | 0 |
| vehiculos | 8 | id_vehiculo | 1 |
| avances_ot | 5 | id_avance | 2 |
| pruebas_ot | 7 | id_prueba | 2 |
| proveedores | 5 | id_proveedor | 0 |
| recepciones | 13 | id_recepcion | 5 |
| cotizaciones | 12 | id_cotizacion | 3 |
| diagnosticos | 13 | id_diagnostico | 2 |
| rol_permisos | 2 | id_rol, id_permiso | 2 |
| usuario_roles | 2 | id_usuario, id_rol | 2 |
| asignaciones_ot | 6 | id_asignacion | 2 |
| ordenes_trabajo | 14 | id_orden | 3 |
| tipos_documento | 4 | id_tipo_documento | 0 |
| unidades_auxilio | 4 | id_unidad_auxilio | 0 |
| contactos_cliente | 9 | id_contacto | 2 |
| detalle_cotizacion | 12 | id_detalle | 3 |
| repuesto_proveedor | 4 | id | 2 |
| historial_estados_ot | 6 | id_historial | 2 |
| historial_propietarios | 8 | id_historial | 3 |

## Entregables y procedencia

- [Lógico vigente editable, sin tipos](Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.drawio), [PDF](Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.pdf) y [PNG](Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.png): conservados sin sobrescribirlos.
- [SQL oficial de entrega](../../database/postgresql/esquema_entrega.sql): exportación de estructura de la base real, no reconstrucción inferida del repositorio. Incluye extensiones, tablas, defaults, PK, FK, CHECK/NOT NULL, índices, secuencias y vinculaciones, vistas, funciones y triggers.
- El inventario técnico V24 con tipos se utilizó para contrastar las propiedades que no se dibujan en un lógico sin tipos. No se considera otra versión lógica oficial.

## Seguridad, alcance y límites

Las consultas al catálogo se ejecutaron dentro de una transacción REPEATABLE READ READ ONLY, con PGOPTIONS default_transaction_read_only=on. El pg_dump estructural utilizó también ese modo, con --schema-only, --no-owner, --no-privileges y --no-comments. No se exportaron filas, contraseñas, cuentas globales, propietarios, ACL, comentarios de usuarios ni valores actuales de secuencias.

No se ejecutaron migraciones, restauraciones, INSERT, UPDATE, DELETE, DROP, TRUNCATE ni operaciones sobre MongoDB. No se crearon bases de prueba ni se probó la carga de este SQL en un servidor; la tarea se limitó a lectura de la instancia existente y preparación documental. El archivo contiene DDL para una base vacía independiente: no debe ejecutarse sobre bases existentes ni mezclarse con el esquema inicial/migraciones del arranque técnico.

Se conserva database/postgresql/esquema_actual.sql como dependencia técnica inicial de Go/Docker/pruebas, junto con todas las migraciones. El README distingue ese archivo de la exportación oficial de entrega. No se eliminaron, renombraron o movieron archivos del proyecto ni se rediseñaron Login, Dashboard o PDF.

La verificación acredita la estructura de esta instancia y fecha, no la de cualquier otra instalación ni el contenido de sus registros. La evidencia verificada está en estructura-real.json, comparacion.json y validacion-entrega.json dentro de backend/.cache/entrega-universitaria-20261105, ignorada por Git y sin filas de negocio. No se hizo commit, push ni merge.
