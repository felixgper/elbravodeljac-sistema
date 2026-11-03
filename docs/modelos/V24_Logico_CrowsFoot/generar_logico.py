"""Diagrama lógico nativo. Lee inventario; escribe exclusivamente esta carpeta."""
from pathlib import Path
from collections import defaultdict
import json,math,heapq,itertools,html,xml.etree.ElementTree as ET
OUT=Path(__file__).resolve().parent;M=json.loads((OUT/'ESQUEMA_VERIFICADO.json').read_text(encoding='utf8'));T=M['tables'];F=M['relationships']
W,H,STEP=8080,5520,10;TW,HEADER,RH=820,60,40;BASE='El_Bravo_del_JAC_V24_Logico_Completo'
SECTIONS=[('01 · CLIENTES Y VEHÍCULOS',(80,280,2520,2280),'#dce8f3'),('02 · RECEPCIÓN, DIAGNÓSTICO Y COTIZACIÓN',(2760,280,2520,2280),'#dcefeb'),('03 · ÓRDENES Y EJECUCIÓN',(5440,280,2520,2280),'#f3eadb'),('04 · AUXILIO Y RECURSOS',(80,2800,2520,2260),'#e9e3f5'),('05 · PERSONAL Y SEGURIDAD',(2760,2800,2520,2260),'#dfe8f5'),('06 · PAGOS Y ENTREGA',(5440,2800,2520,2260),'#f4e5e5')]
POS={'clientes':(180,500),'vehiculos':(1540,500),'contactos_cliente':(180,1440),'historial_propietarios':(1540,1440),'tipos_documento':(180,2180),'recepciones':(2860,500),'diagnosticos':(4220,500),'cotizaciones':(2860,1540),'detalle_cotizacion':(4220,1540),'ordenes_trabajo':(5540,500),'asignaciones_ot':(6900,500),'avances_ot':(5540,1500),'pruebas_ot':(6900,1500),'historial_estados_ot':(5540,2160),'auxilios':(180,3020),'unidades_auxilio':(1540,3020),'servicios':(180,3860),'repuestos':(1540,3860),'proveedores':(180,4520),'repuesto_proveedor':(1540,4520),'usuarios':(2860,3020),'empleados':(4220,3020),'roles':(2860,4040),'usuario_roles':(4220,4040),'permisos':(2860,4720),'rol_permisos':(4220,4720),'pagos':(5540,3020),'entregas':(6900,3020)}
assert set(POS)==set(T)
boxes={n:[*POS[n],TW,HEADER+RH*len(t['columns'])] for n,t in T.items()}
endpoints=[];incoming=defaultdict(list)
for f in F:
 cx,cy,_,ch=boxes[f['child']];px,py,_,ph=boxes[f['parent']]
 if f['child']==f['parent']:cs=ps='R'
 elif cx==px:cs=ps='L'
 elif cx<px:cs,ps='R','L'
 else:cs,ps='L','R'
 incoming[f['parent'],ps].append(f['id']);endpoints.append(dict(fk=f,cs=cs,ps=ps))
# Separate target ports prevent different optional symbols from merging at a PK.
for (n,side),ids in incoming.items():boxes[n][3]=max(boxes[n][3],100+40*len(ids))
obstacles=[(x-12,y-12,x+w+12,y+h+12) for x,y,w,h in boxes.values()]
# Block only section headers; background regions are transparent routing areas.
obstacles += [(x,y,x+w,y+70) for _,(x,y,w,h),_ in SECTIONS]
blocked=set()
for x0,y0,x1,y1 in obstacles:
 for x in range(math.ceil(x0/STEP),math.floor(x1/STEP)+1):
  for y in range(math.ceil(y0/STEP),math.floor(y1/STEP)+1):blocked.add((x,y))
