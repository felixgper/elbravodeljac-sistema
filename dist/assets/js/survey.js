import {surveyQuestions,validateSurvey,answerSurvey,validateSurveyAnswer} from './services/survey-service.js';
import {e} from './components/ui.js';

const token=new URLSearchParams(location.hash.slice(1)).get('t')||'';
// Fragments never reach the server or Referer. Clear history after reading;
// retain the secret only in this page's memory, never in browser storage.
history.replaceState(null,'',location.pathname);
// Reopening a survey link in this same tab is a new invitation/validation,
// not a continuation of the previous form bound to a different token.
window.addEventListener('hashchange',()=>{if(new URLSearchParams(location.hash.slice(1)).get('t'))location.reload();});
const status=document.querySelector('#survey-status'),content=document.querySelector('#survey-content');
const showError=error=>{status.textContent=error.message||'No se pudo consultar la encuesta.';status.setAttribute('role','alert');};
async function start(){
 if(!/^[A-Za-z0-9_-]{43}$/.test(token)){showError(Error('Enlace de encuesta no válido. Solicita el enlace completo al taller.'));return;}
 try{await validateSurvey(token);}catch(error){showError(error);return;}
 status.textContent='0 de 5 preguntas respondidas';
 content.innerHTML=`<form id="survey-form">${surveyQuestions.map((q,i)=>`<fieldset><legend>${i+1}. ${e(q)}</legend><div class="survey-stars">${[1,2,3,4,5].map(n=>`<label><input type="radio" name="rating${i}" value="${n}" required aria-label="${n} ${n===1?'estrella':'estrellas'}"><span aria-hidden="true">★</span></label>`).join('')}</div><small>1: Muy insatisfecho · 5: Muy satisfecho</small></fieldset>`).join('')}<label class="survey-comment" for="comment">¿Tienes alguna sugerencia para mejorar?</label><textarea id="comment" name="comment" rows="4" maxlength="1200" placeholder="Comentario opcional"></textarea><button type="submit">Enviar encuesta</button></form>`;
 const form=content.querySelector('form');let sending=false;
 form.addEventListener('change',()=>{const values=new FormData(form);status.textContent=`${surveyQuestions.filter((_,i)=>values.has('rating'+i)).length} de 5 preguntas respondidas`;});
 form.addEventListener('submit',async event=>{
  event.preventDefault();if(sending)return;
  const data=new FormData(form),ratings=surveyQuestions.map((_,i)=>Number(data.get('rating'+i))),comment=data.get('comment').trim();
  try{validateSurveyAnswer(ratings,comment);}catch(error){showError(error);return;}
  sending=true;form.querySelector('button').disabled=true;form.setAttribute('aria-busy','true');status.textContent='Enviando tu opinión…';
  try{await answerSurvey(token,ratings,comment);content.replaceChildren();status.setAttribute('role','status');status.textContent='¡Gracias por compartir tu experiencia! Tu respuesta se guardó correctamente.';}
  catch(error){showError(error);if([404,409,410].includes(error.status)){content.replaceChildren();}else{sending=false;form.querySelector('button').disabled=false;form.removeAttribute('aria-busy');}}
 });
}
start();
