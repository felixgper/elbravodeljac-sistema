// Official draw.io embed API. Sends structure only, no application data/accounts.
import {createServer} from 'node:http';
import {readFile,writeFile,mkdir} from 'node:fs/promises';
import {createRequire} from 'node:module';
import {resolve,dirname} from 'node:path';
import {fileURLToPath} from 'node:url';
const here=dirname(fileURLToPath(import.meta.url)),base='El_Bravo_del_JAC_V23_Chen_Corregido';
const cache=resolve(here,'../../../backend/.cache/chen-v23');await mkdir(cache,{recursive:true});
const {chromium}=createRequire(import.meta.url)(process.env.PLAYWRIGHT_MODULE||'C:/Users/Usuario/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules/playwright');
const html=`<!doctype html><meta charset="utf-8"><style>html,body{margin:0;width:100%;height:100%}iframe{border:0;width:100%;height:100%}</style><iframe id="editor" src="https://embed.diagrams.net/?embed=1&proto=json&spin=1&saveAndExit=0&noSaveBtn=1&noExitBtn=1"></iframe><script>window.messages=[];window.addEventListener('message',e=>{if(['https://embed.diagrams.net','https://app.diagrams.net'].includes(e.origin)){try{window.messages.push(JSON.parse(e.data))}catch{}}});window.sendToEditor=m=>document.getElementById('editor').contentWindow.postMessage(JSON.stringify(m),'https://embed.diagrams.net');</script>`;
const server=createServer((req,res)=>{res.writeHead(200,{'Content-Type':'text/html;charset=utf-8'});res.end(html)});await new Promise(r=>server.listen(0,'127.0.0.1',r));
const browser=await chromium.launch({headless:true,executablePath:'C:/Program Files/Google/Chrome/Application/chrome.exe'});
const page=await browser.newPage({viewport:{width:1800,height:1200}}),diagnostics=[];
page.on('pageerror',e=>diagnostics.push(e.message));page.on('requestfailed',r=>diagnostics.push({url:r.url(),error:r.failure()?.errorText}));page.on('console',m=>{if(m.type()==='error')diagnostics.push(m.text())});
async function message(action,event){await page.evaluate(m=>{window.messages=[];window.sendToEditor(m)},action);await page.waitForFunction(e=>window.messages.some(m=>m.event===e),event,{timeout:60000});const result=await page.evaluate(e=>window.messages.find(m=>m.event===e),event);if(result.error)throw Error(result.error);return result;}
function decodeData(data){if(!data.startsWith('data:'))throw Error('Sin imagen exportada');const comma=data.indexOf(',');return data.slice(0,comma).endsWith(';base64')?Buffer.from(data.slice(comma+1),'base64'):Buffer.from(decodeURIComponent(data.slice(comma+1)));}
try{
 await page.goto('http://127.0.0.1:'+server.address().port,{waitUntil:'domcontentloaded'});
 await page.waitForFunction(()=>window.messages.some(m=>m.event==='init'),null,{timeout:60000});
 const xml=await readFile(resolve(here,base+'.drawio'),'utf8');
 const loaded=await message({action:'load',xml,fit:1,title:base},'load');
 const raw=await message({action:'export',format:'xml'},'export'),roundtrip=raw.xml||raw.data;
 if(!roundtrip||!roundtrip.includes('tipos_documento')||!roundtrip.includes('R09_b_card'))throw Error('Faltan elementos editables al reexportar');
 await writeFile(resolve(cache,'roundtrip.drawio'),roundtrip);
 const svg=await message({action:'export',format:'svg'},'export');await writeFile(resolve(here,base+'.svg'),decodeData(svg.data));
 const png=await message({action:'export',format:'png',scale:1,border:12,background:'#FFFFFF'},'export');const bytes=decodeData(png.data);
 if(bytes.subarray(1,4).toString()!=='PNG')throw Error('Exportación no es PNG');
 await writeFile(resolve(here,base+'.png'),bytes);
 await page.screenshot({path:resolve(cache,'editor-v23.png')});
 // Test a local geometry edit/reload/re-export through the supported API.
 const edited=await page.evaluate(xml=>{const d=new DOMParser().parseFromString(xml,'application/xml'),g=d.querySelector('mxCell[id="clientes"] mxGeometry');if(!g)throw Error('Sin geometría editable');const x=Number(g.getAttribute('x'));g.setAttribute('x',String(x+40));return {xml:new XMLSerializer().serializeToString(d),x:x+40};},xml);
 await message({action:'load',xml:edited.xml,fit:1,title:'Validación temporal de editabilidad'},'load');
 const second=await message({action:'export',format:'xml'},'export');const editedExport=second.xml||second.data;
 await writeFile(resolve(cache,'edicion-temporal-roundtrip.drawio'),editedExport);
 const actual=await page.evaluate(xml=>Number(new DOMParser().parseFromString(xml,'application/xml').querySelector('mxCell[id="clientes"] mxGeometry')?.getAttribute('x')),editedExport);
 if(actual!==edited.x)throw Error('No se conservó la edición de una entidad');
 const result={passed:true,editor:'https://embed.diagrams.net/',loaded:true,xmlRoundtrip:true,nativeGeometryEdit:true,pngExported:true,svgExported:true,originalFileNotReplaced:true,loadBounds:loaded.bounds||null,diagnostics};
 await writeFile(resolve(here,'VALIDACION_DRAWIO.json'),JSON.stringify(result,null,2));console.log(JSON.stringify(result));
}catch(error){await writeFile(resolve(here,'VALIDACION_DRAWIO.json'),JSON.stringify({passed:false,error:String(error),diagnostics,frames:page.frames().map(f=>f.url()),messages:await page.evaluate(()=>window.messages)},null,2));await page.screenshot({path:resolve(cache,'error-editor.png')});throw error;}
finally{await browser.close();await new Promise(r=>server.close(r));}
