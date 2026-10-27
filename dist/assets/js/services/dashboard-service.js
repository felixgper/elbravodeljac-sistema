import {realMode} from './api-service.js';
import * as real from './real-workshop.js';
import {getStore} from './store-service.js';
import {ACTIVE_STATES,getContext,quoteTotal} from './workshop-service.js';
export function getActiveOrders(data){return data.states.filter(s=>['pending','repair','testing'].includes(s.key)).reduce((n,s)=>n+s.count,0);}
export async function getDashboard(period){if(realMode)return real.getDashboard(period||'Mes');if(period)return demoPeriod(period);
 const s=getStore(),today=new Date(),month=today.getMonth(),year=today.getFullYear();
 const income=Array.from({length:6},(_,i)=>{const d=new Date(year,month-5+i,1);return {month:new Intl.DateTimeFormat('es-PE',{month:'short'}).format(d).replace('.',''),cents:s.payments.filter(p=>{const x=new Date(p.createdAt);return x.getFullYear()===d.getFullYear()&&x.getMonth()===d.getMonth();}).reduce((n,p)=>n+p.amountCents,0)};});
 const states=[['pending','Pendientes','Pendiente','#d89a2a'],['repair','En reparación','En reparación','#791c2b'],['testing','En pruebas','En pruebas','#467bbb'],['completed','Finalizadas','Finalizada','#24906d'],['cancelled','Canceladas','Cancelada','#b85858']].map(([key,label,state,color])=>({key,label,color,count:s.orders.filter(o=>o.state===state).length}));
 const mechanics=s.users.filter(u=>['Mecánico','Supervisor'].includes(u.role)).map(u=>({label:u.name,count:s.orders.filter(o=>o.mechanicId===u.id&&ACTIVE_STATES.includes(o.state)).length})).sort((a,b)=>b.count-a.count).slice(0,4);
 const counts=new Map();s.orders.filter(o=>o.state==='Finalizada').forEach(o=>{getContext(s,o).quote.items.filter(i=>i.type==='Servicio').forEach(i=>counts.set(i.concept,(counts.get(i.concept)||0)+1));});
 const services=[...counts].map(([label,count])=>({label,count})).sort((a,b)=>b.count-a.count).slice(0,4);
 const fmt=d=>new Intl.DateTimeFormat('es-PE',{year:'numeric',month:'long'}).format(d);
 return {monthlyIncomeCents:income.at(-1).cents,completedServices:states.find(x=>x.key==='completed').count,rescuesToday:s.rescues.filter(a=>new Date(a.createdAt).toLocaleDateString('en-CA',{timeZone:'America/Lima'})===today.toLocaleDateString('en-CA',{timeZone:'America/Lima'})).length,income,states,mechanics,services,period:`${income[0].month}–${income.at(-1).month} ${year}`,currentMonth:fmt(today)};
}
export function getReport(period='Este mes'){if(realMode)return real.getReport(period);if(['Semana','Mes','Año'].includes(period)){const d=demoPeriod(period);return {services:d.services,months:d.income.map(i=>({label:i.month,cents:i.cents})),income:d.monthlyIncomeCents,orders:d.completedServices};}
 const s=getStore(),now=new Date(),start=new Date(now.getFullYear(),now.getMonth()-(period==='Últimos 6 meses'?5:0),1),end=new Date(now.getFullYear(),now.getMonth()+1,1);
 const inRange=value=>new Date(value)>=start&&new Date(value)<end;
 const orders=s.orders.filter(o=>o.state==='Finalizada'&&inRange(o.updates.at(-1)?.at||o.createdAt));
 const counts=new Map();orders.forEach(o=>getContext(s,o).quote.items.filter(i=>i.type==='Servicio').forEach(i=>counts.set(i.concept,(counts.get(i.concept)||0)+1)));
 const rescues=s.rescues.filter(r=>inRange(r.createdAt)).length;if(rescues)counts.set('Auxilio mecánico',rescues);
 const months=Array.from({length:period==='Últimos 6 meses'?6:1},(_,i)=>{const date=new Date(start.getFullYear(),start.getMonth()+i,1);return {label:new Intl.DateTimeFormat('es-PE',{month:'long',year:'numeric'}).format(date),cents:s.payments.filter(p=>{const d=new Date(p.createdAt);return d.getMonth()===date.getMonth()&&d.getFullYear()===date.getFullYear();}).reduce((n,p)=>n+p.amountCents,0)};});
 return {services:[...counts].map(([label,count])=>({label,count})),months,orders:orders.length,income:months.reduce((n,m)=>n+m.cents,0)};
}

function demoPeriod(period){
 const s=getStore(),parts=Object.fromEntries(new Intl.DateTimeFormat('en-CA',{year:'numeric',month:'2-digit',day:'2-digit',timeZone:'America/Lima'}).formatToParts(new Date()).map(p=>[p.type,p.value]));
 let start=new Date(`${parts.year}-${parts.month}-${parts.day}T00:00:00-05:00`),end;
 if(period==='Semana'){start.setUTCDate(start.getUTCDate()-(start.getUTCDay()+6)%7);end=new Date(start);end.setUTCDate(end.getUTCDate()+7);}
 else if(period==='Mes'){start.setUTCDate(1);end=new Date(start);end.setUTCMonth(end.getUTCMonth()+1);}
 else{start=new Date(`${parts.year}-01-01T00:00:00-05:00`);end=new Date(start);end.setUTCFullYear(end.getUTCFullYear()+1);}
 const between=t=>new Date(t)>=start&&new Date(t)<end,orders=s.orders.filter(o=>between(o.createdAt)),payments=s.payments.filter(p=>between(p.createdAt));
 const income=[];for(let cursor=new Date(start);cursor<end;){const next=new Date(cursor);if(period==='Año')next.setUTCMonth(next.getUTCMonth()+1);else next.setUTCDate(next.getUTCDate()+1);income.push({month:new Intl.DateTimeFormat('es-PE',{...(period==='Año'?{month:'short'}:{day:'2-digit'}),timeZone:'America/Lima'}).format(cursor),cents:payments.filter(p=>new Date(p.createdAt)>=cursor&&new Date(p.createdAt)<next).reduce((n,p)=>n+p.amountCents,0)});cursor=next;}
 const states=[['pending','Pendientes','Pendiente','#d89a2a'],['repair','En reparación','En reparación','#791c2b'],['testing','En pruebas','En pruebas','#467bbb'],['completed','Finalizadas','Finalizada','#24906d'],['cancelled','Canceladas','Cancelada','#b85858']].map(([key,label,state,color])=>({key,label,color,count:orders.filter(o=>o.state===state).length}));
 const finished=s.orders.filter(o=>o.state==='Finalizada'&&between(o.updates.at(-1)?.at||o.createdAt)),counts=new Map();finished.forEach(o=>getContext(s,o).quote?.items.filter(i=>i.type==='Servicio').forEach(i=>counts.set(i.concept,(counts.get(i.concept)||0)+1)));
 return {monthlyIncomeCents:payments.reduce((n,p)=>n+p.amountCents,0),completedServices:finished.length,rescuesToday:s.rescues.filter(a=>between(a.createdAt)).length,income,states,mechanics:s.users.filter(u=>['Mecánico','Supervisor'].includes(u.role)).map(u=>({label:u.name,count:orders.filter(o=>o.mechanicId===u.id&&ACTIVE_STATES.includes(o.state)).length})),services:[...counts].map(([label,count])=>({label,count})),period,currentMonth:period};
}
