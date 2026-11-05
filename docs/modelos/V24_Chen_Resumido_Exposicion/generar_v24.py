"""Genera una vista Chen resumida V23; solo archivos documentales, sin bases de datos."""
from pathlib import Path
from PIL import Image, ImageDraw, ImageFont
import collections, heapq, html, itertools, json, math, re, os, hashlib, random
import xml.etree.ElementTree as ET
HERE=Path(__file__).resolve().parent
BASE='El_Bravo_del_JAC_V24_Chen_Resumido'
W,H,STEP=3900,2850,20
BLUE,LINE,INK='#3269B8','#6384AD','#193B66'
SPEC=json.loads((HERE/'ESPECIFICACION_V24.json').read_text(encoding='utf-8'))
positions={k:tuple(v) for k,v in SPEC['positions'].items()}
names=SPEC['names'];relations=[dict(r) for r in SPEC['relations']]
placement={r['verb']:(r['x'],r['y']) for r in relations}
source=relations;omitted=[]
shapes={};edges=[]
def shape(id,kind,label,x,y,w,h):
 shapes[id]=dict(id=id,kind=kind,label=label,x=x,y=y,w=w,h=h)
def box(s,pad=0):return(s['x']-s['w']/2-pad,s['y']-s['h']/2-pad,s['x']+s['w']/2+pad,s['y']+s['h']/2+pad)
for n,(x,y) in positions.items():shape(n,'entity',names[n],x,y,340,180 if n=='usuarios' else 132)
for r in relations:
 x,y=placement[r['verb']];r.update(x=x,y=y)
 shape(r['id'],'relationship',r['verb'],x,y,260,170)
 shapes[r['id']]['proposal']=r.get('proposal','')
 if r.get('proposal'):shape(r['id']+'_proposal','proposal',r['proposal'],x+210,y+110,74,30)
 if r['id']=='R13':
  shape('rol_a','role','base',x-235,y-35,130,30)
  shape('rol_b','role','ampliación',x+235,y-35,130,30)

# Connections leave each rectangle toward its relationship, with separate ports.
def side_toward(p,q):
 dx,dy=q[0]-p[0],q[1]-p[1]
 return ('R' if dx>0 else 'L') if abs(dx)>=abs(dy) else ('B' if dy>0 else 'T')
inc=collections.defaultdict(list)
for r in relations:
 for end in ['a','b']:
  n=r[end];side=side_toward(positions[n],placement[r['verb']])
  if n=='ordenes_trabajo' and r['id'] in ['S08','S11']:side='B'
  if n=='cotizaciones' and r['id'] in ['S13','S14']:side='B'
  if n=='diagnosticos' and r['logical']:side='T'
  if n=='recepciones' and r['id']=='S04':side='B'
  if n=='clientes' and r['verb']=='SOLICITA AUXILIO':side='T'
  if r['id']=='R13':side=os.environ.get('CHEN_REV_SIDE','T')
  if r['id']=='R39' and n=='usuarios':side='T'
  inc[n,side].append((r,end))
# Allocate enough real perimeter for labels: no ports outside their entity.
for (n,side),items in inc.items():
 if side in ['T','B']:shapes[n]['w']=max(shapes[n]['w'],(len(items)-1)*120+120)
 else:shapes[n]['h']=max(shapes[n]['h'],(len(items)-1)*80+100)
anchors={};terminals=set()
for (n,side),items in inc.items():
 items.sort(key=lambda t:(t[0]['y'],t[0]['x']))
 x,y=positions[n]
 for i,(r,end) in enumerate(items):
  horizontal=side in ['L','R']
  offset=(i-(len(items)-1)/2)*(80 if horizontal else 120)
  dx,dy={'L':(-1,0),'R':(1,0),'T':(0,-1),'B':(0,1)}[side]
  a=(x+dx*shapes[n]['w']/2+(offset if not horizontal else 0),y+dy*shapes[n]['h']/2+(offset if horizontal else 0))
  portal=(round((a[0]+dx*70)/STEP)*STEP,round((a[1]+dy*70)/STEP)*STEP)
  if horizontal:cx,cy=a[0]+dx*62,a[1]-150
  else:cx,cy=a[0]+60,a[1]+dy*75
  cid=r['id']+'_'+end+'_card'
  shape(cid,'cardinality',r['c'+end],cx,cy,92,32)
  anchors[r['id'],end]=(a,portal)
  terminals.add(tuple(round(v/STEP) for v in portal))

