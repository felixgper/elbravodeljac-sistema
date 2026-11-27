import {request} from './api-service.js';

export const surveyQuestions=Object.freeze([
 '¿Cómo calificarías la atención recibida?',
 '¿Qué tan clara fue la explicación del trabajo realizado?',
 '¿Qué tan satisfecho estás con el tiempo de atención?',
 '¿Cómo calificarías nuestra comunicación durante el servicio?',
 'En general, ¿qué tan satisfecho estás con el servicio?'
]);
export const surveyAspectLabels=Object.freeze(['Atención','Claridad','Tiempo','Comunicación','Satisfacción general']);
export const surveyStatus=orderId=>request(`/ordenes/${encodeURIComponent(orderId)}/encuesta`);
export const issueSurvey=(orderId,regenerate=false)=>request(`/ordenes/${encodeURIComponent(orderId)}/encuesta`,{method:'POST',body:{regenerate}});
export const surveyStatistics=(start,end)=>request(`/encuestas/estadisticas?desde=${encodeURIComponent(start.toISOString())}&hasta=${encodeURIComponent(end.toISOString())}`);
export const surveyTracking=page=>request(`/encuestas?page=${page}`);
export const validateSurvey=token=>request('/encuestas/publica/validar',{method:'POST',body:{token}});
export const answerSurvey=(token,ratings,comment)=>request('/encuestas/publica/responder',{method:'POST',body:{token,ratings,comment}});

export function validateSurveyAnswer(ratings,comment=''){
 if(!Array.isArray(ratings)||ratings.length!==5||!ratings.every(n=>Number.isInteger(n)&&n>=1&&n<=5))throw Error('Responde las cinco preguntas con estrellas entre 1 y 5.');
 if([...comment].length>1200)throw Error('El comentario admite hasta 1200 caracteres.');
}
export function surveyWhatsApp(url,phone=''){
 // Do not guess country codes or use the recipient's contact without approval.
 const normalized=String(phone).replace(/[\s().-]/g,'');
 const target=/^\+[1-9]\d{7,14}$/.test(normalized)?normalized.slice(1):'';
 const message=`¡Gracias por confiar en El Bravo del JAC! 🚗\n\nNos gustaría conocer tu experiencia con nuestro servicio.\n\nResponde una breve encuesta de satisfacción ingresando al siguiente enlace:\n\n${url}\n\n¡Tu opinión nos ayuda a mejorar!`;
 return `https://wa.me/${target}?text=${encodeURIComponent(message)}`;
}
