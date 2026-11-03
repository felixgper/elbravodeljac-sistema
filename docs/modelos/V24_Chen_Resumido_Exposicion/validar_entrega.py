from pathlib import Path
import json,hashlib,subprocess,xml.etree.ElementTree as ET,unicodedata,re
from pypdf import PdfReader
from PIL import Image
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[2];BASE='El_Bravo_del_JAC_V24_Chen_Resumido'
def load(name):return json.loads((HERE/name).read_text(encoding='utf8'))
spec=load('ESPECIFICACION_V24.json');root=ET.parse(HERE/(BASE+'.drawio')).getroot();cells={c.get('id'):c for c in root.iter('mxCell')}
assert len(root.findall('diagram'))==1
assert len(spec['positions'])==13 and len(spec['relations'])==15
assert sum(c.get('edge')=='1' for c in cells.values())==30
assert not any('ellipse' in c.get('style','') for c in cells.values())
for r in spec['relations']:
 assert 'rhombus' in cells[r['id']].get('style','')
 links=[c for c in cells.values() if c.get('edge')=='1' and c.get('target')==r['id']];assert len(links)==2
 for end in ['a','b']:
  edge=cells[r['id']+'_'+end];card=cells[r['id']+'_'+end+'_card'];assert edge.get('source')==r[end] and edge.get('target')==r['id']
  assert card.get('parent')==r[end] and card.get('value')==r['c'+end]
  assert r['c'+end] in ['(0,1)','(1,1)','(0,N)','(1,N)']
geometry=load('VALIDACION_GEOMETRIA.json');assert geometry['all_entities_connected'] and not geometry['shape_overlaps'] and not geometry['connectors_through_shapes'] and not geometry['text_overflow']
render=load('VALIDACION_RENDER_LOCAL.json');assert render['passed'] and render['pdfExported'] and render['nativeGeometryEdit']
pdf=PdfReader(HERE/(BASE+'.pdf'));assert len(pdf.pages)==1
normal=lambda s:re.sub(r'[^A-Z0-9]','',unicodedata.normalize('NFKD',s.upper()))
text=normal(pdf.pages[0].extract_text());assert all(normal(n) in text for n in spec['names'].values())
width,height=map(float,[pdf.pages[0].mediabox.width,pdf.pages[0].mediabox.height]);assert abs(width-1190.55)<2 and abs(height-841.89)<2
im=Image.open(HERE/(BASE+'.png'));size=im.size;assert size[0]>3500 and size[1]>2500
im.thumbnail((1950,1450));im.save(HERE/'vista_previa.png')
protected=load('PROTECCION_ORIGINALES.json');changed=[n for n,h in protected.items() if not(ROOT/n).exists() or hashlib.sha256((ROOT/n).read_bytes()).hexdigest()!=h];assert not changed,changed
status=subprocess.check_output(['git','status','--porcelain','--untracked-files=all'],cwd=ROOT,text=True);assert all(line[3:].strip('"').startswith('docs/modelos/V24_Chen_Resumido_Exposicion/') for line in status.splitlines()),status
branch=subprocess.check_output(['git','branch','--show-current'],cwd=ROOT,text=True).strip();assert branch=='backend/arquitectura-go-mongodb'
result=dict(passed=True,entities=13,relationships=15,connections=30,cardinalitiesChecked=30,attributes=0,diagramPages=1,pdfPages=1,pdfFormat='A3 horizontal',pdfEntityNamesPresent=True,pngDimensions=size,nativeDrawioRender=True,nativeGeometryEdit=True,protectedFiles=len(protected),changedProtectedFiles=changed,changesOnlyInNewDocumentaryFolder=True,branch=branch,databasesAccessed=False,applicationTestsRun=False)
(HERE/'VALIDACION_FINAL.json').write_text(json.dumps(result,ensure_ascii=False,indent=2),encoding='utf8')
manifest={f.name:hashlib.sha256(f.read_bytes()).hexdigest() for f in HERE.iterdir() if f.is_file() and f.name!='MANIFIESTO_SHA256.json'};(HERE/'MANIFIESTO_SHA256.json').write_text(json.dumps(manifest,indent=2),encoding='utf8');print(json.dumps(result,ensure_ascii=False,indent=2))
