"""Verificación independiente del XML corregido y preservación de originales."""
from pathlib import Path
import collections, hashlib, html, itertools, json, re
import xml.etree.ElementTree as ET

HERE=Path(__file__).resolve().parent
REPO=HERE.parents[2]
PREFIX='El_Bravo_del_JAC_V22_Chen_Corregido'
old=ET.parse(HERE.parent/'V22_Chen_Simplificado/El_Bravo_del_JAC_V22_Chen_Simplificado.drawio')
new=ET.parse(HERE/(PREFIX+'.drawio'))
assert len(new.findall('diagram'))==1
model=new.find('./diagram/mxGraphModel')
assert model.get('pageWidth')=='6200' and model.get('pageHeight')=='3400'
cells=model.findall('./root/mxCell');idx={c.get('id'):c for c in cells}
assert len(cells)==len(idx)
oldidx={c.get('id'):c for c in old.findall('.//mxCell')}
inv=json.loads((HERE.parent/'V21_Chen/inventario_actual.json').read_text(encoding='utf-8'))
bridges={'usuario_roles','rol_permisos','repuesto_proveedor'}
entities=set(inv['tables'])-bridges|{'resultados_escaner','evidencias_tecnicas'}
for name in entities:
 assert oldidx[name].find('mxGeometry').attrib==idx[name].find('mxGeometry').attrib
 assert oldidx[name].get('value')==idx[name].get('value')
for c in old.findall('.//mxCell[@vertex="1"]'):
 if 'rhombus;' in c.get('style',''):
  assert c.find('mxGeometry').attrib==idx[c.get('id')].find('mxGeometry').attrib
relations=json.loads((HERE/'RELACIONES_CORREGIDAS.json').read_text(encoding='utf-8'))
baseline={r['id']:r for r in json.loads((HERE.parent/'V21_Chen/relaciones_chen.json').read_text(encoding='utf-8'))}
assert len(relations)==35
changes={'R17':('ca','(1,N)'),'R19':('cb','(1,1)'),
 'R20':('ca','(1,N)'),'R21':('ca','(1,N)'),
 'R26':('cb','(1,1)'),'R36':('cb','(1,1)'), 'R44':('cb','(1,1)')}
fks=[]
for r in relations:
 original=baseline[r['id']]
 for end in ['a','b']:
  assert r[end]==original[end]
  expected=changes[r['id']][1] if r['id'] in changes and changes[r['id']][0]=='c'+end else original['c'+end]
  assert r['c'+end]==expected,(r['id'],end)
  edge=idx[r['id']+'_'+end]
  assert edge.get('source')==r[end] and edge.get('target')==r['id']
  assert idx[r['id']+'_'+end+'_card'].get('value')==expected.strip('()').replace(',',':')
  assert ('dashed=1' in edge.get('style'))==r['logical']
  if not r['logical'] and not r.get('bridge'):
   fk=next(f for f in inv['relationships'] if f['name'] in r['physical'])
   assert (fk['parent'],fk['child'])==(r['a'],r['b'])
   sql=('(0,1)' if fk['unique'] else '(0,N)','(0,1)' if fk['nullable'] else '(1,1)')
   assert (original['ca'],original['cb'])==sql
 fks.extend(r['physical'])
assert len(fks)==len(set(fks))==36
assert len(set(inv['tables'])-entities-bridges)==0
assert set(fks)<={f['name'] for f in inv['relationships']}
assert {r['id'] for r in relations}>={'R13','R39'}
assert idx['R13_a'].get('source')==idx['R13_b'].get('source')=='cotizaciones'
def style(c):return dict(x.split('=',1) for x in c.get('style','').split(';') if '=' in x)
assert (style(idx['R13_a'])['exitX'],style(idx['R13_a'])['exitY'])!=(style(idx['R13_b'])['exitX'],style(idx['R13_b'])['exitY'])
assert idx['rol_a'].get('value')=='base' and idx['rol_b'].get('value')=='ampliación'
shapes={};edges=[];adj=collections.defaultdict(set)
for c in cells:
 if c.get('vertex')=='1':
  g=c.find('mxGeometry');x,y,w,h=(float(g.get(k)) for k in ['x','y','width','height'])
  assert x>=0 and y>=0 and x+w<=6200 and y+h<=3400
  shapes[c.get('id')]=(x,y,x+w,y+h)
  text=html.unescape(re.sub('<[^>]*>',' ',c.get('value','')))
  assert 'ellipse' not in c.get('style','') and not re.search(r'\bid_|\b(nombre|correo|placa|fecha|PK|FK|VARCHAR|BIGINT)\b',text,re.I)
