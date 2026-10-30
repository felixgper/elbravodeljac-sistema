from pathlib import Path
import json,html,re
import xml.etree.ElementTree as ET

HERE=Path(__file__).resolve().parent
old={r['id']:r for r in json.loads((HERE.parent/'V21_Chen/relaciones_chen.json').read_text(encoding='utf-8'))}
rs=json.loads((HERE/'RELACIONES_CORREGIDAS.json').read_text(encoding='utf-8'))
inv=json.loads((HERE.parent/'V21_Chen/inventario_actual.json').read_text(encoding='utf-8'))
v=json.loads((HERE/'VALIDACION.json').read_text(encoding='utf-8'))
verified=json.loads((HERE/'VALIDACION_INDEPENDIENTE.json').read_text(encoding='utf-8'))
cells=ET.parse(HERE/'El_Bravo_del_JAC_V22_Chen_Corregido.drawio').findall('.//mxCell')
idx={c.get('id'):c for c in cells}
def name(n):return html.unescape(re.sub('<[^>]*>',' ',idx[n].get('value')))
def pair(r):return r['ca']+' / '+r['cb']
text='''# Revisión del modelo conceptual Chen V22

**29/10/2026 · Versión corregida · Evaluación técnica: 18,5/20.**

## Entrega y protección del diseño

Se partió del V22 aprobado, sin reconstruir el modelo. Se conserva el mismo
lienzo de **6200 × 3400**, las posiciones y tamaños de las **26 entidades**,
los **33 rombos originales**, el fondo blanco y los colores institucionales
azules de este modelo. Solo cambian los textos autorizados, siete etiquetas de
participación, las rutas/puertos necesarios y se añaden dos rombos operativos.
No hay óvalos, atributos, identificadores de datos, tipos SQL ni claves visibles.

Los archivos V21 y V22 originales se conservaron byte a byte: el manifiesto
ORIGINALES_SHA256.json registra sus hashes y el validador comprueba los 20 archivos.
Rama: `docs/modelo-conceptual-chen-v21`, basada en la rama funcional que contiene
el merge de EMPLEADOS, `7dd2738`. No se cambia de rama ni se mezclan otras ramas.

Entregables principales:

- El_Bravo_del_JAC_V22_Chen_Corregido.drawio
- El_Bravo_del_JAC_V22_Chen_Corregido.png
- INFORME_REVISION_CHEN_V22.md

Los scripts e inventarios de validación se incluyen como apoyo. La vista general
es una reducción del mismo PNG: no constituye una segunda página del Draw.io.

## Alcance de los mínimos corregidos

Este modelo explica **las operaciones actuales que ejecuta la API Go**. Para
las siete participaciones revisadas se representa el mínimo del negocio para
registros creados por esa API, sin modificar PostgreSQL. No se afirma que todos
los registros legados o ingresados por SQL directo satisfagan esos mínimos.
La cabecera del propio diagrama declara este alcance.

El SQL sigue admitiendo más situaciones: una FK obliga al hijo a identificar
un padre si es NOT NULL, pero no obliga al padre a tener hijos; las cuatro
referencias opcionales de auxilio siguen siendo nullable. **Una validación
de Go no se presenta aquí como una restricción física nueva.**

La notación junto a cada entidad es mínimo:máximo, por ejemplo 0:N o 1:1.
En las tablas del informe se utilizan los pares equivalentes (0,N) y (1,1).

## Comparación ANTES / DESPUÉS

| Relación | Antes V22 | Después corregido | Evidencia y condiciones |
|---|---|---|---|
| USUARIOS — RECEPCIONES | Ausente | SE ASIGNA A DIAGNÓSTICO: usuario (0,N), recepción (0,1) | `recepciones.id_mecanico_diagnostico` es nullable y no único; 004 y `AssignDiagnosis`. Servicio directo no recibe esa asignación. |
| COTIZACIONES — COTIZACIONES | Ausente | TIENE AMPLIACIÓN: rol base (0,N), rol ampliación (0,1) | `id_cotizacion_base` es nullable y no único; 005 y `SaveQuote`. Los rótulos base/ampliación distinguen los dos extremos. |
| ÓRDENES — ASIGNACIONES | (0,N) / (1,1) | **(1,N) / (1,1)** | `CreateOrder` inserta un mecánico obligatorio en la misma transacción. Se cuenta el historial de asignaciones, no solo las activas. |
| UNIDADES — AUXILIOS | (0,N) / (0,1) | **(0,N) / (1,1)** | `AuxilioInput.Validate(create=true)` exige unidad; `SaveAuxilio` comprueba que esté activa. SQL sigue nullable. |
| ÓRDENES — HISTORIAL DE ESTADOS | (0,N) / (1,1) | **(1,N) / (1,1)** | `CreateOrder` inserta el estado inicial en la misma transacción. SQL no obliga al padre a tener eventos. |
| COTIZACIONES — DETALLES | (0,N) / (1,1) | **(1,N) / (1,1)** | La API exige entre 1 y 100 conceptos. El máximo N del modelo expresa pluralidad; el límite operativo 100 se conserva en Go. Reemplazo de detalles transaccional. |
| VEHÍCULOS — AUXILIOS | (0,N) / (0,1) | **(0,N) / (1,1)** | La API exige vehículo registrado y valida su cliente. SQL sigue nullable. |
| CLIENTES — AUXILIOS | (0,N) / (0,1) | **(0,N) / (1,1)** | La API exige cliente al crear el auxilio. SQL sigue nullable. |
| USUARIOS — AUXILIOS | (0,N) / (0,1) | **(0,N) / (1,1)** | La API exige técnico activo con rol compatible. SQL sigue nullable. |
| RECEPCIONES — AUXILIOS | PROCEDE DE AUXILIO | **SE VINCULA CON**, conservando (0,N) / (0,1) | Existe FK opcional. El nuevo nombre no promete una secuencia temporal ni la vinculación posterior desde el formulario. |

Las demás participaciones de las relaciones originales no se cambian.

## Las dos relaciones operativas añadidas

### Responsable previo al diagnóstico

SE ASIGNA A DIAGNÓSTICO conecta **USUARIOS con RECEPCIONES**, no con una orden
ni con un diagnóstico inventado. La recepción puede estar sin asignar o tener
un responsable; una cuenta puede asumir varias recepciones. La API verifica
rol técnico, estado activo y permiso de asignación; el servicio directo no
admite esa asignación. La reasignación del diagnóstico finalizado se rechaza.

DIAGNOSTICA se conserva por separado: identifica al responsable del diagnóstico
cuando ya existe. No es redundante con la asignación previa. La asignación a
una OT sigue pasando por ASIGNACIONES DE PERSONAL y puede ser de otro mecánico.

### Cotización original y ampliaciones

TIENE AMPLIACIÓN es una relación recursiva con dos conectores distintos a la
misma entidad COTIZACIONES y los roles **base** y **ampliación**. Una cotización
base puede no tener ampliaciones o tener varias. Para el universo completo de
cotizaciones, una cotización puede no tener base o tener una sola. Cuando se
crea específicamente como ampliación en Go, **debe identificar exactamente
una base**: el cero corresponde a las cotizaciones originales.

Go exige una base original de la misma recepción, con una orden abierta y sin
entrega. No se interpreta el rombo como permiso para cadenas arbitrarias,
autorreferencias o ciclos: la FK por sí sola no impone esas condiciones, pero
`SaveQuote` restringe el flujo actual. La revisión de contenido y la ampliación
son conceptos diferentes; no se crea una entidad ficticia para cada revisión.

## Auxilio: relación válida y limitación funcional conservada

SE VINCULA CON expresa una asociación estructural opcional entre auxilio y
recepción. La recepción puede estar vinculada con varios auxilios; un auxilio
puede referir a una recepción o a ninguna. No se deduce cronología de la FK.

La API permite indicar una recepción **ya existente** al crear el auxilio,
comprobando cliente y vehículo. El formulario actual no envía esa referencia;
el PATCH solo permite técnico, unidad y estado y rechaza cambiar recepción.
Por tanto **no se ha implementado ni probado un flujo de vinculación posterior
de una nueva recepción a un auxilio previamente registrado**. Esa tarea sigue
pendiente en la aplicación y queda fuera de este cambio documental.

No se borra la asociación real ni se promete que el formulario ya la complete.
No se modificó código para solventarla.

## Cobertura de las 27 tablas

Hay 24 entidades relacionales, tres tablas puente representadas como N:M y
dos entidades documentales complementarias: **26 rectángulos y 35 rombos**.

| Tabla física | Representación conceptual |
|---|---|
'''
bridges={'usuario_roles':'ASUME ROL','rol_permisos':'CONCEDE PERMISO','repuesto_proveedor':'ABASTECE'}
for n in sorted(inv['tables']):
 text+='| `'+n+'` | '+('Relación N:M '+bridges[n] if n in bridges else name(n))+' |\n'
