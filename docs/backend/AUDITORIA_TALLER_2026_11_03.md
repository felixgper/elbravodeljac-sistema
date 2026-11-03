# Auditoría general del taller y limpieza verificada — 3 de noviembre de 2026

## A. Estado general

El entorno de desarrollo quedó **funcional en modo REAL y sin operaciones**, en `http://localhost:8080/`. No se observan errores críticos o altos pendientes en los escenarios ejecutados. Esto es una conclusión acotada a las pruebas descritas, no una certificación de seguridad universal ni una aprobación para producción empresarial.

Se inició en `backend/arquitectura-go-mongodb`, commit `27c6389` (“Mapas actualizados”), con Git limpio. Se conservó la rama y no hubo commit, push ni merge. Los modelos, esquema SQL y migraciones quedaron intactos (90 archivos protegidos por huellas).

Docker Engine 29.7.2 funcionaba. Al principio solo estaba activo un proyecto ajeno, `el-bravo-del-jac` con PostgreSQL 17/MongoDB 8. La inspección completa encontró nuestros contenedores detenidos; se verificaron etiquetas, archivos Compose y volúmenes antes de arrancarlos sin recrear sus bases. No se intervino en el proyecto ajeno ni en el PostgreSQL original ElBravoDelJac.

| Entorno | PostgreSQL | MongoDB | Aplicación |
| --- | --- | --- | --- |
| Desarrollo autorizado `elbravodeljac-fase-e-dev` | elbravodeljac_dev, PostgreSQL 18.6, sin puerto host publicado | elbravodeljac_tecnico_dev, MongoDB 9, sin puerto host publicado | localhost:8080, loopback |
| QA nuevo `elbravodeljac-taller-20261103-test` | elbravodeljac_test, loopback 25435 | elbravodeljac_tecnico_test, loopback 27032 | servidores Go HTTP efímeros de integración |
| Copias nuevas del respaldo, dentro de ese QA | jac_restore_20261103100158 | elbravodeljac_restore_20261103100158 | servidor HTTP efímero de prueba |

Los proyectos QA anteriores no fueron limpiados ni recreados. Las nuevas pruebas escribieron únicamente en QA. Los datos creados al probar la base limpia quedaron en la copia restaurada, nunca se recargaron en desarrollo.

## B. Pruebas por módulo

| Módulo | Verificación realmente ejecutada | Resultado |
| --- | --- | --- |
| Inicio | Lecturas SQL, indicadores derivados, filtros por período y apertura vacía desde navegador | Aprobado |
| Clientes | Alta, consulta, búsqueda normalizada/paginada, edición, documentos, longitudes sin truncar, duplicados y FK; formulario con pérdida de conexión y regreso | Aprobado; corregido borrador perdido |
| Vehículos | Alta/consulta/edición, placas duplicadas, Toyota/Honda y catálogo existente, propietario/historial, kilometraje decimal y actualizaciones concurrentes | Aprobado |
| Recepciones | Cliente/vehículo compatibles, modalidad directa/técnica, asignación, kilometraje, rechazo de empleado inactivo y consulta por cuenta | Aprobado |
| Diagnósticos | Bandejas por responsable, borrador, F5, actualización/finalización, visualización administrativa, IDs ajenos y técnico suplantado | Aprobado |
| Cotizaciones | Conceptos libres y de catálogo, servicio/repuesto excluyentes, descuentos/decimales, revisiones/aprobación, ampliaciones y PDF pendiente/aprobado/rechazado | Aprobado |
| Órdenes | Cotización aprobada, asignación, Mis tareas, permisos por objeto, avances/pruebas, conciliación, cierre e historial | Aprobado |
| Auxilios | Dos incidentes históricos del mismo cliente/vehículo, unidad/técnico, estados y vínculos | Aprobado |
| Pagos y entrega | Pagos parciales, saldo, exceso rechazado, concurrencia/reintentos, receptor/contacto, conformidad, entrega única y ficha de dos páginas | Aprobado |
| Mecánicos/empleados | Registro sin cuenta, vínculo verificado, roles N, rechazo de empleado inactivo para nueva asignación y gestión autorizada | Aprobado |
| Proveedores y catálogos | CRUD permitido, FK proveedor/repuesto, costo/precio/stock, especialidad incompatible no descartada silenciosamente | Aprobado; sin inventar compras/movimientos |
| Reportes | Ingresos reales sin multiplicación por joins, Semana/Mes/Año, evolución/comparación histórica, conciliación y estado vacío | Aprobado |
| Usuarios/permisos | Gerente/Administrador/Mecánico, matriz, roles adicionales, último acceso administrativo, autoescalamiento, JWT/CSRF, revocación y cuenta inactiva | Aprobado |
| MongoDB/GridFS | Alta/consulta/archivo real, validaciones, deduplicación, concurrencia, reserva/compensación, fallo de servicio y autorización PostgreSQL | Aprobado |

