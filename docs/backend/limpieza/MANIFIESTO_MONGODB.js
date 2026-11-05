// mongosh, READ ONLY. Diagnosis IDs must come from the same PostgreSQL backup.
const name=process.env.JAC_RESTORE_MONGO_DB;
if(!/^elbravodeljac_restore_\d{14}$/.test(name||''))throw Error('Solo copia restaurada QA');
const diagnoses=JSON.parse(require('fs').readFileSync('/tmp/jac-cleanup-diagnoses.json','utf8'));
if(!Array.isArray(diagnoses)||diagnoses.some(id=>!/^\d+$/.test(id)))throw Error('Manifiesto PostgreSQL inválido');
const source=db.getSiblingDB(name),filter={diagnosticoId:{$in:diagnoses}};
const scans=source.resultados_escaner.find(filter,{_id:1,diagnosticoId:1}).toArray();
const evidence=source.evidencias_tecnicas.find(filter,{_id:1,diagnosticoId:1,gridfsFileId:1,state:1}).toArray();
const files=[],review=[];const hex=value=>EJSON.serialize(value)?.$oid||'';
for(const e of evidence){
 if(!/^[a-f0-9]{24}$/.test(e.gridfsFileId||'')){review.push({evidenceId:hex(e._id),reason:'Sin identificador GridFS válido'});continue;}
 const f=source.getCollection('archivos_tecnicos.files').findOne({_id:ObjectId(e.gridfsFileId)});
 if(!f||f.metadata?.managedBy!=='elbravodeljac'||f.metadata?.diagnosticoId!==e.diagnosticoId||hex(f.metadata?.evidenceId)!==hex(e._id)){
  review.push({evidenceId:hex(e._id),reason:'Propiedad o enlace de archivo no comprobados'});continue;
 }
 files.push({_id:e.gridfsFileId,evidenceId:hex(e._id),diagnosticoId:e.diagnosticoId,chunks:source.getCollection('archivos_tecnicos.chunks').countDocuments({files_id:f._id})});
}
const allFiles=source.getCollection('archivos_tecnicos.files').countDocuments({});
print(JSON.stringify({dryRun:true,database:name,diagnoses,scans:scans.map(x=>({_id:hex(x._id),diagnosticoId:x.diagnosticoId})),evidence:evidence.map(x=>({_id:hex(x._id),diagnosticoId:x.diagnosticoId})),files,requiresReview:review,unselected:{scans:source.resultados_escaner.countDocuments({})-scans.length,evidence:source.evidencias_tecnicas.countDocuments({})-evidence.length,files:allFiles-files.length},writes:0}));
