import {request,allPages} from './api-service.js';
// Conservar requestId junto al borrador hasta recibir confirmación. Los reintentos
// de escrituras técnicas reutilizan esa clave; nunca se generan reintentos ciegos.
export const newTechnicalRequestID=()=>crypto.randomUUID();
const base=id=>'/diagnosticos/'+encodeURIComponent(id);
export const listScans=id=>allPages(base(id)+'/escaneres');
export const listEvidence=id=>allPages(base(id)+'/evidencias');
export const saveScan=(id,input)=>request(base(id)+'/escaneres',{method:'POST',body:input});
export async function uploadEvidence(id,{file,description='',requestId}) {
  if(!requestId)throw new Error('Conserva una clave de solicitud para poder reintentar la carga.');
  if(!file || file.size===0 || file.size>10*1024*1024)throw new Error('El archivo debe tener entre 1 byte y 10 MiB.');
  if(!['image/jpeg','image/png','application/pdf'].includes(file.type))throw new Error('Formato permitido: JPEG, PNG o PDF. El servidor también valida el contenido.');
  const body=new FormData();body.append('archivo',file);body.append('requestId',requestId);body.append('descripcion',description);
  return request(base(id)+'/evidencias',{method:'POST',body});
}
export const evidenceURL=(diagnosis,id)=>'/api/v1'+base(diagnosis)+'/evidencias/'+encodeURIComponent(id)+'/archivo';

// Persist only an opaque request ID and payload digest, never file bytes or text.
// Closing/reopening the modal or reselecting the same file after refresh keeps
// the same Mongo idempotency key. A changed payload requires explicit reset.
const attempts=new Map();
export function technicalAttempt(kind,diagnosis,actor){
 const key='bravo-jac:technical:'+actor+':'+diagnosis+':'+kind;
 if(attempts.has(key))return attempts.get(key);
 let pending=null,busy=false;
 try{pending=JSON.parse(globalThis.sessionStorage?.getItem(key)||'null');}catch{}
 if(pending&&(!/^[a-f0-9]{64}$/.test(pending.digest)||typeof pending.requestId!=='string'))pending=null;
 const reset=()=>{pending=null;try{globalThis.sessionStorage?.removeItem(key);}catch{}};
 const result={get pending(){return pending!==null;},reset,confirm(rows){if(pending&&rows.some(r=>r.requestId===pending.requestId)){reset();return true;}return false;},async send(payload,operation){
  if(busy)throw Error('Espera a que termine el envío actual.');busy=true;
  try{
   let digestInput;
   if(kind==='evidence'){
    const f=payload.file;if(!f||!f.size||f.size>10*1024*1024)throw Error('Selecciona un archivo de hasta 10 MiB.');
    if(!['image/jpeg','image/png','application/pdf'].includes(f.type))throw Error('Selecciona una fotografía JPG o PNG, o un PDF.');
    const sum=await crypto.subtle.digest('SHA-256',await f.arrayBuffer());
    digestInput=JSON.stringify({file:[...new Uint8Array(sum)],description:payload.description||'',type:f.type});
   }else digestInput=JSON.stringify(payload);
   const bytes=await crypto.subtle.digest('SHA-256',new TextEncoder().encode(digestInput));const digest=[...new Uint8Array(bytes)].map(b=>b.toString(16).padStart(2,'0')).join('');
   if(pending&&pending.digest!==digest)throw Error('Hay un envío pendiente de confirmar. Reintenta con los mismos datos, actualiza los registros o confirma preparar otra solicitud.');
   if(!pending){pending={requestId:newTechnicalRequestID(),digest};try{globalThis.sessionStorage?.setItem(key,JSON.stringify(pending));}catch{reset();throw Error('No se pudo conservar la solicitud. Permite el almacenamiento de sesión antes de subir información.');}}
   const response=await operation({...payload,requestId:pending.requestId});reset();return response;
  }catch(error){if([400,413,422].includes(error.status))reset();throw error;}
  finally{busy=false;}
 }};attempts.set(key,result);return result;
}
