import {mountPeriodPicker,todayLima} from './analysis-periods.js';
import {reportRange} from '../services/dashboard-service.js';
import {realSession} from '../services/api-service.js';
import {surveyStatus,issueSurvey,surveyWhatsApp,surveyStatistics} from '../services/survey-service.js';
import {e,stamp,table} from './ui.js';

export async function mountDeliverySurvey(root,{orderId,phone='',canShare=false}){
 const section=document.createElement('section');section.className='panel separated survey-delivery';root.append(section);
 const owner=realSession()?.id||'';let data=null,url='',busy=false,intent='share';
 const current=()=>section.isConnected&&(realSession()?.id||'')===owner;
 const feedback=message=>{const target=section.querySelector('.survey-feedback');if(target)target.textContent=message;};
 const usable=()=>data?.status==='generada'&&Date.parse(data.expiresAt)>Date.now()&&url;
 const replacement=()=>data&&(data.status==='vencida'||(data.status==='generada'&&Date.parse(data.expiresAt)<=Date.now()));
 function popup(){try{const p=globalThis.open?.('about:blank','_blank');if(p)p.opener=null;return p;}catch{return null;}}
 async function deliver(action,window){
  if(action==='copy'){
   try{await navigator.clipboard.writeText(url);if(current())feedback('Enlace copiado. Puedes pegarlo en un mensaje.');}
   catch{if(!current())return;const label=section.querySelector('.survey-manual-copy');label.hidden=false;const input=label.querySelector('input');input.value=url;input.focus();input.select();feedback('No se pudo copiar automáticamente. Selecciona y copia el enlace.');}
  }else{
   const target=window||popup();if(target&&!target.closed){try{target.location.replace(surveyWhatsApp(url,phone));feedback('Confirma el destinatario y el envío en WhatsApp.');}catch{target.close();feedback('No se pudo abrir WhatsApp. Puedes copiar el enlace.');}}
   else feedback('El navegador bloqueó WhatsApp. Permite la ventana o utiliza Copiar enlace.');
  }
 }
 async function prepare(action,regenerate=false){
  if(busy||!current()||!canShare||data?.status==='respondida')return;busy=true;
  const window=action==='share'?popup():null;section.querySelectorAll('button').forEach(b=>b.disabled=true);feedback('Preparando la encuesta…');
  try{
   if(!usable()||regenerate){const result=await issueSurvey(orderId,regenerate);if(!current()){window?.close();return;}data=result.data;url=result.url||'';}
   render();if(data?.status==='respondida'){window?.close();return;}if(!usable())throw Error('No se pudo recuperar un enlace vigente.');await deliver(action,window);
  }catch(error){window?.close();if(current())feedback(error.status===409?'No se pudo preparar otro enlace. Espera un minuto y vuelve a abrir esta entrega.':error.message);}
  finally{busy=false;if(current())section.querySelectorAll('button').forEach(b=>b.disabled=false);}
 }
 function choose(action){
  if(busy||!current())return;
  if(replacement()){
   intent=action;section.querySelector('.survey-confirmation').hidden=false;section.querySelector('.survey-confirm').textContent=action==='copy'?'Crear otro enlace y copiar':'Crear otro enlace y compartir';section.querySelector('.survey-confirm').focus();return;
  }
  prepare(action,false);
 }
 function render(){
  if(!current())return;
  section.innerHTML='<h2 class="form-section-title">Encuesta de satisfacción</h2><p>Invita al cliente a calificar el servicio recibido.</p><div class="survey-share-actions"></div><p class="survey-feedback" role="status" aria-live="polite"></p>';
  const actions=section.querySelector('.survey-share-actions');
  if(data?.status==='respondida'){actions.innerHTML='<p>Gracias. Ya recibimos la opinión del cliente.</p>';return;}
  if(!canShare)return;
  actions.innerHTML=`<div class="survey-actions-row"><button type="button" class="primary-button survey-share">Compartir por WhatsApp</button><button type="button" class="secondary-button survey-copy">Copiar enlace</button></div><p>Responder es voluntario. Tú decides si envías el mensaje.</p><label class="survey-manual-copy" hidden>Enlace de la encuesta<input class="survey-url" readonly aria-label="Enlace de la encuesta"></label><div class="survey-confirmation" role="group" aria-label="Confirmar un nuevo enlace" hidden><p>Prepararemos otro enlace. Si conservas uno anterior, dejará de funcionar. ¿Deseas continuar?</p><button type="button" class="primary-button survey-confirm">Crear otro enlace y compartir</button><button type="button" class="secondary-button survey-cancel">Cancelar</button></div>`;
  actions.querySelector('.survey-share').onclick=()=>choose('share');actions.querySelector('.survey-copy').onclick=()=>choose('copy');
  actions.querySelector('.survey-confirm').onclick=()=>prepare(intent,true);actions.querySelector('.survey-cancel').onclick=()=>{actions.querySelector('.survey-confirmation').hidden=true;actions.querySelector(intent==='copy'?'.survey-copy':'.survey-share').focus();};
 }
 section.innerHTML='<h2 class="form-section-title">Encuesta de satisfacción</h2><p role="status">Consultando encuesta…</p>';
 try{const result=await surveyStatus(orderId);data=result.data;url=canShare?result.url||'':'';render();}catch(error){if(current())section.innerHTML='<h2 class="form-section-title">Encuesta de satisfacción</h2><p role="status">'+e('No se pudo consultar la encuesta: '+error.message)+'</p>';}
}

