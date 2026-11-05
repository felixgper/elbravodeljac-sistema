# Auditoría funcional del taller — 2 de noviembre de 2026

## Resultado y alcance

Se ejecutaron las fases de inspección, pruebas, correcciones y preparación de
limpieza. **No se ejecutó la limpieza definitiva ni la fase 6.** Ambas requieren
la aprobación explícita solicitada por el usuario.

La rama inicial `backend/arquitectura-go-mongodb` estaba limpia, en `dd68084`
(integración de catálogo documental/tres roles). Se mantuvo esa rama y no se
hicieron commit, push ni merge. Los modelos académicos no se modificaron.

Docker Engine 29.7.2 estaba disponible. Desarrollo autorizado:

| Servicio | Proyecto | Base / dirección |
|---|---|---|
| PostgreSQL | `elbravodeljac-fase-e-dev` | `elbravodeljac_dev`, puerto interno 5432 sin publicación |
| MongoDB | El mismo | `elbravodeljac_tecnico_dev`, sin puerto publicado |
| Go y frontend real | El mismo | `http://localhost:8080/` |
| QA PostgreSQL | `elbravodeljac-taller-20261102-test` | `elbravodeljac_test`, loopback 25434 |
| QA MongoDB | El mismo proyecto QA | `elbravodeljac_tecnico_test`, loopback 27031 |

QA se creó con volúmenes nuevos y persistentes, mediante
`compose.test.yaml` + `compose.auditoria-taller.yaml`. No reutiliza volúmenes de
desarrollo. No se reiniciaron ni eliminaron los otros proyectos existentes.
No se accedió al PostgreSQL original `ElBravoDelJac` ni se restauraron respaldos
sobre él. Las migraciones 001–009 y las 28 tablas / 50 FK se conservaron.

## Pruebas realmente ejecutadas

Primero se ejecutó la suite existente sin correcciones: aprobó. Después se
añadieron escenarios que faltaban; estos reprodujeron fallos y sirvieron como
regresión. La aprobación de la suite anterior no se tomó como prueba de que
todos los comportamientos empresariales estuvieran cubiertos.

Las cuentas de QA son registros reales con bcrypt y roles canónicos de
PostgreSQL: **un Administrador operativo y dos Mecánicos distintos**, cada uno
vinculado a su propio empleado. No se concedieron permisos de Gerente al
Administrador. Se usaron tres contextos de navegador separados, cookies y
sesiones independientes. Los nombres y contraseñas de estas cuentas no se
publican en el informe; los datos introducidos son exclusivamente de prueba.

| Caso | Evidencia ejecutada / resultado |
|---|---|
| A: Administración | Alta de cliente con Pasaporte y número, vehículo Honda asociado, km `150.50`, síntomas, recepción de diagnóstico y asignación al primer mecánico. SQL comprueba `id_recepcionista`, vehículo, cliente y responsable. |
| B: Mecánico | Su bandeja contiene la recepción, la del segundo mecánico no. Registra hallazgos/evaluación/recomendaciones, guarda borrador, F5 conserva el texto y finaliza. Administración consulta el diagnóstico final. |
| C: Separación | Segundo mecánico recibe 403 al editar diagnóstico/orden ajenos. Cambiar el responsable en el cuerpo no suplanta al asignado. Intentar roles/vínculo laboral por petición de un mecánico recibe 403. Las sesiones siguen separadas tras cerrar la de Administración. |
| D: Continuación | Cotización de 2 × (100 − 10) = **180.00**, PDF pendiente, aprobación explícita, orden con mecánico, avances, conceptos realizados en control técnico, prueba satisfactoria, finalización. Pagos reales de **50 + 130**, exceso rechazado y entrega antes de saldo completo rechazada. Entrega final con conformidad y ficha PDF de dos páginas. Reingreso conserva entrega y nueva descarga idéntica. |
| E: Directo | Recepción directa y cotización persistida con `diagnosisId=null`, sin diagnóstico ficticio. La suite previa también recorre el servicio directo hasta entrega/PDF. |
| F: Auxilio | Dos altas HTTP reales del mismo cliente y vehículo, IDs independientes, mecánico correcto y actualización desde el nuevo control de estado. SQL comprueba persistencia de ambos y relación con el usuario vinculado a empleado. |
| MongoDB | Escáner y evidencia PDF cargados mediante servicios reales JS/Go, archivo recuperado de GridFS y reintento de escáner sin duplicar. Segundo mecánico recibe 403 al consultar el escáner ajeno. |

