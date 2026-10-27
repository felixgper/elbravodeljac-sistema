import {spawnSync} from 'node:child_process';
for(const file of ['verify.mjs','workflow.mjs','edge-cases.mjs','real-api.mjs']){
 const result=spawnSync(process.execPath,['checks/'+file],{stdio:'inherit'});
 if(result.status!==0)process.exit(result.status||1);
}
