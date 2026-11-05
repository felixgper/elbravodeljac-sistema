import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import * as api from '../dist/assets/js/services/workshop-service.js';
import {resetDemo,decimalString,decimalMinor} from '../dist/assets/js/services/store-service.js';
import {incomeChart,horizontalBars,donutChart} from '../dist/assets/js/components/charts.js';
const storage=new Map();globalThis.localStorage={getItem:k=>storage.get(k)??null,setItem:(k,v)=>storage.set(k,String(v)),removeItem:k=>storage.delete(k)};
resetDemo();
const input={quoteId:'COT-060',mechanicId:'u-pedro'};
const first=api.createOrder(input,'u-admin');api.deleteOrder(first.id);const second=api.createOrder(input,'u-admin');assert.notEqual(first.id,second.id,'Eliminar una OT no debe permitir reutilizar su código');
const update={mechanicId:'u-pedro',state:'En pruebas',note:'Prueba inicial',test:'Prueba de ruta',result:'Satisfactoria'};
api.updateOrder(second.id,update,'u-admin');api.updateOrder(second.id,{...update,state:'En reparación',note:'Nueva intervención posterior a la prueba',test:'',result:'No aplica'},'u-admin');
assert.throws(()=>api.updateOrder(second.id,{...update,state:'Finalizada',note:'Intento con prueba antigua',test:'',result:'No aplica'},'u-admin'),/prueba satisfactoria/);
api.updateOrder(second.id,{...update,state:'Finalizada',note:'Nueva prueba posterior a reparación'},'u-admin');
const users=api.getStore().users;const jimmy=users.find(u=>u.id==='u-jimmy'),admin=users.find(u=>u.id==='u-admin');await api.saveUser({...jimmy,active:false},jimmy.id);
await assert.rejects(api.saveUser({...admin,active:false},admin.id),/al menos un/);
assert.equal(decimalString('00002023.5'),'2023.50');assert.equal(decimalMinor('99999999.99'),9999999999);
for(const invalid of ['-1','0.001','1,25','NaN','Infinity','100000000.00'])assert.throws(()=>decimalMinor(invalid));
assert(!horizontalBars([{label:'Sin asignaciones',count:0}],'mechanics').includes('NaN'));
assert(!donutChart([{label:'Pendientes',count:0,color:'#999'}]).includes('NaN'));
assert(!incomeChart([{month:'A',cents:99999999},{month:'B',cents:1}],320).includes('NaN'));
const realCrypto=globalThis.crypto;Object.defineProperty(globalThis,'crypto',{value:{getRandomValues:realCrypto.getRandomValues.bind(realCrypto)},configurable:true});
for(const input of ['', 'abc','Contraseña ficticia 2026','ñá😀'.repeat(20)]){
 assert.equal(await api.passwordDigest(input),createHash('sha256').update(input).digest('hex'),'El resumen alternativo debe coincidir con SHA-256');
}
Object.defineProperty(globalThis,'crypto',{value:realCrypto,configurable:true});
// Eliminaciones: respetan la integridad y la trazabilidad
resetDemo();
assert.throws(()=>api.deleteClient('c1'),/vehículo/);
const nc=api.saveClient({documentType:'DNI',documentNumber:'11112222',name:'Temporal',phone:'999888777',email:'',address:'Lima'});api.deleteClient(nc.id);assert(!api.getStore().clients.some(c=>c.id===nc.id));
assert.throws(()=>api.deleteVehicle('v1'),/historial/);
const sp=api.saveSupplier({name:'Temp',ruc:'20999888777',phone:'987654321',parts:'X'});api.deleteSupplier(sp.id);assert(!api.getStore().suppliers.some(p=>p.id===sp.id));
assert.throws(()=>api.deleteUser('u-admin','u-admin'),/propio usuario/);
assert.throws(()=>api.deleteUser('u-pedro','u-admin'),/inactivo/);
const nu=await api.saveUser({name:'Temp',username:'temporal',email:'t@demo.local',role:'Mecánico',active:true,password:'Temporal2026!'});api.deleteUser(nu.id,'u-admin');assert(!api.getStore().users.some(u=>u.id===nu.id));
// Roles y permisos
const perms=await import('../dist/assets/js/services/permissions.js');
resetDemo();
assert(perms.permissionChecker('Gerente')('usuarios','roles'));
assert(perms.permissionChecker('Administrador')('pagos','registrar'));
assert(!perms.permissionChecker('Administrador')('usuarios','roles'));
assert(!perms.permissionChecker('Auxiliar')('inicio','ver'));
assert.equal(perms.homePath('Mecánico'),'inicio');
assert.throws(()=>api.saveRolePermissions('Gerente',{inicio:['ver']}),/Gerente/);
assert.throws(()=>api.saveRolePermissions('Auxiliar',{inicio:['ver']}),/rol válido/);
api.saveRolePermissions('Administrador',{inicio:['ver'],clientes:['crear','eliminar'],pagos:['ver']});
const saved=perms.rolePermissions().Administrador;
assert(!saved.clientes,'Sin "ver" no se conservan acciones sueltas');assert.deepEqual(saved.pagos,['ver']);assert.deepEqual(saved.inicio,['ver']);
api.resetRolePermissions('Administrador');assert(perms.permissionChecker('Administrador')('clientes','editar'));
console.log('Casos límite verificados: códigos sin reutilización, pruebas posteriores a reparación, último administrador, decimales exactos, gráficos vacíos y compatibilidad de claves ficticias en HTTP local.');
