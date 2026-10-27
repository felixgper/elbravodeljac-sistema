import {getStore} from './store-service.js';
import {modules} from '../components/navigation.js';
import {ROLES,defaultPermissions,sanitizePermissions} from '../data/permissions.js';

// Matriz completa de permisos por rol (valores guardados + valores iniciales).
export function rolePermissions(state=getStore()){
  const defaults=defaultPermissions(),saved=state.rolePermissions||{};
  return Object.fromEntries(ROLES.map(role=>[role,saved[role]?sanitizePermissions(saved[role]):defaults[role]||{}]));
}
// Devuelve una función can(módulo, acción) para el rol indicado.
export function permissionChecker(role,matrix=rolePermissions()){
  const own=matrix[role]||{};
  return (module,action='ver')=>Boolean(own[module]?.includes(action));
}
export function allowedModules(role){const can=permissionChecker(role);return modules.filter(m=>can(m.id,'ver'));}
// Primera pantalla a la que puede entrar el rol (Inicio si la tiene).
export function homePath(role){
  const can=permissionChecker(role),list=modules.filter(m=>can(m.id,'ver')),first=list.find(m=>m.id==='inicio')||list[0];
  if(!first)return 'sin-acceso';
  return first.id==='auxilio'&&!can('auxilio','crear')?'auxilio/tareas':first.id;
}
export function countPermissions(permissions){return Object.values(permissions||{}).reduce((n,list)=>n+list.length,0);}