## C. Flujo con cuentas independientes

`TestTallerFourLinkedAccounts` usa un Administrador operativo, dos Mecánicos distintos y un Gerente, cada uno con su contexto/cookies independientes y con vínculo laboral de QA. No se concede permiso de Gerente al Administrador para facilitar la prueba.

Se registró cliente y vehículo, recepción técnica con mecánico, borrador de diagnóstico persistido tras F5 y finalización. Administración recuperó el diagnóstico, creó una cotización de **2 × (100 − 10) = 180.00**, descargó el PDF, registró aprobación y generó la orden. El Mecánico A registró ejecución/control, avance y prueba satisfactoria. Administración registró **50 + 130**, verificó receptor/conformidad y entregó. Se recuperaron nuevamente los datos y la ficha PDF; la nueva descarga coincidió con la archivada.

Mecánico B no ve la asignación en su bandeja y recibe rechazo al modificar diagnóstico/orden ajenos o consultar el escáner no autorizado. Intentar cambiar técnico/roles desde el cuerpo de una petición no suplanta al responsable. La sesión del mecánico permanece independiente al cerrar la de Administración.

También se recorrieron servicio directo sin diagnóstico ficticio, ampliaciones, auxilios y MongoDB. Las fechas distintas de los dos auxilios se prueban desplazando solo la fecha de un registro QA un día: no se afirma haber esperado un día ni se alteró una fecha de desarrollo.

## D. Hallazgos reproducidos y gravedad

| Hallazgo actual | Gravedad | Reproducción y evidencia | Resolución |
| --- | --- | --- | --- |
| Dos envíos simultáneos de un formulario alcanzan la API | MEDIO | Dos requestSubmit en el mismo formulario con respuesta retardada producen 2 peticiones. extra-reproduccion.log falla con 2 frente a 1. No se afirma que un doble clic normal siempre lo reprodujera; los botones ya se deshabilitaban | Bloqueo de envío en curso en bindForm; repetición final emite una sola petición |
| Nuevo cliente pierde campos al volver desde el listado | MEDIO | Tras rellenar y volver a clientes/nuevo, nombre queda vacío. draft-reproduccion.log registra el fallo | Borrador por usuario/modo/sesión; se conserva en navegación/error y se limpia tras éxito |
| Respaldo MongoDB informa diferencia por orden de índices | MEDIO, bloqueo preventivo | Documentos/chunks/hashes y opciones idénticos; los mismos índices aparecen en distinto orden tras mongorestore | Comparar lista por nombre manteniendo todas las definiciones y el orden de claves de cada índice. Respaldo nuevo restaurado y verificado |
| Nota antigua de endpoints dice que pagos no almacena claves de reintento | BAJO/documental | Contradice migración 006 y código vigente requestId | ENDPOINTS.md corregido; no se atribuye idempotencia genérica a otras altas |

No se reabrieron errores antiguos sin comprobarlos: la asignación a empleado inactivo devuelve 422; la UI de auxilio actualiza estado; el Auxiliar histórico no autoriza una escritura técnica nueva (403). Sus regresiones volvieron a pasar. No se detectó un error CRÍTICO o ALTO nuevo pendiente en el alcance ejecutado.

## E. Correcciones y archivos

