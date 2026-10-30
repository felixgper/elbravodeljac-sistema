"""Chen didáctico sin atributos. Solo documentación; no usa red ni bases."""
from pathlib import Path
from PIL import Image, ImageDraw, ImageFont
import collections, heapq, html, itertools, json, math, re
import xml.etree.ElementTree as ET

HERE=Path(__file__).resolve().parent
PREV=HERE.parent/'V21_Chen'
BASE='El_Bravo_del_JAC_V22_Chen_Simplificado'
W,H,STEP=6200,3400,10
BLUE,LINE,INK='#3269B8','#6384AD','#193B66'
positions={
 'contactos_cliente':(300,360),'historial_propietarios':(1000,360),
 'auxilios':(1700,360),'unidades_auxilio':(2400,360),
 'resultados_escaner':(3100,360),'evidencias_tecnicas':(3800,360),
 'clientes':(300,1050),'vehiculos':(1000,1050),'recepciones':(1700,1050),
 'diagnosticos':(2400,1050),'cotizaciones':(3100,1050),
 'ordenes_trabajo':(3800,1050),'pagos':(4500,1050),'entregas':(5200,1050),
 'detalle_cotizacion':(3100,1850),'asignaciones_ot':(3800,1850),
 'avances_ot':(4500,1850),'pruebas_ot':(5200,1850),
 'historial_estados_ot':(5800,1850),'proveedores':(1000,2450),
 'repuestos':(1700,2450),'servicios':(2400,2450),
 'empleados':(2400,3050),'usuarios':(3100,3050),
 'roles':(3800,3050),'permisos':(4500,3050),
}
names={n:n.upper().replace('_',' ') for n in positions}
names.update({'historial_propietarios':'HISTORIAL DE PROPIETARIOS',
 'historial_estados_ot':'HISTORIAL DE ESTADOS', 'detalle_cotizacion':'DETALLES DE COTIZACIÓN',
 'asignaciones_ot':'ASIGNACIONES DE PERSONAL','avances_ot':'AVANCES DE REPARACIÓN',
 'pruebas_ot':'PRUEBAS DE REPARACIÓN','ordenes_trabajo':'ÓRDENES DE TRABAJO',
 'contactos_cliente':'CONTACTOS DEL CLIENTE','vehiculos':'VEHÍCULOS',
 'diagnosticos':'DIAGNÓSTICOS','resultados_escaner':'RESULTADOS DE ESCÁNER',
 'evidencias_tecnicas':'EVIDENCIAS TÉCNICAS','auxilios':'AUXILIOS MECÁNICOS'})
placement={
 'POSEE':(650,1050),'INGRESA':(1350,1050),'RECIBE DIAGNÓSTICO':(2050,1050),
 'SUSTENTA':(2750,1050),'AUTORIZA ORDEN':(3450,1050),'RECIBE PAGO':(4150,1050),
 'CULMINA EN ENTREGA':(4850,1320),'DESIGNA CONTACTO':(300,710),
 'TIENE CUENTA':(2750,3050),'ASUME ROL':(3450,3050),'CONCEDE PERMISO':(4150,3050),
 'ABASTECE':(1350,2450),'CONSERVA TITULARIDAD':(1000,710),
 'PROCEDE DE AUXILIO':(1700,710),'TIENE ASIGNACIÓN':(3800,1450),
 'REFERENCIA SERVICIO':(2750,2180),'SE DESPLIEGA EN':(2050,360),
 'CONSERVA ESTADOS':(5550,1450),'CONTIENE CONCEPTO':(3100,1450),
 'DESEMPEÑA FUNCIÓN':(3450,2750),'DOCUMENTA ESCÁNER':(2750,710),
 'FIGURA COMO TITULAR':(650,360),'REGISTRA AVANCE':(4350,1450),
 'REQUIERE AUXILIO':(1350,710),'ORIGINA COTIZACIÓN':(3100,710),
 'REFERENCIA REPUESTO':(2400,1850),'SOLICITA ATENCIÓN':(1000,1350),
 'DIAGNOSTICA':(2400,2750),'INCLUYE PRUEBA':(4950,1450),
 'REÚNE EVIDENCIA':(3450,710),'SOLICITA AUXILIO':(650,710),
 'ATIENDE AUXILIO':(2050,2750),'RECIBE ENTREGA':(4700,360),
}
source=json.loads((PREV/'relaciones_chen.json').read_text(encoding='utf-8'))
relations=[r.copy() for r in source if r['verb'] in placement]
assert len(relations)==33
omitted=[r for r in source if r['verb'] not in placement]
shapes={};edges=[]
def shape(id,kind,label,x,y,w,h):
 shapes[id]=dict(id=id,kind=kind,label=label,x=x,y=y,w=w,h=h)
