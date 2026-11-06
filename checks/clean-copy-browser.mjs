import {readIsolatedFixture} from './isolated-fixture.mjs';
const f=await readIsolatedFixture();
import assert from 'node:assert/strict';
import {writeFile,mkdir} from 'node:fs/promises';
import {createRequire} from 'node:module';
import {fileURLToPath} from 'node:url';
const {chromium}=createRequire(import.meta.url)(process.env.PLAYWRIGHT_MODULE||'playwright');

assert(/^jac_restore_\d{14}$/.test(f.database));
const dir=new URL('../backend/.cache/auditoria-20261103/',import.meta.url);await mkdir(dir,{recursive:true});
const browser=await chromium.launch({headless:true,executablePath:'C:/Program Files/Google/Chrome/Application/chrome.exe'}),errors=[],result={database:f.database,passed:false,checks:[]};
async function login(username,password){const context=await browser.newContext({viewport:{width:1440,height:1000}}),page=await context.newPage();page.on('pageerror',e=>errors.push(e.message));page.on('response',r=>{if(r.status()>=500)errors.push('HTTP '+r.status()+' '+new URL(r.url()).pathname);});await page.goto(f.url);await fill(page,{username,password});await page.locator('.login-submit').click();await page.locator('.app-shell').waitFor();return page;}
async function fill(page,values){for(const [key,value] of Object.entries(values)){const el=page.locator(`[name="${key}"]`).first();if(key==='brand'){await el.locator('..').locator('[role=combobox]').fill(String(value));await el.locator('..').getByRole('option',{name:String(value),exact:true}).click();}else if(await el.evaluate(x=>x.tagName)==='SELECT')await el.selectOption(String(value));else await el.fill(String(value));}}
async function api(page,path,method='GET',body){return page.evaluate(async({path,method,body})=>{return await(await import('./assets/js/services/api-service.js')).request(path,{method,body});},{path,method,body});}
async function route(page,path){if(path==='reportes')path='inicio?tab=analisis';await page.evaluate(p=>{location.hash='#/'+p;dispatchEvent(new Event('hashchange'));},path);await page.waitForFunction(p=>document.querySelector('#app').dataset.route==='#/'+p&&document.querySelector('#app').getAttribute('aria-busy')==='false',path);assert.equal(await page.locator('.error-state').count(),0);}
try{
 const manager=await login(f.username,f.password);assert((await api(manager,'/auth/me')).data.roles.includes('Gerente'));result.existingManagerLogin=true;
 for(const path of ['/clientes','/vehiculos','/recepciones','/diagnosticos','/cotizaciones','/ordenes','/auxilios']){const response=await api(manager,path);assert.deepEqual(response.data,[],path);assert.equal(response.meta.total,0,path);}
 result.checks.push('Listados operativos HTTP vacíos con cuenta Gerente existente');
 const usersBefore=(await api(manager,'/usuarios')).meta.total,employeesBefore=(await api(manager,'/empleados')).meta.total;assert(usersBefore>0&&employeesBefore>0);result.preservedCounts={users:usersBefore,employees:employeesBefore};
 for(const path of ['inicio','clientes','vehiculos','recepciones','diagnosticos','cotizaciones','ordenes','pagos','auxilio','reportes'])await route(manager,path);
 for(const period of ['Semana','Mes','Año']){await fill(manager,{period});const wait=manager.waitForResponse(r=>new URL(r.url()).pathname==='/api/v1/reportes');await manager.locator('#dashboard-filter [type=submit]').click();assert.equal((await wait).status(),200);await manager.locator('#report-content:not([aria-busy])').waitFor();}
 await manager.screenshot({path:fileURLToPath(new URL('copia-vacia-reportes.png',dir)),fullPage:true});result.checks.push('Dashboard y reportes Semana/Mes/Año sin operaciones ni errores');
 assert((await api(manager,'/tipos-documento')).data.length>0);assert((await api(manager,'/servicios')).data.length>0);assert((await api(manager,'/repuestos')).data.length>0);
 const roles=(await api(manager,'/roles')).data;
 for(const [label,role] of [['admin','Administrador'],['mecanico_a','Mecánico'],['mecanico_b','Mecánico']]){
  const employee=(await api(manager,'/empleados','POST',{name:'QA copia '+label+' '+f.suffix,employeeType:role,active:true})).data;
  const username='clean_'+label+'_'+f.suffix;
  await api(manager,'/usuarios','POST',{name:employee.name,username,password:f.newPassword,employeeId:employee.id,roleId:roles.find(r=>r.name===role).id});
  const page=await login(username,f.newPassword);assert((await api(page,'/auth/me')).data.roles.includes(role));await page.context().close();
 }
 result.checks.push('Creación por API de empleados/cuentas Administrador y dos Mecánicos vinculados; inicios de sesión independientes');
 await route(manager,'clientes/nuevo');await fill(manager,{name:'Cliente copia '+f.suffix,documentType:'Pasaporte',documentNumber:'CLN'+f.suffix,phone:'987654321',address:'Solo copia QA'});let waiting=manager.waitForResponse(r=>new URL(r.url()).pathname==='/api/v1/clientes'&&r.request().method()==='POST');await manager.locator('#edit-form [type=submit]').click();let response=await waiting;assert.equal(response.status(),201);const client=(await response.json()).data;
 await route(manager,'vehiculos/nuevo?cliente='+client.id);await fill(manager,{plate:'CL'+f.suffix.slice(0,5),brand:'Toyota',model:'Prueba copia',year:'2025',mileage:'1250.25'});waiting=manager.waitForResponse(r=>new URL(r.url()).pathname==='/api/v1/vehiculos'&&r.request().method()==='POST');await manager.locator('#edit-form [type=submit]').click();response=await waiting;assert.equal(response.status(),201);const vehicle=(await response.json()).data;assert.equal((await api(manager,'/vehiculos/'+vehicle.id)).data.mileage,'1250.25');
 result.checks.push('Formularios de cliente/vehículo después de limpieza: persistencia SQL con Toyota y kilometraje decimal');
 assert.deepEqual(errors,[]);result.passed=true;await writeFile(new URL('clean-probe.json',dir),JSON.stringify(result,null,2));console.log('Copia restaurada limpia y altas posteriores: APROBADAS.');
}catch(error){result.error=error.message;result.errors=errors;await writeFile(new URL('clean-probe.json',dir),JSON.stringify(result,null,2));throw error;}finally{await browser.close();}
