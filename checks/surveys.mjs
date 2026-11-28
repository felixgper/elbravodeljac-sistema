import assert from 'node:assert/strict';
import fs from 'node:fs';
import {surveyQuestions,validateSurveyAnswer,surveyWhatsApp,issueSurvey,surveyStatus,surveyStatistics,validateSurvey,answerSurvey} from '../dist/assets/js/services/survey-service.js';
import {surveyCharts,satisfactionBand} from '../dist/assets/js/components/surveys.js';
import {setSession,clearSession} from '../dist/assets/js/services/api-service.js';

assert.equal(surveyQuestions.length,5);
validateSurveyAnswer([1,2,3,4,5],'á'.repeat(1200));
for(const values of [[],[1,2,3,4],[1,2,3,4,5,1],[0,2,3,4,5],[1,2,3,4,6],[1,2,3,4,1.2],['1',2,3,4,5]])assert.throws(()=>validateSurveyAnswer(values));
assert.throws(()=>validateSurveyAnswer([1,2,3,4,5],'a'.repeat(1201)));
const link='https://workshop.test/encuesta#t=synthetic-not-a-real-secret';
const whatsapp=new URL(surveyWhatsApp(link,'+51 987-654-321'));
assert.equal(whatsapp.origin,'https://wa.me');assert.equal(whatsapp.pathname,'/51987654321');
assert(whatsapp.searchParams.get('text').includes(link));assert(whatsapp.searchParams.get('text').includes('¡Gracias'));
for(const phone of ['987654321','javascript:alert(1)','+0000','+51<script>'])assert.equal(new URL(surveyWhatsApp(link,phone)).pathname,'/','Never guess country codes or inject URL');
assert.match(surveyCharts({total:0}),/Sin encuestas registradas/);
const charts=surveyCharts({total:2,average:4.4,aspects:[1,1,1,1,4.4],distribution:[0,0,0,1,1]});
assert(charts.includes('Buena satisfacción'));assert(charts.includes('4.4'));
assert.equal((charts.match(/<svg /g)||[]).length,1);assert.equal((charts.match(/<path /g)||[]).length,3);assert(!charts.includes('<meter'));assert(!charts.includes('NPS'));
for(const [n,label]of [[1,'Satisfacción baja'],[2.99,'Satisfacción baja'],[3,'Satisfacción regular'],[3.99,'Satisfacción regular'],[4,'Buena satisfacción'],[5,'Buena satisfacción']])assert.equal(satisfactionBand(n).label,label);
assert.throws(()=>satisfactionBand(0));assert.throws(()=>satisfactionBand(NaN));
assert(!surveyCharts({total:0,average:0}).includes('<svg'),'Unanswered invitations cannot produce a red or zero gauge');

const calls=[];
globalThis.fetch=async(path,options)=>{calls.push({path,options});return {ok:true,status:200,json:async()=>({data:null})};};
setSession({data:{id:'qa',roles:['Administrador'],permissions:['pagos.ver']},csrfToken:'synthetic-csrf'});
await issueSurvey('7',true);await surveyStatus('7');await surveyStatistics(new Date('2026-11-01T05:00:00Z'),new Date('2026-12-01T05:00:00Z'));
await validateSurvey('synthetic-token');await answerSurvey('synthetic-token',[1,2,3,4,5],'Gracias');
assert.equal(calls[0].options.headers['X-CSRF-Token'],'synthetic-csrf');assert.equal(JSON.parse(calls[0].options.body).regenerate,true);
assert(calls[2].path.includes('desde=2026-11-01'));assert(calls.every(c=>!c.path.includes('synthetic-token')),'Tokens must never appear in request URLs');
assert.equal(JSON.parse(calls[4].options.body).ratings.length,5);clearSession();
const publicSource=fs.readFileSync('dist/assets/js/survey.js','utf8');
for(const forbidden of ['localStorage','sessionStorage','console.log','innerHTML=error'])assert(!publicSource.includes(forbidden));
assert(publicSource.includes('history.replaceState'));assert(publicSource.includes('form.querySelector(\'button\').disabled=true'));
const page=fs.readFileSync('dist/encuesta.html','utf8');assert(page.includes('width=device-width'));assert(page.includes('no-referrer'));
console.log('Encuestas: PASS validación, Unicode, WhatsApp manual, estados vacíos, gráficos, HTTP/CSRF y ausencia de tokens en URL; sin bases ni navegador real.');

