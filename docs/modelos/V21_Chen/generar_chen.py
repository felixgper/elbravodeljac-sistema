"""Reproduce el Chen localmente desde metadatos sin registros. No conecta a BD.
Requiere Pillow. XML y PNG comparten exactamente las mismas figuras y rutas.
"""
from pathlib import Path
import collections, hashlib, heapq, html, itertools, json, math, re, subprocess
import textwrap, xml.etree.ElementTree as ET
from PIL import Image, ImageDraw, ImageFont

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[2]
BASE = 'El_Bravo_del_JAC_V21_Chen_Completo'
inventory = json.loads((HERE/'inventario_actual.json').read_text(encoding='utf-8'))
tables = inventory['tables']
fks = inventory['relationships']
assert len(tables) == 27 and len(fks) == 48
assert not next(c for c in tables['cotizaciones']['columns'] if c['name']=='id_recepcion')['nullable']
BLUE, LINE, INK = '#3269B8', '#6F8FB7', '#193B66'
W, H, STEP = 12600, 6150, 10
shapes, edges, notes = {}, [], []

# Renglón principal de izquierda a derecha, sin romper el lienzo en páginas.
positions = {
 'contactos_cliente':(0,0), 'historial_propietarios':(1,0),
 'auxilios':(2,0), 'unidades_auxilio':(3,0),
 'resultados_escaner':(4,0), 'evidencias_tecnicas':(5,0),
 'clientes':(0,1), 'vehiculos':(1,1), 'recepciones':(2,1),
 'diagnosticos':(3,1), 'cotizaciones':(4,1), 'ordenes_trabajo':(5,1),
 'pagos':(6,1), 'entregas':(7,1),
 'proveedores':(0,2), 'repuestos':(1,2), 'servicios':(2,2),
 'detalle_cotizacion':(3,2), 'historial_estados_ot':(4,2),
 'asignaciones_ot':(5,2), 'avances_ot':(6,2), 'pruebas_ot':(7,2),
 'empleados':(3,3), 'usuarios':(4,3), 'roles':(5,3), 'permisos':(6,3),
}
centers = {n:(850+c*1500, 1000+r*1300) for n,(c,r) in positions.items()}
names = {n:n.upper().replace('_',' ') for n in positions}
names.update({'resultados_escaner':'RESULTADOS ESCÁNER', 'evidencias_tecnicas':'EVIDENCIAS TÉCNICAS',
 'historial_propietarios':'HISTORIAL PROPIETARIOS', 'historial_estados_ot':'HISTORIAL ESTADOS OT'})

# Una entrada = un óvalo individual. Los campos compuestos documentados mantienen
# su significado de negocio; no se convierte JSON ni caché binaria en tablas.
attributes = {
 'clientes':['id_cliente','tipo_documento','numero_documento','nombre','telefono','direccion','correo'],
 'vehiculos':['id_vehiculo','placa','marca','modelo','anio','kilometraje_actual','observaciones'],
 'contactos_cliente':['id_contacto','nombre','tipo_documento','numero_documento','telefono','relacion','activo'],
 'historial_propietarios':['id_historial','fecha_inicio','fecha_fin','verificado','observacion'],
 'recepciones':['id_recepcion','fecha_ingreso','km_ingreso','sintomas','tipo_atencion','diagnostico_asignado_en','historial_asignaciones'],
 'diagnosticos':['id_diagnostico','fecha','resumen','estado','evaluacion','recomendaciones','observaciones','finalizado_en','actualizado_en','historial'],
 'cotizaciones':['id_cotizacion','estado','total','fecha_emision','aprobada_en','version','aprobacion','historial_revision'],
 'detalle_cotizacion':['id_detalle','descripcion','cantidad','precio_unitario','subtotal','origen_repuesto','tipo_concepto','tipo_repuesto','descuento_unitario'],
 'servicios':['id_servicio','nombre','precio_referencia','activo'],
 'repuestos':['id_repuesto','nombre','stock','precio_referencia','activo'],
 'proveedores':['id_proveedor','nombre','documento','telefono','contacto'],
 'ordenes_trabajo':['id_orden','codigo_ot','estado','fecha_creacion','fecha_inicio','fecha_fin','control_tecnico','control_fecha','control_historial'],
 'asignaciones_ot':['id_asignacion','rol_en_trabajo','fecha_asignacion','activo'],
 'avances_ot':['id_avance','observacion','fecha'],
 'pruebas_ot':['id_prueba','tipo_prueba','resultado','observaciones','fecha'],
 'historial_estados_ot':['id_historial','estado_anterior','estado_nuevo','fecha'],
 'pagos':['id_pago','monto','metodo_pago','fecha_pago','comprobante','revertido_en','rectificacion'],
 'entregas':['id_entrega','fecha_entrega','tipo_receptor','identidad_receptor','autorizacion_verificada','medio_verificacion','fecha_verificacion','receptor_real','conformidad','ficha_tecnica'],
 'auxilios':['id_auxilio','ubicacion','tipo_incidente','descripcion','fecha','estado'],
 'unidades_auxilio':['id_unidad_auxilio','codigo','descripcion','activo'],
 'empleados':['id_empleado','nombre','tipo_empleado','activo'],
 'usuarios':['id_usuario','nombre','usuario','correo','activo'],
 'roles':['id_rol','nombre_rol'], 'permisos':['id_permiso','codigo','descripcion'],
 'resultados_escaner':['id_resultado','equipo','dtcs','mediciones','capturadoEn','registradoEn'],
 'evidencias_tecnicas':['id_evidencia','descripcion','mimeType','sizeBytes','archivo_GridFS','capturadoEn','registradoEn'],
}
labels = {
 'nombre':'nombre', 'sintomas':'motivo de ingreso', 'rol_en_trabajo':'función en trabajo',
 'historial':'historial diagnóstico','historial_revision':'historial revisiones',
 'aprobacion':'aprobación del cliente', 'control_tecnico':'control técnico',
 'control_historial':'historial del control', 'identidad_receptor':'identidad receptor',
 'ficha_tecnica':'ficha técnica emitida', 'receptor_real':'receptor final',
 'mimeType':'formato de archivo','sizeBytes':'tamaño de archivo',
 'capturadoEn':'fecha de captura','registradoEn':'fecha de registro',
 'archivo_GridFS':'archivo técnico GridFS','equipo':'equipo de escáner',
 'dtcs':'códigos DTC','mediciones':'mediciones',
 'autorizacion_verificada':'autorización verificada','diagnostico_asignado_en':'fecha asignación técnica',
 'historial_asignaciones':'historial asignaciones',
}
def label(a): return labels.get(a,a.replace('_',' '))
def shape(id,kind,text,cx,cy,w,h,**kw):
    assert id not in shapes
    shapes[id]=dict(id=id,kind=kind,text=text,cx=cx,cy=cy,w=w,h=h,**kw)
    return id
