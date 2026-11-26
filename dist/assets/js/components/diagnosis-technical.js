import {showModal,e,field,textarea,form,table,stamp,notify} from './ui.js';
import {permissionChecker} from '../services/permissions.js';
import {realSession} from '../services/api-service.js';
import {saveScan,listScans,uploadEvidence,listEvidence,evidenceURL,technicalAttempt} from '../services/technical-service.js';

const drafts=new Map();let draftOwner='';
export function scanInput(equipment,version,codes,measurements){
 const dtcs=codes.filter(r=>r.code.trim()||r.description.trim()).map(r=>({codigo:r.code.trim(),descripcion:r.description.trim()}));
 if(!equipment.trim())throw Error('Indica el nombre del equipo.');
 if(dtcs.some(r=>!r.codigo))throw Error('Escribe el código de falla junto a su descripción.');
 const values=Object.create(null);
 for(const row of measurements){
  const key=row.name.trim(),value=row.value.trim();if(!key&&!value)continue;
  if(!key||!value||key.length>100||key.startsWith('$')||/[.\0]/.test(key)||Object.hasOwn(values,key))throw Error('Cada medición necesita un nombre único y un valor. No uses puntos ni signos $ en el nombre.');
  values[key]=/^-?\d+(\.\d+)?$/.test(value)&&Number.isFinite(Number(value))?Number(value):value;
 }
 return {equipo:{nombre:equipment.trim(),version:version.trim()},dtcs,mediciones:values};
}

