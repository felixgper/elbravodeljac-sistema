import {realMode,realSession} from '../services/api-service.js';
import {icon} from './icons.js';
export const e=value=>String(value??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
export const money=cents=>new Intl.NumberFormat('es-PE',{style:'currency',currency:'PEN'}).format((cents||0)/100);
export const date=value=>value?new Intl.DateTimeFormat('es-PE',{dateStyle:'medium',timeZone:'America/Lima'}).format(new Date(value)):'—';
export const stamp=value=>value?new Intl.DateTimeFormat('es-PE',{dateStyle:'short',timeStyle:'short',timeZone:'America/Lima'}).format(new Date(value)):'—';
export const norm=value=>String(value??'').normalize('NFD').replace(/[\u0300-\u036f]/g,'').toLowerCase();
export function go(path){const hash='#/'+path;if(location.hash===hash)window.dispatchEvent(new Event('hashchange'));else location.hash=hash;}
export function link(label,path,primary=false){const text=String(label).startsWith('<svg')?label:e(label);return `<a class="${primary?'primary-button':'secondary-button'}" href="#/${e(path)}">${text}</a>`;}
export function button(label,action,id='',kind='secondary'){return `<button type="button" class="${kind}-button" data-action="${e(action)}" data-id="${e(id)}">${e(label)}</button>`;}
let seq=0;
export function field(label,name,value='',type='text',attrs=''){const id='field-'+(++seq);return `<div class="form-field"><label for="${id}">${e(label)}</label><input id="${id}" name="${e(name)}" type="${type}" value="${e(value)}" ${attrs}></div>`;}
export function textarea(label,name,value='',attrs=''){const id='field-'+(++seq);return `<div class="form-field full-width"><label for="${id}">${e(label)}</label><textarea id="${id}" name="${e(name)}" rows="4" ${attrs}>${e(value)}</textarea></div>`;}
export function select(label,name,options,value='',attrs=''){const id='field-'+(++seq);return `<div class="form-field"><label for="${id}">${e(label)}</label><select id="${id}" name="${e(name)}" ${attrs}>${options.map(o=>{const v=typeof o==='object'?o.value:o,t=typeof o==='object'?o.label:o;return `<option value="${e(v)}" ${String(value)===String(v)?'selected':''}>${e(t)}</option>`;}).join('')}</select></div>`;}
export const placeholder=label=>({value:'',label:label||'Seleccione…'});
export function checkbox(label,name,checked=false,attrs=''){const id='field-'+(++seq);return `<label class="form-check full-width" for="${id}"><input id="${id}" type="checkbox" name="${e(name)}" ${checked?'checked':''} ${attrs}><span>${e(label)}</span></label>`;}
export function badge(state){const color={'Pendiente':'amber','En reparación':'copper','En pruebas':'blue','Finalizada':'green','Aprobada':'green','Rechazada':'red','Cancelada':'red','Activo':'green','Inactivo':'gray','Entregado':'gray'}[state]||'gray';return `<span class="status-badge status-${color}">${e(state)}</span>`;}
export function table(headers,rows,empty='No hay registros para mostrar.'){return `<div class="table-scroll" tabindex="0" role="region" aria-label="Tabla de ${e(headers.slice(0,2).join(' y '))}"><table class="data-table"><thead><tr>${headers.map(h=>`<th scope="col">${e(h)}</th>`).join('')}</tr></thead><tbody>${rows.length?rows.map(cells=>`<tr>${cells.map(cell=>`<td>${cell}</td>`).join('')}</tr>`).join(''):`<tr><td colspan="${headers.length}" class="empty-table">${e(empty)}</td></tr>`}</tbody></table></div>`;}
export function pageHeader(title,subtitle,actions=''){return `<div class="module-heading"><div><h2>${e(title)}</h2><p>${e(subtitle)}</p></div><div class="toolbar-actions">${actions}</div></div>`;}
export function hint(message){return `<div class="inline-note">${icon('info')}<span>${e(message)}</span></div>`;}
export function form(title,body,actions,id='edit-form'){return `<section class="panel form-panel"><h2 class="form-section-title">${e(title)}</h2><form id="${id}" class="workshop-form"><div class="form-grid">${body}</div><p class="form-error" role="alert" hidden></p><div class="form-actions">${actions}</div></form></section>`;}
export const submit=(label='Guardar',value='save')=>`<button type="submit" class="primary-button" value="${e(value)}">${e(label)}</button>`;
export function values(form){const result=Object.fromEntries(new FormData(form));form.querySelectorAll('input[type="checkbox"]').forEach(i=>result[i.name]=i.checked);return result;}
const FIELD_HINTS=[[/contraseña/i,'password'],[/correo/i,'email'],[/usuario/i,'username'],[/dni|ruc|documento/i,'documentNumber'],[/tel[eé]fono/i,'phone'],[/placa/i,'plate'],[/\baño\b/i,'year'],[/kilometraje/i,'mileage'],[/monto/i,'amount']];
function fieldFor(form,message){for(const [pattern,name] of FIELD_HINTS){if(pattern.test(message)){const field=form.querySelector(`[name="${name}"]:not([disabled]):not([readonly])`);if(field)return field;}}return null;}
export function bindForm(id,handler){const el=document.getElementById(id);const error=el.querySelector('.form-error');error.id=error.id||'error-'+id;el.addEventListener('input',ev=>{if(ev.target.removeAttribute){ev.target.removeAttribute('aria-invalid');ev.target.removeAttribute('aria-describedby');}});el.addEventListener('submit',async event=>{event.preventDefault();el.querySelectorAll('[aria-invalid]').forEach(i=>i.removeAttribute('aria-invalid'));error.hidden=true;const buttons=[...el.querySelectorAll('[type="submit"]')];buttons.forEach(b=>b.disabled=true);try{await handler(values(el),event.submitter?.value,el);}catch(reason){error.textContent=reason.message||'No se pudo guardar el registro.';error.hidden=false;const field=fieldFor(el,error.textContent);if(field){field.setAttribute('aria-invalid','true');field.setAttribute('aria-describedby',error.id);field.focus();}else{error.tabIndex=-1;error.focus();}error.scrollIntoView({block:'nearest'});}finally{buttons.forEach(b=>b.disabled=false);}});}
let toastTimer;
export function notify(message){let toast=document.getElementById('toast');if(!toast){toast=document.createElement('div');toast.id='toast';toast.setAttribute('role','status');toast.setAttribute('aria-live','polite');document.body.appendChild(toast);}toast.textContent=message;toast.className='toast visible';clearTimeout(toastTimer);toastTimer=setTimeout(()=>toast.className='toast',4500);}
export function showModal(title,body,{onClose}={}){document.getElementById('workshop-dialog')?.remove();const dialog=document.createElement('dialog');dialog.id='workshop-dialog';dialog.className='workshop-dialog';dialog.setAttribute('aria-labelledby','modal-heading');dialog.innerHTML=`<div class="dialog-heading"><h2 id="modal-heading">${e(title)}</h2><button type="button" class="icon-button" aria-label="Cerrar ventana">${icon('close')}</button></div><div class="dialog-content">${body}</div>`;document.body.appendChild(dialog);dialog.querySelector('.dialog-heading button').onclick=()=>dialog.close();dialog.addEventListener('click',ev=>{if(ev.target===dialog){const r=dialog.getBoundingClientRect();if(ev.clientX<r.left||ev.clientX>r.right||ev.clientY<r.top||ev.clientY>r.bottom)dialog.close();}});dialog.addEventListener('close',()=>{dialog.remove();onClose?.();});dialog.showModal();return dialog;}
export function searchableList(root,{title,subtitle,actions,placeholder:searchPlaceholder,headers,records,searchText,row,empty,onAction,pageSize=10}){
 root.innerHTML=pageHeader(title,subtitle,actions)+`<section class="panel list-panel"><div class="list-toolbar"><label class="list-search">${icon('search')}<input type="search" placeholder="${e(searchPlaceholder)}" aria-label="${e(searchPlaceholder)}"></label><span class="list-count"></span></div><div class="list-results"></div><div class="pagination"></div></section>`;
 let query='',page=1;const result=root.querySelector('.list-results'),pagination=root.querySelector('.pagination');
 const draw=()=>{const found=records.filter(record=>norm(searchText(record)).includes(norm(query))),pages=Math.max(1,Math.ceil(found.length/pageSize));page=Math.min(page,pages);root.querySelector('.list-count').textContent=`${found.length} registro${found.length===1?'':'s'}`;result.innerHTML=table(headers,found.slice((page-1)*pageSize,page*pageSize).map(row),empty);pagination.innerHTML=`<span>Página ${page} de ${pages}</span><div><button type="button" class="secondary-button" data-page="prev" ${page===1?'disabled':''}>Anterior</button><button type="button" class="secondary-button" data-page="next" ${page===pages?'disabled':''}>Siguiente</button></div>`;};
 root.querySelector('input[type="search"]').addEventListener('input',ev=>{query=ev.target.value;page=1;draw();});
 pagination.addEventListener('click',ev=>{const b=ev.target.closest('[data-page]');if(b&&!b.disabled){page+=b.dataset.page==='next'?1:-1;draw();}});
 result.addEventListener('click',ev=>{const b=ev.target.closest('[data-action]');if(b)onAction?.(b.dataset.action,b.dataset.id,b);});draw();
}
export function selectOptions(records,label){return [placeholder(),...records.map(r=>({value:r.id,label:label(r)}))];}
export function readDraft(key){try{return JSON.parse(sessionStorage.getItem('bravo-jac:draft:'+(realMode?'real:'+realSession()?.id+':':'demo:')+key))||{};}catch{return {};}}
export function bindDraft(id,key){const el=document.getElementById(id);el.addEventListener('input',()=>{try{sessionStorage.setItem('bravo-jac:draft:'+(realMode?'real:'+realSession()?.id+':':'demo:')+key,JSON.stringify(values(el)));}catch{}});}
export function clearDraft(key){sessionStorage.removeItem('bravo-jac:draft:'+(realMode?'real:'+realSession()?.id+':':'demo:')+key);}

// ── Acciones de fila: lápiz (editar) y tacho (eliminar) ──────────────────────
export function rowActions({editHref='',editAction='',id='',name='registro',extra='',canEdit=true,canDelete=true}={}){
 const label=e(name);
 const edit=editHref
  ?`<a class="row-icon row-edit" href="#/${e(editHref)}" title="Editar" aria-label="Editar ${label}">${icon('pencil')}</a>`
  :`<button type="button" class="row-icon row-edit" data-action="${e(editAction||'edit')}" data-id="${e(id)}" title="Editar" aria-label="Editar ${label}">${icon('pencil')}</button>`;
 const del=`<button type="button" class="row-icon row-delete" data-action="delete" data-id="${e(id)}" title="Eliminar" aria-label="Eliminar ${label}">${icon('trash')}</button>`;
 return `<div class="row-actions">${extra}${canEdit?edit:''}${canDelete?del:''}</div>`;
}
// Ventana de confirmación reutilizable para cualquier eliminación.
export function confirmDelete({title,message,detail='',confirmLabel='Sí, eliminar',onConfirm}){
 const dialog=showModal(title,`<div class="confirm-delete"><span class="confirm-icon">${icon('trash')}</span><div><p class="confirm-message">${message}</p>${detail?`<p class="confirm-detail">${e(detail)}</p>`:''}<p class="confirm-warning">${icon('warning')}<span>Esta acción no se puede deshacer.</span></p></div></div><p class="form-error" role="alert" hidden></p><div class="form-actions"><button type="button" class="secondary-button" data-confirm="cancel">Cancelar</button><button type="button" class="danger-button" data-confirm="ok">${icon('trash')} ${e(confirmLabel)}</button></div>`);
 const error=dialog.querySelector('.form-error'),ok=dialog.querySelector('[data-confirm="ok"]');
 dialog.querySelector('[data-confirm="cancel"]').onclick=()=>dialog.close();
 ok.onclick=async()=>{error.hidden=true;ok.disabled=true;try{await onConfirm();dialog.close();}catch(reason){error.textContent=reason.message||'No se pudo eliminar el registro.';error.hidden=false;ok.disabled=false;}};
 dialog.querySelector('[data-confirm="cancel"]').focus();
 return dialog;
}
export const initials=name=>String(name??'').trim().split(/\s+/).slice(0,2).map(w=>w[0]||'').join('').toUpperCase()||'?';
export const personCell=(name,sub='')=>`<div class="person-cell"><span class="avatar-initials" aria-hidden="true">${e(initials(name))}</span><div><strong>${e(name)}</strong>${sub?`<small>${e(sub)}</small>`:''}</div></div>`;
export const phoneFormat=value=>{const d=String(value??'');return /^\d{9}$/.test(d)?d.replace(/(\d{3})(\d{3})(\d{3})/,'$1 $2 $3'):e(d);};
export const docBadge=type=>`<span class="doc-chip">${e(type)}</span>`;