**Precisión del caso de fechas:** ambas altas de auxilio se hicieron durante
la misma ejecución. Para comprobar historial en fechas diferentes, el test
desplaza únicamente la fecha de su primer registro QA un día hacia atrás y
verifica dos fechas en SQL. No se simuló haber esperado un día ni se modificó
una fecha de desarrollo. El formulario usa la fecha del servidor.

También se ejecutaron los casos existentes de documentos y placas duplicados,
cliente/vehículo incompatible, usuario inactivo, roles N, revocación de
permisos, JWT inválido, CSRF, último Gerente, transacciones, versiones
concurrentes, sobrepago, entrega sin pago completo, catálogos y precios,
stock, reportes, reintentos MongoDB/GridFS y fallo de MongoDB. Son pruebas de
integración con motores reales, distintas de los mocks de las pruebas unitarias.

Comandos ejecutados:

```powershell
# Primera ejecución: crea un proyecto nuevo; no recrea uno existente.
.\backend\scripts\test-audit.ps1 -StartIsolated `
 -TestProject elbravodeljac-taller-20261102-test `
 -TestOverride compose.auditoria-taller.yaml -PostgresPort 25434 -MongoPort 27031
# Reejecución tras correcciones, sin StartIsolated:
.\backend\scripts\test-audit.ps1 `
 -TestProject elbravodeljac-taller-20261102-test `
 -TestOverride compose.auditoria-taller.yaml -PostgresPort 25434 -MongoPort 27031
```

Resultado: `node checks/run.mjs` (cinco suites), `go test ./... -v -count=1`
con integración y navegador habilitados, y `go vet ./...`: **aprobados**.
El recorrido de tres cuentas se repitió para obtener capturas con las pantallas
cargadas. Las pruebas de navegador usaron Chrome y la suite existente también
usa Edge. No hubo errores HTTP 500 ni errores de consola en el recorrido nuevo.
La traza SQL final contabilizó 8 cancelaciones por navegación/F5, una violación
de unicidad esperada y cero errores SQL inesperados. El logger genérico puede
anotar «fallo de persistencia/servicio» al cancelar una lectura; esas entradas
se distinguieron de una caída real de PostgreSQL, sin ocultar errores de escritura.

Evidencias locales, excluidas de Git:

- `backend/.cache/auditoria-taller-base.log`: suite anterior a las correcciones.
- `backend/.cache/taller-reproduccion.log`: empleado inactivo aceptado antes de corregir.
- `backend/.cache/auxiliar-mongo-reproduccion.log`: autorización histórica antes de corregir.
- `backend/.cache/audit-native.log`, `auditoria-taller-final.log`: suite final.
- `backend/.cache/taller-audit/resultado.json`, `browser.log`: recorrido nuevo.
- Capturas `01-recepcion.png`, `02-borrador-tras-f5.png`,
  `03-orden-finalizada.png`, `04-pago-entrega.png`,
  `05-auxilios-independientes.png` en `backend/.cache/taller-audit/`.
- PDF de cotización y ficha técnica en esa misma carpeta.

No se ejecutaron pruebas contra producción. La limpieza definitiva, la
validación de base limpia y el alta sin dejar registros en esa base permanecen
**pendientes de autorización**, no se presentan como aprobadas.

## Hallazgos y correcciones

### H1 — ALTO, corregido: asignación a empleado inactivo

- Reproducir: empleado inactivo, cuenta activa con Mecánico; enviar una nueva
  recepción con diagnóstico asignado a esa cuenta.
- Esperado: rechazo 422 y ausencia de nueva recepción asignada a esa persona.
- Observado: la petición se aceptaba y persistía. La prueba de navegador falló
  precisamente al esperar el rechazo.
- Causa: `technicalUser` revisaba actividad de cuenta y rol, pero no del empleado
  vinculado. El directorio de mecánicos tenía la misma omisión.
- Corrección mínima: excluir empleados vinculados inactivos en ambas consultas.
  La comprobación se reutiliza al asignar diagnóstico, orden y auxilio. No se
  borra ni modifica una cuenta. Las cuentas heredadas sin vínculo conservan la
  compatibilidad previamente aprobada; véase H4.
- Revalidación: la nueva recepción devuelve 422 y el recorrido con el empleado
  activo continúa correctamente.

### H2 — ALTO, corregido: estado de Auxilio sin control en interfaz