shape('title','title','EL BRAVO DEL JAC',1950,80,3700,75)
shape('subtitle','legend','MODELO CONCEPTUAL CHEN | V24 | VISTA RESUMIDA PARA EXPOSICIÓN',1950,155,3700,55)
shape('guia','legend','Clientes y vehículos → Recepción y diagnóstico → Cotización → Trabajo → Pago y entrega',1950,230,3700,50)
shape('legend-notation','legend','LECTURA: (mínimo,máximo) describe a la entidad junto a la etiqueta. Rectángulo = entidad; rombo = relación. Sin atributos. El flujo continúa en la segunda fila de derecha a izquierda.',1950,2540,3700,100)
shape('legend-business','legend','Servicio directo: recepción → cotización, sin diagnóstico. Orden solo con cotización aprobada. Entrega sujeta a cierre y saldo. * Relaciones resumidas mediante detalles, asignaciones o cuentas laborales. Vista de negocio; no modifica el modelo lógico/físico.',1950,2670,3700,120)
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
  tip=(x+dx*130,y+dy*85)
  portal=(round((tip[0]+dx*30)/STEP)*STEP,round((tip[1]+dy*30)/STEP)*STEP)
  goals[r['id'],end]=(tip,portal,side)
  terminals.add(tuple(round(v/STEP) for v in portal))

