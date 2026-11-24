// Prueba ligera: ejecuta bindShell real con un DOM mínimo; no abre Chrome ni hace HTTP.
import assert from 'node:assert/strict';
import fs from 'node:fs';
import {setSession,clearSession} from '../dist/assets/js/services/api-service.js';
import {allowedModules,permissionChecker} from '../dist/assets/js/services/permissions.js';
const source=fs.readFileSync('dist/assets/js/app.js','utf8');
const bindSource=source.slice(source.indexOf('function bindShell(session) {'),source.indexOf('async function renderDashboard('));
assert(bindSource.startsWith('function bindShell(session) {'));
const css=fs.readFileSync('dist/assets/css/gestion-carbon.css','utf8');
const desktop=css.slice(css.indexOf('@media (min-width:901px){'),css.indexOf('/* Solo Inicio:'));
assert(desktop.includes('.management-chrome .sidebar nav{')&&desktop.includes('flex-direction:row')&&desktop.includes('overflow-x:auto'));
assert(desktop.includes('var(--navigation-height) + 24px'),'El espacio del contenido incluye la navegación');
assert(desktop.includes('.management-chrome .workspace,.management-chrome.dashboard-surface .workspace'),'Regla compartida por todos los módulos');
const policy=JSON.parse(fs.readFileSync('backend/internal/domain/permission-policy.json','utf8'));
let scenarios=0;
for(const role of ['Gerente','Administrador','Mecánico']){
 const permissions=Object.entries(policy.defaults[role]).flatMap(([m,actions])=>actions.map(a=>m+'.'+a));
 setSession({data:{id:'unit-user',name:'Cuenta sintética',roles:[role],permissions},csrfToken:'unit-csrf'});
 const allowed=allowedModules(role),ids=allowed.map(m=>m.id);
 if(role==='Mecánico'){assert(!ids.includes('pagos'));assert(!ids.includes('usuarios'));assert(!ids.includes('cotizaciones'));}
 assert.deepEqual(allowedModules('Nombre de rol no confiable'),allowed,'Los permisos efectivos controlan el menú');
 for(const width of [390,768,900,901,1366,1920]){
  let activeElement=null,scrolls=0;
  const listeners=new Map(),attrs=new Map(),classes=new Set();
  const element=id=>({id,hidden:false,inert:false,value:'',innerHTML:'',focus(){activeElement=this;},scrollIntoView(){scrolls++;},setAttribute(k,v){attrs.set(id+':'+k,v);},addEventListener(k,f){listeners.set(id+':'+k,f);},querySelector(){return active;}});
  const active=element('active'),last=element('last'),app=element('app'),sidebar=element('sidebar');
  app.classList={remove:c=>classes.delete(c),contains:c=>classes.has(c),toggle(c){if(classes.has(c)){classes.delete(c);return false;}classes.add(c);return true;}};
  const elements=Object.fromEntries(['menu-toggle','sidebar-scrim','profile-toggle','profile-menu','module-search','search-results','logout'].map(id=>[id,element(id)]));elements.sidebar=sidebar;elements['profile-menu'].hidden=true;
  const root={querySelector:()=>app};
  const document={getElementById:id=>{assert(elements[id],id);return elements[id];},querySelector:()=>active,querySelectorAll:()=>[active,last],get activeElement(){return activeElement;}};
  const media={matches:width<=900,addEventListener(k,f){listeners.set('media:'+k,f);},removeEventListener(k){listeners.delete('media:'+k);}};
  const factory=new Function('root','document','matchMedia','permissionChecker','allowed','icon','logout','notify','location','let disposeShell;'+bindSource+';return {bindShell,dispose:()=>disposeShell()};');
  const shell=factory(root,document,()=>media,permissionChecker,allowed,()=>'',async()=>{},()=>{},{});
  shell.bindShell({role});
  const fire=(id,type,event={})=>{const f=listeners.get(id+':'+type);assert(f,id+':'+type);return f(event);};
  assert.equal(sidebar.inert,media.matches);
  if(width>900){
   assert.equal(scrolls,1);assert.equal(attrs.get('menu-toggle:aria-expanded'),'true');
   fire('menu-toggle','click');assert(!classes.has('sidebar-compact'));assert(!classes.has('sidebar-open'));
   activeElement=elements['module-search'];root.onkeydown({key:'Escape',target:{tagName:'DIV'}});assert.equal(activeElement,elements['module-search'],'Escape de escritorio no enfoca el botón oculto');
  }else{
   fire('menu-toggle','click');assert(classes.has('sidebar-open'));assert.equal(sidebar.inert,false);assert.equal(elements['sidebar-scrim'].hidden,false);assert.equal(activeElement,active);
   activeElement=last;let prevented=false;root.onkeydown({key:'Tab',target:{tagName:'A'},preventDefault(){prevented=true;}});assert(prevented);assert.equal(activeElement,elements['menu-toggle']);
   root.onkeydown({key:'Escape',target:{tagName:'A'}});assert(!classes.has('sidebar-open'));assert(sidebar.inert);assert(elements['sidebar-scrim'].hidden);
   fire('menu-toggle','click');fire('sidebar-scrim','click');assert(!classes.has('sidebar-open'));
   media.matches=false;fire('media','change');assert.equal(sidebar.inert,false);assert.equal(attrs.get('menu-toggle:aria-expanded'),'true');assert.equal(scrolls,1);
  }
  elements['module-search'].value='Reportes';fire('module-search','input');
  assert.equal(elements['search-results'].innerHTML.includes('inicio?tab=analisis'),permissions.includes('reportes.ver'));
  shell.dispose();assert(!listeners.has('media:change'));scenarios++;
 }
}
clearSession();
console.log(`Navegación ligera: PASS ${scenarios} escenarios (3 perfiles × 6 anchos); permisos efectivos, foco, apertura/cierre móvil, transición y contrato CSS. Sin renderizado de navegador.`);