def box(s,pad=0):return(s['x']-s['w']/2-pad,s['y']-s['h']/2-pad,s['x']+s['w']/2+pad,s['y']+s['h']/2+pad)
for n,(x,y) in positions.items():shape(n,'entity',names[n],x,y,250,94)
for r in relations:
 x,y=placement[r['verb']];r.update(x=x,y=y)
 shape(r['id'],'relationship',r['verb'],x,y,190,116)

# Connections leave each rectangle toward its relationship, with separate ports.
def side_toward(p,q):
 dx,dy=q[0]-p[0],q[1]-p[1]
 return ('R' if dx>0 else 'L') if abs(dx)>=abs(dy) else ('B' if dy>0 else 'T')
inc=collections.defaultdict(list)
for r in relations:
 for end in ['a','b']:
  n=r[end];side=side_toward(positions[n],placement[r['verb']])
  if n=='ordenes_trabajo' and r['verb'] in ['TIENE ASIGNACIÓN','REGISTRA AVANCE','INCLUYE PRUEBA','CONSERVA ESTADOS']:side='B'
  if n=='diagnosticos' and r['logical']:side='T'
  if n=='recepciones' and r['verb']=='ORIGINA COTIZACIÓN':side='T'
  if n=='clientes' and r['verb']=='SOLICITA AUXILIO':side='T'
  inc[n,side].append((r,end))
anchors={};terminals=set()
for (n,side),items in inc.items():
 items.sort(key=lambda t:(t[0]['y'],t[0]['x']))
 x,y=positions[n]
 for i,(r,end) in enumerate(items):
  horizontal=side in ['L','R']
  offset=(i-(len(items)-1)/2)*(30 if horizontal else 70)
  dx,dy={'L':(-1,0),'R':(1,0),'T':(0,-1),'B':(0,1)}[side]
  a=(x+dx*125+(offset if not horizontal else 0),y+dy*47+(offset if horizontal else 0))
  portal=(round((a[0]+dx*110)/STEP)*STEP,round((a[1]+dy*110)/STEP)*STEP)
  if horizontal:cx,cy=a[0]+dx*58,a[1]-14
  else:cx,cy=a[0]+38,a[1]+dy*60
  cid=r['id']+'_'+end+'_card'
  shape(cid,'cardinality',r['c'+end].strip('()').replace(',',':'),cx,cy,56,22)
  anchors[r['id'],end]=(a,portal)
  terminals.add(tuple(round(v/STEP) for v in portal))

blocked=set()
for s in shapes.values():
 b=box(s,8 if s['kind']!='cardinality' else 3)
 for x in range(math.ceil(b[0]/STEP),math.floor(b[2]/STEP)+1):
  for y in range(math.ceil(b[1]/STEP),math.floor(b[3]/STEP)+1):blocked.add((x,y))
