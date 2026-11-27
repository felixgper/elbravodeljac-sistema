import {e} from './ui.js';
import {reportRange} from '../services/dashboard-service.js';

export const todayLima=()=>new Intl.DateTimeFormat('en-CA',{year:'numeric',month:'2-digit',day:'2-digit',timeZone:'America/Lima'}).format(new Date());
export function shiftPeriod(period,reference,direction){
 const date=new Date(reportRange(period,reference).start);
 if(period==='Día'||period==='Semana')date.setUTCDate(date.getUTCDate()+direction*(period==='Día'?1:7));
 if(period==='Mes')date.setUTCMonth(date.getUTCMonth()+direction);
 if(period==='Año')date.setUTCFullYear(date.getUTCFullYear()+direction);
 return date.toISOString().slice(0,10);
}
// Each instance owns its state object; no shared/global filter or event bus.
export function mountPeriodPicker(root,{label,periods,state,onChange}){
 state.period=periods.includes(state.period)?state.period:'Mes';state.reference ||= todayLima();
 root.innerHTML=`<div class="analysis-period-picker"><label>${e(label)}<select aria-label="${e(label)}">${periods.map(p=>`<option${p===state.period?' selected':''}>${e(p)}</option>`).join('')}</select></label><button type="button" class="secondary-button period-prev" aria-label="Período anterior">‹</button><label>Fecha de referencia<input type="date" value="${e(state.reference)}" max="${todayLima()}" required></label><button type="button" class="secondary-button period-next" aria-label="Período siguiente">›</button><button type="button" class="secondary-button period-now">Actual</button></div>`;
 const select=root.querySelector('select'),date=root.querySelector('input'),next=root.querySelector('.period-next');
 const reflect=()=>{select.value=state.period;date.value=state.reference;date.setAttribute('value',state.reference);for(const option of select.options)option.toggleAttribute('selected',option.value===state.period);next.disabled=reportRange(state.period,state.reference).end>new Date(todayLima()+'T05:00:00Z');};
 const apply=()=>{try{reportRange(select.value,date.value);if(date.value>todayLima())return;state.period=select.value;state.reference=date.value;date.removeAttribute('aria-invalid');reflect();onChange();}catch{date.setAttribute('aria-invalid','true');}};
 select.onchange=apply;date.onchange=apply;
 root.querySelector('.period-prev').onclick=()=>{date.value=shiftPeriod(state.period,state.reference,-1);apply();};
 next.onclick=()=>{date.value=shiftPeriod(state.period,state.reference,1);apply();};
 root.querySelector('.period-now').onclick=()=>{date.value=todayLima();apply();};reflect();
}
