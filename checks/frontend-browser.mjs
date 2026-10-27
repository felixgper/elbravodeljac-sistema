import {mkdir} from 'node:fs/promises';
import {fileURLToPath} from 'node:url';
import assert from 'node:assert/strict';
import {createRequire} from 'node:module';
const require=createRequire(import.meta.url);
const {chromium}=require(process.env.PLAYWRIGHT_MODULE||'playwright');
let input='';for await(const chunk of process.stdin)input+=chunk;
const f=JSON.parse(input);
const shots=fileURLToPath(new URL('../backend/.cache/frontend-E-'+f.suffix+'/',import.meta.url));await mkdir(shots,{recursive:true});
const browser=await chromium.launch({headless:true,...(process.env.BRAVO_BROWSER?{executablePath:process.env.BRAVO_BROWSER}:{channel:'msedge'})});
const context=await browser.newContext({viewport:{width:1440,height:1000}}),page=await context.newPage();
const errors=[],requests=[];page.on('pageerror',e=>errors.push(e.message));
page.on('response',r=>{if(r.url().includes('/api/v1/'))requests.push([r.request().method(),new URL(r.url()).pathname,r.status()]);});
const route=async path=>{await page.evaluate(path=>{location.hash='#/'+path;window.dispatchEvent(new Event('hashchange'));},path);await page.waitForFunction(path=>document.querySelector('#app').dataset.route==='#/'+path&&document.querySelector('#app').getAttribute('aria-busy')==='false',path);assert.equal(await page.locator('.error-state').count(),0,await page.locator('#app').innerText());};
const fill=async values=>{for(const [name,value] of Object.entries(values)){const el=page.locator(`input[name="${name}"],select[name="${name}"],textarea[name="${name}"]`).first();if(await el.evaluate(e=>e.tagName)==='SELECT'){if(await el.evaluate(e=>e.hidden)){const record=await page.evaluate(async id=>(await (await import('./assets/js/services/api-service.js')).request('/clientes/'+id)).data,String(value));const field=el.locator('..').locator('[role=combobox]');await field.fill(record.name);await el.locator('..').getByRole('option').filter({hasText:record.documentNumber}).click();}else await el.selectOption(String(value));}else await el.fill(String(value));}};
async function save(path,method,button){const response=page.waitForResponse(r=>new URL(r.url()).pathname===path&&r.request().method()===method);await button.click();const r=await response;if(!r.ok())assert.fail(`${method} ${path}: ${r.status()} ${await r.text()}`);return r.status()===204?null:(await r.json()).data;}
async function appAPI(path,options){return page.evaluate(async ({path,options})=>{const api=await import('./assets/js/services/api-service.js');try{return {value:await api.request(path,options)}}catch(e){return {status:e.status,message:e.message}}},{path,options});}
try{
 await page.goto(f.url);await page.locator('#login-form').waitFor();
 assert(await page.locator('.demo-access').innerText().then(t=>t.includes('Modo real')));
 await fill({username:f.username,password:f.password});await page.locator('.login-submit').click();await page.locator('.kpi-grid').waitFor();
 const cookies=await context.cookies();assert(cookies.some(c=>c.name==='bravo_access'&&c.httpOnly&&c.sameSite==='Strict'));
 assert.equal(await page.evaluate(()=>Object.keys(localStorage).some(k=>k.includes('workshop-demo'))),false);
 await route('clientes/nuevo');await fill({name:'Cliente E '+f.suffix,documentType:'Pasaporte',documentNumber:'E'+f.suffix.toUpperCase(),phone:'+51987654321',email:'',address:'Dirección de prueba'});
 const client=await save('/api/v1/clientes','POST',page.locator('#edit-form [type=submit]'));
 await route('clientes/editar/'+client.id);await fill({name:'Cliente actualizado '+f.suffix});await save('/api/v1/clientes/'+client.id,'PATCH',page.locator('#edit-form [type=submit]'));
 await route('vehiculos/nuevo');await fill({clientId:client.id,plate:'E-'+f.suffix.slice(0,6).toUpperCase(),model:'T8',year:2024,mileage:'100.00'});const vehicle=await save('/api/v1/vehiculos','POST',page.locator('#edit-form [type=submit]'));
 await route('vehiculos/editar/'+vehicle.id);await fill({notes:'Observación actualizada'});await save('/api/v1/vehiculos/'+vehicle.id,'PATCH',page.locator('#edit-form [type=submit]'));
 await route('vehiculos');await page.locator('[data-action=history][data-id="'+vehicle.id+'"]').click();await page.locator('#workshop-dialog').waitFor();await page.locator('#workshop-dialog .dialog-heading button').click();
 await route('recepciones?cliente='+client.id+'&vehiculo='+vehicle.id);await fill({symptoms:'Revisión de prueba E'});const reception=await save('/api/v1/recepciones','POST',page.locator('#edit-form [type=submit]'));
 await route('diagnosticos?recepcion='+reception.id);await fill({findings:'Evaluación técnica E'});const diagnosis=await save('/api/v1/diagnosticos','POST',page.locator('#edit-form [type=submit]'));
 await route('cotizaciones?diagnostico='+diagnosis.id);await page.locator('#add-line').click();await page.locator('[data-key=catalogId]').selectOption(f.serviceId);await page.locator('#add-line').click();await page.locator('[data-key=type]').nth(1).selectOption('Repuesto');await page.locator('[data-key=catalogId]').nth(1).selectOption(f.partId);await fill({status:'Pendiente'});const quote=await save('/api/v1/cotizaciones','POST',page.locator('#edit-form [type=submit]'));assert.equal(quote.total,'120.00');
 await route('cotizaciones?id='+quote.id);await fill({status:'Aprobada'});await save('/api/v1/cotizaciones/'+quote.id,'PUT',page.locator('#edit-form [type=submit]'));
 await route('ordenes/nueva?cotizacion='+quote.id);await fill({mechanicId:f.mechanicId});const order=await save('/api/v1/ordenes','POST',page.locator('#edit-form [type=submit]'));
 await route('ordenes/editar/'+order.id);await fill({state:'En reparación',note:'Reparación de prueba'});await save('/api/v1/ordenes/'+order.id,'PATCH',page.locator('#edit-form [type=submit]').first());
 await route('ordenes/editar/'+order.id);await fill({state:'Finalizada',note:'Prueba satisfactoria',test:'Prueba de ruta',result:'Satisfactoria'});await save('/api/v1/ordenes/'+order.id,'PATCH',page.locator('#edit-form [type=submit]').first());
 await route('pagos?orden='+order.id);await fill({amount:'120.00'});await save('/api/v1/ordenes/'+order.id+'/pagos','POST',page.locator('#payment-form [type=submit]'));
 await route('pagos?orden='+order.id);await page.locator('[name=kind][value=Tercero]').check();await page.locator('#new-contact').click();await page.locator('#contact-form [name=name]').fill('Contacto de pruebas');await page.locator('#contact-form [name=documentNumber]').fill('76543210');await page.locator('#contact-form [name=phone]').fill('+51987654321');await save('/api/v1/clientes/'+client.id+'/contactos','POST',page.locator('#contact-form [type=submit]'));await page.locator('#workshop-dialog').waitFor({state:'detached'});await page.locator('#delivery-form [name=method]').selectOption('Documento');await page.locator('[name=verified]').check();await save('/api/v1/ordenes/'+order.id+'/entrega','POST',page.locator('#delivery-form [type=submit]'));
 await route('auxilio');await fill({clientId:client.id});await fill({vehicleId:vehicle.id,location:'Ubicación de prueba',description:'Auxilio de prueba',technicianId:f.mechanicId,unit:f.unitId});await save('/api/v1/auxilios','POST',page.locator('#edit-form [type=submit]'));
 await route('proveedores');await page.locator('[data-action=new]').click();await fill({name:'Proveedor E '+f.suffix,ruc:'20123456789',phone:'+51987654321'});const supplier=await save('/api/v1/proveedores','POST',page.locator('#supplier-form [type=submit]'));await page.waitForTimeout(100);
 await route('proveedores');await page.locator('[data-action=edit][data-id="'+supplier.id+'"]').click();await fill({name:'Proveedor actualizado '+f.suffix});await save('/api/v1/proveedores/'+supplier.id,'PATCH',page.locator('#supplier-form [type=submit]'));
 await route('usuarios/nuevo');await fill({name:'Usuario E '+f.suffix,username:'new_'+f.suffix,email:'e_'+f.suffix+'@example.test',role:'Mecánico',password:f.password});const newUser=await save('/api/v1/usuarios','POST',page.locator('#edit-form [type=submit]'));
 await route('usuarios/editar/'+newUser.id);await fill({name:'Usuario editado '+f.suffix});await save('/api/v1/usuarios/'+newUser.id,'PATCH',page.locator('#edit-form [type=submit]'));
 await route('usuarios');await page.locator('.list-search input').fill('new_'+f.suffix);await page.locator('[data-action=delete][data-id="'+newUser.id+'"]').click();await save('/api/v1/usuarios/'+newUser.id,'DELETE',page.locator('[data-confirm=ok]'));
 await route('proveedores');await page.locator('.list-search input').fill('Proveedor actualizado '+f.suffix);await page.locator('[data-action=delete][data-id="'+supplier.id+'"]').click();await save('/api/v1/proveedores/'+supplier.id,'DELETE',page.locator('[data-confirm=ok]'));
 await route('usuarios/roles?rol='+encodeURIComponent('ELimited_'+f.suffix));await page.locator('[data-module=ordenes][data-action=editar]').uncheck();await save('/api/v1/roles/'+(await appAPI('/roles')).value.data.find(r=>r.name==='ELimited_'+f.suffix).id+'/permisos','PUT',page.locator('#perm-save'));
 await route('clientes/nuevo');await fill({name:'Nuevo propietario '+f.suffix,documentType:'Pasaporte',documentNumber:'N'+f.suffix.toUpperCase(),phone:'+51987654321',address:'Dirección de prueba'});const owner=await save('/api/v1/clientes','POST',page.locator('#edit-form [type=submit]'));
 await route('vehiculos/editar/'+vehicle.id);await page.locator('[name=transfer]').check();await fill({newClientId:owner.id});await page.locator('[name=transferVerified]').check();await save('/api/v1/vehiculos/'+vehicle.id+'/transferencias','POST',page.locator('#edit-form [type=submit]'));
 // MongoDB via the same browser session and the actual JS service; no mock DB.
 const tech=await page.evaluate(async id=>{const t=await import('./assets/js/services/technical-service.js');const scan={requestId:t.newTechnicalRequestID(),equipo:{nombre:'Escáner E'},dtcs:[],mediciones:{rpm:800}};const a=await t.saveScan(id,scan),b=await t.saveScan(id,scan);const file=new File(['%PDF-1.4\n1 0 obj << /Type /Catalog >> endobj\n%%EOF'], 'prueba.pdf',{type:'application/pdf'}),requestId=t.newTechnicalRequestID();const x=await t.uploadEvidence(id,{file,requestId}),y=await t.uploadEvidence(id,{file,requestId});const downloaded=await fetch(t.evidenceURL(id,x.data.id));return {scan:a.data.id===b.data.id,evidence:x.data.id===y.data.id,text:await downloaded.text(),count:(await t.listEvidence(id)).length};},diagnosis.id);
 assert(tech.scan&&tech.evidence&&tech.count===1);assert(tech.text.startsWith('%PDF-1.4')&&tech.text.endsWith('%%EOF'));
 for(const path of ['inicio','clientes','vehiculos','recepciones','diagnosticos','cotizaciones','ordenes','pagos','auxilio','auxilio/tareas','mecanicos','proveedores','reportes','usuarios','usuarios/roles']){await route(path);await page.screenshot({path:shots+'/'+path.replaceAll('/','-')+'.png',fullPage:true});}
 await route('reportes');await fill({period:'Año'});await page.locator('#report-form [type=submit]').click();await page.locator('.report-columns').waitFor();
 // A network error cannot close the form, navigate or show a success toast.
 await route('clientes/nuevo');await fill({name:'No persistir',documentType:'Pasaporte',documentNumber:'FAIL'+f.suffix.toUpperCase(),phone:'+51987654321',address:'Prueba'});await page.evaluate(()=>document.getElementById('toast')?.remove());await page.route('**/api/v1/clientes',r=>r.abort());await page.locator('#edit-form [type=submit]').click();await page.locator('.form-error:not([hidden])').waitFor();assert.equal(await page.locator('#toast.visible').count(),0);assert(page.url().includes('clientes/nuevo'));await page.unroute('**/api/v1/clientes');
 await page.locator('#profile-toggle').click();await page.locator('#logout').click();await page.locator('#login-form').waitFor();assert(!(await context.cookies()).some(c=>c.name==='bravo_access'));
 // Limited user: no privilege by URL, button, forged request or auxiliary query.
 await fill({username:f.limitedUsername,password:f.password});await page.locator('.login-submit').click();await page.locator('.app-shell').waitFor();await page.waitForTimeout(100);
 const denied=await appAPI('/clientes',{method:'POST',body:{}});assert.equal(denied.status,403);
 assert.equal((await appAPI('/frontend/opciones?module=usuarios&resource=clientes')).status,403);
 const scoped=await appAPI('/frontend/opciones?module=ordenes&resource=clientes');assert.equal(scoped.value.data.length,0);
 await page.evaluate(()=>location.hash='#/usuarios');await page.locator('.no-access').waitFor();assert.equal(await page.locator('.nav-item[href="#/usuarios"]').count(),0);
 assert.equal(await page.evaluate(()=>Object.keys(localStorage).some(k=>k.includes('workshop-demo'))),false);
 assert.deepEqual(errors,[]);assert(!requests.some(r=>r[2]===501));
 console.log('E2E real: altas y edición de cliente/vehículo, CRUD usuario/proveedor, transferencia/contacto, vehículo e historial, recepción, diagnóstico, catálogo/cotización, OT/avances/pruebas, pago, entrega, auxilio, 15 vistas, reportes, matriz, cookies/CSRF, 403 y aislamiento, error de red, MongoDB/GridFS y reintentos: OK.');
}finally{await browser.close();}