const {reportRange,getIncomeReport}=await import('../dist/assets/js/services/dashboard-service.js');
const {shiftPeriod}=await import('../dist/assets/js/components/analysis-periods.js');
for(const [p,ref,start,end]of [['Día','2024-02-29','2024-02-29','2024-03-01'],['Semana','2026-01-01','2025-12-29','2026-01-05'],['Mes','2026-01-31','2026-01-01','2026-02-01'],['Año','2026-05-01','2026-01-01','2027-01-01']]){const r=reportRange(p,ref);assert.equal(r.start.toISOString(),start+'T05:00:00.000Z');assert.equal(r.end.toISOString(),end+'T05:00:00.000Z');}
assert.equal(shiftPeriod('Día','2024-03-01',-1),'2024-02-29');assert.equal(shiftPeriod('Mes','2026-01-31',1),'2026-02-01');assert.equal(shiftPeriod('Año','2026-06-01',-1),'2025-01-01');
const financial=[];
globalThis.fetch=async(path,options)=>{financial.push({path,options});return {ok:true,status:200,json:async()=>({data:{incomeCents:'0',months:[{label:'2026-11-24',cents:'0'}],services:[],orders:0}})};};
for(const p of ['Día','Semana','Mes']){const r=await getIncomeReport(p,'2026-11-24');assert.equal(r.income,0);assert.equal(r.comparison.income,0);}
assert.equal(financial.length,6);assert(financial.every(x=>x.path.startsWith('/api/v1/reportes?')&&x.path.includes('desde=')&&x.path.includes('hasta=')));
console.log('PASS: velocímetro solo Q5, límites de color, períodos de Lima, ingresos por intervalos autorizados y períodos sin datos.');

const {dashboardSummary}=await import('../dist/assets/js/views/dashboard-summary.js');
const summaryInput={period:'Mes',cards:[{label:'Ingresos del período',value:'S/ 0.00',icon:'wallet'},{label:'Órdenes activas',value:0,icon:'wrench'},{label:'Servicios finalizados',value:0,icon:'check'}],states:[{label:'Pendientes',count:0,color:'#d89a2a'}],mechanics:[],income:[],services:[],activity:[],canViewOrders:true,satisfactionAllowed:true};
const summary=dashboardSummary(summaryInput,true);
assert.equal((summary.match(/<article class="kpi-card/g)||[]).length,4);
assert(summary.indexOf('dashboard-income')<summary.indexOf('dashboard-satisfaction"'));
assert(summary.indexOf('dashboard-satisfaction"')<summary.indexOf('dashboard-orders'));
for(const title of ['Órdenes y estatus','Trabajo por mecánico','Servicios más realizados'])assert(summary.includes(title));
assert(!dashboardSummary({...summaryInput,cards:summaryInput.cards.slice(1),income:null},false).includes('dashboard-satisfaction-card'));
console.log('PASS: estructura 4/3/3, orden de paneles y restricción del nuevo indicador por permisos.');

// A zero total has no donut; real orders retain the original chart renderer.
assert(summary.includes('dashboard-orders-empty'));
assert(summary.includes('No hay órdenes registradas.'));
assert(!summary.includes('class="donut"'));
assert(!summary.includes('class="chart-legend"'));
const populatedStates=[{label:'Pendientes',count:1,color:'#d89a2a'},{label:'Finalizadas',count:2,color:'#36a376'}];
const populatedSummary=dashboardSummary({...summaryInput,states:populatedStates},true);
const {donutChart}=await import('../dist/assets/js/components/charts.js');
assert(populatedSummary.includes(donutChart(populatedStates)));
assert(!populatedSummary.includes('dashboard-orders-empty'));
assert.equal(summary.split('</section>')[0],populatedSummary.split('</section>')[0]);
const css=fs.readFileSync('dist/assets/css/gestion-carbon.css','utf8');
assert(css.includes('.dashboard-orders .donut'));
assert(css.includes('max-height:184px'));
console.log('PASS: empty orders omit the donut; populated orders retain the original chart and upper indicators.');

// Annual income keeps the authorized ledger and existing Lima boundaries.
financial.length=0;
const annualIncome=await getIncomeReport('Año','2024-02-29');
assert.equal(annualIncome.range.start.toISOString(),'2024-01-01T05:00:00.000Z');
assert.equal(annualIncome.range.end.toISOString(),'2025-01-01T05:00:00.000Z');
assert.equal(annualIncome.range.previous.toISOString(),'2023-01-01T05:00:00.000Z');
assert.equal(financial.length,2);
assert(financial[0].path.includes('desde=2024-01-01')&&financial[0].path.includes('hasta=2025-01-01'));
const {mountPeriodPicker}=await import('../dist/assets/js/components/analysis-periods.js');
const compactSelect={};const compactRoot={innerHTML:'',querySelector:s=>{assert.equal(s,'select');return compactSelect;}};
const independent={period:'Mes',reference:'2024-02-29'};let changed=0;
mountPeriodPicker(compactRoot,{label:'Período de ingresos',periods:['Semana','Mes','Año'],state:independent,onChange:()=>changed++,compact:true});
assert(!compactRoot.innerHTML.includes('<input'));assert(!compactRoot.innerHTML.includes('<button'));assert(!compactRoot.innerHTML.includes('Fecha de referencia'));
assert.deepEqual([...compactRoot.innerHTML.matchAll(/<option[^>]*>(.*?)<\/option>/g)].map(x=>x[1]),['Semana','Mes','Año']);
compactSelect.onchange({target:{value:'Año'}});assert.equal(changed,1);assert.equal(independent.period,'Año');assert.equal(independent.reference,'2024-02-29');
const integratedSource=fs.readFileSync('dist/assets/js/views/dashboard.js','utf8');assert(!integratedSource.includes("field('Fecha de referencia'"));assert(!integratedSource.includes('reference=data.reference'));
console.log('PASS: annual income boundaries, compact period-only controls and internal base date preserved.');
