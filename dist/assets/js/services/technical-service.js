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