def bbox(s,pad=0):
    return (s['cx']-s['w']/2-pad,s['cy']-s['h']/2-pad,
            s['cx']+s['w']/2+pad,s['cy']+s['h']/2+pad)
def overlap(a,b): return a[0]<b[2] and a[2]>b[0] and a[1]<b[3] and a[3]>b[1]
def ellipse_boundary(cx,cy,w,h,towards):
    dx,dy=towards[0]-cx,towards[1]-cy
    k=1/math.sqrt((dx/(w/2))**2+(dy/(h/2))**2)
    return (cx+dx*k,cy+dy*k)

for n,(cx,cy) in centers.items():
    shape(n,'entity',names[n],cx,cy,280,80,mongo=n.startswith(('resultados_','evidencias_')))
    aa=attributes[n]
    top=aa[:math.ceil(len(aa)/2)]; bottom=aa[len(top):]
    for side,items in [(-1,top),(1,bottom)]:
        count=len(items)
        for i,a in enumerate(items):
            x=cx+(i-(count-1)/2)*220
            # Convex fan: none of the outer attribute lines passes through an inner oval.
            y=cy+side*(200+100*(min(i,count-1-i)))
            aid=shape(n+'__'+a,'attribute',label(a),x,y,200,60,key=a.startswith('id_'),field=a,owner=n)
            start=(cx,cy+side*40)
            end=ellipse_boundary(x,y,200,60,start)
            edges.append(dict(id='attr_'+aid,source=n,target=aid,points=[start,end],attribute=True,dashed=False))

# Every physical FK is mapped once. N:M bridges become Chen diamonds, not tables.
verbs = {
 ('vehiculos','id_cliente'):'POSEE', ('contactos_cliente','id_cliente'):'DESIGNA CONTACTO',
 ('recepciones','id_vehiculo'):'INGRESA', ('recepciones','id_cliente_atencion'):'SOLICITA ATENCIÓN',
 ('recepciones','id_recepcionista'):'REGISTRA INGRESO',
 ('recepciones','id_mecanico_diagnostico'):'ASUME DIAGNÓSTICO',
 ('recepciones','diagnostico_asignado_por'):'ASIGNA DIAGNÓSTICO',
 ('diagnosticos','id_recepcion'):'RECIBE DIAGNÓSTICO', ('diagnosticos','id_tecnico'):'DIAGNOSTICA',
 ('cotizaciones','id_recepcion'):'ORIGINA COTIZACIÓN', ('cotizaciones','id_diagnostico'):'SUSTENTA',
 ('cotizaciones','id_cotizacion_base'):'TIENE AMPLIACIÓN',
 ('detalle_cotizacion','id_cotizacion'):'CONTIENE CONCEPTO',
 ('detalle_cotizacion','id_servicio'):'REFERENCIA SERVICIO', ('detalle_cotizacion','id_repuesto'):'REFERENCIA REPUESTO',
 ('ordenes_trabajo','id_cotizacion'):'AUTORIZA ORDEN', ('ordenes_trabajo','id_usuario_creacion'):'CREA ORDEN',
 ('ordenes_trabajo','control_autor'):'REGISTRA CONTROL',
 ('asignaciones_ot','id_orden'):'TIENE ASIGNACIÓN', ('asignaciones_ot','id_usuario'):'DESEMPEÑA FUNCIÓN',
 ('avances_ot','id_orden'):'REGISTRA AVANCE', ('avances_ot','id_usuario'):'INFORMA AVANCE',
 ('pruebas_ot','id_orden'):'INCLUYE PRUEBA', ('pruebas_ot','id_usuario'):'REALIZA PRUEBA',
 ('historial_estados_ot','id_orden'):'CONSERVA ESTADOS', ('historial_estados_ot','id_usuario'):'CAMBIA ESTADO',
 ('pagos','id_orden'):'RECIBE PAGO', ('pagos','id_usuario_registro'):'REGISTRA PAGO',
 ('entregas','id_orden'):'CULMINA EN ENTREGA', ('entregas','id_usuario_entrega'):'EFECTÚA ENTREGA',
 ('entregas','id_usuario_verificador'):'VERIFICA RECEPTOR', ('entregas','id_contacto_receptor'):'RECIBE ENTREGA',
 ('historial_propietarios','id_cliente'):'FIGURA COMO TITULAR',
 ('historial_propietarios','id_vehiculo'):'CONSERVA TITULARIDAD',
 ('historial_propietarios','id_usuario_registro'):'REGISTRA TITULARIDAD',
 ('auxilios','id_cliente_solicitante'):'SOLICITA AUXILIO', ('auxilios','id_vehiculo'):'REQUIERE AUXILIO',
 ('auxilios','id_usuario_registro'):'REGISTRA AUXILIO', ('auxilios','id_tecnico'):'ATIENDE AUXILIO',
 ('auxilios','id_unidad_auxilio'):'SE DESPLIEGA EN', ('auxilios','id_recepcion'):'PROCEDE DE AUXILIO',
 ('usuarios','id_empleado'):'TIENE CUENTA',
}
bridge = {'usuario_roles':('usuarios','roles','ASUME ROL'),
 'rol_permisos':('roles','permisos','CONCEDE PERMISO'),
 'repuesto_proveedor':('proveedores','repuestos','ABASTECE')}
relations=[]
for f in fks:
    if f['child'] in bridge: continue
    p,c=f['parent'],f['child']
    relations.append(dict(a=p,b=c,verb=verbs[c,f['column']],
        ca='(0,1)' if f['unique'] else '(0,N)', cb='(0,1)' if f['nullable'] else '(1,1)',
        physical=[f['name']], column=c+'.'+f['column'],logical=False))
for t,(a,b,verb) in bridge.items():
    relations.append(dict(a=a,b=b,verb=verb,ca='(0,N)',cb='(0,N)',
        physical=[f['name'] for f in fks if f['child']==t],bridge=t,logical=False))
for n in ['resultados_escaner','evidencias_tecnicas']:
    relations.append(dict(a='diagnosticos',b=n,verb='DOCUMENTA ESCÁNER' if n.startswith('resultados') else 'REÚNE EVIDENCIA',ca='(0,N)',cb='(1,1)',physical=[],logical=True))
    relations.append(dict(a='usuarios',b=n,verb='CAPTURA RESULTADO' if n.startswith('resultados') else 'INCORPORA EVIDENCIA',ca='(0,N)',cb='(1,1)',physical=[],logical=True))