goals={}
for r in relations:
 for end in ['a','b']:
  x,y=placement[r['verb']];side=side_toward((x,y),positions[r[end]])
  dx,dy={'L':(-1,0),'R':(1,0),'T':(0,-1),'B':(0,1)}[side]
  # Opposite ends must not attach to the same tip of a diamond.
  if end=='b' and side==goals[r['id'],'a'][2]:
   side={'L':'R','R':'L','T':'B','B':'T'}[side]
   dx,dy={'L':(-1,0),'R':(1,0),'T':(0,-1),'B':(0,1)}[side]
  tip=(x+dx*95,y+dy*58)
  portal=(round((tip[0]+dx*40)/STEP)*STEP,round((tip[1]+dy*40)/STEP)*STEP)
  goals[r['id'],end]=(tip,portal,side)
  terminals.add(tuple(round(v/STEP) for v in portal))

used=set();used_nodes=collections.defaultdict(set)
def seg(a,b):return tuple(sorted([a,b]))
def route(a,b):
 start=tuple(round(v/STEP) for v in a);target=tuple(round(v/STEP) for v in b)
 assert start not in blocked and target not in blocked,(a,b,'blocked endpoint')
 best={(start,0):0};prev={};heap=[(0,0,start,0)]
 while heap:
  _,cost,p,old=heapq.heappop(heap)
  if cost!=best.get((p,old)):continue
  if p==target:final=(p,old);break
  for dx,dy,d in [(1,0,1),(-1,0,1),(0,1,2),(0,-1,2)]:
   q=(p[0]+dx,p[1]+dy)
   if not(8<q[0]<W//STEP-8 and 18<q[1]<H//STEP-12):continue
   if q in blocked or seg(p,q) in used or (q in terminals and q!=target):continue
   h=used_nodes.get(p,set());t=used_nodes.get(q,set())
   if h and (d in h or len(h)>1 or (old and old!=d)):continue
   if t and (d in t or len(t)>1):continue
   nc=cost+10+(14 if old and d!=old else 0)+(160 if t else 0)
   if nc<best.get((q,d),1e15):
    best[q,d]=nc;prev[q,d]=(p,old)
    heapq.heappush(heap,(nc+10*(abs(q[0]-target[0])+abs(q[1]-target[1])),nc,q,d))
 else:raise RuntimeError('Sin ruta '+str((a,b)))
 pp=[];st=final
 while True:
  pp.append(st[0])
  if st==(start,0):break
  st=prev[st]
 pp.reverse()
 for a,b in zip(pp,pp[1:]):
  used.add(seg(a,b));d=1 if a[1]==b[1] else 2
  used_nodes[a].add(d);used_nodes[b].add(d)
 out=[]
 for x,y in pp:
  p=(x*STEP,y*STEP)
  if len(out)>1 and (out[-2][0]==out[-1][0]==p[0] or out[-2][1]==out[-1][1]==p[1]):out[-1]=p
  else:out.append(p)
 return out

# Short local associations first; the longer links use surrounding whitespace.
relations.sort(key=lambda r:math.dist(positions[r['a']],placement[r['verb']])+math.dist(positions[r['b']],placement[r['verb']]))
for r in relations:
 for end in ['a','b']:
  a,p=anchors[r['id'],end];b,q,_=goals[r['id'],end]
  edges.append(dict(id=r['id']+'_'+end,source=r[end],target=r['id'],points=[a]+route(p,q)+[b],dashed=r['logical']))
 print(r['verb'],flush=True)
shape('title','title','EL BRAVO DEL JAC · MODELO CONCEPTUAL CHEN SIMPLIFICADO',3000,60,5900,60)
shape('legend','legend','Solo entidades, relaciones y participaciones · Una sola página · Cada etiqueta indica mínimo:máximo para su entidad',3000,125,5900,40)

# Fit text inside the actual diamond, not merely its enclosing rectangle.
measure=ImageDraw.Draw(Image.new('RGB',(1,1)))
for s in shapes.values():
 kind=s['kind']
 size=24 if kind=='entity' else 20 if kind in ['relationship','cardinality'] else 32 if kind=='title' else 21
 while True:
  ff=ImageFont.truetype('C:/Windows/Fonts/'+('arialbd.ttf' if kind in ['entity','title'] else 'arial.ttf'),size)
  maxwidth=s['w']-(60 if kind=='relationship' else 20)
  lines=[];line=''
  for word in s['label'].split():
   test=(line+' '+word).strip()
   if measure.textlength(test,font=ff)>maxwidth and line:lines.append(line);line=word
   else:line=test
  lines.append(line);lh=size+(2 if kind=='cardinality' else 5)
  fits=len(lines)*lh<=s['h']
  if kind=='relationship':
   y=s['y']-len(lines)*lh/2
   for line in lines:
    x=s['x']-measure.textlength(line,font=ff)/2
    bb=measure.textbbox((x,y),line,font=ff)
    fits= fits and all(abs((px-s['x'])/(s['w']/2))+abs((py-s['y'])/(s['h']/2))<.97 for px,py in itertools.product([bb[0],bb[2]],[bb[1],bb[3]]))
    y+=lh
  if fits:break
  size-=1
  assert size>=15,('Texto sin espacio',s['id'])
 s.update(lines=lines,font_size=size,lineheight=lh)

mx=ET.Element('mxfile',host='app.diagrams.net',type='device')
d=ET.SubElement(mx,'diagram',id='chen-simplificado-unico',name='Chen simplificado · lienzo único')
m=ET.SubElement(d,'mxGraphModel',page='1',pageWidth=str(W),pageHeight=str(H),pageScale='1',grid='1',gridSize='10',background='#FFFFFF')
root=ET.SubElement(m,'root');ET.SubElement(root,'mxCell',id='0');ET.SubElement(root,'mxCell',id='1',parent='0')
for e in edges:
 s,t=shapes[e['source']],shapes[e['target']];a,b=e['points'][0],e['points'][-1]
 style=f'edgeStyle=none;endArrow=none;startArrow=none;rounded=0;strokeColor={LINE};strokeWidth=2;jumpStyle=arc;jumpSize=6;'
 for key,v in [('exitX',(a[0]-s['x'])/s['w']+.5),('exitY',(a[1]-s['y'])/s['h']+.5),('entryX',(b[0]-t['x'])/t['w']+.5),('entryY',(b[1]-t['y'])/t['h']+.5)]:style+=key+'='+str(v)+';'
 style+='exitPerimeter=0;entryPerimeter=0;'
 if e['dashed']:style+='dashed=1;dashPattern=6 4;'
 cell=ET.SubElement(root,'mxCell',id=e['id'],edge='1',parent='1',source=e['source'],target=e['target'],style=style)
 geo=ET.SubElement(cell,'mxGeometry',relative='1',**{'as':'geometry'});arr=ET.SubElement(geo,'Array',**{'as':'points'})
 for x,y in e['points'][1:-1]:ET.SubElement(arr,'mxPoint',x=str(x),y=str(y))
for s in shapes.values():
 kind=s['kind'];style='html=1;whiteSpace=wrap;fontFamily=Arial;align=center;verticalAlign=middle;'
 if kind in ['entity','relationship']:
  style+=f'fillColor={BLUE};strokeColor={BLUE};fontColor=#FFFFFF;fontSize='+str(s['font_size'])
  style+=';fontStyle=1;' if kind=='entity' else ';rhombus;perimeter=rhombusPerimeter;'
 else:style+=f'text;strokeColor=none;fillColor=#FFFFFF;fontColor={INK};fontSize='+str(s['font_size'])+';'+('fontStyle=1;' if kind=='title' else '')
 cell=ET.SubElement(root,'mxCell',id=s['id'],value='<br>'.join(html.escape(t) for t in s['lines']),style=style,vertex='1',parent='1')
 bb=box(s);ET.SubElement(cell,'mxGeometry',x=str(bb[0]),y=str(bb[1]),width=str(s['w']),height=str(s['h']),**{'as':'geometry'})
ET.indent(mx);ET.ElementTree(mx).write(HERE/(BASE+'.drawio'),encoding='utf-8',xml_declaration=True)

im=Image.new('RGB',(W,H),'white');draw=ImageDraw.Draw(im)
segments=[];crossings=[]
for e in edges:
 for a,b in zip(e['points'],e['points'][1:]):
  if e['dashed']:
   length=math.dist(a,b)
   if not length:continue
   for start in range(0,math.ceil(length),16):
    stop=min(length,start+10)
    draw.line([(a[0]+(b[0]-a[0])*start/length,a[1]+(b[1]-a[1])*start/length),(a[0]+(b[0]-a[0])*stop/length,a[1]+(b[1]-a[1])*stop/length)],fill=LINE,width=2)
  else:draw.line([a,b],fill=LINE,width=2)
  for c,d in segments:
   if a[1]==b[1] and c[0]==d[0] and min(a[0],b[0])+7<c[0]<max(a[0],b[0])-7 and min(c[1],d[1])+7<a[1]<max(c[1],d[1])-7:
    x,y=c[0],a[1];crossings.append((x,y));draw.rectangle((x-6,y-6,x+6,y+2),fill='white');draw.arc((x-6,y-6,x+6,y+6),180,360,fill=LINE,width=2)
   elif a[0]==b[0] and c[1]==d[1] and min(c[0],d[0])+7<a[0]<max(c[0],d[0])-7 and min(a[1],b[1])+7<c[1]<max(a[1],b[1])-7:
    x,y=a[0],c[1];crossings.append((x,y));draw.rectangle((x-2,y-6,x+6,y+6),fill='white');draw.arc((x-6,y-6,x+6,y+6),270,450,fill=LINE,width=2)
  segments.append((a,b))
overflow=[]
for s in shapes.values():
 kind=s['kind'];bb=box(s)
 if kind=='entity':draw.rectangle(bb,fill=BLUE)
 elif kind=='relationship':draw.polygon([(s['x'],bb[1]),(bb[2],s['y']),(s['x'],bb[3]),(bb[0],s['y'])],fill=BLUE)
 else:draw.rectangle(bb,fill='white')
 size=s['font_size']
 f=ImageFont.truetype('C:/Windows/Fonts/'+('arialbd.ttf' if kind in ['entity','title'] else 'arial.ttf'),size)
 lines=s['lines'];lh=s['lineheight']
 if len(lines)*lh>s['h']:overflow.append(s['id'])
 y=s['y']-len(lines)*lh/2
 for line in lines:
  draw.text((s['x']-draw.textlength(line,font=f)/2,y),line,font=f,fill='white' if kind in ['entity','relationship'] else INK);y+=lh
im.save(HERE/(BASE+'.png'),dpi=(160,160),optimize=True)
im.resize((2170,1190),Image.Resampling.LANCZOS).save(HERE/'vista_general.png')

def intersect(a,b,r):
 x0,y0,x1,y1=r;t0,t1=0.,1.;dx=b[0]-a[0];dy=b[1]-a[1]
 for p,q in [(-dx,a[0]-x0),(dx,x1-a[0]),(-dy,a[1]-y0),(dy,y1-a[1])]:
  if p==0:
   if q<0:return False
  else:
   t=q/p
   if p<0:t0=max(t0,t)
   else:t1=min(t1,t)
 return t0<t1 and t0<1 and t1>0
hits=[];overlaps=[]
for a,b in itertools.combinations(shapes.values(),2):
 aa,bb=box(a,-1),box(b,-1)
 if aa[0]<bb[2] and aa[2]>bb[0] and aa[1]<bb[3] and aa[3]>bb[1]:overlaps.append([a['id'],b['id']])
for e in edges:
 for s in shapes.values():
  if s['id'] not in [e['source'],e['target']] and any(intersect(a,b,box(s,-2)) for a,b in zip(e['points'],e['points'][1:])):hits.append([e['id'],s['id']])
adj=collections.defaultdict(set)
for e in edges:adj[e['source']].add(e['target']);adj[e['target']].add(e['source'])
seen=set();todo=['clientes']
while todo:
 n=todo.pop()
 if n not in seen:seen.add(n);todo.extend(adj[n]-seen)
valid=dict(pages=len(mx.findall('diagram')),entities=len(positions),relationships=len(relations),
 attributes=0,ovals=0,all_entities_connected=set(positions)<=seen,shape_overlaps=overlaps,
 connectors_through_shapes=hits,text_overflow=overflow,crossings=len(crossings),png=[W,H])
(HERE/'VALIDACION.json').write_text(json.dumps(valid,ensure_ascii=False,indent=2),encoding='utf-8')
print(json.dumps(valid,ensure_ascii=False,indent=2),flush=True)
assert not hits and not overlaps and not overflow and set(positions)<=seen
assert len(mx.findall('diagram'))==1 and not any('ellipse' in c.get('style','') for c in root)

report='''# Chen simplificado V22 — El Bravo del JAC

Versión didáctica sin atributos. Una sola página y un único grafo conectado.
Contiene 24 entidades relacionales, las tres N:M como rombos y dos conceptos
documentales MongoDB. No se muestran IDs, campos, tipos de datos ni claves.

## Qué se simplificó

Se conservan las entidades del negocio, sus asociaciones principales, el servicio
directo sin diagnóstico, las cotizaciones previas a la orden, los pagos y la
entrega asociados a la orden, y los responsables técnicos de diagnóstico,
asignaciones y auxilio. EMPLEADOS sigue separado de la cuenta y del acceso RBAC.
El cargo laboral permanece en el sistema, aunque aquí no se dibuja como atributo.

Se omiten conexiones de autoría y auditoría repetitivas: quién registró cada
ingreso, pago, entrega, prueba o cambio. Los eventos e historiales sí permanecen.
También se abstraen la autoría documental, la revisión de cotizaciones y las
dos conexiones adicionales de asignación de diagnóstico en recepción. No se
elimina ninguna relación del sistema: esta es una vista conceptual simplificada,
no un reemplazo del esquema físico ni una promesa de cobertura de todas sus FK.

## Participaciones y lectura

Cada etiqueta está junto a la entidad a la que pertenece y expresa mínimo:máximo:
`0:N` admite cero o varios, `1:1` exige exactamente uno y `0:1` admite uno opcional.
Una N:M se representa con `0:N` en ambos extremos. Las reglas condicionales no
se convierten en mínimos universales que los datos históricos no garantizan.
Las líneas discontinuas representan vínculos documentales MongoDB, sin FK
entre motores. Los pequeños sobrepasos distinguen cruces de uniones.

Servicio directo omite diagnóstico, pero conserva recepción y cotización.
Crear una orden exige cotización aprobada. La entrega corresponde a una orden
finalizada y con saldo neto cero; no se inventa una relación pago–entrega.
Las funciones de mecánico y auxiliar se mantienen dentro de las asignaciones;
no se confunden con los cargos laborales o con los roles de acceso.

## Relaciones visibles

| Entidad | Relación | Entidad | Participaciones respectivas |
|---|---|---|---|
'''
for r in sorted(relations,key=lambda r:r['id']):
 report+=f"| {names[r['a']]} | {r['verb']} | {names[r['b']]} | {r['ca']} / {r['cb']} |\n"
report+='''
## Validación

Se validaron el XML nativo, una sola página, extremos conectados, ausencia de
óvalos/atributos, cobertura de todas las entidades, participaciones y geometría.
Los resultados exactos figuran en VALIDACION.json. Se revisa la imagen completa
y ampliaciones antes de entregar. No se afirma haber abierto Draw.io Desktop:
la validación del archivo editable es local. Abrir desde Archivo → Abrir desde
→ Dispositivo en Draw.io y seleccionar el archivo .drawio.

Fuentes: inventario y relaciones contrastados del V21 con las migraciones
001–007 y la rama funcional que incluye EMPLEADOS. V21 queda conservado.
Solo se agregan documentos V22; no se modifican bases, backend, frontend ni
migraciones. Sin commit, push o merge.
'''
(HERE/'INFORME_CHEN_SIMPLIFICADO.md').write_text(report,encoding='utf-8')