- Reproducir: abrir Auxilio, registrar un incidente y consultar su tabla.
- Esperado: poder seguir su estado con las autorizaciones existentes.
- Observado: tabla sin estado ni acción de actualización; el backend ya tenía
  `PATCH /auxilios/{id}` funcional.
- Causa: faltaba la conexión de esa operación al formulario.
- Corrección mínima: columnas Estado/Acción y modal Actualizar con estado,
  técnico y unidad. Usa la API y el permiso `auxilio.crear` existentes. No se
  añadió un módulo ni CSS; se reutilizan tabla, modal y formularios actuales.
- Revalidación: dos incidentes independientes se crean y actualizan desde UI,
  sus relaciones permanecen correctas en PostgreSQL.

### H3 — ALTO, corregido: Auxiliar histórico como autorización de escritura técnica

- Reproducir en QA: participación activa histórica `AUXILIAR` en una orden y
  cuenta con `diagnosticos.crear`, sin ser responsable de ese diagnóstico.
- Esperado: 403 al registrar escáner/evidencia; la relación histórica no asigna
  trabajo técnico nuevo.
- Observado: el filtro de PostgreSQL aceptaba cualquier asignación activa. La
  regresión llegó al control de MongoDB y devolvió 503 porque ese servidor de
  prueba no configuraba el adaptador; **no se afirma que esa reproducción
  escribiera documentos**. Demuestra que superó indebidamente autorización.
- Corrección mínima: la autorización por asignación exige `MECANICO` para
  escrituras. Consulta histórica y responsables/supervisión siguen sus reglas.
- Revalidación: 403 antes de acceder a MongoDB. Las pruebas con MongoDB real
  también aprobaron las escrituras y lecturas legítimas.

### H4 — MEDIO, pendiente de revisión de cuentas heredadas

Desarrollo tiene 35 cuentas y un empleado: 34 cuentas carecen de vínculo,
incluidas 20 cuentas activas con Mecánico. Esto proviene de la transición
compatible aprobada, no de una pérdida causada por esta auditoría.
`usuarios.id_empleado` es opcional y único. No se deben inventar empleados ni
unir personas por parecido de nombre. Solución: Gerente revisa cada cuenta,
registra el empleado real y vincula explícitamente. Una obligación global de
vínculo requiere decidir antes cómo tratar las cuentas heredadas.

### H5 — MEDIO, transición de roles todavía pendiente

Se mantienen asignaciones de Supervisor (6) y Auxiliar (14), retiradas para
altas nuevas. Sus permisos efectivos se conservan según la decisión anterior.
Solo tres roles se ofrecen para nuevas asignaciones. No convertir ni retirar
estos accesos automáticamente. El nuevo Administrador tiene cero cuentas en
desarrollo; la cuenta Gerente puede operar. La auditoría creó un Administrador
operativo real únicamente en QA para probar su política de 34 permisos.

### H6 — BAJO/documental: cardinalidades SQL menos estrictas que la API

No se detectó mezcla entre IDs de empleados y cuentas. `id_tecnico`,
`id_recepcionista`, `id_usuario` y asignaciones apuntan a **USUARIOS**; el cargo
laboral se obtiene por `usuarios.id_empleado`. Los historiales de diagnóstico
guardan `by` como usuario actor y el responsable como `id_tecnico`.

| Relación | SQL / comportamiento / modelo |
|---|---|
| Empleado–Usuario | `(0,1)` a `(0,1)`, vínculo único opcional; permite empleado sin cuenta y cuenta heredada sin vínculo. V22 coincide. |
| Usuario–Roles | N:M mediante puente; nuevos usuarios reciben rol por API, SQL puede conservar cuentas sin rol. No convertir cargo en permiso. |
| Cliente–Auxilio | `(0,N)`; dos incidentes del mismo cliente se admiten. V22 coincide, no hay unicidad vitalicia por cliente. |
| Recepción–Diagnóstico | Recepción `(0,1)` diagnósticos; el diagnóstico tiene una recepción. Directo mantiene cero, técnico permite pendiente sin fila de diagnóstico. |
| Recepción–Cotización | Una recepción puede tener cotizaciones/revisiones, sin diagnóstico en servicio directo. No se fabrican diagnósticos. |
| Cotización–Orden | Orden desde una cotización aprobada; una cotización como máximo origina una orden. |
| Cotización–Detalle | API exige uno o más conceptos; la FK SQL por sí sola no impone que el padre tenga hijos. |
| Orden–Pagos/Entrega | Varios pagos; entrega única con saldo completo y condiciones técnicas verificadas en Go. |
| Diagnóstico–MongoDB | `(0,N)` documentos; identidad decimal del diagnóstico, sin FK entre motores. La API comprueba autorización PostgreSQL. |
| Auxilio–Cliente/vehículo/técnico/unidad | El API nuevo exige vínculos; algunas FK físicas son opcionales para compatibilidad histórica. No se cambió su nulabilidad. |