priority=['POSEE','INGRESA','RECIBE DIAGNÓSTICO','SUSTENTA','AUTORIZA ORDEN','RECIBE PAGO','CULMINA EN ENTREGA','DESIGNA CONTACTO','TIENE CUENTA','ASUME ROL','CONCEDE PERMISO','ABASTECE']
relations.sort(key=lambda r:(priority.index(r['verb']) if r['verb'] in priority else 100,
                             math.dist(centers[r['a']],centers[r['b']]),r['verb']))

# Corridors protect complete attribute fans, rather than merely oval outlines.
clouds=[]
for n,(x,y) in centers.items():
    clouds += [(x-680,y-540,x+680,y-145),(x-680,y+145,x+680,y+540)]
baseboxes=[bbox(s,12) for s in shapes.values()]
baseboxes += [(x-520,y-140,x+520,y+140) for x,y in centers.values()]
note_areas=[(100,4900,3800,5700),(6500,420,9100,550)]
reserved=list(note_areas)
candidates=[(x,y) for x in range(400,W-300,100) for y in range(600,H-400,20)
            if not any(a<x<b and c<y<d for a,c,b,d in clouds)]
for i,r in enumerate(relations,1):
    r['id']=rid=f'R{i:02}'
    ax,ay=centers[r['a']];bx,by=centers[r['b']]
    ideal=((ax+bx)/2,(ay+by)/2)
    for x,y in sorted(candidates,key=lambda p:abs(p[0]-ideal[0])+abs(p[1]-ideal[1])):
        # Includes endpoint labels, a little whitespace and the optional cost oval.
        box=(x-230,y-75,x+230,y+(190 if r.get('bridge')=='repuesto_proveedor' else 75))
        if any(overlap(box,b) for b in baseboxes+reserved+clouds):continue
        break
    else: raise RuntimeError('No hay espacio para '+rid)
    reserved.append(box)
    r.update(cx=x,cy=y)
    shape(rid,'relationship',rid+'\n'+r['verb'],x,y,160,100,logical=r['logical'])
    shape(rid+'_ca','cardinality',r['ca'],x-170,y-28,76,24)
    shape(rid+'_cb','cardinality',r['cb'],x+170,y-28,76,24)
    if r.get('bridge')=='repuesto_proveedor':
        aid=shape(rid+'_cost','attribute','costo referencia',x,y+140,200,60,owner=rid,field='costo_referencia')
        edges.append(dict(id='attr_'+aid,source=rid,target=aid,points=[(x,y+50),(x,y+110)],attribute=True,dashed=False))

