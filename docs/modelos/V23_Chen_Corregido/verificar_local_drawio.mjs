import {createServer} from 'node:http';
import {readFile,writeFile} from 'node:fs/promises';
import {createRequire} from 'node:module';
import {resolve} from 'node:path';
const dir=resolve('docs/modelos/V23_Chen_Corregido'),cache=resolve('backend/.cache/chen-v23'),base='El_Bravo_del_JAC_V23_Chen_Corregido';
const xml=await readFile(resolve(dir,base+'.drawio'),'utf8'),js=await readFile(resolve(cache,'viewer-static.min.js'));
const html=`<!doctype html><meta charset="utf-8"><style>body{margin:0;background:white}#diagram{width:9400px;height:5800px}</style><div id="diagram"></div><script src="/viewer.js"></script><script>window.ready=false;const e=document.getElementById('diagram');e.setAttribute('data-mxgraph',JSON.stringify({xml:${JSON.stringify(xml)},nav:false,resize:false,'auto-fit':false,'auto-origin':false,toolbar:''}));GraphViewer.createViewerForElement(e,v=>{window.viewer=v;window.ready=true});</script>`;
const server=createServer((q,s)=>{s.setHeader('Content-Type',q.url==='/viewer.js'?'application/javascript':'text/html');s.end(q.url==='/viewer.js'?js:html)});await new Promise(r=>server.listen(0,'127.0.0.1',r));
const {chromium}=createRequire(import.meta.url)('C:/Users/Usuario/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules/playwright');
const browser=await chromium.launch({headless:true,executablePath:'C:/Program Files/Google/Chrome/Application/chrome.exe'}),page=await browser.newPage({viewport:{width:1800,height:1200}});const errors=[];page.on('pageerror',e=>errors.push(e.message));
try{await page.goto('http://127.0.0.1:'+server.address().port);await page.waitForFunction(()=>window.ready,{timeout:30000});
const result=await page.evaluate(()=>{const g=viewer.graph;g.view.setScale(1);g.view.setTranslate(0,0);g.refresh();const svg=g.getSvg('#ffffff',1,20);const output=new XMLSerializer().serializeToString(svg);const model=g.getModel(),cell=model.getCell('clientes');const old=cell.geometry.x;g.moveCells([cell],40,0);const moved=cell.geometry.x;const codec=new mxCodec(),roundtrip=mxUtils.getXml(codec.encode(model));g.moveCells([cell],-40,0);return {svg:output,roundtrip,old,moved,cells:Object.keys(model.cells).length,edges:Object.values(model.cells).filter(c=>c.edge).length,bounds:g.getGraphBounds()};});
await writeFile(resolve(dir,base+'.svg'),result.svg);await writeFile(resolve(cache,'edicion-local-roundtrip.xml'),result.roundtrip);
if(result.moved!==result.old+40||result.edges!==84)throw Error('Editability/count failed');
await page.setContent(result.svg);const svg=page.locator('svg');await svg.screenshot({path:resolve(dir,base+'.png'),timeout:60000});
const validation={passed:true,method:'Motor oficial draw.io GraphViewer/mxGraph local; Chrome headless',source:'https://viewer.diagrams.net/js/viewer-static.min.js',remoteEditor:'Bloqueado HTTP 403; no se afirma apertura en editor remoto',loadedNativeXml:true,svgExported:true,pngExported:true,nativeGeometryEdit:true,geometryBefore:result.old,geometryAfter:result.moved,edges:result.edges,cells:result.cells,originalFileNotReplaced:true,errors};await writeFile(resolve(dir,'VALIDACION_RENDER_LOCAL.json'),JSON.stringify(validation,null,2));console.log(JSON.stringify(validation));
}finally{await browser.close();await new Promise(r=>server.close(r));}
