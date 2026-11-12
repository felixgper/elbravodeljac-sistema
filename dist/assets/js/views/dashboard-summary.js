import {donutChart,incomeChart,rankedColumns} from '../components/charts.js';
import {badge,e} from '../components/ui.js';
import {icon} from '../components/icons.js';

const empty = message => `<p class="dashboard-empty">${e(message)}</p>`;
const panel = (title,body,extra='',action='') => `<section class="panel ${extra}"><div class="panel-heading"><h2>${title}</h2>${action}</div>${body}</section>`;

function ordersTable(recent,canEditOrders){
 if(!recent)return empty('Tu cuenta no tiene acceso al listado de órdenes.');
 if(recent.error)return empty('No se pudo consultar el listado de órdenes.');
 if(!recent.orders.length)return empty('No hay órdenes disponibles para tu cuenta.');
 return `<div class="dashboard-recent-scroll"><table class="dashboard-recent-table">
  <thead><tr><th scope="col">N.º OT</th><th scope="col">Cliente</th><th scope="col">Placa</th><th scope="col">Estado</th></tr></thead>
  <tbody>${recent.orders.map(order=>{
   const code=e(order.code||order.id);
   return `<tr><td>${canEditOrders?`<a href="#/ordenes/editar/${encodeURIComponent(order.id)}">${code}</a>`:code}</td><td>${e(order.clientName||'—')}</td><td>${e(order.vehiclePlate||'—')}</td><td>${badge(order.state)}</td></tr>`;
  }).join('')}</tbody></table></div>`;
}

export function dashboardSummary({period,cards,states,mechanics,income,activity,recent,services,canEditOrders}, financeAllowed) {
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
  ${panel('Órdenes recientes',ordersTable(recent,canEditOrders)+(recent?.orders?.length?'<p class="dashboard-caption">Últimas órdenes disponibles, sin filtro de período.</p>':''),'dashboard-recent',recent?'<a class="dashboard-small-action" href="#/ordenes">Ver todas</a>':'')}
  ${panel('Trabajo por mecánico',mechanics.length?rankedColumns(mechanics)+'<p class="dashboard-caption">Órdenes activas asignadas</p>':empty('No hay órdenes activas asignadas.'))}
  ${panel('Servicios más realizados',services.length?rankedColumns(services,true)+'<p class="dashboard-caption">Cantidades confirmadas en el período</p>':empty('No hay servicios finalizados en el período.'))}
 </div>
 <section class="panel dashboard-activity" aria-label="Actividad del taller">
  <dl>${activity.map(([label,value])=>`<div><dt>${e(label)}</dt><dd>${e(value)}</dd></div>`).join('')}</dl>
  ${financeAllowed?'<button type="button" class="dashboard-small-action" id="open-analysis">Ver análisis financiero</button>':''}
 </section>`;
}