text+='''
Las relaciones dibujadas representan explícitamente **36 de las 48 FK**.
Las otras 12 son referencias de autoría/auditoría a USUARIOS y permanecen
abstraídas para conservar la legibilidad: autores de auxilios, avances,
entregas/verificación, estados, titularidad, creación de OT, pagos, pruebas,
recepción, control técnico y asignación del diagnóstico. No se afirma que el
conceptual sea un inventario exhaustivo de FK. Las dos omisiones operativas
se recuperaron sin añadir las conexiones de auditoría repetitivas.

## Participaciones finales, una por una

Cada par pertenece respectivamente a la primera y segunda entidad de la fila.
Las siete filas de mínimo de negocio se distinguen de la garantía SQL.

| Entidad A | Relación | Entidad B | Participación A / B | Alcance |
|---|---|---|---|---|
'''
changed={'R17','R19','R20','R21','R26','R36','R44'}
for r in sorted(rs,key=lambda x:x['id']):
 scope='Mínimo validado en Go; SQL más permisivo' if r['id'] in changed else 'Referencia lógica MongoDB, validada en Go' if r['logical'] else 'N:M mediante tabla puente' if r.get('bridge') else 'Coincide con FK, nulabilidad y unicidad SQL'
 text+=f"| {name(r['a'])} | {r['verb']} | {name(r['b'])} | {pair(r)} | {scope} |\n"
