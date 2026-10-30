"""Validación independiente, sin Pillow, editor, red ni conexión a bases."""
from pathlib import Path
import collections, hashlib, html, json, re, xml.etree.ElementTree as ET

p=Path(__file__).resolve().parent
repo=p.parents[2]
tree=ET.parse(p/'El_Bravo_del_JAC_V21_Chen_Completo.drawio')
assert len(tree.findall('diagram'))==1
diagram=tree.find('diagram');model=diagram.find('mxGraphModel')
assert model is not None and model.get('page')=='1'
assert model.get('pageWidth')=='12600' and model.get('pageHeight')=='6150'
cells=model.find('root').findall('mxCell')
ids=[c.get('id') for c in cells]
assert len(ids)==len(set(ids))
byid={c.get('id'):c for c in cells}
vertices=[c for c in cells if c.get('vertex')=='1']
connectors=[c for c in cells if c.get('edge')=='1']
adj=collections.defaultdict(set)
for e in connectors:
    a,b=e.get('source'),e.get('target')
    assert a in byid and b in byid
    assert byid[a].get('vertex')=='1' and byid[b].get('vertex')=='1'
    assert 'endArrow=none' in e.get('style') and 'startArrow=none' in e.get('style')
    assert e.find('mxGeometry') is not None
    assert all(k+'=' in e.get('style') for k in ['entryX','entryY','exitX','exitY'])
    adj[a].add(b);adj[b].add(a)
assert not any('image=' in c.get('style','') or 'data:image' in c.get('value','') for c in cells)
inv=json.loads((p/'inventario_actual.json').read_text(encoding='utf-8'))
relations=json.loads((p/'relaciones_chen.json').read_text(encoding='utf-8'))
v=json.loads((p/'VALIDACION.json').read_text(encoding='utf-8'))
bridges={'usuario_roles','rol_permisos','repuesto_proveedor'}
entities=set(inv['tables'])-bridges
assert len(entities)==24 and entities<=set(ids)
seen=set();stack=['clientes']
while stack:
    a=stack.pop()
    if a in seen:continue
    seen.add(a);stack.extend(adj[a]-seen)
assert entities|{'resultados_escaner','evidencias_tecnicas'}<=seen
covered=[f for r in relations for f in r['physical']]
assert len(covered)==len(set(covered))==48
assert set(covered)=={r['name'] for r in inv['relationships']}
for r in relations:
    assert 'rhombus;' in byid[r['id']].get('style')
    if not r['logical'] and not r.get('bridge'):
        assert len(r['physical'])==1
        fk=next(f for f in inv['relationships'] if f['name']==r['physical'][0])
        assert (r['a'],r['b'])==(fk['parent'],fk['child'])
        assert r['ca']==('(0,1)' if fk['unique'] else '(0,N)')
        assert r['cb']==('(0,1)' if fk['nullable'] else '(1,1)')
    elif r.get('bridge'):
        assert r['ca']==r['cb']=='(0,N)' and len(r['physical'])==2
        assert all(f['child']==r['bridge'] for f in inv['relationships'] if f['name'] in r['physical'])
    else:
        assert r['ca']=='(0,N)' and r['cb']=='(1,1)'
    for which in ['a','b']:
        edge=byid[r['id']+'_'+which]
        assert edge.get('source')==r[which] and edge.get('target')==r['id']
        assert byid[r['id']+'_c'+which].get('value')==r['c'+which]
        assert r['c'+which] in ['(0,1)','(1,1)','(0,N)','(1,N)']
        assert ('dashed=1' in edge.get('style'))==r['logical']
for name in entities:
    idattrs=[c for c in vertices if c.get('id','').startswith(name+'__id_')]
    assert len(idattrs)==1,(name,'identificador no representado')
    assert 'fontStyle=4;' in idattrs[0].get('style')
assert len([c for c in vertices if c.get('id','').endswith(('_ca','_cb'))])==98
assert not v['shape_overlaps'] and not v['connector_shape_hits']
assert not v['text_overflow'] and not v['outside_canvas']
for s in v['sources']:
    assert hashlib.sha256((repo/s['path']).read_bytes()).hexdigest()==s['sha256']
result=dict(valid=True,pages=1,vertices=len(vertices),connectors=len(connectors),
    relational_entities=24,bridge_relationships=3,physical_fks_covered=48,
    diamonds=len(relations),cardinality_labels=98,source_hashes_unchanged=True,
    local_validation_only=True)
result['cardinalities_recomputed_from_catalog']=True
(p/'VALIDACION_XML.json').write_text(json.dumps(result,indent=2),encoding='utf-8')
print(json.dumps(result,indent=2))
