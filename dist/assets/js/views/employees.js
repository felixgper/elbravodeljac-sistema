import {realMode,request} from '../services/api-service.js';
import {realStore} from '../services/real-state.js';
import {permissionChecker} from '../services/permissions.js';
import {e,field,checkbox,form,submit,link,pageHeader,hint,table,badge,bindForm,notify,go,searchableList} from '../components/ui.js';

export function renderEmployees({root,path,params,session}){
 if(!realMode){root.innerHTML=pageHeader('Empleados','Gestión laboral')+hint('La gestión de empleados está disponible en modo real; no modifica los datos de demostración.');return;}
 const records=realStore().employees||[],can=permissionChecker(session.role),id=path[2]==='editar'?path[3]:null;
 if(['nuevo','editar'].includes(path[2])){
  if(!can('usuarios',id?'editar':'crear'))throw Error('No tienes permiso para gestionar empleados.');
  const employee=id?records.find(x=>x.id===id):{};if(!employee)throw Error('Empleado no encontrado.');
  root.innerHTML=pageHeader(id?'Editar empleado':'Registrar empleado','El cargo laboral es independiente de los roles de acceso.')+form('DATOS LABORALES',field('Nombre completo','name',employee.name||'','text','required maxlength="150"')+field('Tipo de empleado / cargo laboral','employeeType',employee.employeeType||'','text','required maxlength="40" list="employee-types"')+'<datalist id="employee-types">'+['Mecánico','Auxiliar','Supervisor','Administrador','Practicante'].map(v=>`<option value="${e(v)}">`).join('')+'</datalist>'+checkbox('Empleado activo','active',id?employee.active:true),submit('Guardar empleado')+link('Volver a empleados','usuarios/empleados'))+hint('Registrar un empleado no crea una cuenta ni concede permisos. El estado laboral no sustituye la habilitación de su cuenta.');
  bindForm('edit-form',async data=>{await request('/empleados'+(id?'/'+id:''),{method:id?'PATCH':'POST',body:{name:data.name,employeeType:data.employeeType,active:Boolean(data.active)}});notify(id?'Empleado actualizado.':'Empleado registrado sin crear una cuenta.');go('usuarios/empleados');});return;
 }
 searchableList(root,{title:'Empleados',subtitle:'Trabajadores con o sin cuenta. El cargo no asigna permisos.',actions:link('Volver a usuarios','usuarios')+(can('usuarios','crear')?link('Registrar empleado','usuarios/empleados/nuevo',true):''),placeholder:'Buscar por nombre o cargo…',headers:['Empleado','Cargo laboral','Estado laboral','Cuenta','Acción'],records,searchText:x=>x.name+' '+x.employeeType,row:x=>[e(x.name),e(x.employeeType),badge(x.active?'Activo':'Inactivo'),e(x.username||'Sin cuenta'),(can('usuarios','editar')?link('Editar','usuarios/empleados/editar/'+x.id):'')+(!x.userId&&x.active&&can('usuarios','crear')&&can('usuarios','roles')?' '+link('Crear cuenta','usuarios/nuevo?empleado='+x.id):'')]});
 root.insertAdjacentHTML('beforeend',hint('Para vincular una cuenta existente, abre Usuarios → Editar y selecciona el empleado verificado. Se conservan sus roles y su historial. Las asignaciones técnicas siguen usando las cuentas actuales.'));
}