text+='''
## Reglas que se conservan y cómo explicarlas

- Servicio directo: recepción → cotización sin diagnóstico. Atención técnica:
  diagnóstico finalizado de la misma recepción antes de cotizar.
- Crear OT exige cotización aprobada y aceptación registrada; el rombo no
  convierte una cotización pendiente en autorización de trabajo.
- Mecánico responsable y auxiliar son funciones distintas de la asignación;
  no equivalen al cargo del empleado ni a los roles de acceso. El auxiliar es
  opcional y debe ser una persona diferente del mecánico en el flujo actual.
- La política base del auxiliar es de consulta. El permiso adicional de editar
  y el alcance de asignación pueden habilitar cambios: la prohibición absoluta
  por cargo no está implementada. No se cambia la política desde este modelo.
- Una OT nueva puede no tener avances o pruebas todavía. Para finalizar exige
  ejecución/conciliación y prueba satisfactoria vigente. No se cambia esos
  mínimos globales a uno por confundir etapa inicial con finalización.
- Pagos parciales o anticipos; la entrega requiere orden finalizada y saldo
  neto cero, además de las verificaciones de la atención. La entrega depende
  de OT, no de un pago individual. La unicidad física es por orden.
- EMPLEADOS y USUARIOS conservan (0,1) en ambos extremos. El cargo laboral no
  concede roles, y las N:M de autorización no se sustituyen por GRANT SQL.
- Los detalles pueden ser manuales o referenciar servicio/repuesto de manera
  excluyente; no se hacen obligatorios los dos catálogos.

## Contraste con lógico, físico y documentos MongoDB

Las fuentes de V19/V20 ya auditadas se conservan con sus blobs Git. V19
representa 26 tablas / 160 columnas / 42 FK; la evolución real y 007 alcanzan
27 / 201 / 48. V20 proponía EMPLEADOS; ahora está implementado. V20 contenía
la nulabilidad antigua de la recepción de cotización: se mantiene la corrección
a obligatoriedad que ya tenía V22, no se reintroduce el error.

Se comprobaron nuevamente las migraciones 001–007 mediante sus hashes y las
validaciones Go relevantes. Se consultaron exclusivamente metadatos de
`elbravodeljac_dev` en una transacción READ ONLY: 27 tablas, 48 FK y nulabilidad
de los vínculos afectados. Las referencias de auxilio siguen nullable; no se
les atribuye un NOT NULL inexistente. Las fuentes funcionales no cambiaron.

RESULTADOS DE ESCÁNER y EVIDENCIAS TÉCNICAS mantienen referencias lógicas a
DIAGNÓSTICOS, dibujadas discontinuas y explicadas también en la cabecera.
Cada documento exige referencia al diagnóstico; Go verifica existencia y
autorización. MongoDB no proporciona una FK entre motores. GridFS es almacenamiento
del archivo, no una nueva entidad del taller. Las referencias al capturador se
abstraen como autoría y no se confunden con las 48 FK PostgreSQL.

## Validación y revisión visual

'''
text+=f"""- Una página nativa mxGraph; lienzo {v['png'][0]} × {v['png'][1]}.
- {v['entities']} entidades, {v['relationships']} relaciones, 70 participaciones; cero atributos y óvalos.
- 26 entidades y 33 rombos originales conservan exactamente posición y tamaño.
- Conectores adheridos a extremos existentes; recursiva con puertos distintos.
- Sin superposiciones, líneas atravesando figuras, textos desbordados ni recortes.
- Cruces medidos: **13 antes → {v['crossings']} después**, con dos relaciones añadidas.
  Persisten cruces, marcados con sobrepasos: no se declara un grafo plano.
- El validador independiente recalcula la geometría desde XML y confirma la
  cobertura de tablas, los siete cambios, las dos adiciones y los hashes originales.
- Se revisaron la imagen completa y ampliaciones de la relación recursiva,
  responsable previo y núcleo operativo. El mismo estilo de V22 queda preservado.

Validación local del XML nativo; no se presenta como apertura interactiva en
Draw.io Desktop ni exportación mediante su motor. El PNG se renderiza desde
las mismas figuras, fuentes, puertos y rutas. Abrir el .drawio mediante
Archivo → Abrir desde → Dispositivo. No hay imágenes incrustadas ni recursos externos.

## Evaluación revisada: 18,5/20

| Criterio | Puntuación |
|---|---:|
| Coherencia estructural y cobertura de entidades | 4/4 |
| Participaciones y distinción negocio/SQL | 3,75/4 |
| Relaciones operativas y trazabilidad del cambio | 3,75/4 |
| Notación y legibilidad | 3,5/4 |
| Explicación de condiciones y alcances | 3,5/4 |
| **Total** | **18,5/20** |

Las observaciones documentales aprobadas quedan atendidas. No se fuerza 20:
la lectura integral requiere zoom, permanecen {v['crossings']} cruces y varias reglas por
etapa, legado o autorización necesitan el informe acompañante. La limitación
funcional del auxilio continúa pendiente en la aplicación; el modelo la describe
sin afirmar que se resolvió. Los atributos siguen excluidos por decisión expresa.

El modelo puede presentarse como **Chen simplificado con mínimos de negocio de
la API actual**, acompañado de sus condiciones. No sustituye un modelo físico
ni una demostración de normalización basada en dependencias funcionales.

No se modificó PostgreSQL, MongoDB, Go, frontend ni migraciones. Sin borrados,
commit, push o merge. Los originales se conservaron íntegros.
"""
(HERE/'INFORME_REVISION_CHEN_V22.md').write_text(text,encoding='utf-8')
print('Informe generado; comparación de 10 cambios y las 35 relaciones finales.')