- `dist/assets/js/components/ui.js`: exclusión de reenvío mientras hay guardado pendiente; vuelve a habilitar el formulario tras error/éxito.
- `dist/assets/js/views/clients-vehicles.js`: reutiliza borradores existentes para nuevos clientes; no cambia el formulario ni sus validaciones.
- `backend/scripts/backup-audit.ps1`: permite seleccionar el override del QA nuevo y ordena índices MongoDB por nombre al comparar; no relaja definiciones.
- `docs/backend/ENDPOINTS.md`: actualiza la descripción de requestId y respuestas implementadas.
- `backend/integration/taller_audit_test.go`, `checks/taller-browser.mjs`: cuatro sesiones, 1.200 altas de carga exclusivamente QA, móvil, offline, duplicación y borrador.
- `backend/integration/clean_copy_test.go`, `checks/clean-copy-browser.mjs`: prueba opt-in en copia restaurada vacía, host/puerto/nombre restringidos, cuenta existente y nuevas altas solo en copia.
- `compose.auditoria-20261103.yaml`: proyecto QA nuevo con puertos/volúmenes propios.
- Este informe. Las herramientas de limpieza/manifiestos, respaldos y logs de esta ejecución están en caché privada, excluida de Git.

No se modificó CSS, menú, colores, disposición de formularios, backend funcional Go, tablas ni migraciones. La imagen de API/frontend se reconstruyó y solo se recreó su contenedor; no los contenedores/volúmenes de las bases.

## F. Pendientes y límites

1. **MEDIO: cuentas heredadas.** Hay 35 cuentas, 34 sin empleado vinculado; 20 cuentas activas con rol Mecánico conservan compatibilidad legada. No se inventaron personas ni vínculos.
2. **MEDIO: roles heredados.** Se conservan Supervisor (6 asignaciones) y Auxiliar (14), además de Administrador legado. No se cambian permisos ni se convierten cuentas automáticamente. La revisión corresponde a Gerencia antes de operación empresarial.
3. P1/P2/P3 y renombrado de id_recepcionista requieren decisiones y propuesta compatible; no se ejecutó evolución estructural.
4. Los catálogos contienen ejemplos anteriores. Se preservaron completos porque no era necesario eliminarlos y su clasificación definitiva no se deduce solo por el nombre; el proveedor de origen dudoso se mantiene.
5. No se ejecutó un ensayo de carga de producción con muchos usuarios concurrentes ni una prueba de penetración externa. Los tiempos descritos son locales.
6. Se verificaron errores 400/401/403/404/409/422 en integración y fallos MongoDB/compensación (503), además de errores controlados en pruebas unitarias. No se provocó un 500 destruyendo desarrollo; el recorrido de navegador no observó respuestas 500 ni errores JS inesperados.
7. La expiración/invalidación JWT y cambios de credenciales/permisos se cubren en autenticación e integración. En la nueva prueba UI se retiró la cookie para probar vuelta a login; no se afirma haber esperado manualmente una sesión completa hasta su vencimiento.

## G. Coherencia con modelos y esquema

Se consultaron auditoría anterior, V23 y su inventario/pendientes, V24 conceptual resumido, V24 lógico detallado y lógico sin tipos. La consulta del catálogo PostgreSQL real confirmó **28 tablas, 208 columnas y 50 FK**. Las 50 referencias, columnas origen/destino, NULL y UNIQUE individual coinciden con ESQUEMA_VERIFICADO.json. Los 90 archivos protegidos de modelos/esquema/migraciones mantienen sus hashes.

| Relación | Estado vigente, conservado |
| --- | --- |
| Cliente–vehículo | Titular actual registrado y varios vehículos; historial de traspasos separado |
| Cliente–auxilio / vehículo–auxilio | Incidentes múltiples históricos; NULL físico más permisivo que altas nuevas por API |
| Vehículo/cliente–recepción | Una atención conserva vehículo y cliente de ese momento |
| Recepción–diagnóstico | Cero o uno por UNIQUE; directo no lleva diagnóstico y técnico pendiente puede no tenerlo todavía |
| Recepción–cotización | FK obligatoria; base/ampliaciones controladas, incluso directo |
| Diagnóstico–cotización | FK nullable; si se usa, finalizado y de la misma recepción (Go y triggers) |
| Cotización–orden | Cotización aprobada y aprobación identificada; máximo una orden por cotización |
| Orden–pago / entrega | Varios pagos; entrega única y condicionada a estado/saldo/conformidad |
| Empleado–usuario (P1) | Vínculo nullable y UNIQUE, opcional en ambos sentidos |
| Responsable (P2) | FK técnicas/auditoría siguen hacia USUARIOS; no equivalencia numérica de IDs con empleados |
| Auxilio–recepción (P3) | FK opcional sin unicidad inversa: varios auxilios pueden referir la misma recepción |
| Usuarios–roles–permisos | Puentes N:M y permisos efectivos conservados; cargo no equivale a rol |

