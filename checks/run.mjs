import {spawnSync} from 'node:child_process';
const suites=[['verify.mjs','DEMO'],['workflow.mjs','DEMO'],['edge-cases.mjs','DEMO/UNIDAD'],['real-api.mjs','UNIDAD: fetch simulado'],['report-periods.mjs','UNIDAD'],['diagnosis-text.mjs','UNIDAD'],['isolation.mjs','UNIDAD: guardas sin bases'],['prototype.mjs','HISTÓRICO: prototipo']];
for(const [file,kind] of suites){
 console.log(`[${kind}] ${file}`);
 const result=spawnSync(process.execPath,['checks/'+file],{stdio:'inherit'});
 if(result.status!==0)process.exit(result.status||1);
}
