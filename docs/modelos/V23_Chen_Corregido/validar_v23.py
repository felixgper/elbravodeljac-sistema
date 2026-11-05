"""Valida XML editable y preservación; no accede a las bases ni modifica la aplicación."""
from pathlib import Path
import json,hashlib,subprocess,xml.etree.ElementTree as ET,collections,itertools
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[2]
def read(name): return json.loads((HERE/name).read_text(encoding='utf-8-sig'))
provenance=read('PROCEDENCIA_V22.json');spec=read('ESPECIFICACION_V23.json')
original=subprocess.check_output(['git','show',provenance['commit']+':'+provenance['file']],cwd=ROOT)
assert hashlib.sha256(original).hexdigest()==provenance['sha256']
old=ET.fromstring(original);doc=ET.parse(HERE/'El_Bravo_del_JAC_V23_Chen_Corregido.drawio').getroot()
cells={c.get('id'):c for c in doc.iter('mxCell')};oldcells={c.get('id'):c for c in old.iter('mxCell')}
entities=set(spec['positions']);relations={r['id']:r for r in spec['relations']};edges=[c for c in cells.values() if c.get('edge')=='1']
oldrels={i for i,c in oldcells.items() if 'rhombus' in c.get('style','')}
assert len(oldrels)==35 and oldrels<=relations.keys()
assert len(entities)==27 and len(relations)==42 and len(edges)==84
assert all(i in oldcells for i in entities-{'tipos_documento'})
assert len(doc.findall('diagram'))==1
assert not any('ellipse' in c.get('style','') for c in cells.values())
styles=lambda c:dict(x.split('=',1) for x in c.get('style','').split(';') if '=' in x)
def geom(c):
 g=c.find('mxGeometry');return tuple(float(g.get(k,0)) for k in ['x','y','width','height'])
def absolute(c):
 x,y,w,h=geom(c)
 if c.get('parent') in entities:
  px,py,_,_=geom(cells[c.get('parent')]);x+=px;y+=py
 return x,y,w,h
segments=[];adj=collections.defaultdict(set)
for rid,r in relations.items():
 assert 'rhombus' in cells[rid].get('style','')
 connections=[e for e in edges if e.get('target')==rid];assert len(connections)==2
 for end in ['a','b']:
  entity=r[end];edge=cells[rid+'_'+end];card=cells[rid+'_'+end+'_card']
  assert edge.get('source')==entity and edge.get('target')==rid
  assert card.get('parent')==entity and card.get('value')==r['c'+end]
  assert card.get('value') in ['(0,1)','(1,1)','(0,N)','(1,N)']
  x,y,w,h=absolute(cells[entity]);cx,cy,cw,ch=absolute(card)
  dx=max(x-(cx+cw),cx-(x+w),0);dy=max(y-(cy+ch),cy-(y+h),0)
  assert max(dx,dy)<=100,(rid,end,'Etiqueta lejana')
  adj[entity].add(rid);adj[rid].add(entity)
for e in edges:
 st=styles(e);a=absolute(cells[e.get('source')]);b=absolute(cells[e.get('target')])
 start=(a[0]+a[2]*float(st['exitX']),a[1]+a[3]*float(st['exitY']))
 end=(b[0]+b[2]*float(st['entryX']),b[1]+b[3]*float(st['entryY']))
 points=[start]+[(float(p.get('x')),float(p.get('y'))) for p in e.findall('mxGeometry/Array/mxPoint')]+[end]
 for p,q in zip(points,points[1:]):
  if p!=q:segments.append((e.get('id'),p,q))
overlaps=[]
for (ea,a,b),(eb,c,d) in itertools.combinations(segments,2):
 if ea==eb:continue
 for axis in [0,1]:
  other=1-axis
  if abs(a[axis]-b[axis])<.001 and abs(c[axis]-d[axis])<.001 and abs(a[axis]-c[axis])<.001:
   length=min(max(a[other],b[other]),max(c[other],d[other]))-max(min(a[other],b[other]),min(c[other],d[other]))
   if length>.01:overlaps.append([ea,eb,length])
assert not overlaps,overlaps
seen=set();todo=['clientes']
while todo:
 node=todo.pop()
 if node not in seen:seen.add(node);todo.extend(adj[node]-seen)
assert entities|relations.keys()<=seen
protected=json.loads((ROOT/'backend/.cache/chen-v23/proteccion-antes.json').read_text(encoding='utf-8-sig'))
changed=[name for name,h in protected.items() if hashlib.sha256((ROOT/name).read_bytes()).hexdigest()!=h]
assert not changed,changed
branch=subprocess.check_output(['git','branch','--show-current'],cwd=ROOT,text=True).strip();assert branch=='backend/arquitectura-go-mongodb'
status=subprocess.check_output(['git','status','--porcelain','--untracked-files=all'],cwd=ROOT,text=True)
assert all(line[3:].replace('"','').startswith('docs/modelos/V23_Chen_Corregido/') for line in status.splitlines()),status
render=read('VALIDACION_RENDER_LOCAL.json');assert render['passed'] and render['nativeGeometryEdit']
geometry=read('VALIDACION_GEOMETRIA.json');assert not geometry['shape_overlaps'] and not geometry['connectors_through_shapes'] and not geometry['text_overflow']
result=dict(passed=True,pages=1,entities=27,originalEntitiesPreserved=26,relations=42,originalRelationsPreserved=35,connections=84,cardinalitiesChecked=84,attributes=0,ovals=0,graphConnected=True,collinearOverlappingSegments=overlaps,protectedFiles=len(protected),changedProtectedFiles=changed,originalV22Sha256=provenance['sha256'],nativeRenderPassed=True,nativeEditPassed=True,branch=branch,changesOnlyInV23=True,databasesAccessed=False)
(HERE/'VALIDACION_INDEPENDIENTE.json').write_text(json.dumps(result,ensure_ascii=False,indent=2),encoding='utf8')
print(json.dumps(result,ensure_ascii=False,indent=2))