export function satisfactionBand(average){
 if(!Number.isFinite(average)||average<1||average>5)throw Error('Calificación no válida.');
 return average<3?{label:'Satisfacción baja',color:'#c95459'}:average<4?{label:'Satisfacción regular',color:'#d8ab48'}:{label:'Buena satisfacción',color:'#36a376'};
}
export function surveyCharts(data){
 if(!data.total)return '<p class="empty-state">Sin encuestas registradas</p>';
 const mean=Number(data.average),band=satisfactionBand(mean),angle=Math.PI*(1-(mean-1)/4),x=180+105*Math.cos(angle),y=170-105*Math.sin(angle);
 const value=new Intl.NumberFormat('es-PE',{minimumFractionDigits:1,maximumFractionDigits:2}).format(mean);
 return `<div class="satisfaction-gauge"><svg viewBox="0 0 360 210" role="img" aria-label="${e(band.label)}: ${e(value)} de 5 estrellas; ${e(data.total)} encuestas respondidas"><path d="M50 170 A130 130 0 0 1 180 40" stroke="#c95459"/><path d="M180 40 A130 130 0 0 1 271.924 78.076" stroke="#d8ab48"/><path d="M271.924 78.076 A130 130 0 0 1 310 170" stroke="#36a376"/><g class="satisfaction-scale"><text x="38" y="197">1</text><text x="176" y="20">3</text><text x="285" y="62">4</text><text x="316" y="197">5</text></g><line x1="180" y1="170" x2="${x.toFixed(3)}" y2="${y.toFixed(3)}"/><circle cx="180" cy="170" r="8"/></svg><p class="satisfaction-value">${e(value)} <span>/ 5</span></p><p class="satisfaction-interpretation" style="color:${band.color}">${e(band.label)}</p><p>${e(data.total)} ${data.total===1?'encuesta respondida':'encuestas respondidas'}</p><div class="satisfaction-legend"><span><i class="satisfaction-red"></i>Baja: 1 a menos de 3</span><span><i class="satisfaction-yellow"></i>Regular: 3 a menos de 4</span><span><i class="satisfaction-green"></i>Buena: 4 a 5</span></div></div>`;
}
export async function mountSurveyAnalysis(root,range,state={},options={}){
 const section=options.section||document.createElement('section');
 if(!options.section)section.className='panel separated survey-analysis';
 section.innerHTML=(options.section?'<div class="panel-heading"><h2>Satisfacción del cliente</h2><div class="satisfaction-filter"></div></div>':'<h2 class="form-section-title">SATISFACCIÓN DEL CLIENTE</h2>')+(options.section?'':'<div class="satisfaction-filter"></div>')+'<div class="survey-results" role="status"></div>';
 if(!options.section)root.append(section);let requestId=0;state.period ||= 'Mes';state.reference ||= todayLima();
 const target=section.querySelector('.survey-results');
 async function draw(){
  const id=++requestId;options.onUpdate?.({loading:true,state});target.innerHTML='<p>Consultando opiniones…</p>';target.setAttribute('aria-busy','true');
  try{const r=reportRange(state.period,state.reference),{data}=await surveyStatistics(r.start,r.end);if(!section.isConnected||id!==requestId)return;
   const fmt=d=>new Intl.DateTimeFormat('es-PE',{dateStyle:'short',timeZone:'America/Lima'}).format(d);
   target.innerHTML=`<p class="analysis-period-caption">${fmt(r.start)} – ${fmt(new Date(r.end.getTime()-1))}</p>`+surveyCharts(data);
   options.onUpdate?.({data,state,range:r});
  }catch(error){if(section.isConnected&&id===requestId){target.textContent='No se pudo consultar la satisfacción: '+error.message;options.onUpdate?.({error,state});}}
  finally{if(id===requestId)target.removeAttribute('aria-busy');}
 }
 mountPeriodPicker(section.querySelector('.satisfaction-filter'),{label:options.section?'Período de satisfacción':'Ver por:',periods:['Semana','Mes','Año'],state,onChange:draw,compact:Boolean(options.section)});await draw();
}
