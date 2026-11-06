import {readIsolatedFixture} from './isolated-fixture.mjs';
const f=await readIsolatedFixture();
import assert from 'node:assert/strict';
import {createRequire} from 'node:module';
import {mkdir,writeFile} from 'node:fs/promises';
import {fileURLToPath} from 'node:url';
const {chromium}=createRequire(import.meta.url)(process.env.PLAYWRIGHT_MODULE||'playwright');

assert(/^http:\/\/127\.0\.0\.1:\d+$/.test(f.url));
const out=new URL('../backend/.cache/diseno-login-dashboard/',import.meta.url);await mkdir(out,{recursive:true});
const browser=await chromium.launch({headless:true,executablePath:process.env.BRAVO_BROWSER||'C:/Program Files/Google/Chrome/Application/chrome.exe'});
const errors=[],checks=[];let populated;
const ok=name=>{checks.push(name);console.log('PASS:',name);};
async function shot(p,name){await p.evaluate(()=>{document.activeElement?.blur();scrollTo(0,0);});await p.screenshot({path:fileURLToPath(new URL(name+'.png',out)),fullPage:true});}
async function fits(p){assert(await p.evaluate(()=>document.documentElement.scrollWidth<=innerWidth+2),'Desbordamiento horizontal');}
async function route(p,path){await p.evaluate(path=>{location.hash='#/'+path;dispatchEvent(new Event('hashchange'));},path);await p.waitForFunction(path=>document.querySelector('#app').dataset.route==='#/'+path&&document.querySelector('#app').getAttribute('aria-busy')==='false',path);assert.equal(await p.locator('.error-state').count(),0);}
async function call(p,path,method='GET',body){return p.evaluate(async({path,method,body})=>{try{return {status:200,...await (await import('./assets/js/services/api-service.js')).request(path,{method,body})};}catch(e){return {status:e.status,message:e.message};}},{path,method,body});}
async function login(p,username,password=f.password){await p.goto(f.url);await p.locator('[name=username]').fill(username);await p.locator('[name=password]').fill(password);await p.locator('.login-submit').click();await p.locator('.app-shell').waitFor();await p.waitForFunction(()=>document.querySelector('#app').getAttribute('aria-busy')==='false');}
async function stateReady(p){await p.locator('.kpi-card').first().waitFor();await p.locator('#page-content[aria-busy=true]').waitFor({state:'detached'}).catch(()=>{});}
try {
 const admin=await browser.newPage({viewport:{width:1440,height:900}});admin.on('pageerror',e=>errors.push(e.message));
 await admin.goto(f.url);await admin.locator('.login-submit').waitFor();await admin.evaluate(()=>document.fonts.ready);
 assert.equal(await admin.locator('#login-form input:not([hidden])').count(),2);
 await admin.locator('#password').fill('PruebaVisual');await admin.locator('#password-toggle').click();assert.equal(await admin.locator('#password').getAttribute('type'),'text');await admin.locator('#password-toggle').click();assert.equal(await admin.locator('#password').getAttribute('type'),'password');await admin.locator('#password').fill('');
 await shot(admin,'login-escritorio');await fits(admin);
 for(const [width,height,name] of [[820,1180,'login-tablet'],[390,844,'login-movil']]){await admin.setViewportSize({width,height});await fits(admin);await shot(admin,name);}
 await admin.setViewportSize({width:1440,height:900});
 await admin.locator('#username').fill(f.adminUsername);await admin.locator('#password').fill('incorrecta');await admin.locator('.login-submit').click();await admin.locator('#login-error:not([hidden])').waitFor();assert(await admin.locator('#login-error').innerText());ok('Contraseña incorrecta mantiene el error; mostrar/ocultar funciona; solo dos campos visibles');
 await login(admin,f.adminEmail);await stateReady(admin);ok('Login real con correo del Administrador');
 const before=await call(admin,'/dashboard');const activeBefore=before.data.states.filter(s=>['pending','repair','testing'].includes(s.key)).reduce((n,s)=>n+Number(s.count),0);
 await admin.locator('[name=reference]').fill('2000-01-01');const emptyResponse=admin.waitForResponse(r=>new URL(r.url()).pathname==='/api/v1/dashboard'&&r.status()===200);await admin.locator('#dashboard-filter [type=submit]').click();await emptyResponse;await stateReady(admin);
 const empty=await call(admin,'/dashboard?period=Mes&desde=2000-01-01T05:00:00Z&hasta=2000-02-01T05:00:00Z');assert.equal(Number(empty.data.completedServices),0);assert.equal(Number(empty.data.rescuesToday),0);assert.equal(Number(empty.data.monthlyIncomeCents),0);await shot(admin,'dashboard-vacio');ok('Dashboard real en período sin registros, sin estadísticas inventadas');
 const client=(await call(admin,'/clientes','POST',{name:'Cliente visual QA '+f.suffix,documentType:'Pasaporte',documentNumber:'CV'+f.suffix,phone:'987654321',address:'QA aislado'})).data;
 assert(client?.id);
 const car=(await call(admin,'/vehiculos','POST',{clientId:client.id,plate:'CV'+f.suffix.slice(0,5),brand:'Honda',model:'QA visual',year:2025,mileage:'100'})).data;
 const reception=(await call(admin,'/recepciones','POST',{clientId:client.id,vehicleId:car.id,mileage:'100',attentionType:'Directo',symptoms:'Servicio directo QA'})).data;
 const quote=(await call(admin,'/cotizaciones','POST',{receptionId:reception.id,status:'Aprobada',approval:{name:'Cliente visual QA',method:'Presencial'},items:[{type:'Servicio',concept:'Revisión autorizada QA',quantity:'1',unitPrice:'100',discount:'0'}]})).data;
 assert(quote?.id);
 const order=(await call(admin,'/ordenes','POST',{quoteId:quote.id,mechanicId:f.mechanicId})).data;assert(order?.id);
 await route(admin,'clientes');await route(admin,'inicio');await stateReady(admin);populated=await call(admin,'/dashboard');
 const active=populated.data.states.filter(s=>['pending','repair','testing'].includes(s.key)).reduce((n,s)=>n+Number(s.count),0);
 const card=admin.locator('.kpi-card').filter({has:admin.locator('h2',{hasText:'Órdenes activas'})});assert.equal(Number(await card.locator('.kpi-value').innerText()),active);assert.equal(active,activeBefore+1);ok('Indicadores coinciden con PostgreSQL y aumentan en una orden creada por API');
 await shot(admin,'dashboard-escritorio');
 for(const period of ['Semana','Mes','Año']){
  await admin.locator('[name=period]').selectOption(period);await admin.locator('[name=reference]').fill('2026-11-03');
  const res=admin.waitForResponse(r=>new URL(r.url()).pathname==='/api/v1/dashboard'&&r.status()===200);await admin.locator('#dashboard-filter [type=submit]').click();await res;await stateReady(admin);
 }
 await admin.locator('#tab-analisis').click();await admin.locator('.income-chart').waitFor();await shot(admin,'dashboard-analisis');await admin.locator('#tab-resumen').click();await stateReady(admin);ok('Filtros Semana/Mes/Año/fecha y gráficos de Análisis siguen consultando APIs');
 for(const [width,height,name] of [[820,1180,'dashboard-tablet'],[390,844,'dashboard-movil']]){await admin.setViewportSize({width,height});await fits(admin);await shot(admin,name);await admin.locator('#menu-toggle').click();await admin.locator('#sidebar .nav-item').first().waitFor();await admin.locator('#sidebar-scrim').click();}
 await admin.setViewportSize({width:1440,height:900});
 const modules=await admin.locator('.sidebar .nav-item').evaluateAll(links=>links.map(a=>a.hash.slice(2)));
 for(const mod of modules.filter(m=>m!=='inicio')){
  await route(admin,mod);assert.equal(await admin.locator('.dashboard-surface').count(),0);
  const props=()=>admin.locator('#page-content').evaluate(el=>[...el.querySelectorAll('.form-panel,.data-table,input,select,textarea')].slice(0,12).map(n=>{const s=getComputedStyle(n);return [n.tagName,s.backgroundColor,s.color,s.fontSize,s.padding];}));
  const styled=await props();await admin.evaluate(()=>document.querySelector('link[href$="gestion-carbon.css"]').disabled=true);const baseline=await props();assert.deepEqual(styled,baseline,'Cambio visual involuntario en '+mod);await admin.evaluate(()=>document.querySelector('link[href$="gestion-carbon.css"]').disabled=false);
 }
 await route(admin,'clientes/nuevo');await shot(admin,'clientes-formulario-claro');ok('Navegación completa y estilos de formularios/tablas de otros módulos idénticos con/sin el CSS nuevo');
 const mechanic=await browser.newPage({viewport:{width:1440,height:900}});mechanic.on('pageerror',e=>errors.push(e.message));await login(mechanic,f.mechanicUsername);await stateReady(mechanic);
 assert.equal(await mechanic.locator('.kpi-card').filter({hasText:'Ingresos'}).count(),0);assert(await mechanic.locator('#tab-analisis').isDisabled());assert.equal((await call(mechanic,'/reportes')).status,403);await shot(mechanic,'dashboard-mecanico');ok('Mecánico sin KPI financiero, Análisis deshabilitado y API financiera 403');
 const manager=await browser.newPage({viewport:{width:1440,height:900}});manager.on('pageerror',e=>errors.push(e.message));await login(manager,f.managerUsername);await stateReady(manager);assert(!(await manager.locator('#tab-analisis').isDisabled()));ok('Gerente conserva Análisis autorizado; Administrador y Mecánico mantienen sus roles');
 await admin.locator('#profile-toggle').click();await admin.locator('#logout').click();await admin.locator('#login-form').waitFor();await login(admin,f.adminUsername);await stateReady(admin);ok('Cierre de sesión y nuevo login por nombre de usuario');
 assert.deepEqual(errors,[]);ok('Sin errores JavaScript; escritorio, tablet y móvil sin desbordamiento global');
 await writeFile(new URL('resultado.json',out),JSON.stringify({passed:true,checks,errors,dashboardAPI:populated.data},null,2));
} finally {await browser.close();}