print(f'Figuras: {len(shapes)}; relaciones: {len(relations)}. Enrutando...',flush=True)
# Grid A*: no coincident connector segments, no turns at crossings, no arrowheads.
blocked=set()
def block(box):
    a,b,c,d=box
    for x in range(max(0,math.ceil(a/STEP)),min(W//STEP,math.floor(c/STEP))+1):
        for y in range(max(0,math.ceil(b/STEP)),min(H//STEP,math.floor(d/STEP))+1):
            blocked.add((x,y))
for box in clouds:block(box)
for box in note_areas:block(box)
for s in shapes.values():
    if s['kind']=='relationship':block(bbox(s,12))
    elif s['kind']=='cardinality':block(bbox(s,2))
    elif s['kind']=='entity':block(bbox(s,8))
used_segments=set(); used_nodes=collections.defaultdict(set); terminals=set()
def segment(a,b):return tuple(sorted([a,b]))
def route(start,end):
    st=tuple(round(v/STEP) for v in start);en=tuple(round(v/STEP) for v in end)
    assert st not in blocked and en not in blocked,(start,end,'blocked endpoint')
    best={(st,0):0}; parent={}; heap=[(0,0,st,0)]
    final=None
    while heap:
        _,cost,p,old=heapq.heappop(heap)
        state=(p,old)
        if cost!=best.get(state):continue
        if p==en:final=state;break
        for dx,dy,direction in [(1,0,1),(-1,0,1),(0,1,2),(0,-1,2)]:
            q=(p[0]+dx,p[1]+dy)
            if not(10<q[0]<W//STEP-10 and 25<q[1]<H//STEP-12):continue
            if q in blocked or segment(p,q) in used_segments:continue
            if q in terminals and q!=en:continue
            here=used_nodes.get(p,set());there=used_nodes.get(q,set())
            if here and (direction in here or len(here)>1 or (old and direction!=old)):continue
            if there and (direction in there or len(there)>1):continue
            nc=cost+10+(5 if old and old!=direction else 0)+(60 if there else 0)
            ns=(q,direction)
            if nc<best.get(ns,10**12):
                best[ns]=nc;parent[ns]=state
                h=10*(abs(q[0]-en[0])+abs(q[1]-en[1]))
                heapq.heappush(heap,(nc+h,nc,q,direction))
    if final is None:raise RuntimeError(f'Sin ruta {start} -> {end}')
    pp=[];state=final
    while True:
        pp.append(state[0])
        if state==(st,0):break
        state=parent[state]
    pp.reverse()
    for a,b in zip(pp,pp[1:]):
        used_segments.add(segment(a,b))
        dr=1 if a[1]==b[1] else 2
        used_nodes[a].add(dr);used_nodes[b].add(dr)
    # Keep only bends in editable Draw.io polylines.
    out=[]
    for p in pp:
        p=(p[0]*STEP,p[1]*STEP)
        if len(out)>=2 and ((out[-2][0]==out[-1][0]==p[0]) or (out[-2][1]==out[-1][1]==p[1])):out[-1]=p
        else:out.append(p)
    return out

incidence=collections.defaultdict(list)
for r in relations:
    for which in ['a','b']:incidence[r[which]].append((r,which))
ports={}
for entity,items in incidence.items():
    x,y=centers[entity]
    # Balance heavily connected USUARIOS; all ports are still individually glued.
    sides={-1:[],1:[]}
    for r,which in sorted(items,key=lambda rw:rw[0]['cy']):
        desired=-1 if r['cx']<x else 1
        if len(sides[desired])>=11:desired=-desired
        sides[desired].append((r,which))
    for side,items in sides.items():
        assert len(items)<=13,(entity,len(items))
        for i,(r,which) in enumerate(items):
            off=(i-(len(items)-1)/2)*20
            off=round(off/20)*20
            # Ensure uniqueness when rounding a half-step.
            off=(i-len(items)//2)*20
            anchor=(x+side*140,y+(i-(len(items)-1)/2)*min(5,60/max(1,len(items)-1)))
            portal=(x+side*360,y+off)
            ports[r['id'],which]=(anchor,portal,side)

for r in relations:
    for which,sign in [('a',-1),('b',1)]:
        terminals.add(tuple(round(v/STEP) for v in ports[r['id'],which][1]))
        terminals.add((round((r['cx']+sign*120)/STEP),round(r['cy']/STEP)))
for i,r in enumerate(relations):
    for which,sign in [('a',-1),('b',1)]:
        anchor,portal,side=ports[r['id'],which]
        end=(r['cx']+sign*80,r['cy'])
        goal=(r['cx']+sign*120,r['cy'])
        pp=route(portal,goal)
        edges.append(dict(id=r['id']+'_'+which,source=r[which],target=r['id'],
            points=[anchor]+pp+[end],attribute=False,dashed=r['logical'],
            exitX=1 if side==1 else 0,exitY=(anchor[1]-centers[r[which]][1]+40)/80,
            entryX=0 if sign==-1 else 1,entryY=.5,cardinality=r['c'+which]))
    print(r['id']+' '+r['verb'],flush=True)

# All explanatory notes remain editable, in the same page as the model.
def note(id,text,x,y,w,h,size=22):
    shape(id,'note',text,x+w/2,y+h/2,w,h,size=size)
note('title','EL BRAVO DEL JAC · MODELO CONCEPTUAL ENTIDAD–RELACIÓN · CHEN V21',100,65,12100,70,38)
note('subtitle','Una sola página · 24 entidades relacionales + 3 asociaciones N:M · 48 FK cubiertas · Complemento MongoDB separado conceptualmente',100,145,12100,45,25)
note('legend','RECTÁNGULO = entidad    ÓVALO = atributo    ROMBO = relación    Clave subrayada = identificador\nCada par (mín., máx.) se refiere a la entidad del conector rotulado; N = varios. Los sobrepasos de líneas NO son uniones. Sin flechas.\nLos mínimos generales reflejan restricciones actuales; las condiciones por estado y las reglas de Go se explican en el informe.',100,230,12100,130,23)
note('mongo_note','COMPLEMENTO DOCUMENTAL · MongoDB\nResultados y evidencias: referencias lógicas discontinuas a DIAGNÓSTICOS y USUARIOS. No hay FK entre motores. Archivos: GridFS.',6560,430,2500,110,23)
note('rules','REGLAS DEL TALLER\nServicio directo: sin diagnóstico; siempre cotización aprobada antes de crear la orden.\nAtención técnica: diagnóstico finalizado por el responsable asignado.\nAsignación: MECÁNICO responsable y AUXILIAR son funciones distintas.\nAuxiliar: consulta con su rol base; revisar excepción de permisos en el informe.\nPagos: anticipos o parciales; entrega de una orden finalizada con saldo neto cero.\nCargos laborales y roles de acceso son independientes.',120,4930,3650,340,25)
note('scope','LECTURA DEL MODELO\nN:M: ASUME ROL, CONCEDE PERMISO y ABASTECE (con costo).\nLos detalles, asignaciones y eventos tienen identidad e información propia.\nUn vínculo opcional no equivale a un dato inventado: los mínimos 0 son deliberados.\nLas asociaciones de responsabilidad conservan USUARIOS; EMPLEADOS no sustituye sus referencias históricas.\nAprobación, control técnico, conformidad e historiales son atributos compuestos.\nLas cuentas legadas o técnicas pueden permanecer sin empleado.',120,5320,3650,340,25)
note('footer','Fuentes: respaldo estructural + migraciones 001–007 + reglas Go · Rama base 7dd2738 · 29/10/2026 · Solo documentación; sin cambios al sistema ni a los datos.',100,5980,12100,60,22)

# Editable native mxGraph XML. No bitmap is embedded in the drawing.
mxfile=ET.Element('mxfile',host='app.diagrams.net',version='24.7.17',type='device')
diagram=ET.SubElement(mxfile,'diagram',id='chen-v21-unico',name='Chen completo · lienzo único')
model=ET.SubElement(diagram,'mxGraphModel',dx=str(W),dy=str(H),grid='1',gridSize='10',guides='1',tooltips='1',connect='1',arrows='0',fold='1',page='1',pageScale='1',pageWidth=str(W),pageHeight=str(H),math='0',shadow='0',background='#FFFFFF')
root=ET.SubElement(model,'root');ET.SubElement(root,'mxCell',id='0');ET.SubElement(root,'mxCell',id='1',parent='0')
for edge in edges:
    style='edgeStyle=none;rounded=0;html=1;endArrow=none;startArrow=none;strokeColor='+LINE+';strokeWidth=1.4;jumpStyle=arc;jumpSize=8;'
    if edge['dashed']:style+='dashed=1;dashPattern=6 4;'
    if not edge['attribute']:
        for p in ['exitX','exitY','entryX','entryY']:style+=p+'='+str(edge[p])+';'
        style+='exitPerimeter=1;entryPerimeter=1;'
    else:
        src,tgt=shapes[edge['source']],shapes[edge['target']]
        p,q=edge['points'][0],edge['points'][-1]
        for name,value in [('exitX',(p[0]-src['cx'])/src['w']+.5),('exitY',(p[1]-src['cy'])/src['h']+.5),('entryX',(q[0]-tgt['cx'])/tgt['w']+.5),('entryY',(q[1]-tgt['cy'])/tgt['h']+.5)]:
            style+=name+'='+str(value)+';'
        style+='exitPerimeter=0;entryPerimeter=0;'
    cell=ET.SubElement(root,'mxCell',id=edge['id'],style=style,edge='1',parent='1',source=edge['source'],target=edge['target'])
    geo=ET.SubElement(cell,'mxGeometry',relative='1',**{'as':'geometry'})
    arr=ET.SubElement(geo,'Array',**{'as':'points'})
    for x,y in edge['points'][1:-1]:ET.SubElement(arr,'mxPoint',x=str(round(x,3)),y=str(round(y,3)))
for s in shapes.values():
    kind=s['kind'];text=html.escape(s['text']).replace('\n','<br>')
    if kind in ['entity','relationship','attribute']:
        style=f'html=1;whiteSpace=wrap;fillColor={BLUE};strokeColor={BLUE};fontColor=#FFFFFF;fontFamily=Arial;fontSize='+('20' if kind=='entity' else '16')+';align=center;verticalAlign=middle;'
        if kind=='entity':style+='rounded=0;fontStyle=1;'
        if kind=='relationship':style+='rhombus;perimeter=rhombusPerimeter;'
        if kind=='attribute':style+='ellipse;perimeter=ellipsePerimeter;'
        if s.get('key'):style+='fontStyle=4;'
    else:
        style='text;html=1;whiteSpace=wrap;strokeColor=none;fillColor='+('#FFFFFF' if kind=='cardinality' else 'none')+f';fontColor={INK};fontFamily=Arial;align='+('center' if kind=='cardinality' else 'left')+';verticalAlign=middle;fontSize='+str(s.get('size',17))+';'
    cell=ET.SubElement(root,'mxCell',id=s['id'],value=text,style=style,vertex='1',parent='1')
    a,b,c,d=bbox(s);ET.SubElement(cell,'mxGeometry',x=str(round(a,3)),y=str(round(b,3)),width=str(s['w']),height=str(s['h']),**{'as':'geometry'})
ET.indent(mxfile)
ET.ElementTree(mxfile).write(HERE/(BASE+'.drawio'),encoding='utf-8',xml_declaration=True)

fontdir=Path('C:/Windows/Fonts')
def font(size,bold=False):return ImageFont.truetype(str(fontdir/('arialbd.ttf' if bold else 'arial.ttf')),size)
im=Image.new('RGB',(W,H),'white');draw=ImageDraw.Draw(im)
def polyline(pp,color=LINE,width=2,dashed=False):
    for a,b in zip(pp,pp[1:]):
        if not dashed:draw.line([a,b],fill=color,width=width);continue
        length=math.dist(a,b)
        for st in range(0,math.ceil(length),15):
            en=min(length,st+9)
            draw.line([(a[0]+(b[0]-a[0])*st/length,a[1]+(b[1]-a[1])*st/length),(a[0]+(b[0]-a[0])*en/length,a[1]+(b[1]-a[1])*en/length)],fill=color,width=width)

drawn=[]; crossings=[]
for e in edges:
    polyline(e['points'],dashed=e['dashed'])
    if e['attribute']:continue
    for a,b in zip(e['points'],e['points'][1:]):
        if a[0]!=b[0] and a[1]!=b[1]:continue
        for c,d in drawn:
            if a[1]==b[1] and c[0]==d[0] and min(a[0],b[0])+9<c[0]<max(a[0],b[0])-9 and min(c[1],d[1])+9<a[1]<max(c[1],d[1])-9:
                x,y=c[0],a[1];crossings.append((x,y));draw.rectangle((x-8,y-8,x+8,y+3),fill='white');draw.arc((x-8,y-8,x+8,y+8),180,360,fill=LINE,width=2)
            elif a[0]==b[0] and c[1]==d[1] and min(c[0],d[0])+9<a[0]<max(c[0],d[0])-9 and min(a[1],b[1])+9<c[1]<max(a[1],b[1])-9:
                x,y=a[0],c[1];crossings.append((x,y));draw.rectangle((x-3,y-8,x+8,y+8),fill='white');draw.arc((x-8,y-8,x+8,y+8),270,450,fill=LINE,width=2)
        drawn.append((a,b))

def text_lines(text,ff,maxwidth):
    lines=[]
    for para in text.split('\n'):
        if not para:lines.append('');continue
        line=''
        for word in para.split():
            test=(line+' '+word).strip()
            if draw.textlength(test,font=ff)>maxwidth and line:lines.append(line);line=word
            else:line=test
        lines.append(line)
    return lines
text_errors=[]
for s in shapes.values():
    bb=bbox(s);kind=s['kind']
    if kind=='attribute':draw.ellipse(bb,fill=BLUE)
    elif kind=='entity':draw.rectangle(bb,fill=BLUE)
    elif kind=='relationship':draw.polygon([(s['cx'],bb[1]),(bb[2],s['cy']),(s['cx'],bb[3]),(bb[0],s['cy'])],fill=BLUE)
    elif kind=='cardinality':draw.rectangle(bb,fill='white')
    size=s.get('size',20 if kind=='entity' else 16 if kind in ['attribute','relationship'] else 17)
    maxwidth=s['w']-(65 if kind=='relationship' else 20)
    while True:
        ff=font(size,kind=='entity');lines=text_lines(s['text'],ff,maxwidth)
        lh=size+5
        if len(lines)*lh<=s['h']-6 or size<=13:break
        size-=1
    if len(lines)*lh>s['h']-6:text_errors.append(s['id'])
    y=s['cy']-len(lines)*lh/2
    for line in lines:
        width=draw.textlength(line,font=ff)
        x=bb[0]+5 if kind=='note' else s['cx']-width/2
        draw.text((x,y),line,font=ff,fill=INK if kind in ['note','cardinality'] else 'white')
        if s.get('key'):draw.line((x,y+size+2,x+width,y+size+2),fill='white',width=1)
        y+=lh
im.save(HERE/(BASE+'.png'),dpi=(160,160),optimize=True)
im.resize((2520,1230),Image.Resampling.LANCZOS).save(HERE/'vista_general.png')
for name,box in {'detalle_personal':(4550,4350,10800,5570),'detalle_flujo':(350,1850,12350,2860),'detalle_diagnostico':(3300,1850,7900,2800),'detalle_mongo':(6200,420,9200,1550)}.items():
    im.crop(box).save(HERE/(name+'.png'))

def intersects_segment_box(a,b,box):
    # Liang–Barsky against a slightly inset rectangle (no false corner tangency).
    x0,y0,x1,y1=box;t0,t1=0.,1.;dx=b[0]-a[0];dy=b[1]-a[1]
    for p,q in [(-dx,a[0]-x0),(dx,x1-a[0]),(-dy,a[1]-y0),(dy,y1-a[1])]:
        if p==0:
            if q<0:return False
        else:
            u=q/p
            if p<0:t0=max(t0,u)
            else:t1=min(t1,u)
    return t0<t1 and t0<1 and t1>0

overlaps=[];hits=[]
ss=[s for s in shapes.values() if s['kind']!='note']
for a,b in itertools.combinations(ss,2):
    if overlap(bbox(a,-1),bbox(b,-1)):overlaps.append([a['id'],b['id']])
for e in edges:
    for s in shapes.values():
        if s['id'] in [e['source'],e['target']]:continue
        box=bbox(s,-3)
        if any(intersects_segment_box(a,b,box) for a,b in zip(e['points'],e['points'][1:])):hits.append([e['id'],s['id']])
adj=collections.defaultdict(set)
for e in edges:adj[e['source']].add(e['target']);adj[e['target']].add(e['source'])
seen=set();stack=['clientes']
while stack:
    p=stack.pop()
    if p in seen:continue
    seen.add(p);stack.extend(adj[p]-seen)
physical=[p for r in relations for p in r['physical']]
outside=[s['id'] for s in shapes.values() if bbox(s)[0]<0 or bbox(s)[1]<0 or bbox(s)[2]>W or bbox(s)[3]>H]
sources=[]
for p in [REPO/'database/postgresql/esquema_actual.sql',*sorted((REPO/'backend/internal/platform/postgres/migrations').glob('*.sql'))]:
    sources.append({'path':p.relative_to(REPO).as_posix(),'sha256':hashlib.sha256(p.read_bytes()).hexdigest()})
validation=dict(page_count=len(mxfile.findall('diagram')),tables=27,foreign_keys=48,
 relational_entities=24,mongo_entities=2,nm_relationships=3,total_diamonds=len(relations),
 attributes=sum(s['kind']=='attribute' for s in shapes.values()),editable_vertices=len(shapes),
 connectors=len(edges),all_entities_connected=set(positions)<=seen,
 physical_fk_coverage=len(set(physical)),duplicated_fk_coverage=len(physical)-len(set(physical)),
 cardinality_labels=2*len(relations),shape_overlaps=overlaps,connector_shape_hits=hits,
 text_overflow=text_errors,outside_canvas=outside,crossing_bridges=len(crossings),
 canvas=[W,H],png=[W,H],sources=sources)
(HERE/'VALIDACION.json').write_text(json.dumps(validation,ensure_ascii=False,indent=2),encoding='utf-8')
(HERE/'relaciones_chen.json').write_text(json.dumps(relations,ensure_ascii=False,indent=2),encoding='utf-8')
print(json.dumps({k:v for k,v in validation.items() if k!='sources'},ensure_ascii=False,indent=2),flush=True)
assert not overlaps and not hits and not text_errors and not outside
assert len(set(physical))==48 and len(physical)==48 and set(positions)<=seen
assert len(mxfile.findall('diagram'))==1

report = '''# El Bravo del JAC — Modelo conceptual Chen V21

**29/10/2026. Una página, un diagrama, completamente editable.**

Rama documental: `docs/modelo-conceptual-chen-v21`, creada desde
`backend/arquitectura-go-mongodb`, commit `7dd2738`, que ya incluye el merge
de EMPLEADOS (`ac0b12a`). El árbol estaba limpio antes de crear la rama.
Los V19/V20 permanecen sin cambios en `docs/propuesta-modelo-v20`, commit
`086f166`. Se leyeron desde Git sin mezclar ramas ni copiar sus cambios.

## Archivos y apertura

- `El_Bravo_del_JAC_V21_Chen_Completo.drawio`: abrir en Draw.io mediante
  **Archivo → Abrir desde → Dispositivo**. Tiene exactamente una página,
  «Chen completo · lienzo único». Cada figura, texto y conector es editable.
- `El_Bravo_del_JAC_V21_Chen_Completo.png`: vista íntegra de 12600 × 6150 píxeles.
  Ampliar para leer los atributos; no está concebido para reducirse a una hoja A4.
- `vista_general.png` y los PNG `detalle_*`: recortes de revisión visual del
  MISMO lienzo. No son páginas adicionales ni partes necesarias para completarlo.
- `generar_chen.py`, `inventario_actual.json`, `relaciones_chen.json` y
  `VALIDACION.json`: reproducción, cobertura y comprobaciones locales.

El generador utiliza Pillow y solo lee el inventario guardado: no conecta a
bases ni ejecuta migraciones. No incorpora un PNG dentro del `.drawio`.

## Fuentes contrastadas

Se revisaron el respaldo estructural `database/postgresql/esquema_actual.sql`,
las siete migraciones 001–007, V19/V20 y su inventario, y los archivos Go
`workshop_writes.go`, `diagnosis_workflow.go`, `technical_sheet.go`,
`employees.go`, `identity.go`, `technical.go` y el dominio MongoDB.
Las reglas documentadas se contrastaron con la implementación, no solo con
los nombres de las FK. Se consultó únicamente el catálogo de estructura de
`elbravodeljac_dev`, PostgreSQL aislado: **27 tablas, 201 columnas y 48 FK**.
No se leyeron registros personales. Los hashes del SQL consultado están en
`VALIDACION.json`; no se modificó ninguno.

## Cómo explicar el dibujo

Rectángulos azules: entidades. Óvalos individuales: atributos. Identificadores
subrayados: claves de identidad. Rombos: asociaciones verbales. No contiene
tipos SQL, listas de columnas dentro de rectángulos, patas de cuervo ni UML.

Cada par `(mínimo,máximo)` está junto al extremo del conector que llega al rombo
y se refiere a la **entidad conectada por esa línea**: cuántas participaciones
puede tener una instancia de esa entidad en esa relación. Los códigos R01…R49
permiten contrastar los dos extremos con la tabla de este informe. Las líneas
no tienen flechas; sus sobrepasos indican cruces sin unión.

Ejemplo: en POSEE, CLIENTES `(0,N)` significa que un cliente puede no tener
vehículos registrados o tener varios; VEHÍCULOS `(1,1)` exige exactamente
un titular actual. En TIENE CUENTA hay `(0,1)` en ambos extremos. En las
relaciones N:M ambos extremos admiten `(0,N)`.

Los mínimos globales conservadores evitan inventar obligatoriedad para datos
históricos. Una FK obligatoria exige un padre al hijo; no exige hijos a cada
padre. Una FK única limita al padre a un hijo. Los mínimos adicionales
condicionales del negocio se explican más abajo. No se presenta `(1,N)` como
restricción universal si la base y sus registros legados no lo garantizan.

## Cobertura de las 27 tablas físicas

Las tres tablas puente se transforman en relaciones N:M. Las otras 24 conservan
entidad propia. Los detalles, asignaciones y eventos no se colapsan en N:M:
tienen identidad, fechas, estados o información que necesita conservarse.

| Tabla física | Representación Chen | Motivo |
|---|---|---|
'''
for name in sorted(tables):
    if name in bridge:
        r=next(r for r in relations if r.get('bridge')==name)
        motive='Asociación con costo de referencia; su ID físico es una clave técnica.' if name=='repuesto_proveedor' else 'Puente sin atributos propios; su identidad es el par de participantes.'
        rep=f"Relación N:M **{r['verb']}** ({r['id']})"
    else:
        rep='Entidad **'+names[name]+'**'
        motive='Entidad asociativa con atributos propios.' if name in ['detalle_cotizacion','asignaciones_ot','historial_propietarios'] else 'Entidad de negocio o evento con identidad propia.'
    report+=f'| `{name}` | {rep} | {motive} |\n'
report+='''
`repuesto_proveedor.costo_referencia` es un óvalo unido a ABASTECE. No se
pierde el costo al sustituir la tabla por la asociación. El ID sustituto `id`
es exclusivamente físico; la pareja proveedor–repuesto es única.

## Relaciones y cardinalidades completas

En esta tabla, el par de cada columna pertenece a la entidad indicada en ella.
Las 45 asociaciones relacionales cubren las 48 FK sin omisiones ni duplicaciones;
cuatro asociaciones documentales son lógicas y se diferencian por línea discontinua.
La numeración se ordena por claridad visual, no por motor.

| Código | Relación | Entidad A y participación | Entidad B y participación | Origen |
|---|---|---|---|---|
'''
for r in relations:
    origin='Referencia lógica MongoDB, sin FK' if r['logical'] else ('Puente `'+r['bridge']+'`' if r.get('bridge') else '`'+r['column']+'`')
    report+=f"| {r['id']} | {r['verb']} | {names[r['a']]} **{r['ca']}** | {names[r['b']]} **{r['cb']}** | {origin} |\n"
report+='''
El mapa exacto de nombres de restricciones está en `relaciones_chen.json`.
Las tres N:M representan seis FK; las otras 42 asociaciones relacionales
representan una FK cada una: 6 + 42 = **48**. Sumadas las cuatro lógicas MongoDB,
el dibujo tiene **49 rombos y 98 etiquetas de participación**.

## Atributos y nivel conceptual

No se duplican como óvalos las columnas que ya representan una asociación.
Los atributos con identidad propia se subrayan. Los códigos de negocio únicos
adicionales (placa, código OT, documento, nombre de rol, etc.) siguen siendo
atributos; que no estén subrayados no elimina sus restricciones físicas.

El modelo muestra 154 óvalos: 140 corresponden a entidades relacionales,
13 al complemento MongoDB y uno al costo de abastecimiento. Los objetos
compuestos se resumen por su significado para mantener legible todo el lienzo:

- **Aprobación del cliente:** nombre, medio, fecha, versión e importes/conceptos
  aprobados; `aprobacion`. El cliente del vehículo llega a través de RECEPCIONES,
  pero el aprobante consignado no es una FK independiente a CLIENTES.
- **Control técnico:** inspecciones, estados Bien/Atención/Urgente/Sin evaluar,
  neumáticos, fluidos, observaciones, mantenimiento recomendado y ejecución de
  conceptos; `control_tecnico`. Se conserva su historial, autor y fecha.
- **Identidad receptor:** nombre, tipo/número de documento y teléfono históricos
  (`*_receptor_snap`). **Receptor final** y **conformidad** representan sus
  objetos registrados. No se inventan firmas ni personas receptoras.
- **Ficha técnica emitida:** copia de los datos y documento conservado
  (`ficha_snapshot`, `ficha_pdf`). No se crea una nueva tabla conceptual de fichas.
- **Historiales:** versiones anteriores de asignación, diagnóstico, cotización
  y control almacenadas en sus agregados. El historial de estados de OT sí
  conserva su entidad porque es una tabla/evento independiente.
- **Escáner:** equipo (nombre/versión), conjunto de códigos DTC (código y
  descripción) y mediciones; se muestran como atributos de contenido documental.

Estos óvalos representan contenido agregado, no un desglose de todos los
subcampos de almacenamiento. No son rectángulos con columnas. Los anexos
documentan la correspondencia de **cada una de las 201 columnas** para que los
campos técnicos no se confundan con conceptos faltantes.

Se omiten deliberadamente hashes de credenciales/contenido, claves de reintento
y contadores de concurrencia. Son mecanismos de almacenamiento o seguridad,
no nuevas entidades. El nombre de USUARIOS se conserva por compatibilidad con
legado; cuando hay EMPLEADO vinculado, su nombre laboral es el principal y los
triggers 007 los sincronizan. No se incluyen datos personales ni contraseñas.

## Reglas del taller y límites de la cardinalidad general

1. **Servicio directo:** RECEPCIÓN de tipo Directo no tiene DIAGNÓSTICO. La
   relación ORIGINA COTIZACIÓN permite cotizar directamente desde la recepción;
   no se sintetizan diagnósticos. En una atención técnica, cotizar exige el
   diagnóstico finalizado y perteneciente a esa misma recepción. Así, SUSTENTA
   tiene `(0,1)` desde cotización globalmente, con obligatoriedad condicional
   para tipo Diagnostico (migraciones 003/004).
2. **Responsables del diagnóstico:** ASUME DIAGNÓSTICO es la cuenta asignada a
   RECEPCIÓN; ASIGNA DIAGNÓSTICO es la cuenta que hizo la asignación;
   DIAGNOSTICA es la cuenta responsable del diagnóstico. No son tres empleados
   nuevos. Hay como máximo un diagnóstico por recepción, por índice único.
   La API acepta recepciones técnicas todavía sin asignar por compatibilidad;
   no admite diagnosticar hasta que exista responsable válido. No se impone un
   mínimo uno incondicional a esa asignación.
3. **Cotizaciones:** toda cotización corresponde a exactamente una recepción.
   Su orden solo se crea con estado aprobado y evidencia de aprobación vigente
   (`CreateOrder`). Una cotización puede no producir OT y como máximo produce
   una, por unicidad. La ampliación es otra cotización relacionada a su base;
   las revisiones de contenido siguen en el atributo historial/versiones.
4. **Conceptos:** las nuevas cotizaciones de la API exigen al menos un detalle.
   La FK, por sí misma, no obliga a todas las cotizaciones históricas a tener
   hijos: el mínimo general no se endurece sin esa garantía. Cada detalle puede
   referenciar un servicio O un repuesto, nunca ambos; un concepto manual
   validado puede no referenciar ninguno (CHECK de migración 001). No se
   dibuja una dependencia obligatoria del catálogo que el sistema ya no exige.
   Cantidad, precio, descuento y subtotal conservan su significado actual.
5. **Asignaciones:** cada registro tiene una orden, una cuenta, una función,
   fecha y vigencia. `rol_en_trabajo` distingue MECANICO y AUXILIAR; no es el
   rol RBAC ni `tipo_empleado`. Go exige mecánico activo para crear OT y valida
   el auxiliar opcional. Se conservan todas las asignaciones históricas, por
   eso el máximo es N. La base no tiene unicidad universal de mecánico activo
   por orden ni obliga mediante FK a una asignación en todo registro legado.
6. **Estados, avances y pruebas:** se conservan como eventos ligados a OT y
   a la cuenta autora. La finalización exige en Go la última prueba satisfactoria,
   posterior a la última reparación. El control técnico tiene su autor/fecha;
   no se completa automáticamente al pagar.
7. **Pagos y entrega:** una orden puede tener varios pagos, anticipos o ninguno
   todavía. El importe neto descuenta reversiones y la entrega exige saldo cero,
   orden finalizada y verificación de identidad/autorización (006). Hay como
   máximo una entrega por orden. ENTREGAS depende de OT, no de cada pago.
   Un contacto receptor es opcional; también se conservan identidad histórica
   y receptor real. No hay FK independiente para cada nombre dentro de una copia.
8. **Propiedad:** POSEE identifica el titular actual. HISTORIAL PROPIETARIOS
   conserva varios períodos por cliente/vehículo. El índice parcial admite
   como máximo un período abierto por vehículo; no vuelve 1:1 todo el historial
   ni garantiza que siempre exista un período abierto para cada vehículo.
9. **Auxilio:** puede registrarse sin cliente, vehículo, técnico, unidad o
   recepción vinculados todavía. El usuario de registro es obligatorio. No se
   impone en el dibujo una asignación inmediata que el esquema permite posponer.
10. **Empleados y acceso:** TIENE CUENTA es opcional y única en ambos lados.
    El cargo no otorga permisos. Las N:M ASUME ROL y CONCEDE PERMISO conservan
    varias asignaciones. Una cuenta legada/técnica puede no tener empleado.
    Estado laboral y habilitación de cuenta son distintos. Las asignaciones
    históricas siguen relacionadas con USUARIOS, tal como está implementado.
    No se confunden filas de ROLES con GRANT del motor PostgreSQL.

## Discrepancias y decisiones para revisión

- **Auxiliar y permisos:** la política base de Auxiliar es de consulta. Sin
  embargo, `canEditOrder` comprueba asignación activa, sin filtrar su función
  MECANICO/AUXILIAR; el endpoint comprueba además `ordenes.editar`. Por tanto,
  una cuenta auxiliar a la que se conceda ese permiso u otro rol habilitante
  puede editar dentro del alcance permitido. La prohibición absoluta por ser
  auxiliar **no está garantizada** por esa función. El control técnico sí
  comprueba MECANICO o permiso de asignación. El dibujo no cambia la seguridad:
  esta diferencia se presenta para revisión y no se corrige en una rama documental.
- **V19:** refleja 26 tablas, 160 columnas, 42 FK; no incluye EMPLEADOS ni las
  ampliaciones posteriores de atención directa, control, asignación y pagos.
- **V20:** incluyó 27/201/48 como propuesta. EMPLEADOS ya está implementado
  por 007; dejó de ser solo una hipótesis para este V21. La tabla de cuentas
  mantiene el vínculo opcional, no una supuesta obligación global de cuenta humana.
- **Error de nulabilidad del inventario V20:** `cotizaciones.id_recepcion`
  aparecía nullable a pesar de que 003 ejecuta SET NOT NULL. V21 lo corrige a
  `(1,1)` desde COTIZACIONES en ORIGINA COTIZACIÓN. El SQL original no se altera.
- **Unicidades revalidadas:** se comprobaron también los índices únicos separados
  (diagnóstico por recepción, OT por cotización y entrega por OT), no solo
  constraints. Se conservan esos máximos del V20; no son restricciones nuevas.
- **Catálogos/vistas:** marcas, datos del taller y movimientos netos son vistas,
  no tres nuevas entidades persistentes. Especialidad de proveedor no se inventa.
- **MongoDB:** cada resultado/evidencia referencia un diagnóstico y una cuenta
  de captura por identificadores lógicos. Go comprueba existencia/autorización
  en PostgreSQL, pero MongoDB no impone una FK entre motores. Una escritura
  externa que eluda Go podría dejar referencias huérfanas: no se promete
  integridad referencial distribuida automática. GridFS es almacenamiento
  técnico del archivo, no otro cliente, vehículo, orden o pago duplicado.

## Validación y revisión visual

El archivo XML se volvió a analizar y se comprobaron IDs únicos, extremos
existentes, geometrías, una sola página, conectividad y cobertura exacta.
Los 98 pares se derivaron de nulabilidad y unicidad reales, más las N:M y
referencias documentales explícitas. La cobertura se puede repetir con el
inventario sin conectarse a PostgreSQL. No se ejecutaron pruebas de Go ni
migraciones: el sistema no cambió en esta tarea.

Se revisaron visualmente la vista íntegra y ampliaciones de recepción/diagnóstico/
cotización, empleados/cuentas/roles, MongoDB y el flujo de pago/entrega.
Se ajustaron el espacio de salida de los conectores y sus rutas antes de cerrar
la entrega. No hay figuras superpuestas, conectores atravesando figuras, textos
fuera de sus cajas ni objetos cortados por el lienzo, según las verificaciones
geométricas. Las relaciones de auditoría hacen que el grafo no sea una cadena:
existen cruces, representados con sobrepasos, sin unir conectores distintos.
No se afirma que el dibujo sea un grafo plano sin cruces.

**Validación Draw.io realizada: local, del XML nativo mxGraph.** No se encontró
Draw.io Desktop en las rutas habituales y no se subió el modelo a un servicio externo.
No se presenta esta validación como una apertura interactiva del editor. El
archivo se entrega sin comprimir, con figuras nativas y sin recursos remotos.

## Correspondencia de todas las columnas

Esta tabla permite revisar qué se visualiza y qué se abstrae como mecanismo
técnico. No se elimina ninguna columna del sistema.

| Tabla | Óvalos directos | Relaciones (no repetir como óvalos) | Agrupación conceptual / campos técnicos |
|---|---|---|---|
'''
for n,t in sorted(tables.items()):
    fkcols={f['column'] for f in fks if f['child']==n}
    direct=set(attributes.get(n,[]))
    cols=[c['name'] for c in t['columns']]
    attr=[c for c in cols if c in direct and c not in fkcols]
    rel=[c for c in cols if c in fkcols]
    other=[c for c in cols if c not in attr and c not in rel]
    if n=='repuesto_proveedor':attr=['costo_referencia'];other=['id (identidad técnica de la asociación)']
    report+='| `'+n+'` | '+', '.join('`'+c+'`' for c in attr)+' | '+', '.join('`'+c+'`' for c in rel)+' | '+(', '.join('`'+c+'`' for c in other) or '—')+' |\n'
report+='''
Los cuatro campos `*_receptor_snap` se agrupan como identidad receptor; los
dos `ficha_*` como ficha técnica emitida. `password_hash` se mantiene en el
sistema y se omite como detalle criptográfico del modelo conceptual.
`request_id`, `contenido_hash`, `asignacion_version`, `control_version`,
`flujo_version` y la `version` técnica del diagnóstico son controles técnicos.
La versión de cotización sí se muestra porque identifica lo autorizado.

**Seguridad de la entrega:** solo archivos bajo `docs/modelos/V21_Chen/`.
Sin cambios a Go, frontend, PostgreSQL, MongoDB o migraciones. Sin borrados,
recreación de bases, commit, push o merge. Los diagramas V19/V20 conservan sus blobs Git.
'''
(HERE/'INFORME_CHEN_V21.md').write_text(report,encoding='utf-8')
