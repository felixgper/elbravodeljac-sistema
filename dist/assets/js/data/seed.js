// Datos completamente ficticios para recorrer todos los procesos del prototipo.
export function createSeed(now = new Date()) {
  const year = now.getFullYear(), month = now.getMonth();
  const dateAt = (offset, day = 10) => new Date(year, month + offset, day, 12).toISOString();
  const clients = [
    ['c1','DNI','12345678','Juan Pérez','987654321'], ['c2','DNI','87654321','María Torres','912345678'],
    ['c3','RUC','20123456789','Transportes JAC SAC','987000111'], ['c4','CE','001294555','María López','988111222'],
    ['c5','Pasaporte','AB567889','Martin Reyes','988222333'], ['c6','DNI','43659874','Pedro Ramírez','987124356'],
  ].map(([id,documentType,documentNumber,name,phone])=>({id,documentType,documentNumber,name,phone,email:'',address:'Los Olivos, Lima',createdAt:dateAt(-5)}));
  const models = ['T8','S3','T6','JS4','S2','T9'];
  const plates = ['ABC-123','DEF-456','GHI-789','JKL-234','MNO-567','PQR-890','STU-345','VWX-678','YZA-901','BCD-234','EFG-567','HIJ-890'];
  const vehicles = plates.map((plate,i)=>({id:`v${i+1}`,clientId:clients[i%clients.length].id,plate,brand:'JAC',model:models[i%6],year:2020+i%5,mileage:(25300+i*1790).toFixed(2),notes:'Vehículo de demostración',transfers:[]}));
  const users = [
    ['u-admin','Admin','admin','admin@demo.local','Administrador'],['u-jimmy','Jimmy','jimmy','jimmy@demo.local','Propietario'],
    ['u-carlos','Carlos','carlos','carlos@demo.local','Supervisor'],['u-pedro','Pedro','pedro','pedro@demo.local','Mecánico'],
    ['u-luis','Luis','luis','luis@demo.local','Mecánico'],['u-maria','María','maria','maria@demo.local','Mecánico'],
    ['u-kevin','Kevin','kevin','kevin@demo.local','Auxiliar'],['u-office','Administración','administracion','oficina@demo.local','Administrativo'],
  ].map(([id,name,username,email,role])=>({id,name,username,email,role,active:true,passwordDigest:null,createdAt:dateAt(-5)}));
  const receptions=[],diagnoses=[],quotes=[],orders=[],payments=[];
  const states=['Pendiente','Pendiente',...Array(5).fill('En reparación'),...Array(3).fill('En pruebas'),...Array(24).fill('Finalizada'),'Cancelada'];
  const monthlyAmounts=[620000,780000,850000,1020000,1130000,1245000];
  states.forEach((state,i)=>{
    const vehicle=vehicles[i%vehicles.length],code=String(25+i).padStart(3,'0');
    const at = i>=10&&i<16 ? dateAt(i-15) : dateAt(0,Math.min(5+i%15,now.getDate()));
    const rid=`RX-${code}`,did=`DX-${code}`,qid=`COT-${code}`,oid=`OT-${code}`;
    const amount=i>=10&&i<16?monthlyAmounts[i-10]:15500+(i%5)*2500;
    const mechanicId=['u-pedro','u-luis','u-maria'][i%3];
    receptions.push({id:rid,clientId:vehicle.clientId,vehicleId:vehicle.id,mileage:vehicle.mileage,symptoms:i%2?'Revisión del sistema de frenos':'Ruido en el motor al acelerar',createdAt:at,createdBy:'u-office'});
    diagnoses.push({id:did,receptionId:rid,findings:i%2?'Desgaste de pastillas. Se requiere reparación y prueba de ruta.':'Revisión y mantenimiento del motor. Verificar funcionamiento tras intervenir.',createdAt:at,createdBy:mechanicId});
    quotes.push({id:qid,diagnosisId:did,items:[{concept:i>=10&&i<16?'Mantenimiento y reparación integral':['Cambio de aceite','Servicio de frenos','Diagnóstico técnico','Suspensión'][i%4],type:'Servicio',origin:'—',quantity:'1',unitCents:amount}],status:'Aprobada',createdAt:at,updatedAt:at});
    const delivered=i>=10&&i<16;
    orders.push({id:oid,quoteId:qid,mechanicId,assistantId:i%3===0?'u-kevin':'',state,createdAt:at,createdBy:'u-admin',updates:state==='Finalizada'?[{at,by:mechanicId,note:'Servicio concluido y verificado.',test:'Prueba de ruta',result:'Satisfactoria',state}]:[],delivery:delivered?{at,by:'u-office',kind:'Titular / representante',contactId:'',method:'Confirmación del titular',verified:true}:null});
    if(delivered)payments.push({id:`PAG-${code}`,orderId:oid,amountCents:amount,method:'Transferencia',createdAt:at,createdBy:'u-office'});
  });
  for(let i=0;i<3;i++){
    const code=String(60+i).padStart(3,'0'),v=vehicles[10+i%2];
    receptions.push({id:`RX-${code}`,clientId:v.clientId,vehicleId:v.id,mileage:v.mileage,symptoms:['Cambio de aceite y filtro','Revisión de frenos','Revisión general'][i],createdAt:dateAt(0,now.getDate()),createdBy:'u-office'});
    if(i<2)diagnoses.push({id:`DX-${code}`,receptionId:`RX-${code}`,findings:'Mantenimiento requerido. Vehículo evaluado y listo para cotizar.',createdAt:dateAt(0,now.getDate()),createdBy:'u-pedro'});
    if(i===0)quotes.push({id:`COT-${code}`,diagnosisId:`DX-${code}`,items:[{concept:'Cambio de aceite',type:'Servicio',origin:'—',quantity:'1',unitCents:12000},{concept:'Filtro de aceite',type:'Repuesto',origin:'Taller',quantity:'1',unitCents:3500}],status:'Aprobada',createdAt:dateAt(0,now.getDate()),updatedAt:dateAt(0,now.getDate())});
  }
  return {version:2,sequences:{RX:62,DX:61,COT:60,OT:59,PAG:40,AUX:3},clients,vehicles,users,receptions,diagnoses,quotes,orders,payments,contacts:[{id:'ct1',clientId:'c3',name:'Carlos Mendoza',documentType:'DNI',documentNumber:'44556677',phone:'+51987654320',relation:'Conductor'}],
    rescues:Array.from({length:3},(_,i)=>({id:`AUX-00${i+1}`,clientId:clients[i].id,vehicleId:vehicles[i].id,location:['Comas, av. Túpac Amaru','Los Olivos, av. Universitaria','Independencia, Panamericana Norte'][i],incident:'Falla mecánica',unit:'Unidad 01',description:'Vehículo inmovilizado; atención de demostración.',technicianId:'u-pedro',createdAt:dateAt(0,now.getDate()),createdBy:'u-office'})),
    suppliers:[{id:'pr1',name:'Repuestos Lima',ruc:'20112233445',phone:'+51987654321',parts:'Filtros'},{id:'pr2',name:'Autopartes Norte',ruc:'20445566778',phone:'+51986543210',parts:'Frenos'}]};
}
