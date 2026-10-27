import assert from 'node:assert/strict';
import {setDefaultResultOrder} from 'node:dns';
import {readFile,writeFile,mkdir} from 'node:fs/promises';
import {fileURLToPath} from 'node:url';
import {createRequire} from 'node:module';
import {randomBytes} from 'node:crypto';
setDefaultResultOrder('ipv4first');
const {chromium}=createRequire(import.meta.url)(process.env.PLAYWRIGHT_MODULE||'playwright');
const root=new URL('../',import.meta.url),dir=new URL('backend/.cache/review-browser/',root);await mkdir(dir,{recursive:true});
const settings=Object.fromEntries((await readFile(new URL('.env.fase-e.local',root),'utf8')).trim().split(/\r?\n/).map(line=>{const i=line.indexOf('=');return [line.slice(0,i),line.slice(i+1)];}));
const browser=await chromium.launch({headless:true,executablePath:process.env.BRAVO_BROWSER||'C:/Program Files/Google/Chrome/Application/chrome.exe'});
const context=await browser.newContext({viewport:{width:1440,height:1000}}),page=await context.newPage();
const suffix=randomBytes(4).toString('hex').toUpperCase(),errors=[],clientQueries=[];
const report={date:new Date().toISOString(),url:'http://localhost:8080',checks:[]};let step='inicio';
page.on('pageerror',error=>errors.push(error.message));page.on('request',r=>{const u=new URL(r.url());if(u.pathname==='/api/v1/clientes'&&r.method()==='GET')clientQueries.push(u.searchParams.get('pageSize'));});
async function route(path){step=path;await page.evaluate(p=>{location.hash='#/'+p;window.dispatchEvent(new Event('hashchange'));},path);await page.waitForFunction(p=>document.querySelector('#app').dataset.route==='#/'+p&&document.querySelector('#app').getAttribute('aria-busy')==='false',path);assert.equal(await page.locator('.error-state').count(),0,'Error al abrir '+path);}
async function fill(data){for(const [name,value]of Object.entries(data)){const input=page.locator(`input[name="${name}"],select[name="${name}"],textarea[name="${name}"]`).first();if(name==='password'){await input.waitFor();await input.evaluate((el,value)=>el.value=value,String(value));}else if(await input.evaluate(el=>el.tagName)==='SELECT')await input.selectOption(String(value));else await input.fill(String(value));}}
async function save(path,method='POST',button='#edit-form [type=submit]',status=201){const pending=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/v1'+path&&r.request().method()===method);await page.locator(button).first().click();const r=await pending;assert.equal(r.status(),status,method+' '+path+' devolvió '+r.status());return status===204?null:(await r.json()).data;}
async function api(path,options){return page.evaluate(async({path,options})=>(await(await import('./assets/js/services/api-service.js')).request(path,options)).data,{path,options});}
async function selectClient(client){const field=page.locator('.remote-picker [role=combobox]').first();await field.fill('F');await page.locator('.remote-results [role=option]').first().waitFor();await field.fill('felix gonzalez '+suffix.toLowerCase());await page.locator('.remote-results [role=option]').filter({hasText:client.documentNumber}).click();assert.equal(await page.locator('select[name=clientId]').inputValue(),client.id);}
async function login(target,username,password){await target.goto(report.url);await target.locator('#login-form').waitFor();await target.locator('[name=username]').fill(username);await target.evaluate(p=>document.querySelector('[name=password]').value=p,password);await target.locator('.login-submit').click();await target.locator('.app-shell').waitFor();await target.waitForFunction(()=>document.querySelector('#app').getAttribute('aria-busy')==='false');}
try{
 let ready=false;for(let i=0;i<90;i++){try{const r=await context.request.get(report.url+'/readyz',{timeout:2000});ready=r.ok()&&(await r.json()).technicalConfigured;}catch{}if(ready)break;await new Promise(resolve=>setTimeout(resolve,1000));}assert(ready,'Servicios no disponibles en readyz');
 await login(page,settings.DEV_ADMIN_USERNAME,settings.DEV_ADMIN_PASSWORD);
 await route('clientes/nuevo');await fill({name:'Félix González '+suffix,documentType:'Pasaporte',documentNumber:'REV'+suffix,phone:'+51987654321',address:'Verificación aislada de revisión'});const client=await save('/clientes');report.clientId=client.id;
 await route('clientes');await page.locator('.list-search input').fill('F');await page.locator('.list-results .data-table').waitFor();await page.locator('.list-search input').fill('felix gonzalez '+suffix.toLowerCase());await page.locator('.list-results').getByText(client.name,{exact:true}).waitFor();
 assert(await page.locator('.list-results a[title=Editar]').count());assert(await page.locator('.list-results button[title=Eliminar]').count());report.checks.push('Búsqueda F y nombre sin tildes, paginación y patrón lápiz/papelera');
 const vehicles=[];
 for(const [i,brand]of ['Ford','Toyota'].entries()){
  await route('vehiculos/nuevo');await selectClient(client);await fill({plate:'Q'+suffix.slice(0,4)+'-'+i,brand,model:'Prueba',year:2025,mileage:'10.00'});vehicles.push(await save('/vehiculos'));
 }
 const brands=await api('/vehiculos/marcas');assert(brands.includes('Ford')&&brands.includes('Toyota'));report.checks.push('Marcas libres y sugerencias unificadas desde vehículos existentes');
 await route('clientes');await page.locator('.list-search input').fill('felix gonzalez '+suffix);await page.locator('[data-action=delete][data-id="'+client.id+'"]').click();await save('/clientes/'+client.id,'DELETE','[data-confirm=ok]',409);await page.locator('#workshop-dialog .form-error:not([hidden])').waitFor();await page.locator('[data-confirm=cancel]').click();report.checks.push('Eliminación con confirmación rechazada por relaciones sin perder cliente/vehículos');
 await route('recepciones');await selectClient(client);await page.waitForFunction(()=>document.querySelector('select[name=vehicleId]').options.length===3);assert((await page.locator('#new-vehicle').getAttribute('href')).includes('cliente='+client.id));
 await fill({vehicleId:vehicles[0].id});assert.equal(await page.locator('[name=mileage]').inputValue(),'10.00');await fill({vehicleId:vehicles[1].id,symptoms:'Selección del segundo vehículo del mismo cliente'});
 const expectedToday=await page.evaluate(()=> 'Hoy — '+new Intl.DateTimeFormat('es-PE',{day:'2-digit',month:'2-digit',year:'numeric',timeZone:'America/Lima'}).format(new Date()));assert.equal(await page.locator('[name=createdAt]').inputValue(),expectedToday);
 await fill({mileage:'15.00'});await page.reload();await page.waitForFunction(()=>document.querySelector('#app').getAttribute('aria-busy')==='false');assert.equal(await page.locator('[name=mileage]').inputValue(),'15.00');assert.equal(await page.locator('[name=vehicleId]').inputValue(),vehicles[1].id);
 const reception=await save('/recepciones');assert.equal(reception.vehicleId,vehicles[1].id);report.checks.push('Fecha Lima automática y selección de cualquiera de dos vehículos reales');
 await route('diagnosticos?recepcion='+reception.id);await fill({findings:'Diagnóstico de revisión '+suffix});const diagnosis=await save('/diagnosticos');
 const catalog=await api('/servicios',{method:'POST',body:{name:'Sugerencia revisión '+suffix,referencePrice:'999.00'}});
 await route('cotizaciones?diagnostico='+diagnosis.id);await page.locator('#add-line').click();
 await page.locator('[data-key=concept]').nth(0).fill('Mano de obra negociada '+suffix);await page.locator('[data-key=quantity]').nth(0).fill('2');await page.locator('[data-key=unitPrice]').nth(0).fill('100.00');await page.locator('[data-key=discount]').nth(0).fill('10.00');assert(await page.locator('[data-key=partKind]').nth(0).isDisabled());
 await page.locator('#add-line').click();await page.locator('[data-key=type]').nth(1).selectOption('Repuesto');await page.locator('[data-key=partKind]').nth(1).selectOption('Alternativo');await page.locator('[data-key=concept]').nth(1).fill('Repuesto manual '+suffix);await page.locator('[data-key=unitPrice]').nth(1).fill('60.00');await page.locator('[data-key=discount]').nth(1).fill('5.00');
 await page.locator('#add-line').click();await page.locator('[data-key=unitPrice]').nth(2).fill('70.00');await page.locator('[data-key=catalogId]').nth(2).selectOption(catalog.id);assert.equal(await page.locator('[data-key=unitPrice]').nth(2).inputValue(),'70.00');
 await page.locator('#add-line').click();await page.locator('[data-remove="3"]').click();assert.equal(await page.locator('[data-key=concept]').count(),3);
 await fill({status:'Aprobada'});const quote=await save('/cotizaciones');assert.equal(quote.total,'305.00');assert.equal(quote.items[0].serviceId,null);assert.equal(quote.items[1].sparePartId,null);assert.equal(quote.items[1].partKind,'Alternativo');
 await route('cotizaciones?id='+quote.id);await page.reload();await page.waitForFunction(()=>document.querySelector('#app').getAttribute('aria-busy')==='false');assert.equal(await page.locator('[data-key=discount]').first().inputValue(),'10.00');assert.equal(await page.locator('[data-key=unitPrice]').nth(2).inputValue(),'70.00');await page.screenshot({path:fileURLToPath(new URL('cotizacion.png',dir)),fullPage:true});report.checks.push('Conceptos manuales, original/alternativo, descuento exacto, precio negociado sin sobreescritura y persistencia con F5');
 const workers=[];
 for(const [role,link]of [['Mecánico','Registrar mecánico'],['Auxiliar','Registrar auxiliar']]){
  await route('mecanicos');await page.getByRole('link',{name:link,exact:true}).click();await page.locator('#edit-form [name=username]').waitFor();assert.equal(await page.locator('[name=role]').inputValue(),role);
  const password=randomBytes(16).toString('hex'),username=(role==='Mecánico'?'mec_':'aux_')+suffix;
  await fill({name:role+' revisión '+suffix,username,email:username+'@example.test',password});const user=await save('/usuarios');workers.push({...user,password});
 }
 await route('usuarios/editar/'+workers[0].id);await fill({role:'Supervisor'});const multi=await save('/usuarios/'+workers[0].id,'PATCH',undefined,200);assert(multi.roles.some(r=>r.name==='Mecánico')&&multi.roles.some(r=>r.name==='Supervisor'));report.checks.push('Registro centralizado desde Mecánicos/Usuarios y roles N adicionales conservados');
 await route('ordenes/nueva?cotizacion='+quote.id);await fill({mechanicId:workers[0].id,assistantId:workers[1].id});const order=await save('/ordenes');report.orderId=order.id;
 await route('ordenes/editar/'+order.id);await fill({state:'En reparación',note:'Avance permitido sin prueba'});await save('/ordenes/'+order.id,'PATCH',undefined,200);
 await route('ordenes/editar/'+order.id);await fill({note:'Intento de finalizar sin prueba'});await page.evaluate(()=>document.getElementById('toast')?.remove());await save('/ordenes/'+order.id,'PATCH','button[value=finalize]',422);await page.locator('#edit-form .form-error:not([hidden])').waitFor();assert.equal(await page.locator('#toast.visible').count(),0);
 await fill({state:'En pruebas',note:'Prueba satisfactoria documentada',test:'Prueba de ruta',result:'Satisfactoria'});await save('/ordenes/'+order.id,'PATCH',undefined,200);
 await route('ordenes/editar/'+order.id);await fill({note:'Finalización con prueba vigente'});await save('/ordenes/'+order.id,'PATCH','button[value=finalize]',200);report.checks.push('Asignación de mecánico/auxiliar; avance sin prueba; rechazo explicado y finalización con prueba vigente');
 await route('auxilio');const unitsBefore=await api('/unidades-auxilio');if(!unitsBefore.length)await page.getByText('No hay unidades de auxilio registradas.',{exact:false}).waitFor();
 await page.locator('#new-unit').click();await page.locator('#unit-form [name=code]').fill('REV-'+suffix);const unit=await save('/unidades-auxilio','POST','#unit-form [type=submit]');await page.locator('#workshop-dialog').waitFor({state:'detached'});
 await selectClient(client);await page.waitForFunction(()=>document.querySelector('select[name=vehicleId]').options.length===3);await fill({vehicleId:vehicles[0].id,location:'Ubicación de prueba '+suffix,description:'Auxilio con unidad registrada real',technicianId:workers[0].id});assert.equal(await page.locator('[name=unit]').inputValue(),unit.id);const rescue=await save('/auxilios');report.rescueId=rescue.id;report.unitId=unit.id;
 assert.equal((await api('/auxilios/'+rescue.id)).unitId,unit.id);report.checks.push('Alta autorizada de unidad de auxilio y selección persistida sin datos ficticios');
 for(const period of ['Semana','Mes','Año']){
  await route('inicio?period='+encodeURIComponent(period));assert.deepEqual(await page.locator('[data-period]').allTextContents(),['Semana','Mes','Año']);assert.equal(await page.locator('[data-period="'+period+'"]').getAttribute('aria-pressed'),'true');
  const data=await api('/dashboard?period='+encodeURIComponent(period));const days=Number(new Intl.DateTimeFormat('en-CA',{day:'numeric',timeZone:'America/Lima'}).format(new Date(Date.parse(data.until)-1)));assert.equal(data.income.length,period==='Semana'?7:period==='Año'?12:days);
  await route('reportes');await fill({period});const response=page.waitForResponse(r=>r.url().includes('/api/v1/reportes?')&&r.request().method()==='GET');await page.locator('#report-form [type=submit]').click();assert.equal((await response).status(),200);
 }
 report.checks.push('Semana/Mes/Año consultados realmente en dashboard y reportes');
 await route('inicio?period=Semana');await page.screenshot({path:fileURLToPath(new URL('dashboard-semana.png',dir)),fullPage:true});
 await page.reload();await page.locator('.kpi-grid').waitFor();assert.equal((await api('/cotizaciones/'+quote.id)).total,'305.00');assert.equal((await api('/ordenes/'+order.id)).state,'Finalizada');
 const workerContext=await browser.newContext(),worker=await workerContext.newPage();await login(worker,workers[1].username,workers[1].password);const status=await worker.evaluate(async()=>{try{await(await import('./assets/js/services/api-service.js')).request('/clientes',{method:'POST',body:{}});return 200;}catch(e){return e.status;}});assert.equal(status,403);await workerContext.close();
 assert(clientQueries.every(size=>size&&Number(size)<=20),'Descarga de clientes sin paginación acotada');assert.deepEqual(errors,[]);
 report.checks.push('Permisos efectivos, recarga persistente y clientes siempre consultados en páginas acotadas');
 report.quoteId=quote.id;report.receptionId=reception.id;report.diagnosisId=diagnosis.id;report.passed=true;
 await writeFile(new URL('resultado.json',dir),JSON.stringify(report,null,2));console.log('Revisión funcional en Chrome contra Go/PostgreSQL Docker: APROBADA. Reporte sin credenciales: backend/.cache/review-browser/resultado.json');
}catch(error){await writeFile(new URL('fallo.txt',dir),step+'\n'+String(error.stack||error));throw Error('Falló la revisión; consultar backend/.cache/review-browser/fallo.txt (credenciales omitidas).');}
finally{await browser.close();}
