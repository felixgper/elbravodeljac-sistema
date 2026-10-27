import { createSeed } from '../data/seed.js';
export const STORE_KEY='bravo-jac:workshop-demo:v2';
const tables=['clients','vehicles','users','receptions','diagnoses','quotes','orders','payments','contacts','rescues','suppliers'];
export function getStore(){
  let raw;try{raw=localStorage.getItem(STORE_KEY);}catch{throw new Error('Habilita el almacenamiento local del navegador para usar la demostración.');}
  if(!raw){const seed=createSeed();persist(seed);return structuredClone(seed);}
  try{const value=JSON.parse(raw);if(value.version!==2||tables.some(k=>!Array.isArray(value[k])))throw Error();return value;}
  catch{throw new Error('No se pudieron leer los datos de demostración. Exporta una copia o reinícialos desde el perfil.');}
}
function persist(store){try{localStorage.setItem(STORE_KEY,JSON.stringify(store));}catch{throw new Error('No se pudo guardar. El navegador no permite almacenar más datos; no se aplicaron los cambios.');}}
export function transaction(action){const state=getStore();const result=action(state);persist(state);return structuredClone(result??null);}
export function resetDemo(){const state=createSeed();persist(state);return state;}
export function nextCode(list,prefix,state){let max=list.reduce((n,r)=>Math.max(n,Number(String(r.id).split('-').at(-1))||0),0);if(state){state.sequences??={};max=Math.max(max,state.sequences[prefix]||0);state.sequences[prefix]=max+1;}return `${prefix}-${String(max+1).padStart(3,'0')}`;}
export const newId=prefix=>{const bytes=new Uint32Array(4);crypto.getRandomValues(bytes);return `${prefix}-${[...bytes].map(n=>n.toString(16).padStart(8,'0')).join('')}`;};
export const nowISO=()=>new Date().toISOString();
export function requireRecord(list,id,label){const value=list.find(r=>r.id===id);if(!value)throw new Error(`Selecciona ${label} válido.`);return value;}
export function decimalMinor(value,label='Importe'){
  const text=String(value??'').trim();if(!/^\d{1,8}(\.\d{1,2})?$/.test(text))throw new Error(`${label}: ingresa un número positivo con hasta 8 enteros y 2 decimales, usando punto.`);
  const [whole,part='']=text.split('.');return Number(whole)*100+Number(part.padEnd(2,'0'));
}
export function decimalString(value,label='Kilometraje'){const n=decimalMinor(value,label);return `${Math.floor(n/100)}.${String(n%100).padStart(2,'0')}`;}
export function text(value,label,max=250){const v=String(value??'').trim();if(!v)throw new Error(`Completa ${label}.`);if(v.length>max)throw new Error(`${label}: máximo ${max} caracteres.`);return v;}
export function documentNumber(type,value){const n=text(value,'el número de documento',20).toUpperCase();if(type==='DNI'&&!/^\d{8}$/.test(n))throw new Error('El DNI debe tener 8 dígitos.');if(type==='RUC'&&!/^\d{11}$/.test(n))throw new Error('El RUC debe tener 11 dígitos.');if(!['DNI','RUC','CE','Pasaporte'].includes(type)||! /^[A-Z0-9]{5,20}$/.test(n))throw new Error('Revisa el tipo y número de documento.');return n;}
export function phone(value){const p=text(value,'el teléfono',25).replace(/[\s()-]/g,'');if(!/^\+?\d{7,15}$/.test(p))throw new Error('Ingresa un teléfono válido de 7 a 15 dígitos.');return p;}
export function email(value,required=false){const e=String(value??'').trim().toLowerCase();if((required||e)&&! /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(e))throw new Error('Ingresa un correo electrónico válido.');if(e.length>254)throw new Error('El correo es demasiado largo.');return e;}
export function audit(by){return {at:nowISO(),by:by||'u-admin'};}
