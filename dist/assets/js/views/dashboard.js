import {getDashboard,getActiveOrders,getReport} from '../services/dashboard-service.js';
import {permissionChecker} from '../services/permissions.js';
import {realMode} from '../services/api-service.js';
import {donutChart,horizontalBars} from '../components/charts.js';
import {mountDiagnosisDashboard} from '../components/diagnosis-inbox.js';
import {renderReports} from './administration.js';
import {e,money,field,select,form,submit,bindForm,hint} from '../components/ui.js';
import {icon} from '../components/icons.js';
export async function renderIntegratedDashboard({root,session,period='Mes',reference,tab,isCurrent=()=>true}) {
 const can=permissionChecker(session.role),summaryAllowed=can('inicio','ver'),financeAllowed=can('reportes','ver');
 reference ||= new Intl.DateTimeFormat('en-CA',{year:'numeric',month:'2-digit',day:'2-digit',timeZone:'America/Lima'}).format(new Date());
 if(!['Semana','Mes','Año'].includes(period))period='Mes';
 let selected=(tab==='analisis'||!summaryAllowed)&&financeAllowed?'analisis':'resumen',loadedSummary='',loadedAnalysis='';
 const cache=new Map(),key=()=>period+'|'+reference;
 root.innerHTML=`<nav class="tabs dashboard-tabs" role="tablist" aria-label="Contenido de Inicio"><button class="tab" id="tab-resumen" role="tab" aria-controls="dashboard-summary" ${summaryAllowed?'':'disabled'}>Resumen</button><button class="tab" id="tab-analisis" role="tab" aria-controls="dashboard-analysis" ${financeAllowed?'':'disabled title="Requiere permiso de análisis financiero"'}>Análisis ${financeAllowed?'':icon('lock')}</button></nav>`+form('PERÍODO DE CONSULTA',select('Periodo','period',['Semana','Mes','Año'],period)+field('Fecha de referencia','reference',reference,'date','required'),submit('Aplicar filtros'),'dashboard-filter')+`<section id="dashboard-summary" role="tabpanel" aria-labelledby="tab-resumen"></section><section id="dashboard-analysis" role="tabpanel" aria-labelledby="tab-analisis"></section>`;
 const summary=root.querySelector('#dashboard-summary'),analysis=root.querySelector('#dashboard-analysis');
 const updateLocation=()=>{const q=new URLSearchParams({tab:selected,period,reference});history.replaceState(null,'','#/inicio?'+q);};
 const fresh=()=>isCurrent()&&root.isConnected;
 const cached=(name,fetcher)=>{const id=name+'|'+key();if(!cache.has(id))cache.set(id,fetcher().catch(error=>{cache.delete(id);throw error;}));return cache.get(id);};
 async function show(){
  summary.hidden=selected!=='resumen';analysis.hidden=selected!=='analisis';
  for(const name of ['resumen','analisis']){const button=root.querySelector('#tab-'+name);button.classList.toggle('active',selected===name);button.setAttribute('aria-selected',String(selected===name));button.tabIndex=selected===name?0:-1;}
  const ownKey=key(),target=selected==='resumen'?summary:analysis;
  if(selected==='resumen'&&loadedSummary===ownKey||selected==='analisis'&&loadedAnalysis===ownKey)return;
  target.innerHTML='<div class="loading-state" role="status">Consultando información del taller…</div>';target.setAttribute('aria-busy','true');
  try{
   if(selected==='analisis'){
    if(!financeAllowed)throw Error('Tu cuenta no tiene permiso de análisis financiero.');
    const data=await cached('analysis',()=>getReport(period,reference));if(!fresh()||key()!==ownKey)return;
    await renderReports({root:analysis,period,reference,embedded:true,reportData:data});loadedAnalysis=ownKey;
   }else{
    if(!summaryAllowed)throw Error('Tu cuenta no tiene permiso de resumen operativo.');
    const data=await cached('summary',()=>getDashboard(period,reference));if(!fresh()||key()!==ownKey)return;
    const cards=[...(financeAllowed&&data.financialVisible!==false?[{label:'Ingresos del período',value:money(data.monthlyIncomeCents),icon:'wallet',class:'copper',note:data.period}]:[]),{label:'Órdenes activas',value:getActiveOrders(data),icon:'wrench',class:'copper',note:'Pendientes, reparación y pruebas'},{label:'Servicios finalizados',value:data.completedServices,icon:'check',class:'green',note:'Cantidades confirmadas'},{label:'Auxilios del período',value:data.rescuesToday,icon:'rescue',class:'blue',note:realMode?'Solicitudes registradas':'Demostración separada'}];
    summary.innerHTML=`<div class="dashboard-note"><span class="live-dot"></span>Resumen general · ${e(data.period)}</div><section class="kpi-grid" aria-label="Resumen de actividad">${cards.map(c=>`<article class="kpi-card"><span class="kpi-icon ${c.class}">${icon(c.icon)}</span><div><strong class="kpi-value">${e(c.value)}</strong><h2>${e(c.label)}</h2><p>${e(c.note)}</p></div></article>`).join('')}</section><div class="main-charts"><section class="panel"><div class="panel-heading"><h2>Actividad del taller</h2>${icon('wrench')}</div>${[['Recepciones registradas',data.counts?.receptions??0],['Diagnósticos registrados',data.counts?.diagnoses??0],['Entregas registradas',data.counts?.deliveries??0]].map(([label,value])=>`<div class="report-total"><span>${e(label)}</span><strong>${e(value)}</strong></div>`).join('')}${financeAllowed?'<button type="button" class="secondary-button" id="open-analysis">Ver análisis del período</button>':hint('Los importes y comparaciones requieren permiso de análisis financiero.')}</section><section class="panel"><div class="panel-heading"><h2>Estados de órdenes</h2></div>${donutChart(data.states)}</section></div><section class="panel separated"><div class="panel-heading"><h2>Trabajo por mecánico</h2></div>${horizontalBars(data.mechanics,'mechanics')}</section>`;
    summary.querySelector('#open-analysis')?.addEventListener('click',()=>selectTab('analisis'));
    await mountDiagnosisDashboard(summary);loadedSummary=ownKey;
   }
  }catch(error){if(fresh()&&key()===ownKey)target.innerHTML=`<section class="panel error-state" role="alert"><h2>No se pudo completar la consulta</h2><p>${e(error.message)}</p><p>Vuelve a aplicar los filtros para intentarlo.</p></section>`;}
  finally{target.removeAttribute('aria-busy');}
 }
 async function selectTab(name){if(name==='analisis'&&!financeAllowed||name==='resumen'&&!summaryAllowed)return;selected=name;updateLocation();await show();}
 root.querySelector('#tab-resumen').onclick=()=>selectTab('resumen');root.querySelector('#tab-analisis').onclick=()=>selectTab('analisis');
 root.querySelector('.dashboard-tabs').onkeydown=event=>{if(!['ArrowLeft','ArrowRight','Home','End'].includes(event.key))return;const tabs=[...root.querySelectorAll('[role=tab]:not(:disabled)')],index=tabs.indexOf(document.activeElement),next=event.key==='Home'?0:event.key==='End'?tabs.length-1:(index+(event.key==='ArrowRight'?1:-1)+tabs.length)%tabs.length;event.preventDefault();tabs[next].click();tabs[next].focus();};
 bindForm('dashboard-filter',async data=>{period=data.period;reference=data.reference;loadedSummary='';loadedAnalysis='';cache.clear();updateLocation();await show();});
 await show();
}
