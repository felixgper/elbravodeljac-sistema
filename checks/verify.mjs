import assert from 'node:assert/strict';
import fs from 'node:fs';
import {dashboardDemo} from '../dist/assets/js/data/demo.js';
import {getDashboard,getActiveOrders} from '../dist/assets/js/services/dashboard-service.js';
import {modules} from '../dist/assets/js/components/navigation.js';
import {getSession,getRememberedUser,login,logout} from '../dist/assets/js/services/auth-service.js';
const memory = () => {const map=new Map();return {getItem:key=>map.get(key)??null,setItem:(key,value)=>map.set(key,String(value)),removeItem:key=>map.delete(key),values:()=>[...map.values()]}};
globalThis.sessionStorage=memory();globalThis.localStorage=memory();
assert.equal(getSession(),null);
await assert.rejects(login({username:'admin',password:'incorrecta',remember:false}));
assert.equal(getSession(),null,'Un acceso incorrecto no debe crear sesión');
await login({username:' ADMIN ',password:'Bravo2026!',remember:true});
assert.equal(getSession().username,'admin');assert.equal(getRememberedUser(),'admin');
assert(!JSON.stringify(sessionStorage.values()).includes('Bravo2026!'));
logout();assert.equal(getSession(),null);assert.equal(getRememberedUser(),'admin');
await login({username:'admin',password:'Bravo2026!',remember:false});assert.equal(getRememberedUser(),'');
sessionStorage.setItem('bravo-jac:demo-session:v1','{invalid');assert.equal(getSession(),null);
assert.equal(getActiveOrders(dashboardDemo),10);
const cloned=await getDashboard();cloned.states[0].count=999;assert.equal(dashboardDemo.states[0].count,2,'El consumidor no debe mutar la fuente demo');
assert.equal(dashboardDemo.states.reduce((n,s)=>n+s.count,0),35);
assert.equal(dashboardDemo.income.at(-1).cents,dashboardDemo.monthlyIncomeCents);
const screens=JSON.parse(fs.readFileSync('docs/balsamiq-v14.1.json','utf8'));
assert.equal(screens.length,23);
const dashboard=screens.find(s=>s.name==='02 - Dashboard');
const sourceNav=dashboard.controls.filter(c=>c.typeID==='Label'&&Number(c.x)<220).map(c=>c.properties.text.trim().replace(/^>\s*/,'')).filter(t=>t!=='TALLER AUTOMOTRIZ');
assert.deepEqual(modules.map(m=>m.label),sourceNav,'El menú debe conservar las opciones del Balsamiq');
const mapped=['01 - Inicio de sesión','02 - Dashboard',...modules.flatMap(m=>m.screens??[])];
assert.deepEqual([...new Set(mapped)].sort(),screens.map(s=>s.name).sort(),'Las 23 pantallas deben estar contempladas');
const texts=dashboard.controls.map(c=>c.properties?.text);
assert(texts.includes('S/ 12,450'));assert(texts.includes('10'));assert(texts.includes('24'));assert(texts.includes('3'));
const create=screens.find(s=>s.name==='11 - Nueva orden de trabajo');
const update=screens.find(s=>s.name==='11B - Actualizar orden');
assert.notEqual(create.id,update.id);assert(create.controls.some(c=>c.properties?.text==='Cotización aprobada (obligatoria)'));
assert(update.controls.some(c=>c.properties?.text==='Observación / avance de esta actualización'));
for(const folder of ['dist/assets/js','dist/assets/js/components','dist/assets/js/services','dist/assets/js/data']){
 for(const name of fs.readdirSync(folder).filter(n=>n.endsWith('.js'))){
  const source=fs.readFileSync(`${folder}/${name}`,'utf8');assert(!source.includes('fetch(')||name==='api-service.js','HTTP centralizado: cookies y CSRF en api-service');
 }
}
console.log('Correcto: acceso demo, sesión, recordatorio, aislamiento de datos, totales y correspondencia con las 23 pantallas y 13 opciones del Balsamiq.');
