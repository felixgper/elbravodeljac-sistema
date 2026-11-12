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

export function dashboardSummary({period,cards,states,mechanics,income,activity,canViewOrders,services}, financeAllowed) {
 const hasIncome=income?.some(point=>point.cents!==0);
 const hasOrders=states.some(state=>state.count>0);
 const stateChart=donutChart(states);
 return `<section class="kpi-grid" aria-label="Resumen de actividad" style="--dashboard-cards:${cards.length}">
  ${cards.map(card=>`<article class="kpi-card"><span class="kpi-icon ${card.class}">${icon(card.icon)}</span><div><h2>${e(card.label)}</h2><strong class="kpi-value">${e(card.value)}</strong></div><p>${e(card.note)}</p></article>`).join('')}
 </section>
 <div class="dashboard-grid dashboard-overview ${income?'':'dashboard-overview-restricted'}">
  ${income?panel('Evolución de ingresos',`${hasIncome?incomeChart(income,840,'summary-income',172):empty('El ingreso neto del período es cero.')}<p class="dashboard-caption">${e(period)} · Pagos netos registrados</p>`,'dashboard-income'):''}
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
