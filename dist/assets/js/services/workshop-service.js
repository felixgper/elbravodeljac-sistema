import {realMode} from './api-service.js';
import * as real from './real-workshop.js';
import {getStore,transaction,newId,nextCode,nowISO,requireRecord,decimalMinor,decimalString,text,documentNumber,phone,email,audit} from './store-service.js';
export {getStore,decimalMinor,decimalString} from './store-service.js';
export const ORDER_STATES=['Pendiente','En reparación','En pruebas','Finalizada','Cancelada'];
export const ACTIVE_STATES=ORDER_STATES.slice(0,3);
import {ROLES,LOCKED_ROLE,REQUIRED_FOR_ADMIN,defaultPermissions,sanitizePermissions} from '../data/permissions.js';
export {ROLES};
export function getContext(state,order){
  const quote=state.quotes.find(q=>q.id===order?.quoteId),diagnosis=state.diagnoses.find(d=>d.id===quote?.diagnosisId),reception=state.receptions.find(r=>r.id===(quote?.receptionId||diagnosis?.receptionId));
  const client=state.clients.find(c=>c.id===(reception?.clientId||order?.clientId))||(realMode?{id:reception?.clientId||order?.clientId,name:reception?.clientName||order?.clientName}:undefined),vehicle=state.vehicles.find(v=>v.id===(reception?.vehicleId||order?.vehicleId))||(realMode?{id:reception?.vehicleId||order?.vehicleId,plate:reception?.vehiclePlate||order?.vehiclePlate,brand:reception?.vehicleBrand||order?.vehicleBrand,model:reception?.vehicleModel||order?.vehicleModel}:undefined);return {quote,diagnosis,reception,client,vehicle};
}
export function itemSubtotal(item){return Number((BigInt(decimalMinor(item.quantity,'Cantidad'))*BigInt(item.unitCents-(item.discountCents||0))+50n)/100n);}
export function quoteTotal(quote){if(realMode&&quote?.total!==undefined)return real.cents(quote.total);return (quote?.items??[]).reduce((n,item)=>n+itemSubtotal(item),0);}
export function paidTotal(state,orderId){return state.payments.filter(p=>p.orderId===orderId&&!p.reversedAt).reduce((n,p)=>n+p.amountCents,0);}
export function balance(state,order){return (realMode&&order.amountDue!==undefined?real.cents(order.amountDue):quoteTotal(getContext(state,order).quote))-paidTotal(state,order.id);}
export function saveClient(input,id){if(realMode)return real.saveClient(input,id);return transaction(s=>{
  const type=input.documentType,number=documentNumber(type,input.documentNumber);
  if(s.clients.some(c=>c.id!==id&&c.documentType===type&&c.documentNumber===number))throw new Error('Ya existe un cliente con ese tipo y número de documento.');
  const data={documentType:type,documentNumber:number,name:text(input.name,'el nombre o razón social',180),phone:phone(input.phone),email:email(input.email),address:text(input.address,'la dirección',250)};
  if(id){const current=requireRecord(s.clients,id,'un cliente');Object.assign(current,data);return current;}
  const record={id:newId('cli'),...data,createdAt:nowISO()};s.clients.push(record);return record;
});}
export function saveVehicle(input,id,by){if(realMode)return real.saveVehicle(input,id,by);return transaction(s=>{
  const current=id?requireRecord(s.vehicles,id,'un vehículo'):null;
  const plate=current?.plate??text(input.plate,'la placa',12).toUpperCase().replace(/\s/g,'');
  if(!/^[A-Z0-9-]{5,12}$/.test(plate))throw new Error('Revisa la placa: utiliza letras, números y guion.');
  if(s.vehicles.some(v=>v.id!==id&&v.plate.replace(/-/g,'')===plate.replace(/-/g,'')))throw new Error('La placa ya está registrada.');
  const year=Number(input.year);if(!Number.isInteger(year)||year<1950||year>new Date().getFullYear()+1)throw new Error('Ingresa un año válido para el vehículo.');
  const mileage=decimalString(input.mileage);if(current&&decimalMinor(mileage)<decimalMinor(current.mileage))throw new Error('El kilometraje no puede disminuir respecto al registrado.');
  const data={plate,brand:text(input.brand,'la marca',40),model:text(input.model,'el modelo',60),year,mileage,notes:String(input.notes??'').trim().slice(0,1000)};
  let owner=current?.clientId??input.clientId;requireRecord(s.clients,owner,'un cliente propietario');
  if(current&&input.transfer){if(!input.transferVerified)throw new Error('Confirma la verificación de la transferencia.');const next=requireRecord(s.clients,input.newClientId,'un nuevo propietario');if(next.id===owner)throw new Error('El nuevo propietario debe ser diferente del actual.');current.transfers.push({previousClientId:owner,newClientId:next.id,...audit(by)});owner=next.id;}
  if(current){Object.assign(current,data,{clientId:owner});return current;}
  const record={id:newId('veh'),...data,clientId:owner,transfers:[]};s.vehicles.push(record);return record;
});}
export function saveReception(input,by){if(realMode)return real.saveReception(input,by);return transaction(s=>{
  const client=requireRecord(s.clients,input.clientId,'un cliente'),vehicle=requireRecord(s.vehicles,input.vehicleId,'un vehículo');
  if(vehicle.clientId!==client.id)throw new Error('El vehículo debe pertenecer al cliente seleccionado.');
  const mileage=decimalString(input.mileage);if(decimalMinor(mileage)<decimalMinor(vehicle.mileage))throw new Error('El kilometraje de ingreso no puede ser menor al actual.');
  const record={id:nextCode(s.receptions,'RX',s),clientId:client.id,vehicleId:vehicle.id,mileage,attentionType:input.attentionType||'Diagnostico',technicianId:input.technicianId||'',symptoms:text(input.symptoms,'el motivo de ingreso / síntomas',4000),createdAt:nowISO(),createdBy:by};vehicle.mileage=mileage;s.receptions.push(record);return record;
});}
export function saveDiagnosis(input,by){if(realMode)return real.saveDiagnosis(input,by);return transaction(s=>{
 const reception=requireRecord(s.receptions,input.receptionId,'una recepción guardada');if(reception.attentionType==='Directo')throw Error('Servicio directo sin diagnóstico.');
 const old=s.diagnoses.find(d=>d.receptionId===reception.id);if(old&&!input.id)throw Error('La recepción ya tiene un diagnóstico guardado.');if(old&&(old.status||'Finalizado')==='Finalizado')throw Error('El diagnóstico finalizado no se sobrescribe.');
 const data={receptionId:reception.id,findings:input.status==='Borrador'?String(input.findings||''):text(input.findings,'los hallazgos',4000),evaluation:input.evaluation||'',recommendations:input.recommendations||'',notes:input.notes||'',status:input.status||'Finalizado',technicianId:reception.technicianId||by,version:(old?.version||0)+1,finishedAt:input.status==='Borrador'?null:nowISO(),createdBy:by};
 if(old){Object.assign(old,data);return old;}const record={id:nextCode(s.diagnoses,'DX',s),createdAt:nowISO(),...data};s.diagnoses.push(record);return record;
 });}

