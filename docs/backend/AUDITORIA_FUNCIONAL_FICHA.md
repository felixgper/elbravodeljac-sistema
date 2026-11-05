# Auditoría funcional de ficha técnica — 3 de noviembre de 2026

## Resultado y alcance

Las comprobaciones funcionales ejecutadas pasaron. No se detectó un error
reproducible que exigiera cambiar el código de producción, el diseño aprobado,
el esquema o las reglas de negocio. No se realizó commit, push ni merge.

Se probaron el código Go actual del repositorio y el frontend real, servido
por Go en servidores HTTP efímeros de pruebas. No se sustituyeron endpoints por
mocks para el recorrido integrado. Las seis suites JavaScript también incluyen
comprobaciones unitarias y del modo demo; la evidencia persistida procede de las
pruebas HTTP/Chrome con bases aisladas, no de esos datos demo.

**Resultado global:** 45 pruebas Go principales y 52 subcasos aprobados en siete
paquetes; cero fallos y una prueba opcional omitida. Se ejecutó además el caso
específico de esta auditoría con Chrome para comprobar guardar, F5 y reabrir
el control técnico. `go vet ./...` y las seis suites JavaScript terminaron con
código 0. Se verificaron 50 filas DB–API–PDF.

No se reconstruyó ni reinició la API de desarrollo de localhost:8080. Los
servidores Go de auditoría se cerraron al terminar sus pruebas.

## Aislamiento y preservación

- Docker Desktop: servidor 29.7.2. Se inventariaron los proyectos antes de actuar.
- Proyecto nuevo y exclusivo: `elbravodeljac-ficha-funcional-qa`.
- PostgreSQL 18.6: `127.0.0.1:25439`, base `elbravodeljac_test`.
- MongoDB 9.0: `127.0.0.1:27039`, base `elbravodeljac_tecnico_test`.
- Se inicializó exclusivamente ese PostgreSQL nuevo mediante la configuración
  existente de `compose.test.yaml`. No se editaron migraciones ni modelos.
- Se mantienen los contenedores QA activos para revisión, con almacenamiento
  tmpfs exclusivo. No se detuvieron, recrearon ni borraron contenedores o
  volúmenes anteriores. Los servidores HTTP del test sí son efímeros.
- La prueba nueva exige explícitamente el nombre de base anterior y el puerto
  25439; no puede ejecutarse contra desarrollo usando su configuración normal.
- No se conectó al PostgreSQL original `ElBravoDelJac` del puerto 5432 del host.
- Las cuentas iniciales de QA se crean como fixtures independientes, vinculadas
  a empleados y a los roles canónicos Administrador/Mecánico. No se modifican
  sus permisos para hacer pasar la prueba nueva. Después del bootstrap, las
  operaciones del taller del caso específico se realizan por HTTP autenticado.

Se calcularon huellas de todas las tablas públicas de PostgreSQL de desarrollo
y del contenido de sus colecciones MongoDB antes y después. **Coinciden.**
También coinciden los identificadores de los contenedores preexistentes.
Las huellas no incluyen ni publican registros o credenciales en este informe.

Los 17 archivos inicialmente pendientes se respaldaron con sus rutas originales
en `backend/.cache/auditoria-ficha-funcional/antes/`; se conservaron sus contenidos.
La configuración privada `.env.ficha-audit.local` está excluida de Git. Las
contraseñas de prueba se generan aleatoriamente y no se incluyen en la entrega.

## Recorrido integrado y resultados