for rr in relations:
 for ee in ['a','b']:
  for point in [anchors[rr['id'],ee][1],goals[rr['id'],ee][1]]:
   if tuple(round(v/STEP) for v in point) in blocked:
    print('PUERTO BLOQUEADO',rr['id'],ee,point,[k for k,ss in shapes.items() if box(ss,8)[0]<=point[0]<=box(ss,8)[2] and box(ss,8)[1]<=point[1]<=box(ss,8)[3]],flush=True)
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
   if not(4<q[0]<W//STEP-4 and 10<q[1]<H//STEP-4):continue
   if q in blocked or seg(p,q) in used or (q in terminals and q!=target):continue
   h=used_nodes.get(p,set());t=used_nodes.get(q,set())
   if h and (d in h or len(h)>1 or (old and old!=d)):continue
   if t and (d in t or len(t)>1):continue
   nc=cost+10+(14 if old and d!=old else 0)+(int(os.environ.get('CHEN_CROSS_COST','800')) if t else 0)
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
def route_length(r):return math.dist(positions[r['a']],placement[r['verb']])+math.dist(positions[r['b']],placement[r['verb']])
mode=os.environ.get('CHEN_ROUTE_ORDER','short')
if mode=='long':relations.sort(key=lambda r:-route_length(r))
elif mode=='core-long':relations.sort(key=lambda r:(0 if r['id'] in ['R01','R02','R03','R04','R05','R06','R07'] else 1,-route_length(r)))
elif mode=='core-users':relations.sort(key=lambda r:(0 if r['id'] in ['R01','R02','R03','R04','R05','R06','R07'] else 1 if 'usuarios' in [r['a'],r['b']] else 2,route_length(r)))
elif mode=='shuffle':
 random.Random(int(os.environ.get('CHEN_ROUTE_SEED','0'))).shuffle(relations)
 relations.sort(key=lambda r:0 if r['id'] in ['R01','R02','R03','R04','R05','R06','R07','R09','R10','R11'] else 1)
else:relations.sort(key=route_length)
for r in relations:
 for end in ['a','b']:
  a,p=anchors[r['id'],end];b,q,_=goals[r['id'],end]
  points=[]
  for point in [a]+route(p,q)+[b]:
   if points and point==points[-1]:continue
   while len(points)>1:
    u,v=points[-2],points[-1]
    if abs((v[0]-u[0])*(point[1]-v[1])-(v[1]-u[1])*(point[0]-v[0]))>1e-8:break
    points.pop()
   points.append(point)
  edges.append(dict(id=r['id']+'_'+end,source=r[end],target=r['id'],points=points,dashed=r['logical']))
 print(r['verb'],flush=True)
# Fit text inside the actual diamond, not merely its enclosing rectangle.
measure=ImageDraw.Draw(Image.new('RGB',(1,1)))
for s in shapes.values():
 kind=s['kind']
 size=34 if kind=='entity' else 28 if kind in ['relationship','cardinality'] else 56 if kind=='title' else 26 if kind in ['role','proposal','area'] else 30
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
 s['fill']='#237F80' if s['id'] in ['resultados_escaner','evidencias_tecnicas'] else '#B77818' if s.get('proposal') else BLUE

mx=ET.Element('mxfile',host='app.diagrams.net',type='device')
d=ET.SubElement(mx,'diagram',id='chen-simplificado-unico',name='V24 Chen resumido · exposición')
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
  style+=f'fillColor={s['fill']};strokeColor={s['fill']};fontColor=#FFFFFF;fontSize='+str(s['font_size'])
  style+=';fontStyle=1;' if kind=='entity' else ';rhombus;perimeter=rhombusPerimeter;'
 else:style+=f'text;strokeColor=none;fillColor=#FFFFFF;fontColor={INK};fontSize='+str(s['font_size'])+';'+('fontStyle=1;' if kind=='title' else '')
 parent='1';bb=box(s);gx,gy=bb[0],bb[1]
 if kind=='cardinality':
  rid,end=s['id'].rsplit('_',2)[:2];rel=next(r for r in relations if r['id']==rid);parent=rel[end];pb=box(shapes[parent]);gx-=pb[0];gy-=pb[1]
 cell=ET.SubElement(root,'mxCell',id=s['id'],value='<br>'.join(html.escape(t) for t in s['lines']),style=style,vertex='1',parent=parent,connectable='0' if kind not in ['entity','relationship'] else '1')
 ET.SubElement(cell,'mxGeometry',x=str(gx),y=str(gy),width=str(s['w']),height=str(s['h']),**{'as':'geometry'})
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
 if kind=='entity':draw.rectangle(bb,fill=s['fill'])
 elif kind=='relationship':draw.polygon([(s['x'],bb[1]),(bb[2],s['y']),(s['x'],bb[3]),(bb[0],s['y'])],fill=s['fill'])
 else:draw.rectangle(bb,fill='white')
 size=s['font_size']
 f=ImageFont.truetype('C:/Windows/Fonts/'+('arialbd.ttf' if kind in ['entity','title'] else 'arial.ttf'),size)
 lines=s['lines'];lh=s['lineheight']
 if len(lines)*lh>s['h']:overflow.append(s['id'])
 y=s['y']-len(lines)*lh/2
 for line in lines:
  draw.text((s['x']-draw.textlength(line,font=f)/2,y),line,font=f,fill='white' if kind in ['entity','relationship'] else INK);y+=lh
im.save(HERE/(BASE+'.png'),dpi=(160,160),optimize=True)
im.resize((1950,1425),Image.Resampling.LANCZOS).save(HERE/'vista_general.png')

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
valid['connector_length']=round(sum(math.dist(a,b) for e in edges for a,b in zip(e['points'],e['points'][1:])))
(HERE/'VALIDACION_GEOMETRIA.json').write_text(json.dumps(valid,ensure_ascii=False,indent=2),encoding='utf-8')
print(json.dumps(valid,ensure_ascii=False,indent=2),flush=True)
assert not hits and not overlaps and not overflow and set(positions)<=seen
assert len(mx.findall('diagram'))==1 and not any('ellipse' in c.get('style','') for c in root)


(HERE/'RELACIONES_V24.json').write_text(json.dumps(relations,ensure_ascii=False,indent=2),encoding='utf-8')
