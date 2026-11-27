import {getIncomeReport} from '../services/dashboard-service.js';
import {mountPeriodPicker,todayLima} from '../components/analysis-periods.js';
import {donutChart,incomeChart,rankedColumns} from '../components/charts.js';
import {e} from '../components/ui.js';
import {icon} from '../components/icons.js';

const empty = message => `<p class="dashboard-empty">${e(message)}</p>`;
const panel = (title,body,extra='',action='') => `<section class="panel ${extra}"><div class="panel-heading"><h2>${title}</h2>${action}</div>${body}</section>`;

function orderStatus(states){
 const total=states.reduce((sum,state)=>sum+Number(state.count),0);
 const maximum=Math.max(1,...states.map(state=>Number(state.count)));
 return `<ul class="dashboard-status-bars">${states.map(state=>{
  const color=/^#[0-9a-f]{6}$/i.test(state.color)?state.color:'#a55068';
  return `<li><span>${e(state.label)}</span><span class="dashboard-status-track" aria-hidden="true"><span style="width:${Number(state.count)/maximum*100}%;background:${color}"></span></span><strong>${e(state.count)}</strong></li>`;
 }).join('')}</ul><p class="dashboard-caption">Total: <strong>${total} órdenes</strong></p>`;
}

export function dashboardSummary({period,cards,states,mechanics,income,activity,canViewOrders,services,satisfactionAllowed=false}, financeAllowed) {
 satisfactionAllowed=Boolean(satisfactionAllowed&&financeAllowed);
 const hasOrders=states.some(state=>state.count>0);
 const stateChart=donutChart(states);
 return `<section class="kpi-grid" aria-label="Resumen de actividad" style="--dashboard-cards:${cards.length+(satisfactionAllowed?1:0)}">
  ${cards.map(card=>`<article class="kpi-card"><span class="kpi-icon ${card.class}">${icon(card.icon)}</span><div><h2>${e(card.label)}</h2><strong class="kpi-value">${e(card.value)}</strong></div><p>${e(card.note)}</p></article>`).join('')}
  ${satisfactionAllowed?`<article class="kpi-card dashboard-satisfaction-card"><span class="kpi-icon amber">${icon('star')}</span><div><h2>Satisfacción del cliente</h2><strong class="kpi-value" data-satisfaction-value>Consultando…</strong></div><p data-satisfaction-note>Promedio de satisfacción general</p></article>`:''}
 </section>
 <div class="dashboard-grid dashboard-overview ${income?'':'dashboard-overview-restricted'} ${satisfactionAllowed?'dashboard-overview-satisfaction':''}">
  ${income?panel('Evolución de ingresos','<div class="summary-income-filter"></div><div class="summary-income-result" role="status"></div>','dashboard-income'):''}
  ${satisfactionAllowed?panel(`${icon('star')} Satisfacción del cliente`,'','dashboard-satisfaction'):''}
  ${panel('Estados de órdenes',`${stateChart}${hasOrders?'':'<p class="dashboard-caption">Sin órdenes registradas.</p>'}`,'dashboard-orders')}
 </div>
 <div class="dashboard-grid dashboard-detail">
  ${panel('Órdenes y estatus',orderStatus(states),'dashboard-status',canViewOrders?'<a class="dashboard-small-action" href="#/ordenes">Ver todas</a>':'')}
  ${panel('Trabajo por mecánico',mechanics.length?rankedColumns(mechanics)+'<p class="dashboard-caption">Órdenes activas asignadas</p>':empty('No hay órdenes activas asignadas.'))}
  ${panel('Servicios más realizados',services.length?rankedColumns(services,true)+'<p class="dashboard-caption">Cantidades confirmadas en el período</p>':empty('No hay servicios finalizados en el período.'))}
 </div>
 <section class="panel dashboard-activity" aria-label="Actividad del taller">
  <dl>${activity.map(([label,value])=>`<div><dt>${e(label)}</dt><dd>${e(value)}</dd></div>`).join('')}</dl>
  ${financeAllowed?'<button type="button" class="dashboard-small-action" id="open-analysis">Ver análisis financiero</button>':''}
 </section>`;
}

export async function mountSummaryIncome(section,state){
 if(!section)return;state.period ||= 'Mes';state.reference ||= todayLima();let requestId=0;
 const target=section.querySelector('.summary-income-result');
 async function draw(){
  const id=++requestId;target.setAttribute('aria-busy','true');target.innerHTML='<p class="dashboard-caption">Consultando ingresos…</p>';
  try{const data=await getIncomeReport(state.period,state.reference);if(!section.isConnected||id!==requestId)return;
   const points=data.months.map(p=>({month:p.label,cents:p.cents}));
   const fmt=d=>new Intl.DateTimeFormat('es-PE',{dateStyle:'short',timeZone:'America/Lima'}).format(d);
   target.innerHTML=(points.some(p=>p.cents!==0)?incomeChart(points,Math.max(260,section.clientWidth-30),'summary-income',172,true):empty('No hay ingresos netos registrados en el período.'))+`<p class="dashboard-caption">${fmt(data.range.start)} – ${fmt(new Date(data.range.end.getTime()-1))} · Pagos netos registrados</p>`;
  }catch(error){if(section.isConnected&&id===requestId)target.textContent='No se pudieron consultar los ingresos: '+error.message;}
  finally{if(id===requestId)target.removeAttribute('aria-busy');}
 }
 mountPeriodPicker(section.querySelector('.summary-income-filter'),{label:'Ver ingresos por:',periods:['Día','Semana','Mes'],state,onChange:draw});await draw();
}