terminals=set()
for e in endpoints:
 f=e['fk'];x,y,w,h=boxes[f['child']];column=next(i for i,c in enumerate(T[f['child']]['columns']) if c['name']==f['column'])
 a=(x if e['cs']=='L' else x+w,y+HEADER+column*RH+RH/2)
 x,y,w,h=boxes[f['parent']];slot=incoming[f['parent'],e['ps']].index(f['id']);total=len(incoming[f['parent'],e['ps']]);yy=y+100+slot*40
 b=(x if e['ps']=='L' else x+w,yy)
 pa=(a[0]+(-80 if e['cs']=='L' else 80),a[1]);pb=(b[0]+(-80 if e['ps']=='L' else 80),b[1]);e.update(a=a,b=b,pa=pa,pb=pb)
 for v in [pa,pb]:terminals.add((int(v[0]/STEP),int(v[1]/STEP)))
used=set();usednodes=defaultdict(set)
def segment(a,b):return tuple(sorted((a,b)))
# Reserve all terminal stubs, including their directions, before routing.
for e in endpoints:
 for a,b in [(e['a'],e['pa']),(e['b'],e['pb'])]:
  start=tuple(int(v/STEP) for v in a);target=tuple(int(v/STEP) for v in b);dx=1 if target[0]>start[0] else -1;pp=[(x,start[1]) for x in range(start[0],target[0]+dx,dx)]
  for u,v in zip(pp,pp[1:]):used.add(segment(u,v))
  for point in pp[1:-1]:usednodes[point].add(1)
def route(a,b):
 start=tuple(int(v/STEP) for v in a);target=tuple(int(v/STEP) for v in b);assert start not in blocked and target not in blocked,(a,b)
 best={(start,0):0};prev={};queue=[(0,0,start,0)]
 while queue:
  _,cost,p,old=heapq.heappop(queue)
  if cost!=best.get((p,old)):continue
  if p==target:state=(p,old);break
  for dx,dy,d in [(1,0,1),(-1,0,1),(0,1,2),(0,-1,2)]:
   q=(p[0]+dx,p[1]+dy)
   if not (2<q[0]<W/STEP-2 and 12<q[1]<H/STEP-10):continue
   if q in blocked or segment(p,q) in used or (q in terminals and q!=target):continue
   h=usednodes[p];t=usednodes[q]
   if h and (d in h or len(h)>1 or (old and old!=d)):continue
   if t and (d in t or len(t)>1):continue
   nc=cost+10+(20 if old and old!=d else 0)+(350 if t else 0)
   if nc<best.get((q,d),math.inf):best[q,d]=nc;prev[q,d]=(p,old);heapq.heappush(queue,(nc+10*(abs(q[0]-target[0])+abs(q[1]-target[1])),nc,q,d))
 else:raise RuntimeError('Sin ruta '+str((a,b)))
 pp=[]
 while True:
  pp.append(state[0])
  if state==(start,0):break
  state=prev[state]
 pp.reverse()
 for a,b in zip(pp,pp[1:]):
  used.add(segment(a,b));d=1 if a[1]==b[1] else 2;usednodes[a].add(d);usednodes[b].add(d)
 return [(x*STEP,y*STEP) for x,y in pp]
endpoints.sort(key=lambda e:(0 if e['fk']['id'] in ['F22','F02'] else 1,math.dist(e['pa'],e['pb'])))
for e in endpoints:
 full=[e['a']]+route(e['pa'],e['pb'])+[e['b']];points=[]
 for p in full:
  if points and p==points[-1]:continue
  if len(points)>1 and (points[-2][0]==points[-1][0]==p[0] or points[-2][1]==points[-1][1]==p[1]):points[-1]=p
  else:points.append(p)
 e['points']=points;print(e['fk']['id'],flush=True)
