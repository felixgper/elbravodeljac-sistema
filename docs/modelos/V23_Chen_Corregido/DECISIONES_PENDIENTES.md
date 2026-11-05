# Decisiones pendientes — V23

V23 sirve para revisión académica; no autoriza migraciones.

| Marca / asunto | Confirmación necesaria | Implementación actual y alternativa segura |
| --- | --- | --- |
| P1: cuenta interna obligatoriamente laboral | Determinar cuentas internas, técnicas e históricas; política de excepciones | `usuarios.id_empleado` es nullable y único. Inventariar y vincular cuentas sin suprimirlas; solo después proponer obligatoriedad o excepciones expresas. |
| P2: empleado ejecutor / cuenta registradora | Confirmar semántica de responsables y regularizar cuentas técnicas sin vínculo laboral | Mantener FK a usuarios. Resolver empleado vía cuenta vinculada; no reemplazar identificadores de usuarios por identificadores de empleados. Un empleado sin cuenta puede existir, pero hoy no se asigna como técnico operativo sin cuenta técnica habilitada. |
| P3: auxilio–recepción (0,1)/(0,1) | ¿Una recepción puede derivar de varios auxilios? | Hoy la referencia nullable permite varios auxilios por recepción. Si se confirma unicidad, auditar vínculos y diseñar control compatible; si se necesitan varios, ratificar recepción (0,N) y auxilio (0,1). No borrar registros para encajar la propuesta. |
| Titularidad legal | ¿Se verificará documentalmente propiedad legal? | Mantener “titular registrado”. No afirmar comprobación registral que el software no realiza. |
| Solicitante distinto del titular / quien trae vehículo | ¿Debe registrarse independientemente? | El alta actual exige cliente registrado del vehículo; los contactos autorizados se utilizan en entregas. No reutilizarlos como propietario o solicitante sin una decisión. |
| Ejecutores de pruebas / avances | ¿Debe conservarse además el ejecutor físico si captura Administración? | Actualmente se conserva usuario actor. La autoría no prueba ejecución física; proponer dato adicional solo si se confirma necesidad. |
| Inventario y compras | ¿Se requiere trazabilidad de compras, recepciones y consumos? | Hoy catálogo, stock y costo de referencia de proveedor. Un futuro libro de movimientos exige propuesta independiente; no inventar compras ni consumos retrospectivos. |
| Mínimos en historiales | ¿Existen registros antiguos sin trazas o períodos? | No se consultaron bases. Los mínimos del alta por API no certifican el legado. Revisar con acceso autorizado antes de cualquier endurecimiento. |

Mientras estas decisiones no se resuelvan, no presentar P1/P2/P3 como restricciones físicas vigentes ni declarar el modelo perfecto.
