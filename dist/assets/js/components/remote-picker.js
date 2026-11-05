import {e} from './ui.js';
import {lookupClients,clientByID,clientVehicles} from '../services/real-workshop.js';
// A native value control remains the form's source of truth. Typing invalidates
// the previous ID immediately; stale/aborted responses cannot select a record.
export function remotePicker(select,{search,label,selectedId='',initialLabel='',onSelect}={}){
 if(!select)return;
 const required=select.required;select.required=false;select.hidden=true;
 const wrapper=document.createElement('div');wrapper.className='remote-picker';
 const oldID=select.id;select.id=oldID+'-value';
 wrapper.innerHTML=`<input id="${e(oldID)}" role="combobox" aria-autocomplete="list" aria-expanded="false" aria-controls="${e(oldID)}-results" autocomplete="off" placeholder="Escriba para buscar…" ${required?'required':''}><div id="${e(oldID)}-results" class="remote-results" role="listbox" hidden></div>`;
 select.after(wrapper);const input=wrapper.querySelector('input'),results=wrapper.querySelector('[role=listbox]');
 input.disabled=select.disabled;input.value=initialLabel;
 let page=1,query='',ticket=0,controller,timer,active=-1;
 function selected(record){initialLabel=label(record);select.innerHTML=`<option value="${e(record.id)}" selected>${e(label(record))}</option>`;input.value=label(record);input.setCustomValidity('');results.hidden=true;input.setAttribute('aria-expanded','false');select.dispatchEvent(new Event('change',{bubbles:true}));onSelect?.(record);}
 async function draw(){
  const current=++ticket;controller?.abort();controller=new AbortController();results.hidden=false;input.setAttribute('aria-expanded','true');results.innerHTML='<p role="status">Buscando…</p>';
  try{const response=await search(query,page,controller.signal);if(current!==ticket||!wrapper.isConnected)return;const rows=response.data;active=-1;
   results.innerHTML=rows.map((r,i)=>`<button type="button" role="option" data-result="${i}">${e(label(r))}</button>`).join('')||'<p role="status">No hay coincidencias.</p>';
   for(const b of results.querySelectorAll('[data-result]'))b.onclick=()=>selected(rows[Number(b.dataset.result)]);
   if(response.meta?.total>(response.meta.pageSize||20)){results.insertAdjacentHTML('beforeend',`<div class="picker-pages"><button type="button" data-page="prev" ${page===1?'disabled':''}>Anterior</button><span>Página ${page}</span><button type="button" data-page="next" ${page*(response.meta.pageSize||20)>=response.meta.total?'disabled':''}>Siguiente</button></div>`);results.querySelectorAll('[data-page]').forEach(b=>b.onclick=()=>{page+=b.dataset.page==='next'?1:-1;draw();});}
  }catch(error){if(error.name!=='AbortError'&&current===ticket)results.innerHTML=`<p role="alert">${e(error.message)}</p>`;}
 }
 input.addEventListener('input',()=>{select.innerHTML='<option value=""></option>';input.setCustomValidity(required?'Seleccione un resultado.':'');query=input.value;page=1;active=-1;results.hidden=false;results.innerHTML='<p role="status">Buscando…</p>';input.setAttribute('aria-expanded','true');clearTimeout(timer);controller?.abort();ticket++;timer=setTimeout(draw,250);});
 input.addEventListener('focus',()=>{query=input.value===initialLabel?'':input.value;page=1;draw();});
 input.addEventListener('keydown',event=>{const options=[...results.querySelectorAll('[data-result]')];if(event.key==='Escape'){results.hidden=true;input.setAttribute('aria-expanded','false');}else if(['ArrowDown','ArrowUp'].includes(event.key)){event.preventDefault();active=Math.max(0,Math.min(options.length-1,active+(event.key==='ArrowDown'?1:-1)));options.forEach((o,i)=>o.setAttribute('aria-selected',String(i===active)));}else if(event.key==='Enter'&&active>=0&&options[active]){event.preventDefault();options[active].click();}});
 wrapper.addEventListener('focusout',()=>setTimeout(()=>{if(!wrapper.contains(document.activeElement)){results.hidden=true;input.setAttribute('aria-expanded','false');}},0));
 new MutationObserver(()=>input.disabled=select.disabled).observe(select,{attributes:true,attributeFilter:['disabled']});
 if(selectedId&&initialLabel){select.innerHTML=`<option value="${e(selectedId)}" selected>${e(initialLabel)}</option>`;}
 return {input,set:record=>{initialLabel=label(record);select.innerHTML=`<option value="${e(record.id)}" selected>${e(initialLabel)}</option>`;input.value=initialLabel;input.setCustomValidity('');}};
}
export async function clientPicker(root,module,{name='clientId',selectedId=''}={}){
 const select=root.querySelector(`select[name="${name}"]`);if(!select)return;
 const current=selectedId||select.value,record=current?await clientByID(current,module):null;
 if(!root.isConnected)return;
 const label=c=>`${c.name} · ${c.documentType} ${c.documentNumber}`;
 return remotePicker(select,{search:(q,p,signal)=>lookupClients(q,p,signal,module),label,selectedId:current,initialLabel:record?label(record):''});
}
export async function ownerVehicles(root,module,state,selectedId=''){
 const client=root.querySelector('select[name=clientId]'),vehicle=root.querySelector('select[name=vehicleId]');if(!client||!vehicle)return;
 let version=0;
 async function refresh(id=''){
  const current=++version,clientId=client.value;const newVehicle=root.querySelector('#new-vehicle');if(newVehicle)newVehicle.href='#/vehiculos/nuevo?volver=recepciones&cliente='+encodeURIComponent(clientId);vehicle.disabled=true;vehicle.innerHTML='<option value="">Cargando vehículos…</option>';
  try{const rows=await clientVehicles(clientId,module);if(current!==version||!root.isConnected)return;state.vehicles=rows;
   vehicle.innerHTML='<option value="">'+(rows.length?'Seleccione…':'No hay vehículos registrados para este cliente.')+'</option>'+rows.map(v=>`<option value="${e(v.id)}">${e(v.plate+' · '+v.brand+' '+v.model)}</option>`).join('');
   vehicle.value=rows.some(v=>v.id===id)?id:'';vehicle.disabled=false;vehicle.dispatchEvent(new Event('change',{bubbles:true}));
  }catch(error){if(current===version){vehicle.innerHTML='<option value="">No se pudieron consultar los vehículos.</option>';const note=root.querySelector('.form-error');if(note){note.hidden=false;note.textContent=error.message;}}}
 }
 client.onchange=()=>refresh();await refresh(selectedId);
}
