# Inconsistencias, limitaciones y propuestas pendientes

No se aplica ninguna de las propuestas siguientes. Las diferencias son explícitas: declarativo SQL, disparadores SQL, reglas Go y objetivo conceptual no son equivalentes.

| Hallazgo | Evidencia / estado vigente | Impacto y propuesta separada |
| --- | --- | --- |
| P1: cuenta con empleado obligatorio no aplicado | 007 define usuarios.id_empleado nullable y UNIQUE | Representado 0..1 por ambos extremos. Antes de endurecer, inventariar cuentas internas/técnicas y vincular de manera verificada; nunca inferir por nombres/IDs ni suprimir cuentas |
| P2: responsable conceptual empleado, FK técnica usuario | recepciones.id_mecanico_diagnostico, diagnosticos.id_tecnico, asignaciones_ot.id_usuario, auxilios.id_tecnico → usuarios | Lógico conserva FK reales. technicalUser admite cuenta técnica legada sin vínculo laboral, aunque rechaza empleado vinculado inactivo. No se afirma migración completa a responsables empleados |
| Renombrado aprobado no implementado | recepciones.id_recepcionista sigue en esquema base, CreateReception y lecturas | Mostrar nombre vigente. Propuesta: cambio compatible y coordinado de columna/referencias en Go, consultas, fixtures y documentación; revisar dependencias antes de una futura migración. No se crea script en esta tarea |
| P3: auxilio–recepción sin UNIQUE inversa | auxilios.id_recepcion nullable y sin índice único | Muchos auxilios pueden referir una recepción. Si se decide uno solo, auditar duplicados sin borrarlos, confirmar negocio y planificar UNIQUE individual para no nulos (o UNIQUE normal según política). Go por sí solo no evita carreras; no aplicar aún |
| Obligaciones API mayores que NULL del SQL | auxilios: cliente/vehículo/técnico/unidad nullable; Go los exige en nuevas altas | No dibujar NOT NULL inexistente. Historias legadas pueden seguir sin esos vínculos; no se consultaron datos para certificarlos |
| Mínimos de hijos no declarativos | Go crea detalle, asignación e historial inicial en transacciones | El lógico SQL muestra padre con cero o más hijos. Conceptual puede exigir uno en altas válidas; no afirmar restricción SQL de existencia de hijos |
| Condiciones entre procesos no son simples FK | 003/004 y Go: modalidad explícita, diagnóstico de misma recepción y finalizado | Dibujar FK nullable a diagnóstico más recepción obligatoria; explicar condición. Técnica pendiente no pasa a directa por carecer aún de diagnóstico |
| Autorelación de ampliaciones no previene todo ciclo en SQL | cotizaciones.id_cotizacion_base → cotizaciones | Go controla operaciones admitidas; FK sola permitiría ciclos si se escribiera SQL fuera de API. No se inventa CHECK anticíclos |
| Titularidad legal no acreditada por software | vehículo con cliente registrado e historial verificado internamente | No atribuir propiedad jurídica probada. Si se requiere prueba externa, definir alcance independiente |
| Idempotencia acotada a pagos | 006 agrega pagos.request_id e índice parcial; ENDPOINTS.md describe el contrato vigente | No generalizar esta deduplicación a todos los endpoints |
| Doble CHECK de fechas en historial | esquema base incluye chk_fechas_propietario y ck_historial_fechas equivalentes | Redundancia declarativa existente, sin cambio de cardinalidad. No se elimina ninguno |
| Catálogo/etiqueta y nombre laboral redundantes | 007 sincroniza nombres; 008 sincroniza etiqueta documental | Compatibilidad deliberada, no cuenta/persona duplicada a eliminar automáticamente |
| Stock/costo no equivale a contabilidad de compras | repuestos.stock, precio_referencia; repuesto_proveedor.costo_referencia | No inventar compras ni libro de movimientos. Futuro módulo solo con requisitos confirmados |
| Instancia viva no inspeccionada | Se usó DDL y migraciones del repositorio; documentación histórica de pruebas aisladas | No se certifica estado de una BD actual ni aplicación exitosa de 001–009 en producción. No se establecieron conexiones para esta tarea |

## Alcance

Hallazgos de la revisión estructural del 2/11/2026; no son instrucciones para modificar la base. La instancia consultada el 5/11/2026 se documenta en [COMPARACION_BASE_ACTUAL.md](../COMPARACION_BASE_ACTUAL.md). Los manifiestos de protección puntuales permanecen en Git; se retiran sin alterar el modelo ni las restricciones.
