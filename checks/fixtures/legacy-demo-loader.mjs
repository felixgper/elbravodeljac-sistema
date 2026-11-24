// Solo Node: conserva las regresiones históricas en memoria sin habilitar demo
// en los archivos publicados. Cargar explícitamente con --import; no usar en UI.
import {register} from 'node:module';
import {isMainThread} from 'node:worker_threads';
import assert from 'node:assert/strict';
if(isMainThread)register(import.meta.url);
const target=new URL('../../dist/assets/js/services/api-service.js',import.meta.url).href;
export async function load(url,context,nextLoad){
 const result=await nextLoad(url,context);
 if(url!==target)return result;
 const source=Buffer.from(result.source).toString('utf8');
 assert(source.includes('export const realMode = true;'),'Revisar la guarda del fixture antes de ejecutar regresiones');
 return {...result,source:source.replace('export const realMode = true;','export const realMode = false;')};
}
