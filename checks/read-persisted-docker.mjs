import {setDefaultResultOrder} from 'node:dns';
import assert from 'node:assert/strict';
import {readFile,writeFile} from 'node:fs/promises';
import {createRequire} from 'node:module';
setDefaultResultOrder('ipv4first');
const require=createRequire(import.meta.url),{chromium}=require(process.env.PLAYWRIGHT_MODULE||'playwright');
const root=new URL('../',import.meta.url),reportPath=new URL('backend/.cache/live-docker/resultado.json',root);
const report=JSON.parse(await readFile(reportPath,'utf8'));
const settings=Object.fromEntries((await readFile(new URL('.env.fase-e.local',root),'utf8')).trim().split(/\r?\n/).map(line=>{const i=line.indexOf('=');return [line.slice(0,i),line.slice(i+1)];}));
const browser=await chromium.launch({headless:true,executablePath:process.env.BRAVO_BROWSER||'C:/Program Files/Google/Chrome/Application/chrome.exe'});
try{
 const context=await browser.newContext(),page=await context.newPage();
 const ready=await context.request.get(report.url+'/readyz');assert.equal(ready.status(),200);assert.equal((await ready.json()).technicalConfigured,true);
 await page.goto(report.url);await page.locator('#login-form').waitFor();await page.locator('[name=username]').fill(settings.DEV_ADMIN_USERNAME);
 await page.evaluate(password=>{document.querySelector('[name=password]').value=password;},settings.DEV_ADMIN_PASSWORD);await page.locator('.login-submit').click();await page.locator('.kpi-grid').waitFor();
 const persisted=await page.evaluate(async report=>{
  const {request}=await import('./assets/js/services/api-service.js'),technical=await import('./assets/js/services/technical-service.js');
  const get=async path=>(await request(path)).data;
  const scans=await technical.listScans(report.diagnosisId),evidence=await technical.listEvidence(report.diagnosisId);
  const file=await fetch(technical.evidenceURL(report.diagnosisId,report.evidenceId));
  return {client:await get('/clientes/'+report.clientId),order:await get('/ordenes/'+report.orderId),payments:await get('/ordenes/'+report.orderId+'/pagos'),delivery:await get('/ordenes/'+report.orderId+'/entrega'),scans:scans.map(s=>s.id),evidence:evidence.map(e=>e.id),fileStatus:file.status,fileBytes:(await file.arrayBuffer()).byteLength};
 },report);
 assert.equal(persisted.client.id,report.clientId);assert.equal(persisted.order.code,report.orderCode);assert.equal(persisted.order.state,'Finalizada');assert.equal(persisted.payments.paid,'120.00');assert.equal(persisted.payments.balance,'0.00');assert.equal(persisted.delivery.id,report.deliveryId);
 assert(persisted.scans.includes(report.scanId));assert(persisted.evidence.includes(report.evidenceId));assert.equal(persisted.fileStatus,200);assert(persisted.fileBytes>0);
 report.persistedAfterContainerRestart=true;report.persistenceCheckedAt=new Date().toISOString();await writeFile(reportPath,JSON.stringify(report,null,2));
 console.log('Lectura posterior al reinicio de Go, PostgreSQL y MongoDB: cliente, OT, pago, entrega, escáner y archivo GridFS conservados. APROBADO.');
}catch(error){await writeFile(new URL('backend/.cache/live-docker/fallo-persistencia.txt',root),String(error.stack||error));throw Error('Falló la consulta de persistencia; revisar el archivo local de diagnóstico (sin imprimir credenciales).');}
finally{await browser.close();}