Se consultó la especificación `RELACIONES_CORREGIDAS.json` del V22 corregido en
la rama documental, sin alterar esos archivos. TIPOS_DOCUMENTO y las políticas
de tres roles deben incorporarse en la próxima revisión académica, no mediante
sustitución silenciosa de V19–V22. No se propone una tabla nueva.

## Respaldos y preparación de limpieza

Se ejecutó `backend/scripts/backup-audit.ps1`. Copias privadas recuperables:

`backend/.cache/respaldo-auditoria-20261102193213/`

- `postgres.dump`: pg_dump custom de **elbravodeljac_dev**.
- `mongo.archive.gz`: mongodump de **elbravodeljac_tecnico_dev**.
- `SHA256.json`, inventarios, logs y `resultado.json` con `verified=true`.

Restauración real en el proyecto QA, en bases nuevas que se conservan:
`jac_restore_20261102193213` y `elbravodeljac_restore_20261102193213`.
Coinciden filas y huellas de las **28 tablas**, restricciones e índices SQL,
documentos/chunks GridFS, índices y validadores MongoDB. Cuatro CHECK aparecen
con otra escritura de casts de arrays de literales al restaurar: se verificó
la equivalencia de esa conversión concreta; no se ignoraron diferencias
arbitrarias. Las FK/PK/UNIQUE se compararon completas.

La comparación posterior confirmó que las 28 tablas de desarrollo seguían
idénticas al respaldo. La copia restaurada también quedó idéntica tras la
simulación con rollback. Los hashes incluyen las filas completas de usuarios
(y por tanto sus hashes de contraseña), empleados, roles y puentes, sin
publicar esos datos personales.

### Inventario propuesto para borrar, todavía no ejecutado

| Tabla | Filas del respaldo |
|---|---:|
| entregas | 14 |
| pagos | 15 |
| historial_estados_ot | 68 |
| pruebas_ot | 21 |
| avances_ot | 46 |
| asignaciones_ot | 36 |
| ordenes_trabajo | 22 |
| detalle_cotizacion | 47 |
| cotizaciones | 25 |
| diagnosticos | 21 |
| auxilios | 6 |
| recepciones | 30 |
| historial_propietarios | 36 |
| vehiculos | 36 |
| contactos_cliente | 1 |
| clientes | 27 |
| **Total** | **451** |

MongoDB: 6 resultados de escáner, 6 documentos de evidencia, 6 archivos GridFS y
6 chunks asociados. El manifiesto vincula cada archivo con su evidencia,
diagnóstico y `managedBy=elbravodeljac`. No hay archivos seleccionados por
nombre o por fecha solamente. En esta copia no hubo elementos ajenos o sin
pertenencia comprobada; cualquier caso así en un inventario nuevo se detiene
para revisión. No se borró ninguno de esos documentos.

### Información que se conserva

| Tabla/configuración | Filas / decisión |
|---|---|
| usuarios | 35; conservar completos, hashes y estado |
| empleados | 1; conservar vínculo con su cuenta |
| roles | 7 almacenados, incluidos legados; conservar |
| permisos | 40; conservar |
| usuario_roles | 42 asignaciones; conservar |
| rol_permisos | 163 asignaciones; conservar |
| tipos_documento | 4; conservar |
| servicios | 14; conservar y revisar ejemplos manualmente |
| repuestos | 7; conservar stock y precios, sin reinicializarlos |
| proveedores | 1; conservar hasta clasificación explícita |
| repuesto_proveedor | 0; conservar estructura |
| unidades_auxilio | 6; conservar hasta clasificación explícita |
| Estructura / configuración | Tablas, 50 FK, índices, vistas (incluida marcas), funciones, migraciones, .env privados y volúmenes intactos |

No se deduce que un servicio/repuesto/proveedor/unidad sea descartable por su
nombre. Las cuentas y configuración no dependen de esa clasificación.
Las secuencias no se reiniciarán: un nuevo registro no tiene por qué empezar en 1.