export function saveQuote(input,id){if(realMode)return real.saveQuote(input,id);return transaction(s=>{
  const current=id?requireRecord(s.quotes,id,'una cotización'):null;
  if(current&&s.orders.some(o=>o.quoteId===id))throw new Error('La cotización ya originó una orden y conserva sus trabajos autorizados.');
  const diagnosis=input.diagnosisId?requireRecord(s.diagnoses,input.diagnosisId,'un diagnóstico guardado'):null;const reception=requireRecord(s.receptions,input.receptionId||diagnosis?.receptionId,'una recepción');if(!diagnosis&&reception.attentionType!=='Directo')throw Error('La atención requiere diagnóstico.');
  if(s.quotes.some(q=>q.receptionId===reception.id&&q.id!==id||diagnosis&&q.diagnosisId===diagnosis.id&&q.id!==id))throw new Error('Este diagnóstico ya tiene una cotización. Ábrela desde las cotizaciones guardadas.');
  if(!['Pendiente','Aprobada','Rechazada'].includes(input.status))throw new Error('Selecciona el estado de autorización.');
  if(!Array.isArray(input.items)||!input.items.length)throw new Error('Agrega al menos un servicio o repuesto.');
  const items=input.items.map(line=>{
    if(!['Servicio','Repuesto'].includes(line.type))throw new Error('Selecciona el tipo de concepto.');
    const quantity=decimalString(line.quantity,'Cantidad');if(decimalMinor(quantity)===0)throw new Error('La cantidad debe ser mayor que cero.');
    let origin=line.type==='Servicio'?'—':line.origin;if(line.type==='Repuesto'&&!['Taller','Cliente'].includes(origin))throw new Error('Selecciona el origen del repuesto.');
    const unitCents=line.type==='Repuesto'&&origin==='Cliente'?0:decimalMinor(line.unitPrice,'Precio unitario');
    const discountCents=decimalMinor(line.discount||'0','Descuento');if(discountCents>unitCents)throw new Error('El descuento supera el precio de venta.');return {concept:text(line.concept,'el concepto',4000),type:line.type,partKind:line.partKind||'',discountCents,origin,quantity,unitCents};
  });
  if(quoteTotal({items})>999999999999)throw new Error('El total de esta cotización excede el límite de la demostración.');
  const data={diagnosisId:diagnosis?.id||'',receptionId:reception.id,items,status:input.status,updatedAt:nowISO()};
  if(current){Object.assign(current,data);return current;}
  const record={id:nextCode(s.quotes,'COT',s),...data,createdAt:nowISO()};s.quotes.push(record);return record;
});}
function technicalUser(s,id){const u=requireRecord(s.users,id,'un mecánico');if(!u.active||u.role!=='Mecánico')throw new Error('Selecciona un mecánico activo.');return u;}
export function createOrder(input,by){if(realMode)return real.createOrder(input,by);return transaction(s=>{
  const quote=requireRecord(s.quotes,input.quoteId,'una cotización aprobada');
  if(quote.status!=='Aprobada')throw new Error('Solo se crean órdenes desde cotizaciones aprobadas.');
  if(s.orders.some(o=>o.quoteId===quote.id))throw new Error('Esta cotización ya tiene una orden de trabajo.');
  technicalUser(s,input.mechanicId);if(input.assistantId)throw new Error('No se permiten nuevas asignaciones de Auxiliar.');
  const record={id:nextCode(s.orders,'OT',s),quoteId:quote.id,mechanicId:input.mechanicId,assistantId:input.assistantId||'',state:'Pendiente',createdAt:nowISO(),createdBy:by,updates:[],delivery:null};s.orders.push(record);return record;
});}
export function updateOrder(id,input,by){if(realMode)return real.updateOrder(id,input,by);return transaction(s=>{
  const order=requireRecord(s.orders,id,'una orden');if(order.delivery)throw new Error('El vehículo ya fue entregado; la orden conserva su historial cerrado.');
  if(order.state==='Cancelada')throw new Error('Una orden cancelada conserva su historial y no se reactiva.');
  if(!ORDER_STATES.includes(input.state))throw new Error('Selecciona un estado válido.');
  if(order.state==='Finalizada'&&input.state!=='Finalizada')throw new Error('Una orden finalizada no puede retroceder de estado.');
  if(input.state==='Cancelada'&&paidTotal(s,id)>0)throw new Error('La orden tiene pagos registrados y no puede cancelarse desde esta demostración.');
  technicalUser(s,input.mechanicId);if(input.assistantId)throw new Error('No se permiten nuevas asignaciones de Auxiliar.');
  const test=String(input.test??'').trim(),result=input.result||'No aplica';if(!['No aplica','Satisfactoria','Fallida'].includes(result))throw new Error('Selecciona un resultado de prueba válido.');
  if(result!=='No aplica'&&!test)throw new Error('Indica qué prueba se realizó.');
  const lastRepair=order.updates.findLastIndex(u=>u.state==='En reparación');const lastTestIndex=order.updates.findLastIndex(u=>u.result!=='No aplica');
  const lastTest=result!=='No aplica'?{test,result}:lastTestIndex>lastRepair?order.updates[lastTestIndex]:null;
  if(input.state==='Finalizada'&&(!lastTest?.test||lastTest.result!=='Satisfactoria'))throw new Error('Para finalizar registra una prueba satisfactoria. Si falló, corrige la reparación y vuelve a probar.');
  const note=text(input.note,'la observación / avance de esta actualización',4000);
  Object.assign(order,{mechanicId:input.mechanicId,state:input.state});order.updates.push({...audit(by),note,test,result,state:input.state});return order;
});}
export function deleteOrder(id){if(realMode)return real.deleteOrder(id);return transaction(s=>{const order=requireRecord(s.orders,id,'una orden');if(order.delivery||s.payments.some(p=>p.orderId===id))throw new Error('Esta orden tiene pagos o una entrega registrada. Se conserva por trazabilidad.');s.orders=s.orders.filter(o=>o.id!==id);return order;});}
export function saveRescue(input,by){if(realMode)return real.saveRescue(input,by);return transaction(s=>{
  const client=requireRecord(s.clients,input.clientId,'un cliente'),vehicle=requireRecord(s.vehicles,input.vehicleId,'un vehículo');if(vehicle.clientId!==client.id)throw new Error('Selecciona un vehículo del cliente.');technicalUser(s,input.technicianId);
  if(!['Falla mecánica','Vehículo inmovilizado','Otro'].includes(input.incident))throw new Error('Selecciona el tipo de incidente.');
  const record={id:nextCode(s.rescues,'AUX',s),clientId:client.id,vehicleId:vehicle.id,location:text(input.location,'la ubicación',300),incident:input.incident,unit:text(input.unit,'la unidad de auxilio',50),description:text(input.description,'la descripción del problema',4000),technicianId:input.technicianId,createdAt:nowISO(),createdBy:by};s.rescues.push(record);return record;
});}
export function registerPayment(input,by){if(realMode)return real.registerPayment(input,by);return transaction(s=>{
  const order=requireRecord(s.orders,input.orderId,'una orden');if(order.state!=='Finalizada')throw new Error('Finaliza la orden con pruebas satisfactorias antes de registrar el pago.');
  const amountCents=decimalMinor(input.amount,'Monto pagado');if(amountCents<=0)throw new Error('El monto pagado debe ser mayor que cero.');if(amountCents>balance(s,order))throw new Error('El pago no puede superar el saldo pendiente.');
  if(!['Transferencia','Efectivo','Tarjeta'].includes(input.method))throw new Error('Selecciona el medio de pago.');
  const record={id:nextCode(s.payments,'PAG',s),orderId:order.id,amountCents,method:input.method,createdAt:nowISO(),createdBy:by};s.payments.push(record);return record;
});}
export function saveContact(input){if(realMode)return real.saveContact(input);return transaction(s=>{
  requireRecord(s.clients,input.clientId,'un cliente');if(!['DNI','CE','Pasaporte'].includes(input.documentType))throw new Error('Selecciona el documento personal del contacto.');
  const number=documentNumber(input.documentType,input.documentNumber);if(s.contacts.some(c=>c.clientId===input.clientId&&c.documentType===input.documentType&&c.documentNumber===number))throw new Error('Este contacto ya está vinculado al cliente.');
  if(!['Conductor','Representante','Familiar','Otro'].includes(input.relation))throw new Error('Selecciona la relación con el cliente.');
  const telephone=phone(input.phone);if(!telephone.startsWith('+'))throw new Error('Incluye el prefijo internacional del teléfono, por ejemplo +51.');
  const record={id:newId('ct'),clientId:input.clientId,name:text(input.name,'el nombre completo',180),documentType:input.documentType,documentNumber:number,phone:telephone,relation:input.relation};s.contacts.push(record);return record;
});}
export function confirmDelivery(input,by){if(realMode)return real.confirmDelivery(input,by);const recipientKind=input.kind;input={...input,kind:input.kind==='Propietario o titular'?'Titular / representante':['Persona autorizada','Representante de empresa'].includes(input.kind)?'Tercero':input.kind};return transaction(s=>{
  const order=requireRecord(s.orders,input.orderId,'una orden');if(order.delivery)throw new Error('La entrega ya está registrada.');if(order.state!=='Finalizada')throw new Error('La orden debe estar finalizada para entregar el vehículo.');if(balance(s,order)!==0)throw new Error('Completa el pago pendiente antes de entregar el vehículo.');if(!input.verified)throw new Error('Verifica la identidad y autorización específica del receptor.');
  if(!['Titular / representante','Tercero'].includes(input.kind))throw new Error('Selecciona quién retira el vehículo.');
  if(!['Confirmación del titular','Documento','Representante facultado'].includes(input.method))throw new Error('Selecciona cómo se verificó la autorización.');
  let contactId='';if(input.kind==='Tercero'){const contact=requireRecord(s.contacts,input.contactId,'un contacto');if(contact.clientId!==getContext(s,order).client.id)throw new Error('El contacto debe estar vinculado al cliente de la orden.');contactId=contact.id;}
  order.delivery={...audit(by),kind:input.kind,recipientKind,contactId,method:input.method,verified:true,conformity:{accepted:Boolean(input.conformity),notes:input.conformityNotes||''}};return order;
});}
export function saveSupplier(input,id){if(realMode)return real.saveSupplier(input,id);return transaction(s=>{
  const ruc=documentNumber('RUC',input.ruc);if(s.suppliers.some(p=>p.id!==id&&p.ruc===ruc))throw new Error('Ya existe un proveedor con ese RUC.');
  const data={name:text(input.name,'el nombre del proveedor',180),ruc,phone:phone(input.phone),parts:text(input.parts,'los repuestos o especialidad',200)};
  if(id){const current=requireRecord(s.suppliers,id,'un proveedor');Object.assign(current,data);return current;}const record={id:newId('pr'),...data};s.suppliers.push(record);return record;
});}
export async function passwordDigest(password){
  const bytes=new TextEncoder().encode(password);
  if(crypto.subtle){const result=await crypto.subtle.digest('SHA-256',bytes);return [...new Uint8Array(result)].map(n=>n.toString(16).padStart(2,'0')).join('');}
  // Compatibilidad con vistas locales HTTP. Solo para claves ficticias de demostración.
  const primes=[];for(let n=2;primes.length<64;n++){if(!primes.some(p=>n%p===0))primes.push(n);}
  const fraction=n=>Math.floor((n-Math.floor(n))*4294967296)>>>0;
  const k=primes.map(p=>fraction(Math.cbrt(p))),hash=primes.slice(0,8).map(p=>fraction(Math.sqrt(p)));
  const data=new Uint8Array(Math.ceil((bytes.length+9)/64)*64);data.set(bytes);data[bytes.length]=128;
  const view=new DataView(data.buffer);view.setUint32(data.length-4,bytes.length*8);const rotate=(v,n)=>(v>>>n)|(v<<(32-n));
  for(let offset=0;offset<data.length;offset+=64){const w=new Uint32Array(64);for(let i=0;i<16;i++)w[i]=view.getUint32(offset+i*4);for(let i=16;i<64;i++){const x=w[i-15],y=w[i-2];w[i]=(w[i-16]+(rotate(x,7)^rotate(x,18)^(x>>>3))+w[i-7]+(rotate(y,17)^rotate(y,19)^(y>>>10)))>>>0;}
    let [a,b,c,d,f,g,h,j]=hash;for(let i=0;i<64;i++){const t1=(j+(rotate(f,6)^rotate(f,11)^rotate(f,25))+((f&g)^(~f&h))+k[i]+w[i])>>>0;const t2=((rotate(a,2)^rotate(a,13)^rotate(a,22))+((a&b)^(a&c)^(b&c)))>>>0;j=h;h=g;g=f;f=(d+t1)>>>0;d=c;c=b;b=a;a=(t1+t2)>>>0;}
    [a,b,c,d,f,g,h,j].forEach((v,i)=>hash[i]=(hash[i]+v)>>>0);
  }
  return hash.map(n=>n.toString(16).padStart(8,'0')).join('');
}
export async function saveUser(input,id){if(realMode)return real.saveUser(input,id);
  const digest=id?null:await passwordDigest(text(input.password,'la contraseña inicial',128));
  if(!id&&String(input.password).length<8)throw new Error('La contraseña inicial debe tener al menos 8 caracteres.');
  return transaction(s=>{
    const username=text(input.username,'el usuario',60).toLowerCase(),mail=email(input.email,true);if(!/^[a-z0-9_.-]{3,60}$/.test(username))throw new Error('El usuario debe tener 3 a 60 letras, números, punto, guion o guion bajo.');
    if(s.users.some(u=>u.id!==id&&(u.username===username||u.email===mail)))throw new Error('El usuario o el correo ya están registrados.');if(!ROLES.includes(input.role))throw new Error('Selecciona un rol válido.');
    const current=id?requireRecord(s.users,id,'un usuario'):null;
    const isAdmin=role=>role==='Gerente';if(current?.active&&isAdmin(current.role)&&(!input.active||!isAdmin(input.role))&&!s.users.some(u=>u.id!==id&&u.active&&isAdmin(u.role)))throw new Error('Debe conservarse al menos un Gerente activo.');
    const data={name:text(input.name,'el nombre del trabajador',180),username,email:mail,role:input.role,active:Boolean(input.active)};
    if(current){Object.assign(current,data);return current;}
    const record={id:newId('usr'),...data,passwordDigest:digest,createdAt:nowISO()};s.users.push(record);return record;
  });
}

