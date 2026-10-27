import {remotePicker} from './remote-picker.js';

const normalized=value=>value.normalize('NFD').replace(/\p{M}/gu,'').toLocaleLowerCase('es').trim();
// The real catalogue comes from PostgreSQL through GET /vehiculos/marcas.
// No second list of brand names is maintained in JavaScript.
export function bindBrandPicker(root,brands,selected=''){
 const select=root.querySelector('select[name=brand]');
 const custom=root.querySelector('[name=otherBrand]'),field=custom.closest('.form-field');
 const names=[...new Set([...brands,...(selected?[selected]:[])])];
 const choices=names.map(name=>({id:name,name}));
 const other={id:'__other__',name:'Otra marca'};
 const toggle=record=>{const manual=record.id===other.id;field.hidden=!manual;custom.disabled=!manual;custom.required=manual;if(manual)custom.focus();};
 remotePicker(select,{
  selectedId:selected,initialLabel:selected,label:r=>r.name,
  search:async query=>{const rows=choices.filter(r=>normalized(r.name).startsWith(normalized(query)));rows.push(other);return {data:rows,meta:{total:rows.length,pageSize:rows.length,page:1}};},
  onSelect:toggle
 });
 toggle({id:selected});
 custom.addEventListener('input',()=>{
  const value=custom.value.trim();
  custom.setCustomValidity(value&&value.length<=40&&/^[\p{L}\p{N}][\p{L}\p{N} .&’'-]*$/u.test(value)&&normalized(value)!=='otra marca'&&value!=='__other__'?'':'Escribe una marca válida de hasta 40 caracteres.');
 });
}
