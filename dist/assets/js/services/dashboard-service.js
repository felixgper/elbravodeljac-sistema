import {realMode} from './api-service.js';
import * as real from './real-workshop.js';
import {getStore} from './store-service.js';
import {ACTIVE_STATES,getContext,quoteTotal} from './workshop-service.js';
export function getActiveOrders(data){return data.states.filter(s=>['pending','repair','testing'].includes(s.key)).reduce((n,s)=>n+s.count,0);}
export async function getDashboard(){if(realMode)return real.getDashboard();
 const s=getStore(),today=new Date(),month=today.getMonth(),year=today.getFullYear();
 const income=Array.from({length:6},(_,i)=>{const d=new Date(year,month-5+i,1);return {month:new Intl.DateTimeFormat('es-PE',{month:'short'}).format(d).replace('.',''),cents:s.payments.filter(p=>{const x=new Date(p.createdAt);return x.getFullYear()===d.getFullYear()&&x.getMonth()===d.getMonth();}).reduce((n,p)=>n+p.amountCents,0)};});
 const states=[['pending','Pendientes','Pendiente','#d89a2a'],['repair','En reparación','En reparación','#791c2b'],['testing','En pruebas','En pruebas','#467bbb'],['completed','Finalizadas','Finalizada','#24906d'],['cancelled','Canceladas','Cancelada','#b85858']].map(([key,label,state,color])=>({key,label,color,count:s.orders.filter(o=>o.state===state).length}));
 const mechanics=s.users.filter(u=>['Mecánico','Supervisor'].includes(u.role)).map(u=>({label:u.name,count:s.orders.filter(o=>o.mechanicId===u.id&&ACTIVE_STATES.includes(o.state)).length})).sort((a,b)=>b.count-a.count).slice(0,4);
 const counts=new Map();s.orders.filter(o=>o.state==='Finalizada').forEach(o=>{getContext(s,o).quote.items.filter(i=>i.type==='Servicio').forEach(i=>counts.set(i.concept,(counts.get(i.concept)||0)+1));});
 const services=[...counts].map(([label,count])=>({label,count})).sort((a,b)=>b.count-a.count).slice(0,4);
 const fmt=d=>new Intl.DateTimeFormat('es-PE',{year:'numeric',month:'long'}).format(d);
 return {monthlyIncomeCents:income.at(-1).cents,completedServices:states.find(x=>x.key==='completed').count,rescuesToday:s.rescues.filter(a=>new Date(a.createdAt).toLocaleDateString('en-CA',{timeZone:'America/Lima'})===today.toLocaleDateString('en-CA',{timeZone:'America/Lima'})).length,income,states,mechanics,services,period:`${income[0].month}–${income.at(-1).month} ${year}`,currentMonth:fmt(today)};
}
export function getReport(period='Este mes'){if(realMode)return real.getReport(period);
 const s=getStore(),now=new Date(),start=new Date(now.getFullYear(),now.getMonth()-(period==='Últimos 6 meses'?5:0),1),end=new Date(now.getFullYear(),now.getMonth()+1,1);
 const inRange=value=>new Date(value)>=start&&new Date(value)<end;
 const orders=s.orders.filter(o=>o.state==='Finalizada'&&inRange(o.updates.at(-1)?.at||o.createdAt));
 const counts=new Map();orders.forEach(o=>getContext(s,o).quote.items.filter(i=>i.type==='Servicio').forEach(i=>counts.set(i.concept,(counts.get(i.concept)||0)+1)));
 const rescues=s.rescues.filter(r=>inRange(r.createdAt)).length;if(rescues)counts.set('Auxilio mecánico',rescues);
 const months=Array.from({length:period==='Últimos 6 meses'?6:1},(_,i)=>{const date=new Date(start.getFullYear(),start.getMonth()+i,1);return {label:new Intl.DateTimeFormat('es-PE',{month:'long',year:'numeric'}).format(date),cents:s.payments.filter(p=>{const d=new Date(p.createdAt);return d.getMonth()===date.getMonth()&&d.getFullYear()===date.getFullYear();}).reduce((n,p)=>n+p.amountCents,0)};});
 return {services:[...counts].map(([label,count])=>({label,count})),months,orders:orders.length,income:months.reduce((n,m)=>n+m.cents,0)};
}
