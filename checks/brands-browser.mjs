import assert from 'node:assert/strict';
import {setDefaultResultOrder} from 'node:dns';
import {readFile,writeFile,mkdir} from 'node:fs/promises';
import {fileURLToPath} from 'node:url';
import {createRequire} from 'node:module';
import {randomBytes} from 'node:crypto';
setDefaultResultOrder('ipv4first');
const {chromium}=createRequire(import.meta.url)(process.env.PLAYWRIGHT_MODULE||'playwright');
const root=new URL('../',import.meta.url),dir=new URL('backend/.cache/brands-browser/',root);await mkdir(dir,{recursive:true});
const settings=Object.fromEntries((await readFile(new URL('.env.fase-e.local',root),'utf8')).trim().split(/\r?\n/).map(line=>{const i=line.indexOf('=');return [line.slice(0,i),line.slice(i+1)];}));
const browser=await chromium.launch({headless:true,executablePath:process.env.BRAVO_BROWSER||'C:/Program Files/Google/Chrome/Application/chrome.exe'});
const context=await browser.newContext({viewport:{width:1440,height:1000}}),page=await context.newPage();
const suffix=randomBytes(4).toString('hex').toUpperCase(),errors=[];
const report={date:new Date().toISOString(),url:'http://localhost:8080',checks:[]};let step='inicio';
page.on('pageerror',error=>errors.push(error.message));
async function route(path){step=path;await page.evaluate(p=>{location.hash='#/'+p;window.dispatchEvent(new Event('hashchange'));},path);await page.waitForFunction(p=>document.querySelector('#app').dataset.route==='#/'+p&&document.querySelector('#app').getAttribute('aria-busy')==='false',path);assert.equal(await page.locator('.error-state').count(),0,'Error al abrir '+path);}
async function fill(data){for(const [name,value]of Object.entries(data)){const input=page.locator(`input[name="${name}"],select[name="${name}"],textarea[name="${name}"]`).first();if(name==='password'){await input.waitFor();await input.evaluate((el,value)=>el.value=value,String(value));}else if(name==='brand'){await input.locator('..').locator('[role=combobox]').fill(String(value));await input.locator('..').getByRole('option',{name:String(value),exact:true}).click();}else if(await input.evaluate(el=>el.tagName)==='SELECT')await input.selectOption(String(value));else await input.fill(String(value));}}
async function save(path,method='POST',button='#edit-form [type=submit]',status=201){const pending=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/v1'+path&&r.request().method()===method);await page.locator(button).first().click();const r=await pending;assert.equal(r.status(),status,method+' '+path+' devolvió '+r.status());return status===204?null:(await r.json()).data;}
async function api(path,options){return page.evaluate(async({path,options})=>(await(await import('./assets/js/services/api-service.js')).request(path,options)).data,{path,options});}
async function login(target,username,password){await target.goto(report.url);await target.locator('#login-form').waitFor();await target.locator('[name=username]').fill(username);await target.evaluate(p=>document.querySelector('[name=password]').value=p,password);await target.locator('.login-submit').click();await target.locator('.app-shell').waitFor();await target.waitForFunction(()=>document.querySelector('#app').getAttribute('aria-busy')==='false');}
try{
 let ready=false;for(let i=0;i<90;i++){try{const r=await context.request.get(report.url+'/readyz',{timeout:2000});ready=r.ok()&&(await r.json()).technicalConfigured;}catch{}if(ready)break;await new Promise(resolve=>setTimeout(resolve,1000));}assert(ready);
 await login(page,settings.DEV_ADMIN_USERNAME,settings.DEV_ADMIN_PASSWORD);
 const expected=['JAC','Toyota','Nissan','Hyundai','Kia','Chevrolet','Ford','Volkswagen','Suzuki','Mitsubishi','Mazda','Renault','Chery','Changan','Haval','BYD','MG','Great Wall','BMW','Mercedes-Benz','Audi','Volvo','Peugeot','Citroën','BAIC','Jinbei','DFM','Brilliance','Changhe','Mahindra','SsangYong','Daewoo','Land Rover','Honda'];
 const catalogue=await api('/vehiculos/marcas');for(const brand of expected)assert(catalogue.includes(brand),'Falta '+brand);assert.equal(new Set(catalogue.map(v=>v.toLowerCase())).size,catalogue.length);
 report.checks.push('34 marcas disponibles desde API/PostgreSQL antes de registrar vehículos');
 await route('clientes/nuevo');await fill({name:'Marcas navegador '+suffix,documentType:'Pasaporte',documentNumber:'BR'+suffix,phone:'+51987654321',address:'Prueba aislada de catálogo'});const client=await save('/clientes');report.clientId=client.id;
 const brandSelect=()=>page.locator('select[name=brand]'),picker=()=>brandSelect().locator('..').locator('[role=combobox]');
 async function choose(brand,query=brand.slice(0,2)){await picker().fill(query);await brandSelect().locator('..').getByRole('option',{name:brand,exact:true}).click();}
 async function checkFullCatalogue(){await picker().click();for(const brand of expected)await brandSelect().locator('..').getByRole('option',{name:brand,exact:true}).waitFor();await brandSelect().locator('..').getByRole('option',{name:'Otra marca',exact:true}).waitFor();}
 report.vehicles=[];
 for(const [i,brand]of ['Toyota','Hyundai','BMW','Honda'].entries()){
  await route('vehiculos/nuevo?cliente='+client.id);assert.equal(await picker().inputValue(),'');await checkFullCatalogue();
  await choose(brand);await fill({plate:'BR'+suffix.slice(0,4)+i,model:'Prueba marcas',year:2025,mileage:'0.00'});
  const vehicle=await save('/vehiculos');assert.equal(vehicle.brand,brand);report.vehicles.push({id:vehicle.id,brand});
  await route('vehiculos/editar/'+vehicle.id);assert.equal(await picker().inputValue(),brand);await checkFullCatalogue();await choose(brand);
  await fill({model:'Editado '+brand});const edited=await save('/vehiculos/'+vehicle.id,'PATCH',undefined,200);assert.equal(edited.brand,brand);
  await route('vehiculos/editar/'+vehicle.id);await page.reload();await page.waitForFunction(()=>document.querySelector('#app').getAttribute('aria-busy')==='false');assert.equal(await picker().inputValue(),brand);assert.equal((await api('/vehiculos/'+vehicle.id)).model,'Editado '+brand);
 }
 report.checks.push('Alta, consulta, edición y F5: Toyota, Hyundai, BMW y Honda');
 const first=report.vehicles[0];await route('vehiculos/editar/'+first.id);await choose('Honda');assert.equal((await save('/vehiculos/'+first.id,'PATCH',undefined,200)).brand,'Honda');await route('vehiculos/editar/'+first.id);await choose('Toyota');await save('/vehiculos/'+first.id,'PATCH',undefined,200);
 report.checks.push('Cambiar una marca existente en edición y volver a consultarla');
 await route('vehiculos/nuevo?cliente='+client.id);await picker().click();await brandSelect().locator('..').getByRole('option',{name:'Otra marca',exact:true}).click();const other=page.locator('[name=otherBrand]');assert(await other.isVisible());assert(await other.evaluate(el=>el.required&&!el.checkValidity()));
 const custom='Marca Técnica '+suffix;await other.fill(custom);await fill({plate:'BR'+suffix.slice(0,4)+'X',model:'Personalizada',year:2025,mileage:'0.00'});const vehicle=await save('/vehiculos');assert.equal(vehicle.brand,custom);report.customVehicleId=vehicle.id;
 assert.equal((await api('/vehiculos/marcas')).filter(b=>b===custom).length,1);
 await route('vehiculos/editar/'+vehicle.id);await page.reload();await page.waitForFunction(()=>document.querySelector('#app').getAttribute('aria-busy')==='false');assert.equal(await picker().inputValue(),custom);
 await route('vehiculos/nuevo?cliente='+client.id);await choose(custom,'Marca T');assert.equal(await brandSelect().inputValue(),custom);
 await picker().fill('No existe '+suffix);await brandSelect().locator('..').getByRole('option',{name:'Otra marca',exact:true}).waitFor();assert.equal(await picker().evaluate(el=>el.checkValidity()),false);
 report.checks.push('Otra marca obligatoria, validada y persistida; reutilizable desde el mismo catálogo después de F5');
 assert.deepEqual(errors,[]);report.passed=true;await writeFile(new URL('resultado.json',dir),JSON.stringify(report,null,2));console.log('Chrome + API + PostgreSQL: 34 marcas, búsqueda, alta/edición Toyota/Hyundai/BMW/Honda y Otra marca persistida: APROBADO.');
}catch(error){await writeFile(new URL('fallo.txt',dir),step+'\n'+String(error.stack||error));throw Error('Falló marcas: consultar backend/.cache/brands-browser/fallo.txt, sin credenciales.');}
finally{await browser.close();}