id_recepcionista sigue vigente; su renombrado aprobado a id_usuario_registro está pendiente. No se alteraron los modelos para ocultar diferencias. TIPOS_DOCUMENTO continúa implementada. No se agregan compras, movimientos de stock ni FK PostgreSQL hacia MongoDB.

## H. Simplicidad, móvil y rendimiento

Mejoras comprobadas: un guardado pendiente no se duplica mediante reenvío, los datos se conservan ante error de conexión y el nuevo cliente conserva borrador al volver. Se reutilizaron componentes, sin nuevo módulo ni formulario adicional. Las validaciones y campos de integridad se mantuvieron.

El Gerente recorrió los trece módulos con viewport **390×844**, sin desbordamiento horizontal global ni error JS; las tablas mantienen su área de desplazamiento propia. Se comprobó navegación, búsqueda paginada, acciones existentes y permisos visibles. Esto no equivale a una evaluación de todos los teléfonos ni a una sesión de usabilidad con trabajadores reales.

Sigue siendo necesario un ensayo guiado con personal del taller para medir tiempo y clics humanos por tarea; los clics automatizados no representan esa métrica. La conciliación técnica y la matriz de permisos tienen complejidad justificada: no se eliminaron controles de integridad o autorización por simplificar.

Se midieron **20 consultas secuenciales** desde navegador a `/clientes?pageSize=20&page=1` con al menos **1200 registros QA**: mediana **66 ms**, percentil 95 **89 ms**, máximo **112 ms**. Incluye recorrido cliente/API, no solo SQL. No es un SLA ni una prueba de concurrencia comercial. Las pruebas de transacciones/pagos/versiones concurrentes son escenarios funcionales aparte.

## I–J. Limpieza ejecutada e inventario eliminado

Identidad verificada por nombres reales, etiquetas Compose, rutas de configuración, volúmenes y catálogo. La API propia se pausó; PostgreSQL mostró cero conexiones ajenas de aplicación. Se generó el respaldo nuevo y se compararon huellas antes/después para detectar escrituras. El borrado PostgreSQL adquirió locks, ejecutó DELETE explícitos en orden inverso de dependencias y comparó tablas protegidas dentro de la transacción. No hubo CASCADE, TRUNCATE, DROP de bases/esquemas, reinicio de secuencias ni eliminación de volúmenes.

Las tablas temporales de verificación se descartan al confirmar la transacción; no son nuevas tablas del modelo ni migraciones.

| Tabla operativa | Filas eliminadas | Estado final |
| --- | ---: | ---: |
| clientes | 27 | 0 |
| contactos_cliente | 1 | 0 |
| vehiculos | 36 | 0 |
| historial_propietarios | 36 | 0 |
| recepciones | 30 | 0 |
| diagnosticos | 21 | 0 |
| cotizaciones | 25 | 0 |
| detalle_cotizacion | 47 | 0 |
| ordenes_trabajo | 22 | 0 |
| asignaciones_ot | 36 | 0 |
| avances_ot | 46 | 0 |
| pruebas_ot | 21 | 0 |
| historial_estados_ot | 68 | 0 |
| pagos | 15 | 0 |
| entregas | 14 | 0 |
| auxilios | 6 | 0 |
| **Total** | **451** | **0** |

Este conteo nuevo coincidió con 451, pero no se tomó de la auditoría anterior: proviene del respaldo actual, los DELETE y la comprobación posterior.

| Colección técnica | Elementos eliminados | Estado final |
| --- | ---: | ---: |
| archivos_tecnicos.chunks | 6 | 0 |
| archivos_tecnicos.files | 6 | 0 |
| evidencias_tecnicas | 6 | 0 |
| resultados_escaner | 6 | 0 |

