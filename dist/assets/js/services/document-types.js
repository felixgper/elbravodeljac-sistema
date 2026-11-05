import {realMode} from './api-service.js';
import {realStore} from './real-state.js';
const demo=[{documentType:'DNI',label:'DNI',active:true},{documentType:'RUC',label:'RUC',active:true},{documentType:'CE',label:'Carné de Extranjería',active:true},{documentType:'Pasaporte',label:'Pasaporte',active:true}];
export function documentOptions(current='',contact=false){
 const list=realMode?(realStore().documentTypes||[]):demo;
 return list.filter(d=>(d.active||d.documentType===current)&&(!contact||d.code!=='RUC'&&d.documentType!=='RUC')).map(d=>({value:d.documentType,label:d.active?d.label:d.documentType+' (histórico, conservado)'}));
}
