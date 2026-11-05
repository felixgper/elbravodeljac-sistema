import {request,realMode,realSession} from '../services/api-service.js';
import {e,table,link,date,stamp,badge} from './ui.js';
export function pendingDiagnosisTable(rows){return table(['Recepción','Cliente','Vehículo','Motivo','Ingreso','Estado','Acción'],rows.map(({reception:r,diagnosis:d})=>[e(r.id),e(r.clientName),e([r.vehiclePlate,r.vehicleBrand,r.vehicleModel].filter(Boolean).join(' ')),e(r.symptoms),date(r.createdAt),badge(d?.status||r.evaluationState),r.technicianId?link('Abrir evaluación','diagnosticos?recepcion='+r.id):link('Asignar mecánico','recepciones?guardada='+r.id)]),'No hay evaluaciones pendientes autorizadas.');}
export async function refreshDiagnosisInbox(root,ready=false){
 if(!realMode||!realSession()?.permissions.includes('diagnosticos.ver'))return;
 try{const result=await request((ready?'/diagnosticos/pendientes-cotizacion':'/me/diagnosticos')+'?pageSize=10');if(!root.isConnected)return;root.innerHTML=`<h2 class="form-section-title">${ready?'Diagnósticos pendientes de cotización':'Mis diagnósticos pendientes'} · ${result.meta.total}</h2>`+(ready?table(['Recepción','Cliente','Vehículo','Mecánico','Finalización','Hallazgos','Acción'],result.data.map(({reception:r,diagnosis:d})=>[e(r.id),e(r.clientName),e([r.vehiclePlate,r.vehicleBrand,r.vehicleModel].filter(Boolean).join(' ')),e(d.technicianName),stamp(d.finishedAt),e(d.findings),link('Crear cotización','cotizaciones?diagnostico='+d.id)]),'No hay diagnósticos finalizados pendientes de cotización.'):pendingDiagnosisTable(result.data))+(result.meta.total>10?link('Ver todas las evaluaciones','diagnosticos'):'');}
 catch(error){if(root.isConnected)root.innerHTML=`<p class="form-error" role="alert">${e(error.message)}</p>`;}
}
export async function mountDiagnosisDashboard(root){
 if(!realMode||!realSession()?.permissions.includes('diagnosticos.ver'))return;
 const permissions=realSession().permissions,ready=permissions.includes('cotizaciones.crear');root.insertAdjacentHTML('beforeend',`<section class="panel separated" id="diagnosis-inbox-pending"></section>${ready?'<section class="panel separated" id="diagnosis-inbox-ready"></section>':''}`);
 await Promise.all([refreshDiagnosisInbox(root.querySelector('#diagnosis-inbox-pending')),ready?refreshDiagnosisInbox(root.querySelector('#diagnosis-inbox-ready'),true):Promise.resolve()]);
}
let lastRefresh=0;
export function refreshVisibleDiagnosisInboxes(){if(!realMode||document.hidden||Date.now()-lastRefresh<3000)return;lastRefresh=Date.now();for(const [id,ready]of [['diagnosis-inbox-pending',false],['diagnosis-inbox-ready',true]]){const root=document.getElementById(id);if(root)refreshDiagnosisInbox(root,ready);}}