Se comprobó cada vínculo evidencia–archivo–diagnóstico y metadata managedBy antes de seleccionar archivos. Solo se borraron IDs enumerados en el manifiesto; documentos y chunks tenían huellas por registro. No quedan archivos/chunks ni evidencias huérfanas; índices y validadores siguen iguales.

## K. Información conservada

| Tabla protegida | Filas conservadas | Comprobación |
| --- | ---: | --- |
| empleados | 1 | Conteo y huella idénticos |
| permisos | 40 | Conteo y huella idénticos |
| proveedores | 1 | Conteo y huella idénticos |
| repuesto_proveedor | 0 | Conteo y huella idénticos |
| repuestos | 7 | Conteo y huella idénticos |
| rol_permisos | 164 | Conteo y huella idénticos |
| roles | 7 | Conteo y huella idénticos |
| servicios | 14 | Conteo y huella idénticos |
| tipos_documento | 4 | Conteo y huella idénticos |
| unidades_auxilio | 6 | Conteo y huella idénticos |
| usuario_roles | 42 | Conteo y huella idénticos |
| usuarios | 35 | Conteo y huella idénticos |

Son **35 cuentas (34 activas)**, un empleado y sus vínculos, 7 roles almacenados, 40 permisos, 42 asignaciones usuario–rol y **164** asignaciones rol–permiso (conteo actual). Se conservaron también hashes de contraseña, nombres, estado, catálogos, configuración privada y las estructuras existentes.

Desarrollo no tiene una cuenta del rol canónico Administrador; no se transformó Administrador legado ni se creó una cuenta ficticia para ocultarlo. Se comprobó el inicio de sesión del **Gerente existente** tanto en copia como en desarrollo. Las cuentas Administrador y Mecánicos nuevas se crearon y autenticaron por API **solo en la copia**. No se conocen ni se probaron las contraseñas individuales de las otras cuentas antiguas; su conservación se acreditó por filas completas/huellas, sin cambiar hashes ni estado.

## L. Respaldo y restauración

Respaldo válido, privado y persistente fuera de los volúmenes Docker:

`backend/.cache/respaldo-auditoria-20261103100158/`

- `postgres.dump`: formato custom de elbravodeljac_dev.
- `mongo.archive.gz`: elbravodeljac_tecnico_dev completo, incluidas colecciones y GridFS.
- `SHA256.json`: hashes de ambos archivos, recomprobados antes de limpiar.
- Inventarios, manifests por ID, recibos de eliminación y resultados protegidos, todos excluidos de Git.

Se restauró en bases nuevas del QA actual. Coincidieron filas y huellas de las 28 tablas, claves/constraints/índices PostgreSQL, documentos/chunks, validadores e índices MongoDB. La comparación SQL conserva la normalización previa de casts equivalentes de literales de CHECK; no ignora otros cambios. La corrección MongoDB solo ordena **la lista de índices por nombre**, sin alterar sus claves/definiciones.

El primer intento se bloqueó por orden diferente de índices y **no produjo borrados de desarrollo**. Se guardó como evidencia y se hizo un respaldo posterior nuevo que pasó la restauración completa. La copia primero pasó una simulación con rollback y huellas intactas; después se limpió deliberadamente para el ensayo funcional de base vacía.

PostgreSQL y MongoDB no comparten una transacción. Se guardó el manifiesto antes del DELETE SQL y el recibo SQL antes de aplicar MongoDB. La fase Mongo valida contenido de cada registro y admite reanudación sobre el subconjunto pendiente de IDs del plan, sin aceptar documentos nuevos o alterados. Si falla entre motores, la API debe permanecer en mantenimiento; completar el manifiesto verificado o restaurar ambos archivos a **destinos aislados nuevos**, comprobar coherencia y decidir la recuperación conjunta. No restaurar ciegamente sobre una base real/existente. Esta ejecución no sufrió fallo entre motores; no se afirma haber ensayado una caída de host en ese punto exacto.

