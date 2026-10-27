import {request} from '../services/api-service.js';
import {e,field,textarea,select,checkbox,form,submit,showModal,bindForm,hint,notify} from './ui.js';

export async function openInspection(order,quote){
 const response=(await request('/ordenes/'+order.id+'/control-tecnico')).data,{record,catalogue}=response,c=record.control||{},locked=record.readOnly||!record.canEdit,attrs=locked?'disabled':'';
 let body=catalogue.groups.map(g=>`<fieldset class="full-width"><legend>${e(g.label)}</legend><div class="form-grid">${g.items.map(i=>select(i.label,'inspect_'+i.id,['Sin evaluar','Bien','Atención','Urgente'],c.inspections?.[i.id]||'Sin evaluar',attrs)).join('')}</div></fieldset>`).join('');
 body+=`<fieldset class="full-width"><legend>Neumáticos · Presión PSI</legend><div class="form-grid">${[['frontLeft','Delantero piloto'],['frontRight','Delantero copiloto'],['rearLeft','Trasero piloto'],['rearRight','Trasero copiloto']].map(([id,label])=>field(label,'tire_'+id,c.tires?.[id]||'','number',attrs+' min="0" max="150" step="0.1"')).join('')}</div></fieldset>`;
 body+=field('Otros líquidos','otherFluids',c.otherFluids||'','text',attrs+' maxlength="1200"');
 body+=`<fieldset class="full-width"><legend>Mantenimiento preventivo y correctivo · Utilizado en este servicio</legend><div class="form-grid">${catalogue.maintenance.map(g=>`<div class="form-field"><strong>${e(g.label)}</strong>${g.options.map(i=>checkbox(i.label,'maint_'+g.id+'_'+i.id,c.maintenance?.[g.id]?.includes(i.id),attrs)).join('')}</div>`).join('')}</div></fieldset>`;
 body+=field('Bobinas','coils',c.coils||'','text',attrs+' maxlength="1200"');
 body+=`<fieldset class="full-width"><legend>Otros servicios y reparaciones realizados</legend><div class="form-grid">${catalogue.otherServices.map(i=>checkbox(i.label,'service_'+i.id,c.otherServices?.includes(i.id),attrs)).join('')}</div></fieldset>`;
 body+=`<fieldset class="full-width"><legend>Trabajos de la cotización realmente realizados</legend>${(quote?.items||[]).map(i=>checkbox(i.concept,'performed_'+i.id,c.performedItems?.includes(i.id),attrs)).join('')}</fieldset>`;
 body+=textarea('Observaciones del cliente','clientNotes',c.clientNotes||'',attrs+' maxlength="1200"')+textarea('Observaciones técnicas','technicalNotes',c.technicalNotes||'',attrs+' maxlength="1200"')+field('Próximo mantenimiento recomendado (km)','nextMileage',c.nextMileage||'','number',attrs+' min="0" max="999999999"')+field('Fecha recomendada','nextDate',c.nextDate||'','date',attrs);
 const dialog=showModal('Control técnico · '+order.code,hint(locked?'Consulta del control guardado. La orden está cerrada o no tienes asignación para modificarlo.':'Completa durante la reparación. No marques servicios que no se realizaron. Dejar casillas vacías conserva Sin evaluar.')+form('FICHA TÉCNICA · PÁGINA 2',body,locked?'':submit('Guardar control técnico'),'inspection-form'));
 if(!locked)bindForm('inspection-form',async data=>{
  const control={version:record.version,inspections:{},maintenance:{},otherServices:[],tires:{},performedItems:[]};
  for(const g of catalogue.groups)for(const i of g.items)control.inspections[i.id]=data['inspect_'+i.id];
  for(const g of catalogue.maintenance)control.maintenance[g.id]=g.options.filter(i=>data['maint_'+g.id+'_'+i.id]).map(i=>i.id);
  control.otherServices=catalogue.otherServices.filter(i=>data['service_'+i.id]).map(i=>i.id);
  for(const id of ['frontLeft','frontRight','rearLeft','rearRight'])control.tires[id]=data['tire_'+id]||'';
  control.performedItems=(quote?.items||[]).filter(i=>data['performed_'+i.id]).map(i=>i.id);
  for(const key of ['otherFluids','coils','clientNotes','technicalNotes','nextMileage','nextDate'])control[key]=data[key]||'';
  await request('/ordenes/'+order.id+'/control-tecnico',{method:'PUT',body:control});dialog.close();notify('Control técnico guardado.');
 });
}
export function sheetButton(id,label='Exportar ficha técnica (PDF)'){return `<button type="button" class="secondary-button" data-sheet="${e(id)}">${e(label)}</button>`;}
export function bindSheetDownloads(root){root.querySelectorAll('[data-sheet]').forEach(button=>button.onclick=async()=>{button.disabled=true;try{const blob=await request('/ordenes/'+encodeURIComponent(button.dataset.sheet)+'/ficha-tecnica.pdf',{responseType:'blob'}),url=URL.createObjectURL(blob),a=document.createElement('a');a.href=url;a.download='ficha-OT-'+button.dataset.sheet+'.pdf';a.click();setTimeout(()=>URL.revokeObjectURL(url),1000);}catch(error){showModal('No se pudo exportar la ficha',hint(error.message));}finally{button.disabled=false;}});}
