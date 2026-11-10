"""Inspección estática del DDL versionado. No ejecuta SQL ni abre conexiones."""
from pathlib import Path
import re,json,hashlib
ROOT=Path(__file__).resolve().parents[3];OUT=Path(__file__).resolve().parent
T={};FK=[];sources=[];evolution=[]
def dtype(raw):
 m=re.match(r'(character varying\(\d+\)|varchar\(\d+\)|numeric\(\d+,\s*\d+\)|timestamp with time zone|bigserial|bigint|integer|smallint|boolean|text|jsonb|bytea|timestamptz)',raw,re.I)
 if not m:raise ValueError('Tipo no reconocido: '+raw)
 return m[1].upper().replace('CHARACTER VARYING','VARCHAR').replace('TIMESTAMP WITH TIME ZONE','TIMESTAMPTZ').replace(' ','')
def cols(raw):return [c.strip().strip('"') for c in raw.split(',')]
def column(table,name,raw,origin):
 assert name not in [c['name'] for c in T[table]['columns']],(table,name)
 c=dict(name=name,type=dtype(raw),not_null='NOT NULL' in raw.upper() or 'PRIMARY KEY' in raw.upper(),definition=raw,origin=origin)
 T[table]['columns'].append(c)
 if 'PRIMARY KEY' in raw.upper():T[table]['pk']=[name]
 if 'UNIQUE' in raw.upper():T[table]['unique'].append(dict(name=table+'_'+name+'_key',columns=[name],predicate='',origin=origin))
 ref=re.search(r'REFERENCES (?:public\.)?(\w+)\s*\((\w+)\)',raw,re.I)
 if ref:FK.append(dict(name=table+'_'+name+'_fkey',child=table,column=name,parent=ref[1],parent_column=ref[2],origin=origin,definition=raw,generatedName=True))
# Historical provenance: baseline 000 comes before migrations 001-009.
for file in sorted((ROOT/'backend/internal/platform/postgres/migrations').glob('*.sql')):
 raw=file.read_text(encoding='utf8');origin=str(file.relative_to(ROOT)).replace('\\','/');sql=re.sub(r'--[^\n]*','',raw)
 sources.append(dict(path=origin,sha256=hashlib.sha256(file.read_bytes()).hexdigest()))
 for m in re.finditer(r'CREATE TABLE (?:IF NOT EXISTS )?(?:public\.)?(\w+) \(\n(.*?)\n\);',sql,re.S|re.I):
  name,body=m.groups();assert name not in T,name;T[name]=dict(columns=[],pk=[],unique=[],checks=[],origin=origin)
  for line in body.splitlines():
   cm=re.match(r'\s*(\w+)\s+(.+)',line)
   if cm and cm[1]!='CONSTRAINT':column(name,cm[1],cm[2].rstrip(','),origin)
   elif cm:T[name]['checks'].append(line.strip().rstrip(','))
 for m in re.finditer(r'ALTER TABLE ONLY public\.(\w+)\s+ADD CONSTRAINT (\w+) (PRIMARY KEY|UNIQUE|FOREIGN KEY) \(([^)]+)\)(?: REFERENCES public\.(\w+)\((\w+)\))?([^;]*);',sql):
  table,name,kind,fields,parent,pc,tail=m.groups()
  if kind=='PRIMARY KEY':T[table]['pk']=cols(fields)
  elif kind=='UNIQUE':T[table]['unique'].append(dict(name=name,columns=cols(fields),predicate='',origin=origin))
  else:
   assert len(cols(fields))==1
   FK.append(dict(name=name,child=table,column=fields,parent=parent,parent_column=pc,origin=origin,definition=m[0],generatedName=False))
 for m in re.finditer(r'ALTER TABLE (?:public\.)?(\w+) ADD COLUMN (?:IF NOT EXISTS )?(\w+) ([^;]+);',sql,re.I):column(m[1],m[2],m[3],origin)
 for m in re.finditer(r'ALTER TABLE (?:public\.)?(\w+) ALTER COLUMN (\w+) (SET NOT NULL|DROP NOT NULL|SET DEFAULT [^;]+);',sql,re.I):
  c=next(c for c in T[m[1]]['columns'] if c['name']==m[2]);action=m[3].upper()
  if action=='SET NOT NULL':c['not_null']=True
  elif action=='DROP NOT NULL':c['not_null']=False
  else:c['effectiveDefault']=m[3][12:]
 for m in re.finditer(r'CREATE UNIQUE INDEX (?:IF NOT EXISTS )?(\w+) ON (?:public\.)?(\w+)(?: USING btree)?\s*\(([^)]+)\)([^;]*);',sql,re.I):
  predicate=m[4].strip();T[m[2]]['unique'].append(dict(name=m[1],columns=cols(m[3]),predicate=predicate,origin=origin))
 for table,name in re.findall(r'ALTER TABLE (?:public\.)?(\w+) DROP CONSTRAINT (?:IF EXISTS )?(\w+)',sql,re.I):
  T[table]['checks']=[c for c in T[table]['checks'] if not (isinstance(c,str) and re.search(r'CONSTRAINT '+re.escape(name)+r'\b',c)) and not (isinstance(c,dict) and c.get('name')==name)]
 for m in re.finditer(r'ALTER TABLE (?:public\.)?(\w+) ADD CONSTRAINT (\w+) CHECK \((.*?)(?=;)',sql,re.S|re.I):T[m[1]]['checks'].append(dict(name=m[2],sql=m[3],origin=origin))
 for table in T.values():
  for pk in table['pk']:next(c for c in table['columns'] if c['name']==pk)['not_null']=True
 evolution.append(dict(source=origin,tables=len(T),columns=sum(len(t['columns']) for t in T.values()),foreignKeys=len(FK)))