La revisión de los valores encontró marcadores inequívocos de las ejecuciones
anteriores en los 14 servicios (verificación Docker / sugerencia de revisión),
los 7 repuestos (verificación Docker) y las 6 unidades (códigos REV con sufijo
de prueba): **27 entradas candidatas de ejemplo**. Se documentan, pero quedan
excluidas del alcance de 451 filas por seguridad; retirarlas requiere aprobar
ese alcance adicional. El único proveedor no se puede clasificar como ficticio
con certeza. La revisión detallada está en el archivo privado
`revision-catalogos.json` del respaldo. No se copiaron nombres personales al informe.

### Procedimiento preparado y límites

1. Pedir aprobación para **este alcance inventariado**. Si cambia el inventario,
   volver a presentar la diferencia; no ampliar el borrado automáticamente.
2. Con aprobación, pausar solo la API propia durante mantenimiento, confirmar
   nombres de motores/proyecto y generar respaldos nuevos verificados sin escrituras
   concurrentes. No parar ni borrar los volúmenes de las bases.
3. Capturar el manifiesto MongoDB y los IDs de diagnósticos **antes** de limpiar
   PostgreSQL, más huellas de seguridad/catálogos y estructura.
4. PostgreSQL: una transacción con DELETE explícitos en el orden de la tabla
   anterior, sin CASCADE ni reinicios de IDs. Comparar dentro de ella las 12
   tablas protegidas. Si falla cualquier restricción, rollback completo.
5. Tras confirmar PostgreSQL, borrar únicamente documentos/archivos MongoDB
   enumerados en el manifiesto: archivos por GridFS ID comprobado, metadatos de
   evidencia y escáneres seleccionados. No ejecutar borrados globales de buckets.
   Conservar el manifiesto para reintento idempotente. No hay transacción
   distribuida entre motores: si falla MongoDB, mantener mantenimiento hasta
   completar el manifiesto o restaurar ambos respaldos coherentes.
6. Verificar conteos cero en operaciones, datos protegidos y hashes intactos,
   referencias/índices, sesiones, catálogos y comportamiento vacío. La prueba de
   nueva alta sin ensuciar la base se hace en otra copia/rollback autorizado.
   Solo entonces reabrir la API, sin cargar fixtures o datos demo al arrancar.

Lo ejecutado en esta auditoría fue únicamente:

- `docs/backend/limpieza/SIMULAR_POSTGRESQL.sql`: permite solo una copia
  `jac_restore_…`, prueba el orden de DELETE, comprueba huellas protegidas y
  **siempre termina con ROLLBACK**. Se comprobó otra vez que no persistió borrado.
- `docs/backend/limpieza/MANIFIESTO_MONGODB.js`: inventario de solo lectura en
  copia restaurada, con `dryRun=true`, `writes=0`.

No se entregó un comando de borrado automático de desarrollo. La fase 6 y la
ejecución definitiva deben esperar aprobación explícita.

## Archivos afectados

- Backend: `workshop_writes.go`, `workshop_reads.go`, `technical.go`.
- Frontend: `real-workshop.js`, `views/operations.js`.
- QA: `catalog_roles_test.go`, nuevo `taller_audit_test.go`,
  nuevo `checks/taller-browser.mjs`, `test-audit.ps1`,
  `compose.auditoria-taller.yaml`.
- Respaldo y preparación: `backup-audit.ps1`, los dos archivos de
  `docs/backend/limpieza/` y este informe.

No cambian tablas, FK, migraciones, JWT, matrices de permisos ni diseño visual.
Las tres correcciones reproducibles de severidad ALTA quedaron cubiertas por
regresión. No se reprodujeron fallos CRÍTICOS pendientes en el alcance probado;
esto no equivale a una certificación de seguridad universal ni sustituye la
revisión de cuentas heredadas antes del uso empresarial.

Las correcciones se incorporaron al contenedor local de API/frontend sin
reiniciar PostgreSQL/MongoDB ni ejecutar migraciones. `/readyz` confirmó estado
`ready` con MongoDB configurado. Al cierre se volvieron a comparar las 28 tablas
y las cuatro colecciones de desarrollo con el respaldo: permanecieron idénticas.
Resultado privado: `conservacion-final.json` dentro de la carpeta de respaldo.
Después del reinicio de Go es necesario volver a iniciar sesión.