file=ET.Element('mxfile',host='app.diagrams.net',type='device');diagram=ET.SubElement(file,'diagram',id='logico-v24',name='V24 · lógico PostgreSQL completo');graph=ET.SubElement(diagram,'mxGraphModel',page='1',pageWidth=str(W),pageHeight=str(H),pageScale='1',grid='1',gridSize='20',background='#ffffff');root=ET.SubElement(graph,'root');ET.SubElement(root,'mxCell',id='0');ET.SubElement(root,'mxCell',id='1',parent='0')
def cell(id,value,x,y,w,h,style,parent='1',**attrs):
 c=ET.SubElement(root,'mxCell',id=id,value=value,style=style,vertex='1',parent=parent,**attrs);ET.SubElement(c,'mxGeometry',x=str(x),y=str(y),width=str(w),height=str(h),**{'as':'geometry'});return c
text='text;html=0;whiteSpace=wrap;overflow=hidden;strokeColor=none;fillColor=none;align=left;verticalAlign=middle;fontFamily=Arial;fontColor=#244563;'
cell('title','EL BRAVO DEL JAC / MODELO LÓGICO RELACIONAL COMPLETO',80,40,7850,70,text+'fontSize=48;fontStyle=1;')
cell('subtitle',f"V24 · PostgreSQL · {len(T)} tablas · {sum(len(t['columns']) for t in T.values())} atributos · {len(F)} FK · SQL base + migraciones 001–009 · Sin cambios al sistema",80,130,7850,55,text+'fontSize=28;')
for i,(title,(x,y,w,h),color) in enumerate(SECTIONS):
 cell('area-'+str(i),'',x,y,w,h,'rounded=1;arcSize=1;fillColor=#ffffff;strokeColor=#dce5ee;strokeWidth=2;connectable=0;')
 cell('area-title-'+str(i),title,x,y,w,70,text+f'fillColor={color};fontSize=30;fontStyle=1;spacingLeft=25;')
marker={'1':'ERmandOne','0..1':'ERzeroToOne','0..N':'ERzeroToMany'}
for e in endpoints:
 f=e['fk'];x,y,w,h=boxes[f['parent']];metadata=f"{f['id']} | {f['child']}.{f['column']} → {f['parent']}.{f['parent_column']} | NULL={f['nullable']} | UQ individual={f['individuallyUnique']}"
 style=f"edgeStyle=none;rounded=0;html=0;startArrow={marker[f['childMultiplicity']]};endArrow={marker[f['parentMultiplicity']]};startSize=18;endSize=18;strokeColor=#587799;strokeWidth=2;exitX={0 if e['cs']=='L' else 1};exitY=0.5;entryX={0 if e['ps']=='L' else 1};entryY={(e['b'][1]-y)/h};exitPerimeter=0;entryPerimeter=0;jumpStyle=arc;jumpSize=8;"
 c=ET.SubElement(root,'mxCell',id=f['id'],value='',edge='1',parent='1',source='row-'+f['child']+'-'+f['column'],target='table-'+f['parent'],style=style,tooltip=metadata)
 geo=ET.SubElement(c,'mxGeometry',relative='1',**{'as':'geometry'});arr=ET.SubElement(geo,'Array',**{'as':'points'})
 for xx,yy in e['points'][1:-1]:ET.SubElement(arr,'mxPoint',x=str(xx),y=str(yy))
 ET.SubElement(c,'Object',**{'as':'metadata','fkId':f['id'],'constraint':f['name'],'sourceColumn':f['column'],'targetColumn':f['parent_column'],'childMultiplicity':f['childMultiplicity'],'parentMultiplicity':f['parentMultiplicity']})
