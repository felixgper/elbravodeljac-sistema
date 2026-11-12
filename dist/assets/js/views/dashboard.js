import {getDashboard,getActiveOrders,getReport,reportRange} from '../services/dashboard-service.js';
import {permissionChecker} from '../services/permissions.js';
import {realMode,request} from '../services/api-service.js';
import {getContext,getStore} from '../services/workshop-service.js';
import {dashboardSummary} from './dashboard-summary.js';
import {mountDiagnosisDashboard} from '../components/diagnosis-inbox.js';
import {renderReports} from './administration.js';
import {e,money,field,select,form,submit,bindForm} from '../components/ui.js';
import {icon} from '../components/icons.js';
async function recentOrders(){
 if(realMode){
  const {data}=await request('/ordenes?page=1&pageSize=4');
  return data.map(order=>({...order,state:order.delivered?'Entregado':order.state}));
 }
 const store=getStore();
 return store.orders.slice().sort((a,b)=>b.createdAt.localeCompare(a.createdAt)).slice(0,4).map(order=>{
  const {client,vehicle}=getContext(store,order);
  return {...order,clientName:client?.name,vehiclePlate:vehicle?.plate,state:order.delivery?'Entregado':order.state};
 });
}

export async function renderIntegratedDashboard({root,session,period='Mes',reference,tab,isCurrent=()=>true}) {
 const can=permissionChecker(session.role),summaryAllowed=can('inicio','ver'),financeAllowed=can('reportes','ver');
 reference ||= new Intl.DateTimeFormat('en-CA',{year:'numeric',month:'2-digit',day:'2-digit',timeZone:'America/Lima'}).format(new Date());
 if(!['Semana','Mes','Año'].includes(period))period='Mes';
 let selected=(tab==='analisis'||!summaryAllowed)&&financeAllowed?'analisis':'resumen',loadedSummary='',loadedAnalysis='';
 let snapshot=null;
 const cache=new Map(),key=()=>period+'|'+reference;
 root.innerHTML=`<div class="dashboard-toolbar"><nav class="tabs dashboard-tabs" role="tablist" aria-label="Contenido de Inicio"><button class="tab" id="tab-resumen" role="tab" aria-controls="dashboard-summary" ${summaryAllowed?'':'disabled'}>Resumen</button><button class="tab" id="tab-analisis" role="tab" aria-controls="dashboard-analysis" ${financeAllowed?'':'disabled title="Requiere permiso de análisis financiero"'}>Análisis ${financeAllowed?'':icon('lock')}</button></nav><details class="dashboard-export"><summary class="secondary-button">Exportar ${icon('down')}</summary><div class="dashboard-export-options"><button type="button" data-export="xlsx" disabled>Exportar a Excel (.xlsx)</button><button type="button" data-export="pdf" disabled>Exportar a PDF (.pdf)</button></div></details></div><p class="dashboard-export-status" role="status" hidden></p>`+form('PERÍODO DE CONSULTA',select('Período','period',['Semana','Mes','Año'],period)+field('Fecha de referencia','reference',reference,'date','required'),submit('Aplicar filtros'),'dashboard-filter')+`<section id="dashboard-summary" role="tabpanel" aria-labelledby="tab-resumen"></section><section id="dashboard-analysis" role="tabpanel" aria-labelledby="tab-analisis"></section>`;
 const heading=root.closest('.workspace')?.querySelector('.page-heading');
 if(heading){
  heading.querySelector('.form-panel:has(>#dashboard-filter)')?.remove();
  heading.append(root.querySelector('#dashboard-filter').closest('.form-panel'));
 }
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
    const [data,recent]=await Promise.all([
     cached('summary',()=>getDashboard(period,reference)),
     can('ordenes','ver')?cached('orders',recentOrders).then(orders=>({orders}),()=>({error:true})):null
    ]);
    if(!fresh()||key()!==ownKey)return;
    const showIncome=financeAllowed&&data.financialVisible!==false;
    const cards=[
     ...(showIncome?[{label:'Ingresos del período',value:money(data.monthlyIncomeCents),amount:data.monthlyIncomeCents/100,icon:'wallet',class:'copper',note:'Pagos netos registrados'}]:[]),
     {label:'Órdenes activas',value:getActiveOrders(data),icon:'wrench',class:'blue',note:'Pendientes, en reparación o pruebas'},
     {label:'Servicios finalizados',value:data.completedServices,icon:'check',class:'green',note:'Servicios confirmados'}
    ];
    const range=reportRange(period,reference);
    const formatDate=value=>new Intl.DateTimeFormat('es-PE',{dateStyle:'short',timeZone:'America/Lima'}).format(value);
    const displayedPeriod=realMode?data.period:`${period} · ${formatDate(range.start)}–${formatDate(range.end.getTime()-1)}`;
    snapshot={period:displayedPeriod,reference,demo:!realMode,cards,states:data.states??[],mechanics:(data.mechanics??[]).filter(item=>item.count>0),
     recent,canEditOrders:can('ordenes','editar'),
     services:(data.services??[]).filter(item=>item.count>0).sort((a,b)=>b.count-a.count).slice(0,4),
     income:showIncome?data.income??[]:null,activity:[['Auxilios registrados',data.rescuesToday],['Recepciones registradas',data.counts?.receptions??0],['Diagnósticos registrados',data.counts?.diagnoses??0],['Entregas registradas',data.counts?.deliveries??0]]};
    summary.innerHTML=dashboardSummary(snapshot,financeAllowed);
    summary.querySelector('#open-analysis')?.addEventListener('click',()=>selectTab('analisis'));
    loadedSummary=ownKey;updateExport();
    await mountDiagnosisDashboard(summary);
   }
  }catch(error){if(fresh()&&key()===ownKey)target.innerHTML=`<section class="panel error-state" role="alert"><h2>No se pudo completar la consulta</h2><p>${e(error.message)}</p><p>Vuelve a aplicar los filtros para intentarlo.</p></section>`;}
  finally{target.removeAttribute('aria-busy');}
 }
 const exportMenu=root.querySelector('.dashboard-export');
 const exportStatus=root.querySelector('.dashboard-export-status');
 let exporting=false;
 function updateExport(){
  exportMenu.hidden=selected!=='resumen'||!summaryAllowed;
  exportMenu.querySelectorAll('button').forEach(button=>{button.disabled=!snapshot||exporting;});
 }
 exportMenu.addEventListener('keydown',event=>{if(event.key==='Escape'&&exportMenu.open){exportMenu.open=false;exportMenu.querySelector('summary').focus();}});
 exportMenu.addEventListener('click',async event=>{
  const format=event.target.closest('[data-export]')?.dataset.export;
  if(!format||!snapshot||exporting)return;
  const data=snapshot;
  exporting=true;updateExport();
  exportStatus.hidden=false;exportStatus.textContent='Preparando descarga…';
  try{
   const {downloadDashboard}=await import('./dashboard-export.js');
   await downloadDashboard(format,data);
   exportMenu.open=false;exportMenu.querySelector('summary').focus();
   exportStatus.textContent='Descarga preparada.';
  }catch{
   exportStatus.textContent='No se pudo generar el archivo. Intenta exportarlo nuevamente.';
  }finally{exporting=false;updateExport();}
 });
 async function selectTab(name){if(name==='analisis'&&!financeAllowed||name==='resumen'&&!summaryAllowed)return;selected=name;updateExport();updateLocation();await show();}
 root.querySelector('#tab-resumen').onclick=()=>selectTab('resumen');root.querySelector('#tab-analisis').onclick=()=>selectTab('analisis');
 root.querySelector('.dashboard-tabs').onkeydown=event=>{if(!['ArrowLeft','ArrowRight','Home','End'].includes(event.key))return;const tabs=[...root.querySelectorAll('[role=tab]:not(:disabled)')],index=tabs.indexOf(document.activeElement),next=event.key==='Home'?0:event.key==='End'?tabs.length-1:(index+(event.key==='ArrowRight'?1:-1)+tabs.length)%tabs.length;event.preventDefault();tabs[next].click();tabs[next].focus();};
 bindForm('dashboard-filter',async data=>{period=data.period;reference=data.reference;snapshot=null;exportStatus.hidden=true;updateExport();loadedSummary='';loadedAnalysis='';cache.clear();updateLocation();await show();});
 updateExport();
 await show();
}