Para operación comercial conviene programar y ensayar respaldos y recuperación de ambos motores con almacenamiento privado independiente. Referencias primarias: [PostgreSQL: SQL dump](https://www.postgresql.org/docs/18/backup-dump.html), [MongoDB: mongodump y metadatos](https://www.mongodb.com/docs/database-tools/mongodump/).

## M. Aplicación operativa vacía

`TestCleanRestoredCopy` fue omitido en la suite general sin su parámetro opt-in y **se ejecutó aparte con su DSN QA explícito**, aprobando en 12,07 s. Verificó listados vacíos, Dashboard, Reportes Semana/Mes/Año, Gerente existente, creación/autenticación de nuevas cuentas Administrador y dos Mecánicos, y nuevos cliente/vehículo Toyota con kilometraje 1250.25 persistido. Todas esas escrituras quedaron en la copia restaurada.

Después de limpiar desarrollo y reconstruir solo la API/frontend, una prueba Chrome de solo lectura comprobó:

- modo REAL, login y logout del Gerente existente;
- trece módulos sin error y siete listados operativos vacíos;
- Reportes Semana/Mes/Año con incomeCents=0;
- `/readyz` listo con PostgreSQL y MongoDB;
- 35 cuentas y un empleado presentes;
- cero escrituras de negocio desde el navegador y cero errores JS/HTTP inesperados.

La comparación posterior a reabrir confirmó **16 tablas operativas en cero y 12 protegidas idénticas**. MongoDB sigue con sus cuatro colecciones en cero. No se cargaron fixtures/demo al iniciar. La aplicación está en `http://localhost:8080/`; usar las credenciales existentes. El archivo de acceso conocido permanece local en `backend/.cache/acceso-desarrollo.txt` y no se publica.

## N. Antes de uso comercial

1. Gerencia debe identificar personas reales y revisar cuentas/roles heredados; crear Administrador operativo si corresponde, sin convertir automáticamente cuentas existentes.
2. Resolver P1/P2/P3 y planificar por separado el renombrado de id_recepcionista con sus dependencias. No endurecer relaciones sin revisar el legado.
3. Clasificar catálogos de demostración y mantener solo datos empresariales aprobados; no deducir borrados por nombre.
4. Definir despliegue HTTPS y revisar configuración de cookies/secretos para ese entorno; este localhost usa HTTP de desarrollo. La [API Cookie de Go](https://pkg.go.dev/net/http#Cookie) expone las opciones usadas; esta auditoría no certifica un despliegue externo.
5. Ensayar recuperación conjunta y control de acceso a los respaldos, así como monitoreo de errores y vencimientos reales de operación.
6. Realizar usabilidad con trabajadores y una prueba de carga dimensionada a concurrencia real. Las mediciones locales no sustituyen ambos ensayos.

## Evidencia y comandos

```powershell
.\backend\scripts\test-audit.ps1 -StartIsolated `
 -TestProject elbravodeljac-taller-20261103-test `
 -TestOverride compose.auditoria-20261103.yaml -PostgresPort 25435 -MongoPort 27032
# Repeticiones sin StartIsolated; no recrear entornos existentes.
.\backend\scripts\test-audit.ps1 `
 -TestProject elbravodeljac-taller-20261103-test `
 -TestOverride compose.auditoria-20261103.yaml -PostgresPort 25435 -MongoPort 27032
```

Ejecuciones reales: cinco suites JavaScript de checks/run.mjs, `go test ./... -v -count=1` con PostgreSQL/MongoDB y navegadores habilitados, `go vet ./...`; integración final 83,637 s. Prueba opt-in de copia limpia ejecutada aparte. Las peticiones de QA usan puertos 25435/27032, nunca el PostgreSQL original del puerto 5432.

Evidencia privada: `backend/.cache/auditoria-20261103/` (base-go.log, extra-reproduccion.log, draft-reproduccion.log, final-suite.log, final-go.log, clean-test.log, clean-probe.json, development-readonly.json, model-validation.json, logs de respaldo/limpieza y capturas), más `backend/.cache/taller-audit/` y el respaldo privado. El tiempo de revisión automática agotado se resolvió reintentando la prueba QA; no se tomó como ejecución exitosa. Una consulta localhost desde el sandbox falló; la verificación autorizada de Chrome confirmó `/readyz` y la aplicación, sin atribuir esa restricción a una caída del servicio.

No se consideran aprobadas pruebas no ejecutadas. No se tocaron bases/proyectos ajenos ni modelos académicos. No hubo commit, push ni merge.
