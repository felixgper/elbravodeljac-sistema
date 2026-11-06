import {readIsolatedFixture} from './isolated-fixture.mjs';
const f=await readIsolatedFixture();
import assert from 'node:assert/strict';
import {createRequire} from 'node:module';
import {mkdir,writeFile} from 'node:fs/promises';
import {fileURLToPath} from 'node:url';
const {chromium}=createRequire(import.meta.url)(process.env.PLAYWRIGHT_MODULE||'playwright');

assert(/^http:\/\/127\.0\.0\.1:\d+$/.test(f.url),'Solo servidor efímero de pruebas');
assert(/^\d+$/.test(f.orderId));
const dir=new URL('../backend/.cache/auditoria-ficha-funcional/casillas-browser/',import.meta.url);
await mkdir(dir,{recursive:true});
const browser=await chromium.launch({headless:true,executablePath:process.env.BRAVO_BROWSER||'C:/Program Files/Google/Chrome/Application/chrome.exe'});
const errors=[];
try {
 const page=await browser.newPage({viewport:{width:1440,height:1000}});
 page.on('pageerror',e=>errors.push(e.message));
 await page.goto(f.url);
 await page.locator('[name=username]').fill(f.username);
 await page.locator('[name=password]').fill(f.password);
 await page.locator('.login-submit').click();
 await page.locator('.app-shell').waitFor();
 await page.evaluate(id=>{location.hash='#/ordenes/editar/'+id;window.dispatchEvent(new Event('hashchange'));},f.orderId);
 async function open(){await page.locator('#technical-control').click();await page.locator('#inspection-form').waitFor();}
 async function verify(){
  for(const [id,state] of Object.entries({fugas_0:'Bien',liquidos_0:'Atención',neumaticos_0:'Urgente',fugas_1:'Sin evaluar'}))assert.equal(await page.locator('[name=inspect_'+id+']').inputValue(),state);
  assert(await page.locator('[name=maint_pastillas_1]').isChecked());
  assert(!(await page.locator('[name=maint_pastillas_0]').isChecked()));
  for(const el of await page.locator('[name^=maint_aceite_],[name^=service_]').all())assert(!(await el.isChecked()));
  const values=await page.locator('[name^=execution_]:is(select)').evaluateAll(els=>els.map(el=>el.value));
  assert.equal(values.filter(x=>x==='Realizado').length,2);
  assert.equal(values.filter(x=>x==='No realizado').length,1);
 }
 await open();await verify();
 const response=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/v1/ordenes/'+f.orderId+'/control-tecnico'&&r.request().method()==='PUT');
 await page.locator('#inspection-form [type=submit]').click();
 assert.equal((await response).status(),200);
 await page.locator('#workshop-dialog').waitFor({state:'detached'});
 await page.reload();await page.locator('#technical-control').waitFor();
 await open();await verify();
 await page.screenshot({path:fileURLToPath(new URL('control-tras-f5.png',dir)),fullPage:true});
 assert.deepEqual(errors,[]);
 await writeFile(new URL('resultado.json',dir),JSON.stringify({passed:true,checks:['Sesión propia del mecánico','Guardar mediante formulario real HTTP 200','F5 y reabrir conserva cuatro estados','Pastillas seleccionadas, aceite y otros servicios vacíos','Dos trabajos realizados y uno no realizado','Sin errores JavaScript']},null,2));
 console.log('PASS: Chrome, formulario real, casillas y F5.');
} finally {await browser.close();}