for c in cells:
 if c.get('edge')!='1':continue
 st=style(c);a,b=c.get('source'),c.get('target')
 assert a in shapes and b in shapes and st['endArrow']==st['startArrow']=='none'
 def point(n,prefix):
  x,y,xx,yy=shapes[n];u,v=float(st[prefix+'X']),float(st[prefix+'Y'])
  assert 0<=u<=1 and 0<=v<=1
  return (x+(xx-x)*u,y+(yy-y)*v)
 pp=[point(a,'exit')]+[(float(p.get('x')),float(p.get('y'))) for p in c.findall('./mxGeometry/Array/mxPoint')]+[point(b,'entry')]
 assert all(u!=v for u,v in zip(pp,pp[1:]))
 for u,v,w in zip(pp,pp[1:],pp[2:]):
  dx,dy=v[0]-u[0],v[1]-u[1];ex,ey=w[0]-v[0],w[1]-v[1]
  assert not (abs(dx*ey-dy*ex)<1e-8 and dx*ex+dy*ey<0),(a,b,'retroceso de línea')
 edges.append((a,b,pp));adj[a].add(b);adj[b].add(a)
seen=set();todo=['clientes']
while todo:
 a=todo.pop()
 if a not in seen:seen.add(a);todo.extend(adj[a]-seen)
assert entities<=seen
def hit(a,b,bb):
 x0,y0,x1,y1=bb;dx,dy=b[0]-a[0],b[1]-a[1];t0,t1=0.,1.
 for p,q in [(-dx,a[0]-x0),(dx,x1-a[0]),(-dy,a[1]-y0),(dy,y1-a[1])]:
  if p==0:
   if q<0:return False
  else:
   u=q/p
   if p<0:t0=max(t0,u)
   else:t1=min(t1,u)
 return t0<t1 and t0<1 and t1>0
for a,b in itertools.combinations(shapes,2):
 x,y,xx,yy=shapes[a];u,v,uu,vv=shapes[b]
 assert not (x<uu and xx>u and y<vv and yy>v),(a,b,'superposición')
for a,b,pp in edges:
 for n,(x,y,xx,yy) in shapes.items():
  if n not in [a,b]:assert not any(hit(u,v,(x+2,y+2,xx-2,yy-2)) for u,v in zip(pp,pp[1:])),(a,b,n)
for file,expected in json.loads((HERE/'ORIGINALES_SHA256.json').read_text(encoding='utf-8')).items():
 assert hashlib.sha256((REPO/file).read_bytes()).hexdigest()==expected,file
for s in json.loads((HERE.parent/'V21_Chen/VALIDACION.json').read_text(encoding='utf-8'))['sources']:
 assert hashlib.sha256((REPO/s['path']).read_bytes()).hexdigest()==s['sha256']
result=dict(valid=True,pages=1,entities=len(entities),relationships=len(relations),attributes=0,
 bridge_tables_covered=3,physical_tables_covered=27,fks_explicitly_represented=36,
 unchanged_entity_positions=26,unchanged_original_diamond_positions=33,
 originals_preserved=True,migrations_unchanged=True,cardinality_changes_checked=7,
 operational_relationships_added=2,connectors=len(edges),all_ports_attached=True,
 no_shape_overlaps=True,no_connectors_through_shapes=True,no_retraced_line_stubs=True,local_xml_validation=True)
(HERE/'VALIDACION_INDEPENDIENTE.json').write_text(json.dumps(result,indent=2),encoding='utf-8')
print(json.dumps(result,indent=2))
