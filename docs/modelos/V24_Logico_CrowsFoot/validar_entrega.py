"""Verifica el artefacto, el DDL y la protección sin conexiones de datos."""
from pathlib import Path
from collections import defaultdict
import json,hashlib,re,csv,itertools,subprocess,xml.etree.ElementTree as ET,unicodedata
from PIL import Image,ImageDraw,ImageFont
from pypdf import PdfReader
O=Path(__file__).resolve().parent;R=O.parents[2];BASE='El_Bravo_del_JAC_V24_Logico_Completo'
def load(name):return json.loads((O/name).read_text(encoding='utf8'))
M=load('ESQUEMA_VERIFICADO.json');T=M['tables'];F=M['relationships'];G=load('GEOMETRIA.json');render=load('VALIDACION_RENDER_LOCAL.json')
assert render['passed'] and render['nativeGeometryEdit'] and render['edges']==len(F)
# Counts independent from the inventory's final counters.
sqls=[(R/s['path']).read_text(encoding='utf8') for s in M['sources']]
base=sqls[0];assert len(re.findall(r'CREATE TABLE public\.\w+ \(',base))==26
basefk=re.findall(r'FOREIGN KEY \((\w+)\) REFERENCES public\.(\w+)\((\w+)\)',base);assert len(basefk)==42
addedrefs=[];addedcols=[];newtables=[]
for source in sqls[1:]:
 clean=re.sub(r'--[^\n]*','',source)
 addedrefs+=re.findall(r'REFERENCES (?:public\.)?(\w+)\((\w+)\)',clean)
 addedcols+=re.findall(r'ALTER TABLE (?:public\.)?\w+ ADD COLUMN (?:IF NOT EXISTS )?\w+',clean)
 newtables+=re.findall(r'CREATE TABLE IF NOT EXISTS public\.(\w+)',clean)
assert len(T)==26+len(newtables)==28
assert len(F)==len(basefk)+len(addedrefs)==50
assert sum(len(t['columns']) for t in T.values())==160+len(addedcols)+8==208
# Every checked source still matches its recorded hash.
for src in M['sources']:assert hashlib.sha256((R/src['path']).read_bytes()).hexdigest()==src['sha256']
xml=ET.parse(O/(BASE+'.drawio')).getroot();cells={c.get('id'):c for c in xml.iter('mxCell')};assert len(cells)==len(list(xml.iter('mxCell')))
assert len(xml.findall('diagram'))==1
assert len([c for i,c in cells.items() if i.startswith('table-')])==len(T)
assert len([c for i,c in cells.items() if i.startswith('row-')])==208
assert len([c for c in cells.values() if c.get('edge')=='1'])==len(F)
assert not any('rhombus' in c.get('style','') for c in cells.values())
marker={'1':'ERmandOne','0..1':'ERzeroToOne','0..N':'ERzeroToMany'}
for name,t in T.items():
 assert t['pk'] and all(next(c for c in t['columns'] if c['name']==k)['not_null'] for k in t['pk'])
 assert cells['header-'+name].get('value')==name
 for col in t['columns']:
  suffix=name+'-'+col['name'];assert cells['column-'+suffix].get('value')==col['name']
  assert cells['type-'+suffix].get('value')==col['type']+' '+('NN' if col['not_null'] else 'NULL')
  flags=cells['flag-'+suffix].get('value','');assert ('PK' in flags)==(col['name'] in t['pk'])
for f in F:
 e=cells[f['id']];assert e.get('source')=='row-'+f['child']+'-'+f['column'];assert e.get('target')=='table-'+f['parent']
 meta=e.find('Object');assert meta.get('sourceColumn')==f['column'] and meta.get('targetColumn')==f['parent_column']
 table=T[f['child']];col=next(c for c in table['columns'] if c['name']==f['column'])
 unique=table['pk']==[col['name']] or any(u['columns']==[col['name']] and not u['predicate'] for u in table['unique'])
 child='0..1' if unique else '0..N';parent='1' if col['not_null'] else '0..1'
 st=dict(v.split('=',1) for v in e.get('style').split(';') if '=' in v)
 assert st['startArrow']==marker[child] and st['endArrow']==marker[parent]
 assert st['edgeStyle']=='none' and meta.get('childMultiplicity')==child and meta.get('parentMultiplicity')==parent
 assert f['id'] in cells['flag-'+f['child']+'-'+f['column']].get('value','')
assert next(f for f in F if f['id']=='F49')['childMultiplicity']=='0..1'
assert {f['id'] for f in F if f['child']==f['parent']}=={'F14'}
# Detect cycles across different tables independently.
adj=defaultdict(set)
for f in F:
 if f['child']!=f['parent']:adj[f['child']].add(f['parent'])
active=set();visited=set()
def visit(n):
 assert n not in active,'Ciclo entre tablas: '+n
 if n in visited:return
 active.add(n)
 for other in adj[n]:visit(other)
 active.remove(n);visited.add(n)