for name,t in T.items():
 x,y,w,h=boxes[name];parent='table-'+name
 cell(parent,'',x,y,w,h,'rounded=0;container=1;collapsible=0;fillColor=#ffffff;strokeColor=#8faac3;strokeWidth=2;')
 cell('header-'+name,name,0,0,w,HEADER,text+'fillColor=#dce8f3;fontSize=29;fontStyle=1;spacingLeft=16;',parent)
 for i,col in enumerate(t['columns']):
  flags=[]
  if col['name'] in t['pk']:flags.append('PK')
  f=next((f for f in F if f['child']==name and f['column']==col['name']),None)
  if f:flags.append(f['id'])
  if [col['name']]==t['pk'] or any(u['columns']==[col['name']] and not u['predicate'] for u in t['unique']):
   if 'PK' not in flags:flags.append('UQ')
  rid='row-'+name+'-'+col['name'];bg='#f2f6fa' if i%2 else '#ffffff'
  cell(rid,'',0,HEADER+i*RH,w,RH,f'rounded=0;fillColor={bg};strokeColor=none;portConstraint=eastwest;',parent)
  cell('flag-'+name+'-'+col['name'],' / '.join(flags),10,0,125,RH,text+'fontSize=19;fontStyle=1;',rid)
  cell('column-'+name+'-'+col['name'],col['name'],140,0,450,RH,text+'fontSize=24;',rid)
  cell('type-'+name+'-'+col['name'],col['type']+' '+('NN' if col['not_null'] else 'NULL'),595,0,220,RH,text+'fontSize=19;',rid)
legend='PK = primaria (si ocupa varias filas, es compuesta) · F01–F50 = columnas FK; trazabilidad completa en inventario · UQ = única individual · NN = obligatorio · NULL = opcional\nCrow’s Foot: || uno · o| cero o uno · o< cero o muchos. Símbolo junto a la tabla = multiplicidad de ese extremo. UNIQUE compuestas/parciales: ver inventario; no implican UQ individual.\nFK operativas y auditoría hacia USUARIOS; EMPLEADOS por vínculo nullable + UQ. id_recepcionista sigue vigente; renombrado pendiente. MongoDB queda fuera: no hay FK PostgreSQL hacia sus colecciones.'
cell('legend',legend,100,5160,7850,240,text+'fillColor=#edf3f9;spacing=20;fontSize=27;')
ET.indent(file);ET.ElementTree(file).write(OUT/(BASE+'.drawio'),encoding='utf-8',xml_declaration=True)
# Independent geometric checks based on explicit orthogonal segments.
def hit(a,b,r):
 x,y,w,h=r
 if a[0]==b[0]:return x<a[0]<x+w and max(min(a[1],b[1]),y)<min(max(a[1],b[1]),y+h)
 return y<a[1]<y+h and max(min(a[0],b[0]),x)<min(max(a[0],b[0]),x+w)
hits=[];overlap=[];cross=[];segments=[]
for e in endpoints:
 for a,b in zip(e['points'],e['points'][1:]):
  assert a[0]==b[0] or a[1]==b[1]
  for n,r in boxes.items():
   if n not in [e['fk']['child'],e['fk']['parent']] and hit(a,b,r):hits.append([e['fk']['id'],n])
  for id,c,d in segments:
   if a[1]==b[1] and c[0]==d[0] and min(a[0],b[0])<c[0]<max(a[0],b[0]) and min(c[1],d[1])<a[1]<max(c[1],d[1]):cross.append([e['fk']['id'],id])
   if a[0]==b[0] and c[1]==d[1] and min(c[0],d[0])<a[0]<max(c[0],d[0]) and min(a[1],b[1])<c[1]<max(a[1],b[1]):cross.append([e['fk']['id'],id])
  segments.append((e['fk']['id'],a,b))
for (n,a),(m,b) in itertools.combinations(boxes.items(),2):
 if a[0]<b[0]+b[2] and a[0]+a[2]>b[0] and a[1]<b[1]+b[3] and a[1]+a[3]>b[1]:overlap.append([n,m])
valid=dict(tables=len(T),attributes=sum(len(t['columns']) for t in T.values()),foreignKeys=len(F),tableOverlaps=overlap,connectorsThroughTables=hits,orthogonal=True,crossings=len(cross),page=[W,H])
(OUT/'GEOMETRIA.json').write_text(json.dumps(dict(boxes=boxes,sections=SECTIONS,routes=endpoints,validation=valid),ensure_ascii=False,indent=2),encoding='utf8');print(json.dumps(valid));assert not hits and not overlap
