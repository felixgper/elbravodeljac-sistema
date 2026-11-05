# Diferencias entre V23 y la implementación

Inspección estática del repositorio al 2026-11-02; no se inspeccionaron registros reales. Se conserva el backend y sus migraciones sin cambios.

| Área | Estado actual contrastado | Lectura en V23 / diferencia |
| --- | --- | --- |
| Empleados | Migración 007: nombre, tipo laboral y activo; cuenta con FK nullable y única | P1 exige exactamente un empleado por cuenta interna como objetivo. No se aplicó NOT NULL. |
| Responsables | Diagnósticos, asignaciones, auxilios y asignación diagnóstica de recepción referencian usuarios técnicos | P2 proyecta la persona EMPLEADO. Exige cuenta vinculada y datos completos; no es una FK física nueva. |
| Autores | Recepcionista, usuario de registro de auxilio/pago/entrega; historial de diagnóstico | R50–R54 distinguen captura de ejecución. R51 se deriva del historial de autores, no de una tabla puente nueva; mínimo uno describe diagnósticos nuevos por API, no audita el legado. |
| Auxilio–recepción | Varios auxilios pueden apuntar a una recepción; Go valida coincidencia cliente/vehículo | P3 propone máximo uno en ambos sentidos; falta confirmación, no está garantizado por SQL. |
| Titular y solicitante | Vehículo con cliente e historial; recepción conserva cliente de esa atención | No hay registro externo que pruebe propiedad legal ni identidad global de personas. “Titular registrado” evita esa afirmación. |
| Concepto cotizado | Validación Go y CHECK de migración 001: servicio o repuesto, referencias excluyentes, alternativa libre tipificada | B1 explica la combinación; las dos relaciones opcionales por sí solas son insuficientes. No se exige crear entradas ficticias de catálogo. |
| Diagnóstico opcional | Recepción directa o técnica; diagnóstico final y misma recepción para cotizar cuando corresponde | No se fuerza diagnóstico ficticio ni se admite omitir el técnico pendiente en una atención que lo requiere. |
| Historial / asignaciones | CreateVehicle y CreateOrder insertan sus hijos en transacciones | Mínimos uno son reglas de alta; no se declara que SQL impida todo padre huérfano ni que se revisaron datos históricos. |
| Proveedor–repuesto | Tabla puente de costo de referencia, repuesto con stock | R12 resume N:M; no equivale a compras ni libro de movimientos. |
| Roles | Migraciones 007/009 separan puesto y autorización; tablas N:M conservadas | Tipo laboral se omite como atributo. No se crean entidad TIPO_EMPLEADO, roles de BD ni GRANT. |
| Documentos | Migración 008 crea catálogo y FK para clientes/contactos | Se añade TIPOS DE DOCUMENTO y R55/R56: componente ya implementado, ausente en V22. |
| MongoDB | metadata y resultados por diagnóstico; archivos GridFS; Go valida permisos en PostgreSQL | Rectángulos verde azulado y líneas discontinuas personalizadas. Referencia lógica sin FK cruzada ni duplicación de entidades transaccionales. |
| Auditoría resumida | Avance/prueba/creación de orden/verificador de entrega tienen referencias de usuario adicionales | No todas las FK de trazabilidad se dibujan para conservar legibilidad; se preservan en código y se documenta esta abstracción. |

## Correspondencia de almacenamiento

Las 27 entidades conceptuales son 25 tablas PostgreSQL representadas individualmente y 2 colecciones MongoDB. Las tres tablas puente restantes (`usuario_roles`, `rol_permisos`, `repuesto_proveedor`) se resumen como relaciones N:M. Así se cubren las 28 tablas resultantes del esquema original y las migraciones 007/008, sin convertir automáticamente cada tabla puente en entidad conceptual. GridFS es el mecanismo de archivos de evidencias, no otro módulo del negocio.

Los nombres internos completos están en ESPECIFICACION_V23.json y RELACIONES_V23.json. La columna de evidencia de AUDITORIA_RELACIONES.csv remite a los archivos Go y vistas existentes. Las cardinalidades representan el negocio indicado con sus reservas, no un volcado exacto de nullable/índices.

## Protección

Se comparan 170 archivos de aplicación, configuración, backend, checks y base de datos contra sus hashes iniciales. No se ejecutó SQL, migración, restauración, borrado, acceso a PostgreSQL/MongoDB ni comandos Docker. La comprobación no afirma que se consultaron o certificaron datos almacenados: afirma que esta tarea no los tocó.
