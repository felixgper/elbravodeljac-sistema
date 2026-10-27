import {setDefaultResultOrder} from 'node:dns';
setDefaultResultOrder('ipv4first');
// Real Chrome -> persistent Go container -> private PostgreSQL/MongoDB network.
// Run only after backend/scripts/start-dev.ps1. Secrets are read from the ignored
// local environment file, never logged or included in the JSON verification report.
import assert from 'node:assert/strict';
import {readFile,writeFile,mkdir} from 'node:fs/promises';
import {createRequire} from 'node:module';
import {fileURLToPath} from 'node:url';
import {randomBytes} from 'node:crypto';
const require=createRequire(import.meta.url);
const {chromium}=require(process.env.PLAYWRIGHT_MODULE||'playwright');
const root=new URL('../',import.meta.url);
const settings=Object.fromEntries((await readFile(new URL('.env.fase-e.local',root),'utf8')).trim().split(/\r?\n/).map(line=>{const i=line.indexOf('=');return [line.slice(0,i),line.slice(i+1)];}));
const url='http://localhost:8080',suffix=randomBytes(5).toString('hex').toUpperCase();
const artifactDir=new URL('backend/.cache/live-docker/',root);await mkdir(artifactDir,{recursive:true});
const browser=await chromium.launch({headless:true,executablePath:process.env.BRAVO_BROWSER||'C:/Program Files/Google/Chrome/Application/chrome.exe'});
const context=await browser.newContext({viewport:{width:1440,height:1000}}),page=await context.newPage();
const runtimeErrors=[];page.on('pageerror',error=>runtimeErrors.push(error.message));
const report={url,checkedAt:new Date().toISOString(),browser:'Chrome',mode:'real',checks:[]};
async function fill(values){for(const [name,value]of Object.entries(values)){const input=page.locator(`input[name="${name}"],select[name="${name}"],textarea[name="${name}"]`).first();if(await input.evaluate(el=>el.tagName)==='SELECT'){if(await input.evaluate(el=>el.hidden)){const record=await page.evaluate(async id=>(await (await import('./assets/js/services/api-service.js')).request('/clientes/'+id)).data,String(value));await input.locator('..').locator('[role=combobox]').fill(record.name);await input.locator('..').getByRole('option').filter({hasText:record.documentNumber}).click();}else await input.selectOption(String(value));}else await input.fill(String(value));}}
async function route(path){await page.evaluate(path=>{location.hash='#/'+path;window.dispatchEvent(new Event('hashchange'));},path);await page.waitForFunction(path=>document.querySelector('#app').dataset.route==='#/'+path&&document.querySelector('#app').getAttribute('aria-busy')==='false',path);assert.equal(await page.locator('.error-state').count(),0,'La vista muestra un error: '+path);}
async function save(path,method='POST',selector='#edit-form [type=submit]'){const pending=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/v1'+path&&r.request().method()===method);await page.locator(selector).first().click();const response=await pending;if(!response.ok())throw Error(`${method} ${path}: HTTP ${response.status()}`);return response.status()===204?null:(await response.json()).data;}
async function login(target,username,password){await target.goto(url);await target.locator('#login-form').waitFor();await target.locator('[name=username]').fill(username);await target.locator('[name=password]').waitFor();await target.evaluate(password=>{document.querySelector('[name=password]').value=password;},password);await target.locator('.login-submit').click();await target.locator('.app-shell').waitFor();await target.waitForFunction(()=>document.querySelector('#app').getAttribute('aria-busy')==='false');}
try{
 const ready=await context.request.get(url+'/readyz');assert.equal(ready.status(),200);assert.equal((await ready.json()).technicalConfigured,true);report.checks.push('Go confirma PostgreSQL y MongoDB en readyz');
 const anonymous=await context.request.get(url+'/api/v1/clientes');assert.equal(anonymous.status(),401);
 await login(page,settings.DEV_ADMIN_USERNAME,settings.DEV_ADMIN_PASSWORD);
 assert((await page.locator('.topbar-demo').innerText()).includes('Modo real'));
 assert((await context.cookies()).some(c=>c.name==='bravo_access'&&c.httpOnly&&c.sameSite==='Strict'));
 const prepared=await page.evaluate(async suffix=>{
  const {request}=await import('./assets/js/services/api-service.js');const roles=(await request('/roles')).data;
  const create=async(path,body)=>(await request(path,{method:'POST',body})).data;
  const mechanic=await create('/usuarios',{name:'Mecánico verificación Docker '+suffix,username:'mecanico_'+suffix,password:crypto.randomUUID(),roleId:roles.find(r=>r.name==='Mecánico').id,active:true});
  const limitedPassword=crypto.randomUUID(),limited=await create('/usuarios',{name:'Auxiliar verificación Docker '+suffix,username:'auxiliar_'+suffix,password:limitedPassword,roleId:roles.find(r=>r.name==='Auxiliar').id,active:true});
  const service=await create('/servicios',{name:'Servicio verificación Docker '+suffix,referencePrice:'100.00'});
  const part=await create('/repuestos',{name:'Repuesto verificación Docker '+suffix,referencePrice:'20.00',stock:'10.00'});
  return {mechanicId:mechanic.id,limitedUsername:limited.username,limitedPassword,serviceId:service.id,partId:part.id,adminRoleId:roles.find(r=>r.name==='Administrador').id};
 },suffix);
 await route('clientes/nuevo');await fill({name:'Verificación Docker '+suffix,documentType:'Pasaporte',documentNumber:'D'+suffix,phone:'+51987654321',address:'Entorno aislado de pruebas'});const client=await save('/clientes');report.clientId=client.id;
 await route('vehiculos/nuevo');await fill({clientId:client.id,plate:'V-'+suffix.slice(0,6),model:'T8',year:2024,mileage:'100.00'});const vehicle=await save('/vehiculos');report.vehicleId=vehicle.id;
 await route('recepciones?cliente='+client.id+'&vehiculo='+vehicle.id);await fill({symptoms:'Verificación de persistencia Docker'});const reception=await save('/recepciones');report.receptionId=reception.id;
 await route('diagnosticos?recepcion='+reception.id);await fill({findings:'Evaluación técnica del entorno de pruebas'});const diagnosis=await save('/diagnosticos');report.diagnosisId=diagnosis.id;
 await route('cotizaciones?diagnostico='+diagnosis.id);await page.locator('#add-line').click();await page.locator('[data-key=catalogId]').selectOption(prepared.serviceId);await page.locator('#add-line').click();await page.locator('[data-key=type]').nth(1).selectOption('Repuesto');await page.locator('[data-key=catalogId]').nth(1).selectOption(prepared.partId);await fill({status:'Aprobada'});const quote=await save('/cotizaciones');assert.equal(quote.total,'120.00');report.quoteId=quote.id;
 await route('ordenes/nueva?cotizacion='+quote.id);await fill({mechanicId:prepared.mechanicId});const order=await save('/ordenes');report.orderId=order.id;report.orderCode=order.code;
 await route('ordenes/editar/'+order.id);await fill({state:'En reparación',note:'Avance verificado desde Chrome'});await save('/ordenes/'+order.id,'PATCH');
 await route('ordenes/editar/'+order.id);await fill({state:'Finalizada',note:'Finalización verificada',test:'Prueba de ruta',result:'Satisfactoria'});await save('/ordenes/'+order.id,'PATCH');
 await route('pagos?orden='+order.id);await fill({amount:'120.00'});const payment=await save('/ordenes/'+order.id+'/pagos','POST','#payment-form [type=submit]');report.paymentId=payment.id;
 await route('pagos?orden='+order.id);await page.locator('#delivery-form [name=method]').selectOption('Documento');await page.locator('[name=verified]').check();const delivery=await save('/ordenes/'+order.id+'/entrega','POST','#delivery-form [type=submit]');report.deliveryId=delivery.id;
 report.checks.push('Formularios cliente -> vehículo -> recepción -> diagnóstico -> cotización -> OT -> avance/prueba -> pago -> entrega');
 const technical=await page.evaluate(async diagnosisId=>{
  const t=await import('./assets/js/services/technical-service.js');
  const scan={requestId:t.newTechnicalRequestID(),equipo:{nombre:'Verificación Docker'},dtcs:[],mediciones:{rpm:800}};
  const a=await t.saveScan(diagnosisId,scan),b=await t.saveScan(diagnosisId,scan);
  const canvas=document.createElement('canvas');canvas.width=2;canvas.height=2;canvas.getContext('2d').fillRect(0,0,2,2);
  const blob=await new Promise(resolve=>canvas.toBlob(resolve,'image/png')),file=new File([blob],'verificacion.png',{type:'image/png'}),requestId=t.newTechnicalRequestID();
  const x=await t.uploadEvidence(diagnosisId,{file,description:'Archivo técnico exclusivo de verificación Docker',requestId}),y=await t.uploadEvidence(diagnosisId,{file,description:'Archivo técnico exclusivo de verificación Docker',requestId});
  const download=await fetch(t.evidenceURL(diagnosisId,x.data.id));
  return {scanId:a.data.id,evidenceId:x.data.id,sameScan:a.data.id===b.data.id,sameEvidence:x.data.id===y.data.id,downloadOK:download.ok&&download.headers.get('Content-Type').includes('image/png'),bytes:(await download.arrayBuffer()).byteLength};
 },diagnosis.id);
 assert(technical.sameScan&&technical.sameEvidence&&technical.downloadOK&&technical.bytes>0);report.scanId=technical.scanId;report.evidenceId=technical.evidenceId;report.checks.push('MongoDB/GridFS: carga, descarga y reintentos sin duplicados');
 await page.reload();await page.waitForFunction(()=>document.querySelector('#app').getAttribute('aria-busy')==='false');
 const persisted=await page.evaluate(async ids=>{const {request}=await import('./assets/js/services/api-service.js');const get=async path=>(await request(path)).data;return {client:await get('/clientes/'+ids.clientId),vehicle:await get('/vehiculos/'+ids.vehicleId),order:await get('/ordenes/'+ids.orderId),payments:await get('/ordenes/'+ids.orderId+'/pagos'),delivery:await get('/ordenes/'+ids.orderId+'/entrega')};},report);
 assert.equal(persisted.client.id,client.id);assert.equal(persisted.vehicle.clientId,client.id);assert.equal(persisted.order.state,'Finalizada');assert.equal(persisted.payments.paid,'120.00');assert.equal(persisted.payments.balance,'0.00');assert.equal(persisted.delivery.id,delivery.id);report.checks.push('Lectura posterior a recarga: registros persistidos, pago exacto y entrega');
 const csrfStatus=await page.evaluate(async()=>{const r=await fetch('/api/v1/clientes',{method:'POST',headers:{'Content-Type':'application/json'},body:'{}'});return r.status;});assert.equal(csrfStatus,403);
 const ownRoleStatus=await page.evaluate(async roleId=>{const {request}=await import('./assets/js/services/api-service.js');try{await request('/roles/'+roleId+'/permisos',{method:'PUT',body:{permissions:{}}});return 200;}catch(error){return error.status;}},prepared.adminRoleId);assert.equal(ownRoleStatus,403);
 const workerContext=await browser.newContext(),worker=await workerContext.newPage();await login(worker,prepared.limitedUsername,prepared.limitedPassword);
 const denied=await worker.evaluate(async()=>{const {request}=await import('./assets/js/services/api-service.js');try{await request('/clientes',{method:'POST',body:{}});return {status:200};}catch(error){return {status:error.status,code:error.code};}});assert.equal(denied.status,403);assert.equal(denied.code,'FORBIDDEN');
 await worker.evaluate(()=>location.hash='#/clientes');await worker.locator('.no-access').waitFor();assert.equal(await worker.locator('.nav-item[href="#/clientes"]').count(),0);await workerContext.close();report.checks.push('401 anónimo, 403 por CSRF, autoedición de permisos bloqueada y auxiliar sin acceso a Clientes');
 for(const screen of ['inicio','clientes','vehiculos','recepciones','diagnosticos','cotizaciones','ordenes','pagos','auxilio','mecanicos','proveedores','reportes','usuarios','usuarios/roles']){await route(screen);await page.screenshot({path:fileURLToPath(new URL(screen.replaceAll('/','-')+'.png',artifactDir)),fullPage:true});}
 assert.equal(await page.evaluate(()=>Object.keys(localStorage).some(k=>k.includes('workshop-demo'))),false);assert.deepEqual(runtimeErrors,[]);
 report.checks.push('14 vistas reales sin errores JavaScript ni mezcla de demo');
 await writeFile(new URL('resultado.json',artifactDir),JSON.stringify(report,null,2));
 console.log('Chrome + Go/Docker + PostgreSQL + MongoDB: recorrido, persistencia, permisos, CSRF y GridFS APROBADOS.');
 console.log('Registros de prueba conservados para consulta. Informe sin credenciales: backend/.cache/live-docker/resultado.json');
}catch(error){await writeFile(new URL('fallo.txt',artifactDir),String(error.stack||error));throw new Error('Falló la verificación real; consultar backend/.cache/live-docker/fallo.txt (no se imprimen credenciales).');}
finally{await browser.close();}
