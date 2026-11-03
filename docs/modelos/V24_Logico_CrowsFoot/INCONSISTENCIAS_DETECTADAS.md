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
| Notas antiguas de endpoints sobre idempotencia | ENDPOINTS.md conserva una frase “El SQL actual no tiene almacenamiento de claves”; 006 ya agrega pagos.request_id e índice parcial | Desfase documental. Esta auditoría toma SQL/Go como evidencia más reciente y lo registra sin editar ENDPOINTS.md. La deduplicación de pagos no se generaliza a todos los endpoints |
| Doble CHECK de fechas en historial | esquema base incluye chk_fechas_propietario y ck_historial_fechas equivalentes | Redundancia declarativa existente, sin cambio de cardinalidad. No se elimina ninguno |
| Catálogo/etiqueta y nombre laboral redundantes | 007 sincroniza nombres; 008 sincroniza etiqueta documental | Compatibilidad deliberada, no cuenta/persona duplicada a eliminar automáticamente |
| Stock/costo no equivale a contabilidad de compras | repuestos.stock, precio_referencia; repuesto_proveedor.costo_referencia | No inventar compras ni libro de movimientos. Futuro módulo solo con requisitos confirmados |
| Instancia viva no inspeccionada | Se usó DDL y migraciones del repositorio; documentación histórica de pruebas aisladas | No se certifica estado de una BD actual ni aplicación exitosa de 001–009 en producción. No se establecieron conexiones para esta tarea |

## Protección

Solo se crean documentos en `docs/modelos/V24_Logico_CrowsFoot/`. PROTECCION_ANTES.json y VALIDACION_FINAL.json comparan todos los archivos versionados existentes. No se ejecutó SQL, migración, limpieza, restauración ni comando Docker. No se consultaron registros personales ni credenciales. No se ejecutaron suites backend/frontend porque no se modificó la aplicación: las pruebas de esta entrega validan el modelo y sus formatos, no funcionalidad HTTP.
