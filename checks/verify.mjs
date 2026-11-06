import assert from 'node:assert/strict';
import fs from 'node:fs';
import {getDashboard,getActiveOrders} from '../dist/assets/js/services/dashboard-service.js';
import {getStore} from '../dist/assets/js/services/store-service.js';
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
const source=await getDashboard(),firstCount=source.states[0].count;
assert.equal(source.states.reduce((n,state)=>n+state.count,0),getStore().orders.length,'El Dashboard demo consulta el store vigente');
assert.equal(getActiveOrders(source),getStore().orders.filter(o=>['Pendiente','En reparación','En pruebas'].includes(o.state)).length);
source.states[0].count=999;
assert.equal((await getDashboard()).states[0].count,firstCount,'Un consumidor no debe mutar los datos del Dashboard demo actual');
assert.equal(source.income.at(-1).cents,source.monthlyIncomeCents);
for(const folder of ['dist/assets/js','dist/assets/js/components','dist/assets/js/services','dist/assets/js/data']){
 for(const name of fs.readdirSync(folder).filter(n=>n.endsWith('.js'))){
  const source=fs.readFileSync(`${folder}/${name}`,'utf8');assert(!source.includes('fetch(')||name==='api-service.js','HTTP centralizado: cookies y CSRF en api-service');
 }
}
console.log('Demo actual: acceso, sesión, recordatorio, store del Dashboard y aislamiento de datos; sin conexión PostgreSQL.');
