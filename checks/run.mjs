import {spawnSync} from 'node:child_process';
const suites=[['verify.mjs','FIXTURE HISTÓRICO'],['workflow.mjs','FIXTURE HISTÓRICO'],['edge-cases.mjs','FIXTURE/UNIDAD'],['real-api.mjs','UNIDAD: fetch simulado'],['report-periods.mjs','UNIDAD'],['diagnosis-text.mjs','UNIDAD'],['isolation.mjs','UNIDAD: guardas sin bases'],['prototype.mjs','HISTÓRICO: prototipo']];
for(const [file,kind] of suites){
 console.log(`[${kind}] ${file}`);
 const legacy=['verify.mjs','workflow.mjs','edge-cases.mjs','report-periods.mjs'].includes(file);
 const options=legacy?['--import','./checks/fixtures/legacy-demo-loader.mjs']:[];
 const result=spawnSync(process.execPath,[...options,'checks/'+file],{stdio:'inherit'});
 if(result.status!==0)process.exit(result.status||1);
}
