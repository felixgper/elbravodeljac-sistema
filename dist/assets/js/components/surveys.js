import {mountPeriodPicker,todayLima} from './analysis-periods.js';
import {reportRange} from '../services/dashboard-service.js';
import {realSession} from '../services/api-service.js';
import {surveyStatus,issueSurvey,surveyWhatsApp,surveyStatistics} from '../services/survey-service.js';
import {e,stamp,table} from './ui.js';

export async function mountDeliverySurvey(root,{orderId,phone='',canShare=false}){
 const section=document.createElement('section');section.className='panel separated survey-delivery';root.append(section);
 const owner=realSession()?.id||'';let data=null,url='',busy=false;
 const current=()=>section.isConnected&&(realSession()?.id||'')===owner;
 const feedback=message=>{const target=section.querySelector('.survey-feedback');if(target)target.textContent=message;};
 const usable=()=>data?.status==='generada'&&Date.parse(data.expiresAt)>Date.now()&&url;
 async function prepare(regenerate){
  if(busy||!current())return;busy=true;
  let popup=null;try{popup=globalThis.open?.('about:blank','_blank');if(popup)popup.opener=null;}catch{}
  section.querySelectorAll('button').forEach(button=>button.disabled=true);feedback('Preparando la encuesta…');
  try{
   const result=await issueSurvey(orderId,regenerate);if(!current()){popup?.close();return;}
   data=result.data;url=result.url;render();feedback('Confirma el destinatario y el envío en WhatsApp.');
   if(popup&&!popup.closed){try{popup.location.replace(surveyWhatsApp(url,phone));}catch{popup.close();}}else section.querySelector('.survey-share')?.focus();
  }catch(error){popup?.close();if(current())feedback(error.status===409?'No se pudo preparar otro enlace. Espera un minuto y vuelve a abrir esta entrega.':error.message);}
  finally{busy=false;if(current())section.querySelectorAll('button').forEach(button=>button.disabled=false);}
 }
 function render(){
  if(!current())return;
  section.innerHTML='<h2 class="form-section-title">Encuesta de satisfacción</h2><p>Invita al cliente a calificar el servicio recibido.</p><div class="survey-share-actions"></div><p class="survey-feedback" role="status" aria-live="polite"></p>';
  const actions=section.querySelector('.survey-share-actions');
  if(data?.status==='respondida'){actions.innerHTML='<p>Gracias. Ya recibimos la opinión del cliente.</p>';return;}
  if(!canShare)return;
  if(usable()){
   actions.innerHTML=`<div class="survey-actions-row"><a class="primary-button survey-share" target="_blank" rel="noopener noreferrer">Compartir encuesta por WhatsApp</a><button type="button" class="secondary-button survey-copy">Copiar enlace</button></div><p>Responder es voluntario. El mensaje no se envía automáticamente.${!/^\+[1-9]\d{7,14}$/.test(String(phone).replace(/[\s().-]/g,''))?' Elige el contacto en WhatsApp.':''}</p><label class="survey-manual-copy" hidden>Enlace de la encuesta<input class="survey-url" readonly aria-label="Enlace de la encuesta"></label>`;
   const share=actions.querySelector('.survey-share');share.href=surveyWhatsApp(url,phone);
   share.onclick=event=>{if(!current()||!usable()){event.preventDefault();if(current()){url='';render();feedback('Confirma la preparación de otro enlace para compartir.');}return;}feedback('Confirma el envío en WhatsApp.');};
   actions.querySelector('.survey-copy').onclick=async()=>{if(!current()||!usable()){url='';render();return;}try{await navigator.clipboard.writeText(url);if(current())feedback('Enlace copiado. Puedes pegarlo en un mensaje.');}catch{const label=actions.querySelector('.survey-manual-copy');label.hidden=false;const input=label.querySelector('input');input.value=url;input.focus();input.select();feedback('Selecciona y copia el enlace.');}};
  }else{
   actions.innerHTML=`<button type="button" class="primary-button survey-share" ${data?'aria-expanded="false"':''}>Compartir encuesta por WhatsApp</button><p>Responder es voluntario. Tú decides si envías el mensaje.</p>${data?'<div class="survey-confirmation" role="group" aria-label="Confirmar un nuevo enlace" hidden><p>Prepararemos otro enlace. Si conservas uno anterior, dejará de funcionar. ¿Deseas continuar?</p><button type="button" class="primary-button survey-confirm">Crear otro enlace y compartir</button><button type="button" class="secondary-button survey-cancel">Cancelar</button></div>':''}`;
   const share=actions.querySelector('.survey-share');share.onclick=()=>{if(!current())return;if(!data){prepare(false);return;}actions.querySelector('.survey-confirmation').hidden=false;share.setAttribute('aria-expanded','true');actions.querySelector('.survey-confirm').focus();};
   if(data){actions.querySelector('.survey-confirm').onclick=()=>prepare(true);actions.querySelector('.survey-cancel').onclick=()=>{actions.querySelector('.survey-confirmation').hidden=true;share.setAttribute('aria-expanded','false');share.focus();};}
  }
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
 section.innerHTML=(options.section?'<div class="panel-heading"><h2>Satisfacción del cliente</h2></div>':'<h2 class="form-section-title">SATISFACCIÓN DEL CLIENTE</h2>')+'<div class="satisfaction-filter"></div><div class="survey-results" role="status"></div>';
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
 mountPeriodPicker(section.querySelector('.satisfaction-filter'),{label:'Ver por:',periods:['Semana','Mes','Año'],state,onChange:draw});await draw();
}