| Comprobación | Resultado | Evidencia o condición verificada |
|---|---|---|
| Login independiente de administradora y dos mecánicos | Aprobada | Cuentas y empleados distintos, roles canónicos, cookie JWT HttpOnly y CSRF |
| Cliente y vehículo | Aprobada | Creación HTTP y lecturas SQL/API concordantes |
| Recepción por falla de frenos | Aprobada | Kilometraje y cliente de atención persistidos |
| Asignación de mecánico | Aprobada | HTTP y FK/actor de recepción comprobados |
| Diagnóstico por el propio mecánico | Aprobada | Login propio, diagnóstico final y autor SQL correcto; otro mecánico no puede consultarlo |
| Cotización pendiente y aprobación | Aprobada | La pendiente no crea orden (422); aprobación explícita con cliente/método de prueba |
| Orden y ejecución | Aprobada | Se crea desde la cotización aprobada, con mecánico asignado |
| Avances y pruebas | Aprobada | Notas y dos pruebas persistidas; transiciones HTTP válidas |
| Inspecciones | Aprobada | Bien, Atención, Urgente y Sin evaluar guardados; recarga API y F5 en Chrome |
| Mantenimiento | Aprobada | Solo pastillas semicerámicas seleccionadas; aceite solo cotizado sin marca |
| Actualización de control | Aprobada | Conserva selecciones; versión obsoleta devuelve 409 |
| Permisos de control | Aprobada | Mecánico ajeno: GET 404 / PUT 403; orden cerrada: PUT 409 |
| Pagos y CSRF | Aprobada | Mecánico sin permiso: 403; escritura sin CSRF: 403; exceso de pago: 422 |
| Pago parcial y entrega | Aprobada | Sin pago o con pago parcial la entrega devuelve 422 |
| Pago completo y entrega | Aprobada | Pagos 50 + 210 = 260; entrega y archivo PDF/instantánea persistidos |
| Generación final anticipada | Aprobada | Sin entrega válida no se descarga ficha final (404) |
| Acceso al PDF | Aprobada | Administradora autorizada: 200; mecánico sin permiso: 403; anónimo: 401 |
| Historial y descarga posterior | Aprobada | Historial consultable, descarga repetida y bytes archivados estables |
| Cambios posteriores de cliente/vehículo | Aprobada | Solo se editan los registros QA; PDF, instantánea y archivo previo quedan idénticos |
| Otros módulos y MongoDB/GridFS | Aprobada | Suites existentes con autorización, deduplicación, archivos y fallos de MongoDB |
| Copia restaurada vacía (`TestCleanRestoredCopy`) | No ejecutada | Falta `INTEGRATION_CLEAN_DATABASE_URL`; exige otra base `jac_restore_YYYYMMDDhhmmss` en puerto 25435. No se restauró ninguna base para esta auditoría |

La prueba omitida no pertenece al recorrido de esta ficha. No se informa como
aprobada ni se modifica su restricción para ejecutarla contra otra base.

## Caso detallado: trazabilidad económica y casillas

El caso de la matriz corresponde a la orden QA **OT-22**, cliente y vehículo
ficticios. La cotización aprobada suma S/ 310:

| Concepto | Cantidad | Precio unitario | Descuento unitario | Importe | Ejecución |
|---|---:|---:|---:|---:|---|
| Mantenimiento de frenos | 1 | 100 | 10 | 90 | Realizado |
| Pastillas delanteras semicerámicas | 1 | 170 | 0 | 170 | Realizado |
| Aceite cotizado | 1 | 50 | 0 | 50 | No realizado, con motivo registrado |

La administradora registra mediante la API la conciliación permitida por el
sistema: se pospone el aceite y se cobran S/ 260. Los precios autorizados no se
alteran. El PDF muestra únicamente las dos líneas realizadas; el marcador del
aceite no aparece en su tabla. La recomendación de revisar alineación aparece
como recomendación del diagnóstico, no como servicio ejecutado.

La matriz compara lecturas SQL directas **anteriores a las ediciones de prueba**,
la instantánea SQL archivada, las respuestas API y el PDF realmente descargado.
Incluye cliente/documento, vehículo/placa, kilometraje, mecánico, diagnóstico,
cantidades, precios, descuentos, importes, avances, pruebas, observaciones,
neumáticos y próximo mantenimiento. Posteriormente se modifica el cliente y el
vehículo QA para comprobar precisamente que la ficha histórica no cambia.

Comprobación del dibujo PDF mediante sus operaciones vectoriales/coordenadas:

- 118 casillas: 63 de inspección, 41 de mantenimiento y 14 de otros servicios.
- Cuatro marcas: aceite de motor Bien; líquido Aceite Atención; presión de aire
  Urgente; pastillas semicerámicas utilizadas.
- Aceite de caja Sin evaluar y todos los controles ausentes: casillas vacías.
- Cero marcas de mantenimiento de aceite y cero marcas de otros servicios.
- Ningún componente tiene dos estados marcados.

Los PDF de prueba no contienen firmas inventadas. Se conservan los espacios
de conformidad.

## Presentación

Se generaron y renderizaron 10 variantes en memoria (31 páginas): ficha básica,
varios trabajos, controles completos/vacíos, multiselección y notas/descripciones
extensas. El caso extremo contiene 45 conceptos y diagnóstico extenso, produce
12 páginas y conserva los marcadores finales de cliente, diagnóstico,
observaciones y último concepto. Se verificaron márgenes A4, numeración,
presencia del logo 19 y contenido final; se revisaron visualmente las vistas.

