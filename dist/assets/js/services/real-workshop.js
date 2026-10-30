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
export const quote=v=>({...v,items:(v.items||[]).map(i=>({...i,origin:i.type==='Servicio'?'—':i.origin==='CLIENTE'?'Cliente':'Taller',unitCents:cents(i.unitPrice),discountCents:cents(i.discount||'0.00')}))});
const stateName=v=>({creada:'Pendiente',en_reparacion:'En reparación',en_pruebas:'En pruebas',finalizada:'Finalizada',cancelada:'Cancelada'})[v]||v;
export function order(v) {
  const assignments=(v.assignments||[]).filter(a=>a.active);
  // Las tres bitácoras se conservan, sin inventar asociaciones entre eventos.
  const updates=[...(v.advances||[]).map(a=>({...a,test:'',result:'No aplica',state:''})),
    ...(v.tests||[]).map(a=>({...a,result:a.result==='satisfactoria'?'Satisfactoria':'Fallida',state:''})),
    ...(v.states||[]).map(a=>({...a,note:'Cambio de estado',test:'',result:'No aplica',state:stateName(a.state)}))].sort((a,b)=>a.at.localeCompare(b.at));
  return {...v,mechanicId:assignments.find(a=>a.role==='MECANICO')?.userId||'',assistantId:assignments.find(a=>a.role==='AUXILIAR')?.userId||'',updates,delivery:v.delivered?{pendingDetails:true}:null};
}
const specs={clients:['clientes','clientes'],vehicles:['vehiculos','vehiculos'],receptions:['recepciones','recepciones'],diagnoses:['diagnosticos','diagnosticos'],quotes:['cotizaciones','cotizaciones'],orders:['ordenes','ordenes'],rescues:['auxilios','auxilio'],suppliers:['proveedores','proveedores'],units:['unidades-auxilio','auxilio'],users:['mecanicos','mecanicos']};
const needs={clientes:['clients','vehicles'],vehiculos:['clients','vehicles'],recepciones:['clients','vehicles','receptions','diagnoses','users'],diagnosticos:['clients','vehicles','receptions','diagnoses','users'],cotizaciones:['quotes'],ordenes:['clients','vehicles','receptions','diagnoses','quotes','orders','users'],pagos:['clients','vehicles','receptions','diagnoses','quotes','orders','users'],auxilio:['clients','vehicles','rescues','units','users','orders','quotes','diagnoses','receptions'],mecanicos:['users','orders','clients','vehicles','receptions','diagnoses','quotes'],proveedores:['suppliers']};
let generation=0;
export async function loadModule(module,params=new URLSearchParams(),path=[]) {
  const current=++generation,sessionRef=realSession(),s={},permissions=sessionRef?.permissions||[],can=m=>permissions.includes(m+'.ver');
  const unavailable=[];
  if(module==='clientes'||module==='pagos')s.documentTypes=(await request('/tipos-documento')).data;
  if(module==='usuarios') {
    s.employees=await allPages('/empleados'); s.users=(await allPages('/usuarios')).map(user); s.roles=await allPages('/roles'); s.permissionCatalog=await allPages('/permisos');
    s.rolePermissions={};
    await Promise.all(s.roles.map(async r=>{s.rolePermissions[r.name]=(await request('/roles/'+r.id+'/permisos')).data.permissions;}));
  } else {
    await Promise.all((needs[module]||[]).map(async key=>{
      if(['clients','vehicles'].includes(key))return;
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
  if(module!=='usuarios'){
    s.clients=[];s.vehicles=[];
    const add=(key,value)=>{if(value&&!s[key].some(v=>v.id===value.id))s[key].push(value);};
    if(module==='clientes'&&path[1]==='editar'&&path[2])add('clients',(await request('/clientes/'+path[2])).data);
    if(module==='vehiculos'&&path[1]==='editar'){
      if(path[2])add('vehicles',(await request('/vehiculos/'+path[2])).data);
      else for(const v of (await request('/vehiculos?pageSize=1')).data)add('vehicles',v);
    }
    if(module==='cotizaciones'){
      s.receptions=[];s.diagnoses=[];s.orders=[];
      const current=s.quotes.find(q=>q.id===(params.get('id')||params.get('ampliacion')));let receptionId=current?.receptionId||params.get('recepcion');
      if(!receptionId&&params.get('diagnostico')){const diagnosis=(await request('/diagnosticos/'+params.get('diagnostico'))).data;receptionId=diagnosis.receptionId;}
      if(receptionId){const found=(await lookupQuoteReceptions('',1,undefined,receptionId)).data[0];if(!found)throw Error('Atención no autorizada o diagnóstico todavía no finalizado.');s.receptions=[found];s.diagnoses=found.diagnosis?[found.diagnosis]:[];}
    }
    const related=[...(s.receptions||[]).slice(0,10),...(s.rescues||[]).slice(0,10),...(s.orders||[]).slice(0,10)];
    for(const r of s.receptions||[])if(r.id===params.get('recepcion')||r.id===params.get('guardada')||r.id===(s.quotes||[]).find(q=>q.id===params.get('id'))?.receptionId)related.push(r);
    const diagnosis=(s.diagnoses||[]).find(d=>d.id===(params.get('diagnostico')||(s.quotes||[]).find(q=>q.id===params.get('id'))?.diagnosisId));
    if(diagnosis){const r=(s.receptions||[]).find(r=>r.id===diagnosis.receptionId);if(r)related.push(r);}
    const selectedOrder=(s.orders||[]).find(o=>o.id===(params.get('orden')||path[2]));if(selectedOrder)related.push(selectedOrder);
    const clients=new Set([...related.map(r=>r.clientId),...s.vehicles.map(v=>v.clientId),params.get('cliente')].filter(Boolean));
    const vehicles=new Set([...related.map(r=>r.vehicleId),params.get('vehiculo')].filter(Boolean));
    await Promise.all([...clients].filter(id=>!s.clients.some(c=>c.id===id)).map(async id=>{
      if(can('clientes'))add('clients',(await request('/clientes/'+id)).data);
      else {const r=await request('/frontend/opciones?module='+module+'&resource=clientes&clientId='+id);for(const c of r.data)add('clients',c);}
    }));
    await Promise.all([...vehicles].filter(id=>!s.vehicles.some(v=>v.id===id)).map(async id=>{
      if(can('vehiculos'))add('vehicles',(await request('/vehiculos/'+id)).data);
    }));
    if(module==='vehiculos')s.brands=(await request('/vehiculos/marcas')).data;
  }
  if(module==='diagnosticos'){s.diagnosisTasks=await allPages('/me/diagnosticos');if(params.get('recepcion')&&!s.receptions?.some(r=>r.id===params.get('recepcion')))throw Error('No se encontró una recepción autorizada para este usuario.');}
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
export async function saveReception(input){return (await request('/recepciones',{method:'POST',body:pick(input,['clientId','vehicleId','mileage','symptoms','attentionType','technicianId'])})).data;}
export async function saveDiagnosis(input){return (await request('/diagnosticos'+(input.id?'/'+input.id:''),{method:input.id?'PUT':'POST',body:pick(input,['receptionId','findings','technicianId','evaluation','recommendations','notes','status','version'])})).data;}
export async function saveQuote(input,id){
  const items=input.items.map(i=>({...pick(i,['concept','type','quantity','unitPrice','partKind','discount','serviceId','sparePartId']),origin:i.type==='Servicio'?'':i.origin}));
  return quote((await request('/cotizaciones'+(id?'/'+id:''),{method:id?'PUT':'POST',body:{diagnosisId:input.diagnosisId||'',receptionId:input.receptionId||'',status:input.status,items,version:input.version||0,parentQuoteId:input.parentQuoteId||'',...(input.status==='Aprobada'?{approval:{name:input.approvalName,method:input.approvalMethod}}:{})}})).data);
}
export async function createOrder(input){return order((await request('/ordenes',{method:'POST',body:pick(input,['quoteId','mechanicId'])})).data);}
export async function updateOrder(id,input){
  const body=pick(input,['state','note','test','result']);
  if(realSession()?.permissions.includes('ordenes.asignar'))Object.assign(body,pick(input,['mechanicId']));
  return order((await request('/ordenes/'+id,{method:'PATCH',body})).data);
}
export async function saveRescue(input){return (await request('/auxilios',{method:'POST',body:{...pick(input,['clientId','vehicleId','location','incident','description','technicianId']),unitId:input.unit}})).data;}
export async function registerPayment(input){
 const key='bravo-payment-pending:'+realSession()?.id+':'+input.orderId,signature=JSON.stringify([cents(input.amount),input.method,input.comprobante||'']);let pending;
 try{pending=JSON.parse(sessionStorage.getItem(key)||'null');}catch{}
 if(pending?.signature!==signature)pending={signature,requestId:input.requestId||crypto.randomUUID()};
 try{sessionStorage.setItem(key,JSON.stringify(pending));}catch{}
 const p=(await request('/ordenes/'+input.orderId+'/pagos',{method:'POST',body:{...pick(input,['amount','method','comprobante']),requestId:pending.requestId}})).data;
 try{sessionStorage.removeItem(key);}catch{}
 return {...p,amountCents:cents(p.amount)};
}
export async function saveContact(input){const c=(await request('/clientes/'+input.clientId+'/contactos',{method:'POST',body:pick(input,['name','documentType','documentNumber','phone','relation'])})).data;updateRealStore(s=>s.contacts.push(c));return c;}
export async function confirmDelivery(input){return (await request('/ordenes/'+input.orderId+'/entrega',{method:'POST',body:{...pick(input,['method','verified','conformity','conformityNotes',...(input.kind!=='Propietario o titular'?['contactId']:[])]),recipientKind:input.kind,kind:input.kind==='Propietario o titular'?'Titular / representante':'Tercero'}})).data;}
export async function saveSupplier(input,id){if(input.parts)throw new Error('PostgreSQL no dispone de especialidad. El campo de repuestos se consulta mediante relaciones de catálogo; no se descartará información.');return (await request('/proveedores'+(id?'/'+id:''),{method:id?'PATCH':'POST',body:pick(input,['name','ruc','phone'])})).data;}
export async function saveUser(input,id){
  const s=realStore(),old=s.users.find(u=>u.id===id),role=s.roles.find(r=>r.active!==false&&r.name===input.role);if(!role&&!id)throw new Error('El rol seleccionado no existe en PostgreSQL.');
  const body=pick(input,['name','username','email','active',...(!id?['password']:[])]),existing=s.users.find(u=>u.id===id);
  // Editar un campo ordinario no vuelve a asignar roles ni revoca sesiones.
  if(role&&(!id||!existing?.roleNames?.includes(role.name)))body.roleId=role.id;
  if(id && body.active===existing?.active)delete body.active;
  if(input.employeeId && input.employeeId!==existing?.employeeId)body.employeeId=input.employeeId;
  if(input.removeRoleId)body.removeRoleId=input.removeRoleId;
  return user((await request('/usuarios'+(id?'/'+id:''),{method:id?'PATCH':'POST',body})).data);
}
const remove=path=>request(path,{method:'DELETE'});
export const deleteClient=id=>remove('/clientes/'+id),deleteVehicle=id=>remove('/vehiculos/'+id),deleteOrder=id=>remove('/ordenes/'+id),deleteSupplier=id=>remove('/proveedores/'+id),deleteUser=id=>remove('/usuarios/'+id);
function roleID(name){const r=realStore().roles.find(r=>r.name===name);if(!r)throw new Error('Rol no disponible.');return r.id;}
export const saveRolePermissions=(role,permissions)=>request('/roles/'+roleID(role)+'/permisos',{method:'PUT',body:{permissions}});
export const resetRolePermissions=role=>request('/roles/'+roleID(role)+'/permisos/restablecer',{method:'POST',body:{}});
export const vehicleHistory=id=>request('/vehiculos/'+id+'/historial');
export async function getDashboard(period='Mes'){
 const d=(await request('/dashboard?period='+encodeURIComponent(period))).data;
 const formatDay=value=>new Intl.DateTimeFormat('es-PE',{day:'2-digit',timeZone:'America/Lima'}).format(new Date(value+'T12:00:00-05:00'));
 const income=d.income.map(m=>({...m,month:m.month.length===10?formatDay(m.month):formatMonth(m.month,true),cents:safeInteger(m.cents)}));
 const fmt=value=>new Intl.DateTimeFormat('es-PE',{dateStyle:'short',timeZone:'America/Lima'}).format(new Date(value));
 return {...d,monthlyIncomeCents:safeInteger(d.monthlyIncomeCents),income,period:period+' · '+fmt(d.from)+'–'+fmt(Date.parse(d.until)-1),currentMonth:formatMonth(d.currentMonth)};
}
export async function getReport(period,range){const args=new URLSearchParams({period,...(range?{desde:range.start.toISOString(),hasta:range.end.toISOString()}:{})});const r=(await request('/reportes?'+args)).data;return {...r,income:safeInteger(r.incomeCents),months:r.months.map(m=>({...m,label:m.label.length===10?new Intl.DateTimeFormat('es-PE',{day:'2-digit',month:'2-digit',timeZone:'America/Lima'}).format(new Date(m.label+'T12:00:00-05:00')):formatMonth(m.label),cents:safeInteger(m.cents)}))};}

function safeInteger(value){const n=Number(value);if(!/^[-]?\d+$/.test(String(value))||!Number.isSafeInteger(n))throw new Error("Total fuera del rango de visualización exacta.");return n;}

function formatMonth(value,short=false){if(!/^\d{4}-(0[1-9]|1[0-2])$/.test(value))throw new Error("Mes inválido recibido del servidor.");return new Intl.DateTimeFormat("es-PE",{month:short?"short":"long",...(short?{}:{year:"numeric"}),timeZone:"America/Lima"}).format(new Date(value+"-15T12:00:00-05:00")).replace(".","");}

export async function lookupClients(query='',page=1,signal,module='recepciones'){
 const args='q='+encodeURIComponent(query)+'&page='+page+'&pageSize=20';
 return request(realSession()?.permissions.includes('clientes.ver')?'/clientes?'+args:'/frontend/opciones?module='+module+'&resource=clientes&'+args,{signal});
}
export async function clientVehicles(clientId,module='recepciones'){
 if(!clientId)return [];
 const path=realSession()?.permissions.includes('vehiculos.ver')?'/vehiculos?clientId='+clientId:'/frontend/opciones?module='+module+'&resource=vehiculos&clientId='+clientId;
 return allPages(path);
}
export async function clientByID(id,module){
 if(realSession()?.permissions.includes('clientes.ver'))return (await request('/clientes/'+id)).data;
 return (await request('/frontend/opciones?module='+module+'&resource=clientes&clientId='+id)).data[0];
}
export async function vehiclePage(query,page=1,signal){return request('/vehiculos?q='+encodeURIComponent(query)+'&page='+page+'&pageSize=10',{signal});}
export async function saveUnit(input){return (await request('/unidades-auxilio',{method:'POST',body:pick(input,['code','description'])})).data;}

export const assignDiagnosis=(id,technicianId,version)=>request('/recepciones/'+id+'/mecanico',{method:'PATCH',body:{technicianId,version}});
export const pendingDiagnoses=()=>allPages('/me/diagnosticos');
export const diagnosesToQuote=()=>allPages('/diagnosticos/pendientes-cotizacion');

export const lookupQuoteReceptions=(q='',page=1,signal,id='')=>request('/cotizaciones/atenciones?q='+encodeURIComponent(q)+'&page='+page+'&pageSize=20'+(id?'&recepcionId='+encodeURIComponent(id):''),{signal});

export const reversePayment=(orderId,paymentId,reason,requestId,replacement)=>request('/ordenes/'+orderId+'/pagos/'+paymentId,{method:'PATCH',body:{reason,requestId,...(replacement?{replacement}:{})}});

export const lookupOrders=(q='',page=1,signal)=>request((realSession()?.permissions.includes('ordenes.ver')?'/ordenes?':'/frontend/opciones?module=pagos&resource=ordenes&')+'q='+encodeURIComponent(q)+'&page='+page+'&pageSize=20',{signal});
export const lookupApprovedQuotes=(q='',page=1,signal)=>request((realSession()?.permissions.includes('cotizaciones.ver')?'/cotizaciones?estado=Aprobada&sinOrden=true&':'/frontend/opciones?module=ordenes&resource=cotizaciones&')+'q='+encodeURIComponent(q)+'&page='+page+'&pageSize=20',{signal});
