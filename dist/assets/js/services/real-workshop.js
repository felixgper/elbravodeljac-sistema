import {request,allPages,realSession} from './api-service.js';
import {realStore,replaceRealStore,updateRealStore} from './real-state.js';
const pick=(v,keys)=>Object.fromEntries(keys.filter(k=>v[k]!==undefined).map(k=>[k,v[k]]));
export function cents(value) {
  if(!/^\d+(\.\d{1,2})?$/.test(String(value)))throw new Error('Importe inválido recibido del servidor.');
  const [a,b='']=String(value).split('.'),n=BigInt(a)*100n+BigInt(b.padEnd(2,'0'));
  if(n>BigInt(Number.MAX_SAFE_INTEGER))throw new Error('Importe fuera del rango de visualización exacta.');
  return Number(n);
}
export const user=v=>({...v,role:(v.roles||[]).map(r=>typeof r==='string'?r:r.name).join(' / '),roleNames:(v.roles||[]).map(r=>typeof r==='string'?r:r.name)});
export const quote=v=>({...v,items:(v.items||[]).map(i=>({...i,origin:i.type==='Servicio'?'—':i.origin==='CLIENTE'?'Cliente':'Taller',unitCents:cents(i.unitPrice)}))});
const stateName=v=>({creada:'Pendiente',en_reparacion:'En reparación',en_pruebas:'En pruebas',finalizada:'Finalizada',cancelada:'Cancelada'})[v]||v;
export function order(v) {
  const assignments=(v.assignments||[]).filter(a=>a.active);
  // Las tres bitácoras se conservan, sin inventar asociaciones entre eventos.
  const updates=[...(v.advances||[]).map(a=>({...a,test:'',result:'No aplica',state:''})),
    ...(v.tests||[]).map(a=>({...a,result:a.result==='satisfactoria'?'Satisfactoria':'Fallida',state:''})),
    ...(v.states||[]).map(a=>({...a,note:'Cambio de estado',test:'',result:'No aplica',state:stateName(a.state)}))].sort((a,b)=>a.at.localeCompare(b.at));
  return {...v,mechanicId:assignments.find(a=>a.role==='MECANICO')?.userId||'',assistantId:assignments.find(a=>a.role==='AUXILIAR')?.userId||'',updates,delivery:v.delivered?{pendingDetails:true}:null};
}
const specs={clients:['clientes','clientes'],vehicles:['vehiculos','vehiculos'],receptions:['recepciones','recepciones'],diagnoses:['diagnosticos','diagnosticos'],quotes:['cotizaciones','cotizaciones'],orders:['ordenes','ordenes'],rescues:['auxilios','auxilio'],suppliers:['proveedores','proveedores'],services:['servicios','cotizaciones'],spareParts:['repuestos','cotizaciones'],units:['unidades-auxilio','auxilio'],users:['mecanicos','mecanicos']};
const needs={clientes:['clients','vehicles'],vehiculos:['clients','vehicles'],recepciones:['clients','vehicles','receptions','diagnoses'],diagnosticos:['clients','vehicles','receptions','diagnoses'],cotizaciones:['clients','vehicles','receptions','diagnoses','quotes','orders','services','spareParts'],ordenes:['clients','vehicles','receptions','diagnoses','quotes','orders','users'],pagos:['clients','vehicles','receptions','diagnoses','quotes','orders','users'],auxilio:['clients','vehicles','rescues','units','users','orders','quotes','diagnoses','receptions'],mecanicos:['users','orders','clients','vehicles','receptions','diagnoses','quotes'],proveedores:['suppliers']};
let generation=0;
export async function loadModule(module,params=new URLSearchParams()) {
  const current=++generation,sessionRef=realSession(),s={},permissions=sessionRef?.permissions||[],can=m=>permissions.includes(m+'.ver');
  const unavailable=[];
  if(module==='usuarios') {
    s.users=(await allPages('/usuarios')).map(user); s.roles=await allPages('/roles'); s.permissionCatalog=await allPages('/permisos');
    s.rolePermissions={};
    await Promise.all(s.roles.map(async r=>{s.rolePermissions[r.name]=(await request('/roles/'+r.id+'/permisos')).data.permissions;}));
  } else {
    await Promise.all((needs[module]||[]).map(async key=>{
      const [route,permission]=specs[key];
      if(!can(permission)||(module==='mecanicos'&&key==='orders')){
        try{s[key]=await allPages('/frontend/opciones?module='+module+'&resource='+route);}catch(error){if(error.status!==403)throw error;unavailable.push(permission);}
        return;
      }
      s[key]=await allPages('/'+route);
    }));
    s.users=(s.users||[]).map(user);
    s.quotes=(s.quotes||[]).map(quote); s.orders=(s.orders||[]).map(order);
    if(module==='auxilio' && can('auxilio')) s.tasks=(await allPages('/me/tareas')).map(order);
    if(module==='pagos' && can('pagos')) {
      // Pagos tiene su propio permiso: no se confunde la falta de acceso con saldo cero.
      const id=params.get('orden');
      if(id){
        const payments=(await request('/ordenes/'+encodeURIComponent(id)+'/pagos')).data;
        s.payments=payments.payments.map(p=>({...p,amountCents:cents(p.amount)}));
        const o=s.orders?.find(o=>o.id===id);
        if(o?.delivered){const d=(await request('/ordenes/'+id+'/entrega')).data;o.delivery={...d,at:d.createdAt,by:d.createdBy};}
        if(o?.clientId)s.contacts=can('clientes')?(await request('/clientes/'+o.clientId+'/contactos')).data:await allPages('/frontend/opciones?module=pagos&resource=contactos&orderId='+id);
      }
    }
  }
  if(module==='proveedores')await Promise.all((s.suppliers||[]).map(async p=>{p.linkedParts=await allPages('/proveedores/'+p.id+'/repuestos');p.parts=p.linkedParts.map(r=>r.name).join(', ');}));
  s.unavailable=[...new Set(unavailable)];
  if(current===generation&&realSession()===sessionRef)replaceRealStore(s);
  return current===generation&&realSession()===sessionRef;
}
export async function saveClient(input,id) {return (await request('/clientes'+(id?'/'+id:''),{method:id?'PATCH':'POST',body:pick(input,['name','documentType','documentNumber','phone','email','address'])})).data;}
export async function saveVehicle(input,id) {
  const body=pick(input,id?['brand','model','year','mileage','notes']:['clientId','plate','brand','model','year','mileage','notes']);body.year=Number(body.year);
  if(id&&input.transfer){
    // Dos comandos públicos separados: no ocultar una actualización parcial.
    const current=realStore().vehicles.find(v=>v.id===id);
    if(Object.keys(body).some(k=>String(body[k])!==String(current?.[k]??'')))throw new Error('Guarda primero los datos del vehículo. Después confirma la transferencia por separado para evitar una actualización parcial.');
    return (await request('/vehiculos/'+id+'/transferencias',{method:'POST',body:pick(input,['newClientId','transferVerified'])})).data;
  }
  return (await request('/vehiculos'+(id?'/'+id:''),{method:id?'PATCH':'POST',body})).data;
}
export async function saveReception(input){return (await request('/recepciones',{method:'POST',body:pick(input,['clientId','vehicleId','mileage','symptoms'])})).data;}
export async function saveDiagnosis(input){return (await request('/diagnosticos',{method:'POST',body:pick(input,['receptionId','findings'])})).data;}
export async function saveQuote(input,id){
  const s=realStore(),items=input.items.map(i=>{
    const key=i.type==='Servicio'?'serviceId':'sparePartId',catalog=i.type==='Servicio'?s.services:s.spareParts;
    // La elección del catálogo se transporta con ID. No se adivinan nombres ni referencias.
    if(!i[key] || (!catalog.some(c=>c.id===i[key]) && !s.quotes.find(q=>q.id===id)?.items.some(c=>c[key]===i[key])))throw new Error('Selecciona cada concepto del catálogo de servicios o repuestos.');
    return {...pick(i,['concept','type','quantity','unitPrice']),[key]:i[key],origin:i.type==='Servicio'?'':i.origin};
  });
  return quote((await request('/cotizaciones'+(id?'/'+id:''),{method:id?'PUT':'POST',body:{diagnosisId:input.diagnosisId,status:input.status,items}})).data);
}
export async function createOrder(input){return order((await request('/ordenes',{method:'POST',body:pick(input,['quoteId','mechanicId','assistantId'])})).data);}
export async function updateOrder(id,input){
  const body=pick(input,['state','note','test','result']);
  if(realSession()?.permissions.includes('ordenes.asignar'))Object.assign(body,pick(input,['mechanicId','assistantId']));
  return order((await request('/ordenes/'+id,{method:'PATCH',body})).data);
}
export async function saveRescue(input){return (await request('/auxilios',{method:'POST',body:{...pick(input,['clientId','vehicleId','location','incident','description','technicianId']),unitId:input.unit}})).data;}
export async function registerPayment(input){const p=(await request('/ordenes/'+input.orderId+'/pagos',{method:'POST',body:pick(input,['amount','method'])})).data;return {...p,amountCents:cents(p.amount)};}
export async function saveContact(input){const c=(await request('/clientes/'+input.clientId+'/contactos',{method:'POST',body:pick(input,['name','documentType','documentNumber','phone','relation'])})).data;updateRealStore(s=>s.contacts.push(c));return c;}
export async function confirmDelivery(input){return (await request('/ordenes/'+input.orderId+'/entrega',{method:'POST',body:pick(input,['kind','method','verified',...(input.kind==='Tercero'?['contactId']:[])])})).data;}
export async function saveSupplier(input,id){if(input.parts)throw new Error('PostgreSQL no dispone de especialidad. El campo de repuestos se consulta mediante relaciones de catálogo; no se descartará información.');return (await request('/proveedores'+(id?'/'+id:''),{method:id?'PATCH':'POST',body:pick(input,['name','ruc','phone'])})).data;}
export async function saveUser(input,id){
  const s=realStore(),old=s.users.find(u=>u.id===id),role=s.roles.find(r=>r.name===(input.role||old?.roleNames?.[0]));if(!role)throw new Error('El rol seleccionado no existe en PostgreSQL.');
  const body=pick(input,['name','username','email','active',...(!id?['password']:[])]),existing=s.users.find(u=>u.id===id);
  // Editar un campo ordinario no vuelve a asignar roles ni revoca sesiones.
  if(!id||!existing?.roleNames?.includes(role.name))body.roleId=role.id;
  if(id && body.active===existing?.active)delete body.active;
  return user((await request('/usuarios'+(id?'/'+id:''),{method:id?'PATCH':'POST',body})).data);
}
const remove=path=>request(path,{method:'DELETE'});
export const deleteClient=id=>remove('/clientes/'+id),deleteVehicle=id=>remove('/vehiculos/'+id),deleteOrder=id=>remove('/ordenes/'+id),deleteSupplier=id=>remove('/proveedores/'+id),deleteUser=id=>remove('/usuarios/'+id);
function roleID(name){const r=realStore().roles.find(r=>r.name===name);if(!r)throw new Error('Rol no disponible.');return r.id;}
export const saveRolePermissions=(role,permissions)=>request('/roles/'+roleID(role)+'/permisos',{method:'PUT',body:{permissions}});
export const resetRolePermissions=role=>request('/roles/'+roleID(role)+'/permisos/restablecer',{method:'POST',body:{}});
export const vehicleHistory=id=>request('/vehiculos/'+id+'/historial');
export async function getDashboard(){const d=(await request('/dashboard')).data,income=d.income.map(m=>({...m,month:formatMonth(m.month,true),cents:safeInteger(m.cents)}));return {...d,monthlyIncomeCents:safeInteger(d.monthlyIncomeCents),income,period:income[0].month+'–'+income.at(-1).month+' '+d.currentMonth.slice(0,4),currentMonth:formatMonth(d.currentMonth)};}
export async function getReport(period){const r=(await request('/reportes?period='+encodeURIComponent(period))).data;return {...r,income:safeInteger(r.incomeCents),months:r.months.map(m=>({...m,label:formatMonth(m.label),cents:safeInteger(m.cents)}))};}

function safeInteger(value){const n=Number(value);if(!/^\d+$/.test(String(value))||!Number.isSafeInteger(n))throw new Error("Total fuera del rango de visualización exacta.");return n;}

function formatMonth(value,short=false){if(!/^\d{4}-(0[1-9]|1[0-2])$/.test(value))throw new Error("Mes inválido recibido del servidor.");return new Intl.DateTimeFormat("es-PE",{month:short?"short":"long",...(short?{}:{year:"numeric"}),timeZone:"America/Lima"}).format(new Date(value+"-15T12:00:00-05:00")).replace(".","");}