No se encontraron recortes, superposiciones ni páginas completamente vacías
en los ejemplos revisados. Poppler avisa de fuentes de sustitución Symbol y
ArialUnicode no disponibles; los textos Helvetica, acentos y casillas vectoriales
se visualizaron correctamente. No se cambió la plantilla por esos avisos.

Dos documentos integrados adicionales se entregan en
`output/pdf/auditoria-ficha/`:

1. `ficha-normal-navegador.pdf`: 2 páginas, descargada desde Chrome durante el
   recorrido real con cuentas separadas; controles sin evaluar visibles vacíos.
2. `ficha-auditada-conciliacion.pdf`: 3 páginas, caso detallado de la matriz,
   con inspecciones, repuestos realizados y conciliación económica. La nota de
   conciliación y observaciones justifican la continuación; no se encogió texto
   ni se eliminó contenido para forzar dos páginas.

Sus PNG numerados son las capturas de cada página. La galería `index.html`
reúne documentos, capturas y matriz.

## Hallazgos y correcciones

**Fallos confirmados de producción:** ninguno en los casos ejecutados.
**Correcciones de producción realizadas:** ninguna. La presentación aprobada
y todos los cambios funcionales anteriores permanecen intactos.

Observaciones informativas del modelo existente, sin cambiar reglas:

- Mantenimiento es multiselección: la interfaz usa checkboxes y Go valida IDs
  y duplicados. Dos viscosidades distintas son aceptadas por el modelo actual;
  no existe una regla de exclusividad que esta auditoría deba inventar. El caso
  de multiselección se prueba por separado, no se presenta como el trabajo de
  frenos realizado.
- `otherServices` registra servicios marcados por el mecánico y es independiente
  de `execution`/`performedItems`. No hay equivalencia automática con líneas
  cobrables. El PDF refleja lo registrado, pero no decide por sí mismo que una
  recomendación fue ejecutada. La ejecución económica sí se contrasta con las
  líneas autorizadas y con la conciliación de conceptos no realizados.
- La descarga actual usa la instantánea congelada, preservando los datos y el
  PDF archivado. Una futura plantilla puede volver a representarla visualmente;
  no se promete identidad gráfica entre versiones futuras de la plantilla.
  Dentro de la versión auditada, editar registros y descargar otra vez conservó
  exactamente los mismos bytes.

Durante la preparación se corrigieron fixtures del nuevo test (tipo/origen de
repuesto y contrato PATCH de cliente/vehículo). No eran fallos de la aplicación:
las peticiones inválidas fueron rechazadas correctamente.

## Archivos nuevos de esta auditoría

- `backend/integration/sheet_functional_audit_test.go`: flujo HTTP/SQL, controles,
  permisos, conservación histórica y captura de evidencia.
- `checks/sheet-functional-browser.mjs`: guardar desde el formulario real del
  mecánico, F5, reabrir y comprobar estados/materiales/ejecución en Chrome.
- `docs/backend/AUDITORIA_FUNCIONAL_FICHA.md`: este informe.
- `output/pdf/auditoria-ficha/`: dos PDF, cinco PNG, matriz, evidencia JSON
  ficticia y galería. No son documentos históricos de la empresa.

Los logs, verificadores auxiliares, respaldo y huellas privadas se conservan
en `backend/.cache/auditoria-ficha-funcional/`, excluido de Git.

## Comandos y revisión del commit

Se ejecutaron:

```text
node checks/run.mjs
go -C backend test ./... -v -count=1
go -C backend test ./integration -run '^TestSheetFunctionalAudit$' -v -count=1
go -C backend test ./internal/platform/technicalsheet -run 'TestBusinessDocumentIdentity|TestSheetCheckboxExamples' -v -count=1
go -C backend vet ./...
git diff --check
```

Para Go integrado se utilizaron variables de proceso con la conexión QA privada,
`INTEGRATION_MONGODB_DATABASE=elbravodeljac_tecnico_test`, `FRONTEND_AUDIT=1`,
`FRONTEND_E2E=1`, `SHEET_FUNCTIONAL_AUDIT=1`; el caso específico de F5 se ejecutó
también con `SHEET_BROWSER_AUDIT=1` y la ruta local de Playwright.

La auditoría funcional ejecutada permite revisar el commit de estos cambios
con la evidencia indicada y la excepción explícita de copia restaurada. No se
ha realizado ninguna operación de Git que publique o confirme los archivos.