export function openDiagnosisTechnical({diagnosisId,closed=false,onBusy=()=>{}}){
 const user=realSession();if(!user||!diagnosisId)return;
 const can=permissionChecker(user.role);if(!can('diagnosticos','ver'))return;
 if(draftOwner!==user.id){drafts.clear();draftOwner=user.id;}
 const key=user.id+':'+diagnosisId;
 let state=drafts.get(key);if(!state){state={tab:'evidence',file:null,description:'',equipment:'',version:'',codes:[{code:'',description:''}],measurements:[],evidence:[],scans:[],busy:false};drafts.set(key,state);}
 const writable=()=>!closed&&can('diagnosticos','crear');
 const evidence=technicalAttempt('evidence',diagnosisId,user.id),scan=technicalAttempt('scan',diagnosisId,user.id);
 const dialog=showModal('Evidencias y escáner',`<p class="dialog-intro">Información complementaria y opcional del diagnóstico. ${closed?'Diagnóstico finalizado: solo consulta.':'Puedes continuar tu diagnóstico sin adjuntar archivos ni registrar códigos.'}</p><div class="tabs technical-tabs" role="tablist" aria-label="Información técnica"><button type="button" role="tab" id="technical-evidence-tab" aria-controls="technical-evidence">Fotografías</button><button type="button" role="tab" id="technical-scan-tab" aria-controls="technical-scan">Escáner</button></div><p class="technical-status" role="status" aria-live="polite"></p><section id="technical-evidence" role="tabpanel" aria-labelledby="technical-evidence-tab"></section><section id="technical-scan" role="tabpanel" aria-labelledby="technical-scan-tab"></section>`,{onClose:()=>{onBusy(false);}});
 dialog.classList.add('diagnosis-technical-dialog');
 const current=()=>dialog.isConnected&&realSession()?.id===user.id;
 const message=(text,error=false)=>{if(!current())return;const n=dialog.querySelector('.technical-status');n.textContent=text;n.setAttribute('role',error?'alert':'status');};
 const toggleBusy=value=>{state.busy=value;onBusy(value);dialog.querySelectorAll('button,input,textarea,a').forEach(n=>{if(n.tagName==='A'){n.setAttribute('aria-disabled',String(value));return;}n.disabled=value;});dialog.setAttribute('aria-busy',String(value));};
 dialog.addEventListener('cancel',event=>{if(state.busy){event.preventDefault();message('Espera a que termine el envío antes de cerrar.');}});
 dialog.addEventListener('click',event=>{if(state.busy&&(event.target.closest('.dialog-heading button')||event.target===dialog)){event.preventDefault();event.stopImmediatePropagation();message('Espera a que termine el envío antes de cerrar.');}},true);
 function selectTab(tab){if(state.busy)return;state.tab=tab;for(const [id,name] of [['evidence','Evidencias'],['scan','Escáner']]){const b=dialog.querySelector('#technical-'+id+'-tab');b.setAttribute('aria-selected',String(id===tab));b.tabIndex=id===tab?0:-1;dialog.querySelector('#technical-'+id).hidden=id!==tab;}}
 for(const id of ['evidence','scan'])dialog.querySelector('#technical-'+id+'-tab').onclick=()=>selectTab(id);
 dialog.querySelector('.technical-tabs').onkeydown=event=>{if(['ArrowLeft','ArrowRight','Home','End'].includes(event.key)){event.preventDefault();selectTab(event.key==='Home'?'evidence':event.key==='End'?'scan':state.tab==='scan'?'evidence':'scan');dialog.querySelector('#technical-'+state.tab+'-tab').focus();}};
 async function refresh(){
  const [files,scans]=await Promise.all([listEvidence(diagnosisId),listScans(diagnosisId)]);if(!current())return;
  state.evidence=files;state.scans=scans;renderLists();
  const fileConfirmed=evidence.confirm(files),scanConfirmed=scan.confirm(scans);
  if(fileConfirmed){state.file=null;state.description='';if(ef){ef.reset();ef.elements.file.required=true;}}
  if(fileConfirmed||scanConfirmed)message('Se encontró el registro enviado anteriormente. Ya está guardado.');
 }
 function renderLists(){
  dialog.querySelector('.technical-evidence-list').innerHTML=table(['Descripción','Formato y fecha','Archivo'],state.evidence.map(v=>[e(v.descripcion||'Sin descripción'),e(v.mimeType)+'<small class="cell-note">'+e(stamp(v.capturadoEn||v.registradoEn))+'</small>',`<a class="secondary-button" href="${e(evidenceURL(diagnosisId,v.id))}" target="_blank" rel="noopener noreferrer">Descargar archivo</a>`]),'Todavía no hay evidencias guardadas.');
  dialog.querySelector('.technical-scan-list').innerHTML=state.scans.length?state.scans.map(v=>`<details class="technical-record"><summary>${e(v.equipo.nombre)}${v.equipo.version?' · '+e(v.equipo.version):''} · ${e(stamp(v.capturadoEn||v.registradoEn))}</summary>${table(['Código','Descripción'],v.dtcs.map(d=>[e(d.codigo),e(d.descripcion||'Sin descripción')]),'Sin códigos registrados.')}${table(['Medición','Valor'],Object.entries(v.mediciones||{}).map(([k,value])=>[e(k),e(typeof value==='object'?JSON.stringify(value):value)]),'Sin mediciones registradas.')}</details>`).join(''):'<p class="empty-table">Todavía no hay escaneos guardados.</p>';
 }
 dialog.querySelector('#technical-evidence').innerHTML=(writable()?form('ADJUNTAR ARCHIVO',field('Fotografía o documento','file','','file','accept="image/jpeg,image/png,application/pdf" required')+textarea('Descripción (opcional)','description',state.description,'maxlength="4000"')+'<p class="form-description full-width">JPG, PNG o PDF. Máximo 10 MiB por archivo.</p>','<button class="primary-button" type="submit">Subir archivo</button><button class="secondary-button technical-new" type="button">Preparar otra carga</button>','technical-evidence-form'):'')+'<button type="button" class="secondary-button technical-refresh">Actualizar registros</button><div class="technical-evidence-list" aria-live="polite">Consultando evidencias…</div>';
 dialog.querySelector('#technical-scan').innerHTML=(writable()?form('REGISTRO MANUAL DEL ESCÁNER',field('Nombre del equipo','equipment',state.equipment,'text','required maxlength="120"')+field('Versión (opcional)','version',state.version,'text','maxlength="80"')+'<div class="full-width"><h3>Códigos de falla</h3><div class="technical-codes"></div><button class="secondary-button" type="button" data-add="code">Agregar código</button></div><div class="full-width"><h3>Mediciones (opcional)</h3><div class="technical-measurements"></div><button class="secondary-button" type="button" data-add="measurement">Agregar medición</button></div>','<button class="primary-button" type="submit">Guardar escaneo</button><button class="secondary-button technical-new" type="button">Preparar otro escaneo</button>','technical-scan-form'):'')+'<div class="technical-scan-list" aria-live="polite">Consultando escaneos…</div>';
 function rows(){
  for(const [kind,items] of [['code',state.codes],['measurement',state.measurements]]){
   const target=dialog.querySelector(kind==='code'?'.technical-codes':'.technical-measurements');if(!target)continue;
   target.innerHTML=items.map((v,i)=>`<div class="technical-row">${kind==='code'?field('Código de falla DTC','code'+i,v.code,'text','maxlength="30"')+field('Descripción del código','description'+i,v.description,'text','maxlength="1000"'):field('Medición','measurement'+i,v.name,'text','maxlength="100"')+field('Valor','value'+i,v.value,'text','maxlength="4000"')}<button type="button" class="secondary-button" data-remove="${kind}" data-index="${i}" aria-label="Quitar ${kind==='code'?'código':'medición'} ${i+1}">Quitar</button></div>`).join('');
   target.querySelectorAll('input').forEach(input=>input.oninput=()=>{target.querySelectorAll('.technical-row').forEach((row,i)=>{const f=row.querySelectorAll('input');Object.assign(items[i],kind==='code'?{code:f[0].value,description:f[1].value}:{name:f[0].value,value:f[1].value});});});
   target.querySelectorAll('[data-remove]').forEach(b=>b.onclick=()=>{items.splice(Number(b.dataset.index),1);rows();});
  }
 }
 rows();
 dialog.querySelectorAll('[data-add]').forEach(b=>b.onclick=()=>{const list=b.dataset.add==='code'?state.codes:state.measurements;if(list.length>=(b.dataset.add==='code'?1000:100)){message('Se alcanzó el máximo de filas.',true);return;}list.push(b.dataset.add==='code'?{code:'',description:''}:{name:'',value:''});rows();});
 dialog.querySelectorAll('.technical-refresh').forEach(b=>b.onclick=()=>refresh().catch(e=>message(e.message,true)));
 async function send(attempt,payload,operation,success){
  if(state.busy||!current())return;toggleBusy(true);message('Guardando…');
  try{await attempt.send(payload,operation);if(!current())return;success();message('Registro guardado correctamente.');notify('Información técnica guardada.');try{await refresh();}catch{message('El registro se guardó. Usa Actualizar registros para consultar la lista.',true);}}
  catch(error){message(error.message,true);if(error.fields?.diagnosisId){closed=true;dialog.querySelectorAll('form').forEach(f=>f.hidden=true);}}
  finally{state.busy=false;onBusy(false);if(current())toggleBusy(false);}
 }
 const ef=dialog.querySelector('#technical-evidence-form'),sf=dialog.querySelector('#technical-scan-form');
 if(ef){ef.elements.file.required=!state.file;ef.elements.file.onchange=()=>{state.file=ef.elements.file.files[0]||null;ef.elements.file.required=!state.file;};ef.elements.description.oninput=()=>state.description=ef.elements.description.value;
  if(state.file)message('La selección anterior se conserva. Puedes reintentar la misma carga.');
  ef.onsubmit=event=>{event.preventDefault();const file=state.file;if(!file){message('Selecciona un archivo para subir.',true);return;}send(evidence,{file,description:state.description},v=>uploadEvidence(diagnosisId,v),()=>{state.file=null;state.description='';ef.reset();ef.elements.file.required=true;});};
 }
 if(sf){sf.elements.equipment.oninput=()=>state.equipment=sf.elements.equipment.value;sf.elements.version.oninput=()=>state.version=sf.elements.version.value;
  sf.onsubmit=event=>{event.preventDefault();let input;try{input=scanInput(state.equipment,state.version,state.codes,state.measurements);}catch(error){message(error.message,true);return;}send(scan,input,v=>saveScan(diagnosisId,v),()=>{});};
 }
 dialog.querySelectorAll('.technical-new').forEach(b=>b.onclick=()=>{const attempt=b.closest('form')===ef?evidence:scan;if(attempt.pending&&!globalThis.confirm('El envío anterior podría haberse guardado. Actualiza los registros antes de preparar otro. ¿Deseas preparar una nueva solicitud?'))return;attempt.reset();message('Puedes preparar un nuevo registro.');});
 selectTab(state.tab);refresh().catch(error=>message(error.message,true));return dialog;
}