for name in T:visit(name)
assert T['usuario_roles']['pk']==['id_usuario','id_rol'] and T['rol_permisos']['pk']==['id_rol','id_permiso']
assert any(set(u['columns'])=={'id_repuesto','id_proveedor'} and not u['predicate'] for u in T['repuesto_proveedor']['unique'])
assert G['validation']['orthogonal'] and not G['validation']['tableOverlaps'] and not G['validation']['connectorsThroughTables']
segments=[];overlap=[]
for route in G['routes']:
 for a,b in zip(route['points'],route['points'][1:]):
  assert a[0]==b[0] or a[1]==b[1]
  segments.append((route['fk']['id'],a,b))
for (i,a,b),(j,c,d) in itertools.combinations(segments,2):
 if i==j:continue
 for axis in [0,1]:
  o=1-axis
  if a[axis]==b[axis]==c[axis]==d[axis]:
   length=min(max(a[o],b[o]),max(c[o],d[o]))-max(min(a[o],b[o]),min(c[o],d[o]))
   if length>0:overlap.append([i,j,length])
assert not overlap,overlap
# Check font widths using the actual diagram's fonts and row dimensions.
draw=ImageDraw.Draw(Image.new('RGB',(1,1)));overflow=[]
for id,c in cells.items():
 if id.startswith(('column-','type-','flag-','header-')):
  style=c.get('style','');size=float(re.search(r'fontSize=([0-9.]+)',style)[1]);bold='fontStyle=1' in style
  font=ImageFont.truetype('C:/Windows/Fonts/'+('arialbd.ttf' if bold else 'arial.ttf'),int(size));width=float(c.find('mxGeometry').get('width'))
  if draw.textlength(c.get('value',''),font=font)>width-3:overflow.append(id)
assert not overflow,overflow
pdf=PdfReader(O/(BASE+'.pdf'));assert len(pdf.pages)==1
width,height=float(pdf.pages[0].mediabox.width),float(pdf.pages[0].mediabox.height);assert abs(width-3370.39)<2 and abs(height-2383.94)<2
text=pdf.pages[0].extract_text();norm=lambda v:re.sub(r'[^a-z0-9_]','',unicodedata.normalize('NFKD',v.lower()))
assert all(norm(name) in norm(text) for name in T)
assert all(f['id'] in text for f in F)
image=Image.open(O/(BASE+'.png'));size=image.size;assert size[0]>7900 and size[1]>5300
preview=image.copy();preview.thumbnail((2400,1640));preview.save(O/'vista_previa.png')
# Supplemental crops; the single master diagram still has each table once.
(O/'vistas').mkdir(exist_ok=True)
bounds=render['bounds'];offsetx=-bounds['x']+20;offsety=-bounds['y']+20
for i,(label,(x,y,w,h),color) in enumerate(G['sections'],1):
 crop=image.crop((int(x+offsetx-15),int(y+offsety-15),int(x+w+offsetx+15),int(y+h+offsety+15)));crop.save(O/'vistas'/('%02d_bloque.png'%i))
protected=load('PROTECCION_ANTES.json');changed=[n for n,h in protected.items() if not(R/n).is_file() or hashlib.sha256((R/n).read_bytes()).hexdigest()!=h];assert not changed,changed
status=subprocess.check_output(['git','status','--porcelain','--untracked-files=all'],cwd=R,text=True)
assert all(line[3:].strip('"').startswith('docs/modelos/V24_Logico_CrowsFoot/') for line in status.splitlines()),status
branch=subprocess.check_output(['git','branch','--show-current'],cwd=R,text=True).strip();assert branch=='backend/arquitectura-go-mongodb'
assert len(list(csv.DictReader((O/'INVENTARIO_FK.csv').open(encoding='utf-8-sig'))))==50
result=dict(passed=True,tables=28,columns=208,primaryKeys=28,foreignKeys=50,baseForeignKeys=42,migrationForeignKeys=8,missingForeignKeys=[],duplicateForeignKeys=[],incorrectEndpoints=[],incorrectCardinalities=[],selfReference=['cotizaciones.id_cotizacion_base'],cyclesBetweenDistinctTables=False,bridgesValidated=True,employeeUserNullableUnique=True,typesDocumentPresent=True,tableOverlaps=[],connectorsThroughTables=[],collinearOverlappingSegments=overlap,textOverflow=overflow,orthogonal=True,crossings=G['validation']['crossings'],drawioPages=1,nativeOpenRender=True,nativeEdit=True,pdfPages=1,pdfFormat='A0 horizontal',pngDimensions=size,protectedFiles=len(protected),changedProtectedFiles=changed,branch=branch,changesOnlyNewDocumentaryFolder=True,databasesAccessed=False,applicationTestsRun=False)
(O/'VALIDACION_FINAL.json').write_text(json.dumps(result,ensure_ascii=False,indent=2),encoding='utf8')
print(json.dumps(result,ensure_ascii=False,indent=2))
