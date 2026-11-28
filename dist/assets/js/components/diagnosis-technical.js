import {showModal,e,field,textarea,form,table,stamp,notify} from './ui.js';
import {icon} from './icons.js';
import {permissionChecker} from '../services/permissions.js';
import {realSession} from '../services/api-service.js';
import {saveScan,listScans,uploadEvidence,listEvidence,evidenceURL,loadEvidence,technicalAttempt} from '../services/technical-service.js';

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

export function manualScanInput(codes){
 const values=codes.map(r=>r.code.trim().toUpperCase()).filter(Boolean);
 if(!values.length)throw Error('Agrega al menos un código de falla DTC.');
 if(values.some(v=>! /^[PCBU][0-3][0-9A-F]{3}(?:[-:][0-9A-F]{2})?$/.test(v)))throw Error('Usa un código DTC válido, como P0300 o U0100.');
 if(new Set(values).size!==values.length)throw Error('No repitas el mismo código de falla.');
 return scanInput('Registro manual','',values.map(code=>({code,description:''})),[]);
}

export function openDiagnosisTechnical({diagnosisId,draftKey=diagnosisId,ensureDiagnosis,closed=false,onBusy=()=>{}}){
 const user=realSession();if(!user||(!diagnosisId&&!ensureDiagnosis))return;
 const can=permissionChecker(user.role);if(!can('diagnosticos','ver'))return;
 if(draftOwner!==user.id){drafts.clear();draftOwner=user.id;}
 const key=user.id+':'+draftKey;
 let state=drafts.get(key);if(!state){state={tab:'evidence',file:null,description:'',codes:[{code:''}],evidence:[],scans:[],busy:false};drafts.set(key,state);}
 const writable=()=>!closed&&can('diagnosticos','crear');let evidence,scan,refreshing,viewer,previewing=false;
 const controller=new AbortController();
 const bindAttempts=()=>{evidence=technicalAttempt('evidence',diagnosisId,user.id);scan=technicalAttempt('scan',diagnosisId,user.id);};
 if(diagnosisId)bindAttempts();
 const dialog=showModal('Evidencias y escáner',`<p class="dialog-intro">Información complementaria y opcional del diagnóstico. ${closed?'Diagnóstico finalizado: solo consulta.':'Puedes continuar tu diagnóstico sin adjuntar archivos ni registrar códigos.'}</p><div class="tabs technical-tabs" role="tablist" aria-label="Información técnica"><button type="button" role="tab" id="technical-evidence-tab" aria-controls="technical-evidence">Fotografías</button><button type="button" role="tab" id="technical-scan-tab" aria-controls="technical-scan">Escáner</button></div><p class="technical-status" role="status" aria-live="polite"></p><section id="technical-evidence" role="tabpanel" aria-labelledby="technical-evidence-tab"></section><section id="technical-scan" role="tabpanel" aria-labelledby="technical-scan-tab"></section>`,{onClose:()=>{controller.abort();viewer?.close();onBusy(false);}});
 dialog.classList.add('diagnosis-technical-dialog');
 const current=()=>dialog.isConnected&&realSession()?.id===user.id;
 const message=(text,error=false)=>{if(!current())return;const n=dialog.querySelector('.technical-status');n.textContent=text;n.setAttribute('role',error?'alert':'status');};
 const toggleBusy=value=>{state.busy=value;onBusy(value);dialog.querySelectorAll('button,input,textarea,a').forEach(n=>{if(n.tagName==='A'){n.setAttribute('aria-disabled',String(value));return;}n.disabled=value;});dialog.setAttribute('aria-busy',String(value));};
 dialog.addEventListener('cancel',event=>{if(state.busy){event.preventDefault();message('Espera a que termine el envío antes de cerrar.');}});
 dialog.addEventListener('click',event=>{if(state.busy&&(event.target.closest('.dialog-heading button')||event.target===dialog)){event.preventDefault();event.stopImmediatePropagation();message('Espera a que termine el envío antes de cerrar.');}},true);
 function selectTab(tab){if(state.busy)return;state.tab=tab;for(const id of ['evidence','scan']){const b=dialog.querySelector('#technical-'+id+'-tab');b.setAttribute('aria-selected',String(id===tab));b.tabIndex=id===tab?0:-1;dialog.querySelector('#technical-'+id).hidden=id!==tab;}}
 for(const id of ['evidence','scan'])dialog.querySelector('#technical-'+id+'-tab').onclick=()=>selectTab(id);
 dialog.querySelector('.technical-tabs').onkeydown=event=>{if(['ArrowLeft','ArrowRight','Home','End'].includes(event.key)){event.preventDefault();selectTab(event.key==='Home'?'evidence':event.key==='End'?'scan':state.tab==='scan'?'evidence':'scan');dialog.querySelector('#technical-'+state.tab+'-tab').focus();}};
 async function refresh(announce=false){
  if(refreshing)return refreshing;
  if(!diagnosisId){renderLists();if(announce)message('No hay registros guardados para este diagnóstico.');return;}
  dialog.querySelectorAll('.technical-refresh').forEach(b=>{b.disabled=true;b.setAttribute('aria-busy','true');});
  refreshing=(async()=>{
   const [files,scans]=await Promise.all([listEvidence(diagnosisId),listScans(diagnosisId)]);if(!current())return;
   state.evidence=files;state.scans=scans;renderLists();
   const fileConfirmed=evidence.confirm(files),scanConfirmed=scan.confirm(scans);
   if(fileConfirmed){state.file=null;state.description='';if(ef){ef.reset();ef.elements.file.required=true;}}
   if(scanConfirmed){state.codes=[{code:''}];rows();}
   if(announce||fileConfirmed||scanConfirmed)message('Registros actualizados.'+(fileConfirmed||scanConfirmed?' Se confirmó el envío anterior.':''));
  })();
  try{await refreshing;}finally{refreshing=null;if(current())dialog.querySelectorAll('.technical-refresh').forEach(b=>{b.disabled=state.busy;b.removeAttribute('aria-busy');});}
 }
 async function preview(id){
  if(previewing||!current()||state.busy)return;previewing=true;message('Abriendo archivo…');
  try{
   const blob=await loadEvidence(diagnosisId,id,{signal:controller.signal});if(!current())return;
   const url=URL.createObjectURL(blob);viewer=document.createElement('dialog');viewer.className='workshop-dialog technical-preview';viewer.setAttribute('aria-label','Vista del archivo');
   viewer.innerHTML=`<div class="dialog-heading"><h2>Vista del archivo</h2><button type="button" class="icon-button" aria-label="Cerrar vista">×</button></div><div class="technical-preview-body">${blob.type==='application/pdf'?'<iframe title="Documento PDF"></iframe>':'<img alt="Evidencia del diagnóstico">'}</div>`;
   document.body.append(viewer);const media=viewer.querySelector('img,iframe');media.src=url;
   viewer.addEventListener('close',()=>{media.removeAttribute('src');URL.revokeObjectURL(url);viewer.remove();viewer=null;},{once:true});
   viewer.querySelector('button').onclick=()=>viewer.close();viewer.showModal();message('');
  }catch(error){if(error.name!=='AbortError')message(error.message,true);}finally{previewing=false;}
 }
 function renderLists(){
  dialog.querySelector('.technical-evidence-list').innerHTML=state.evidence.length?`<div class="technical-gallery">${state.evidence.map(v=>{const image=['image/jpeg','image/png'].includes(v.mimeType);return `<article class="technical-evidence-card"><button type="button" class="technical-thumbnail" data-preview="${e(v.id)}" aria-label="Abrir ${image?'fotografía':'documento PDF'}">${image?`<img src="${e(evidenceURL(diagnosisId,v.id))}" loading="lazy" alt="${e(v.descripcion||'Fotografía del diagnóstico')}">`:icon('quote')+'<span>PDF</span>'}</button><p>${e(v.descripcion||'Sin descripción')}</p><time>${e(stamp(v.capturadoEn||v.registradoEn))}</time><div><button type="button" class="technical-open" data-preview="${e(v.id)}">Abrir</button><a class="technical-download" href="${e(evidenceURL(diagnosisId,v.id))}" download>Descargar</a></div></article>`;}).join('')}</div>`:'<p class="empty-table">Todavía no hay evidencias guardadas.</p>';
  dialog.querySelectorAll('[data-preview]').forEach(b=>b.onclick=()=>preview(b.dataset.preview));
  dialog.querySelectorAll('.technical-thumbnail img').forEach(img=>img.onerror=()=>{if(!img.isConnected||!img.parentNode)return;img.hidden=true;const label=document.createElement('span');label.textContent='Abrir fotografía';img.parentNode.append(label);});
  dialog.querySelector('.technical-scan-list').innerHTML=state.scans.length?state.scans.map(v=>`<details class="technical-record"><summary>${e(v.equipo.nombre)}${v.equipo.version?' · '+e(v.equipo.version):''} · ${e(stamp(v.capturadoEn||v.registradoEn))}</summary>${table(['Código','Descripción'],v.dtcs.map(d=>[e(d.codigo),e(d.descripcion||'—')]),'Sin códigos registrados.')}${Object.keys(v.mediciones||{}).length?table(['Medición','Valor'],Object.entries(v.mediciones).map(([k,value])=>[e(k),e(typeof value==='object'?JSON.stringify(value):value)])):''}</details>`).join(''):'<p class="empty-table">Todavía no hay escaneos guardados.</p>';
 }
 dialog.querySelector('#technical-evidence').innerHTML=(writable()?form('ADJUNTAR ARCHIVO',field('Fotografía o documento','file','','file','accept="image/jpeg,image/png,application/pdf" required')+textarea('Descripción (opcional)','description',state.description,'maxlength="4000"')+'<p class="form-description full-width">JPG, PNG o PDF. Máximo 10 MiB por archivo.</p>','<button class="primary-button" type="submit">Subir archivo</button><button class="secondary-button technical-new" type="button">Preparar otra carga</button>','technical-evidence-form'):'')+'<button type="button" class="secondary-button technical-refresh">Actualizar registros</button><div class="technical-evidence-list" aria-live="polite">Consultando evidencias…</div>';
 dialog.querySelector('#technical-scan').innerHTML=(writable()?form('CÓDIGOS DEL ESCÁNER','<div class="full-width"><div class="technical-codes"></div><button class="secondary-button" type="button" data-add="code">Agregar código</button></div>','<button class="primary-button" type="submit">Guardar escaneo</button><button class="secondary-button technical-new" type="button">Preparar otro escaneo</button>','technical-scan-form'):'')+'<button type="button" class="secondary-button technical-refresh">Actualizar registros</button><div class="technical-scan-list" aria-live="polite">Consultando escaneos…</div>';
 function rows(){
  const target=dialog.querySelector('.technical-codes');if(!target)return;
  target.innerHTML=state.codes.map((v,i)=>`<div class="technical-row technical-dtc-row">${field('Código de falla DTC','code'+i,v.code,'text','maxlength="30" placeholder="Ej. P0300" autocapitalize="characters"')}<button type="button" class="secondary-button" data-remove="${i}" aria-label="Quitar código ${i+1}">Quitar</button></div>`).join('');
  target.querySelectorAll('input').forEach((input,i)=>input.oninput=()=>state.codes[i].code=input.value);
  target.querySelectorAll('[data-remove]').forEach(b=>b.onclick=()=>{state.codes.splice(Number(b.dataset.remove),1);rows();});
 }
 rows();dialog.querySelector('[data-add="code"]')?.addEventListener('click',()=>{if(state.codes.length>=1000){message('Se alcanzó el máximo de códigos.',true);return;}state.codes.push({code:''});rows();});
 dialog.querySelectorAll('.technical-refresh').forEach(b=>b.onclick=()=>refresh(true).catch(error=>message(error.message,true)));
 async function send(kind,payload,operation,success){
  if(state.busy||!current())return;let automaticDraft=false;toggleBusy(true);message('Guardando…');
  try{
   if(!diagnosisId){const record=await ensureDiagnosis();if(!current())return;diagnosisId=record.id;closed=record.status==='Finalizado';automaticDraft=!closed;bindAttempts();if(closed){dialog.querySelectorAll('form').forEach(f=>f.hidden=true);throw Error('El diagnóstico ya está finalizado. Los anexos son de solo consulta.');}}
   await (kind==='evidence'?evidence:scan).send(payload,operation);if(!current())return;success();state.busy=false;dialog.close();queueMicrotask(()=>notify(kind==='evidence'?'Archivo guardado correctamente':'Escaneo guardado correctamente'));
  }catch(error){message((automaticDraft?'El borrador quedó guardado. Puedes reintentar el anexo. ':'')+error.message,true);if(error.fields?.diagnosisId){closed=true;dialog.querySelectorAll('form').forEach(f=>f.hidden=true);}}
  finally{state.busy=false;onBusy(false);if(current())toggleBusy(false);}
 }
 const ef=dialog.querySelector('#technical-evidence-form'),sf=dialog.querySelector('#technical-scan-form');
 if(ef){ef.elements.file.required=!state.file;ef.elements.file.onchange=()=>{state.file=ef.elements.file.files[0]||null;ef.elements.file.required=!state.file;};ef.elements.description.oninput=()=>state.description=ef.elements.description.value;
  if(state.file)message('La selección anterior se conserva. Puedes reintentar la misma carga.');
  ef.onsubmit=event=>{event.preventDefault();const file=state.file;if(!file){message('Selecciona un archivo para subir.',true);return;}send('evidence',{file,description:state.description},v=>uploadEvidence(diagnosisId,v),()=>{state.file=null;state.description='';ef.reset();ef.elements.file.required=true;});};
 }
 if(sf)sf.onsubmit=event=>{event.preventDefault();let input;try{input=manualScanInput(state.codes);}catch(error){message(error.message,true);return;}send('scan',input,v=>saveScan(diagnosisId,v),()=>{state.codes=[{code:''}];});};
 dialog.querySelectorAll('.technical-new').forEach(b=>b.onclick=()=>{const attempt=b.closest('form')===ef?evidence:scan;if(attempt?.pending&&!globalThis.confirm('El envío anterior podría haberse guardado. Actualiza los registros antes de preparar otro. ¿Deseas preparar una nueva solicitud?'))return;attempt?.reset();message('Puedes preparar un nuevo registro.');});
 selectTab(state.tab);refresh().catch(error=>message(error.message,true));return dialog;
}