// ── Eliminaciones con reglas de integridad (se conserva la trazabilidad) ─────
export function deleteClient(id){if(realMode)return real.deleteClient(id);return transaction(s=>{
  const client=requireRecord(s.clients,id,'un cliente');
  const vehicles=s.vehicles.filter(v=>v.clientId===id).length;
  if(vehicles)throw new Error(`${client.name} tiene ${vehicles} vehículo${vehicles===1?'':'s'} registrado${vehicles===1?'':'s'}. Elimina o transfiere sus vehículos antes de eliminar al cliente.`);
  if(s.receptions.some(r=>r.clientId===id)||s.rescues.some(r=>r.clientId===id))throw new Error('El cliente tiene recepciones o auxilios registrados. Se conserva por trazabilidad.');
  s.clients=s.clients.filter(c=>c.id!==id);s.contacts=s.contacts.filter(c=>c.clientId!==id);return client;
});}
export function deleteVehicle(id){if(realMode)return real.deleteVehicle(id);return transaction(s=>{
  const vehicle=requireRecord(s.vehicles,id,'un vehículo');
  if(s.receptions.some(r=>r.vehicleId===id)||s.rescues.some(r=>r.vehicleId===id))throw new Error('El vehículo tiene historial de atenciones o auxilios. Se conserva por trazabilidad.');
  s.vehicles=s.vehicles.filter(v=>v.id!==id);return vehicle;
});}
export function deleteSupplier(id){if(realMode)return real.deleteSupplier(id);return transaction(s=>{
  const supplier=requireRecord(s.suppliers,id,'un proveedor');s.suppliers=s.suppliers.filter(p=>p.id!==id);return supplier;
});}
export function deleteUser(id,by){if(realMode)return real.deleteUser(id,by);return transaction(s=>{
  const user=requireRecord(s.users,id,'un usuario');
  if(id===by)throw new Error('No puedes eliminar tu propio usuario mientras tienes la sesión iniciada.');
  const isAdmin=role=>role==='Gerente';
  if(user.active&&isAdmin(user.role)&&!s.users.some(u=>u.id!==id&&u.active&&isAdmin(u.role)))throw new Error('Debe conservarse al menos un Gerente activo.');
  const used=s.orders.some(o=>o.mechanicId===id||o.assistantId===id||o.createdBy===id)||s.receptions.some(r=>r.createdBy===id)||s.diagnoses.some(d=>d.createdBy===id)||s.payments.some(p=>p.createdBy===id)||s.rescues.some(r=>r.technicianId===id||r.createdBy===id);
  if(used)throw new Error('Este usuario tiene actividad registrada en el taller. Para conservar el historial, edítalo y márcalo como inactivo.');
  s.users=s.users.filter(u=>u.id!==id);return user;
});}

// ── Roles y permisos ─────────────────────────────────────────────────────────
export function saveRolePermissions(role,input){if(realMode)return real.saveRolePermissions(role,input);return transaction(s=>{
  if(!ROLES.includes(role))throw new Error('Selecciona un rol válido.');
  if(role===LOCKED_ROLE)throw new Error('El rol Gerente siempre conserva el acceso completo.');
  const clean=sanitizePermissions(input);

  if(!Object.keys(clean).length)throw new Error('Deja al menos un módulo visible para este rol.');
  const current=s.rolePermissions||{};s.rolePermissions={...Object.fromEntries(ROLES.map(r=>[r,current[r]||defaultPermissions()[r]])),[role]:clean};return clean;
});}
export function resetRolePermissions(role){if(realMode)return real.resetRolePermissions(role);return transaction(s=>{
  if(!ROLES.includes(role))throw new Error('Selecciona un rol válido.');
  if(role===LOCKED_ROLE)throw new Error('El rol Gerente siempre conserva el acceso completo.');
  const defaults=defaultPermissions(),current=s.rolePermissions||{};s.rolePermissions={...Object.fromEntries(ROLES.map(r=>[r,current[r]||defaults[r]])),[role]:defaults[role]};return defaults[role];
});}
