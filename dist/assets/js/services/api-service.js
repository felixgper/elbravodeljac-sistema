import {clearRealStore} from './real-state.js';
// El servidor elige el modo. No hay conmutación ni respaldo automático a la demo.
export const realMode = globalThis.__BRAVO_MODE === 'real' || globalThis.document?.querySelector('meta[name="bravo-mode"]')?.content === 'real';
let session = null, csrfToken = '';
export const realSession = () => session;
export function setSession(response) {
  session = {...response.data, role:(response.data.roles || []).join(' / '), demo:false};
  csrfToken = response.csrfToken;
  return session;
}
export function clearSession() { session=null; csrfToken=''; }
export class APIError extends Error {
  constructor(message,status=0,code='',fields={}) { super(message); Object.assign(this,{status,code,fields}); }
}
export async function request(path,{method='GET',body,signal}={}) {
  if (!path.startsWith('/') || path.startsWith('//')) throw new Error('Ruta API inválida.');
  const headers = {Accept:'application/json'};
  const multipart = body instanceof FormData;
  if(body !== undefined && !multipart) headers['Content-Type']='application/json';
  if(!['GET','HEAD'].includes(method) && csrfToken) headers['X-CSRF-Token']=csrfToken;
  let response;
  try { response = await fetch('/api/v1'+path,{method,headers,credentials:'same-origin',cache:'no-store',body:body===undefined?undefined:multipart?body:JSON.stringify(body),signal}); }
  catch(error) { if(error.name==='AbortError')throw error; throw new APIError('No se pudo contactar al servidor. Comprueba la conexión antes de volver a guardar.'); }
  let result;
  if(response.status!==204) {
    try { result=await response.json(); } catch { throw new APIError('El servidor devolvió una respuesta inválida.',response.status); }
  }
  if(!response.ok) {
    if(response.status===401) { clearSession(); clearRealStore(); globalThis.dispatchEvent?.(new Event('bravo:unauthorized')); }
    throw new APIError(result?.error?.message || 'No se pudo completar la operación.',response.status,result?.error?.code,result?.error?.fields);
  }
  return result;
}
export async function allPages(path,options={}) {
  const rows=[]; let page=1;
  for(;;) {
    const result=await request(path+(path.includes('?')?'&':'?')+'pageSize=100&page='+page,options);
    if(!Array.isArray(result.data))throw new APIError('Listado API inválido.');
    rows.push(...result.data);
    if(!result.meta || rows.length>=result.meta.total) return rows;
    if(!result.data.length)throw new APIError('La paginación cambió; vuelve a cargar la pantalla.');
    page++;
  }
}