for i,f in enumerate(sorted(FK,key=lambda f:(f['child'],f['column'])),1):
 f['id']='F%02d'%i;t=T[f['child']];c=next(c for c in t['columns'] if c['name']==f['column']);pt=T[f['parent']]
 assert f['parent_column'] in [c['name'] for c in pt['columns']]
 assert [f['parent_column']]==pt['pk'] or any(u['columns']==[f['parent_column']] and not u['predicate'] for u in pt['unique']),f
 f['nullable']=not c['not_null'];f['individuallyUnique']=t['pk']==[f['column']] or any(u['columns']==[f['column']] and not u['predicate'] for u in t['unique'])
 f['childMultiplicity']='0..1' if f['individuallyUnique'] else '0..N';f['parentMultiplicity']='0..1' if f['nullable'] else '1'
 f['onDelete']=(re.search(r'ON DELETE (RESTRICT|CASCADE|SET NULL|SET DEFAULT|NO ACTION)',f['definition'],re.I).group(1).upper() if re.search(r'ON DELETE (RESTRICT|CASCADE|SET NULL|SET DEFAULT|NO ACTION)',f['definition'],re.I) else 'NO ACTION')
FK.sort(key=lambda f:f['id'])
assert len({(f['child'],f['column']) for f in FK})==len(FK)
model=dict(tables=T,relationships=FK,sources=sources,evolution=evolution,counts=dict(tables=len(T),columns=sum(len(t['columns']) for t in T.values()),foreignKeys=len(FK)),scope='SQL base más migraciones versionadas; no certifica despliegue de una instancia viva')
(OUT/'ESQUEMA_VERIFICADO.json').write_text(json.dumps(model,ensure_ascii=False,indent=2),encoding='utf8')
print(json.dumps(model['counts']))
for f in FK:print(f['id'],f['child']+'.'+f['column'],'->',f['parent']+'.'+f['parent_column'],f['childMultiplicity']+':'+f['parentMultiplicity'])
