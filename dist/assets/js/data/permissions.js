// Catálogo de módulos, acciones y permisos iniciales de cada rol.
// Los permisos se guardan por rol; cada usuario hereda los de su rol.
export const ROLES=['Gerente','Administrador','Mecánico'];
export const ROLE_ORDER=[...ROLES];
export const ROLE_INFO={Gerente:{icon:'crown',tone:'gold',summary:'Dirección, usuarios y permisos.'},Administrador:{icon:'gear',tone:'copper',summary:'Operación diaria del taller.'},Mecánico:{icon:'wrench',tone:'gray',summary:'Diagnósticos y órdenes asignadas.'}};

// [id de acción, etiqueta]. Solo existen las acciones que el sistema realmente tiene.
export const PERMISSION_MODULES = [
  {id: 'inicio', label: 'Inicio', icon: 'home', actions: [['ver', 'Ver panel de inicio']]},
  {id: 'clientes', label: 'Clientes', icon: 'users', actions: [['ver', 'Ver clientes'], ['crear', 'Crear clientes'], ['editar', 'Editar clientes'], ['eliminar', 'Eliminar clientes']]},
  {id: 'vehiculos', label: 'Vehículos', icon: 'car', actions: [['ver', 'Ver vehículos'], ['crear', 'Crear vehículos'], ['editar', 'Editar vehículos'], ['eliminar', 'Eliminar vehículos']]},
  {id: 'recepciones', label: 'Recepciones', icon: 'reception', actions: [['ver', 'Ver recepciones'], ['crear', 'Crear recepciones']]},
  {id: 'diagnosticos', label: 'Diagnósticos', icon: 'diagnosis', actions: [['ver', 'Ver diagnósticos'], ['crear', 'Guardar y finalizar evaluaciones'], ['asignar','Asignar mecánico de diagnóstico'], ['supervisar','Supervisar diagnósticos de otros trabajadores']]},
  {id: 'cotizaciones', label: 'Cotizaciones', icon: 'quote', actions: [['ver', 'Ver cotizaciones'], ['crear', 'Crear cotizaciones'], ['editar', 'Editar cotizaciones']]},
  {id: 'ordenes', label: 'Órdenes de trabajo', icon: 'orders', actions: [['ver', 'Ver órdenes'], ['crear', 'Crear órdenes'], ['editar', 'Editar órdenes'], ['eliminar', 'Eliminar órdenes'], ['asignar', 'Reasignar mecánicos']]},
  {id: 'auxilio', label: 'Auxilio mecánico', icon: 'rescue', actions: [['ver', 'Ver auxilios y tareas'], ['crear', 'Registrar auxilios']]},
  {id: 'pagos', label: 'Pagos y entrega', icon: 'payment', actions: [['ver', 'Ver pagos'], ['registrar', 'Registrar pagos'], ['entregar', 'Confirmar entrega'], ['rectificar','Revertir pagos con motivo']]},
  {id: 'mecanicos', label: 'Mecánicos', icon: 'wrench', actions: [['ver', 'Ver mecánicos']]},
  {id: 'proveedores', label: 'Proveedores', icon: 'supplier', actions: [['ver', 'Ver proveedores'], ['crear', 'Crear proveedores'], ['editar', 'Editar proveedores'], ['eliminar', 'Eliminar proveedores']]},
  {id: 'reportes', label: 'Reportes', icon: 'chart', actions: [['ver', 'Ver reportes']]},
  {id: 'usuarios', label: 'Usuarios y roles', icon: 'shield', actions: [['ver', 'Ver usuarios'], ['crear', 'Crear usuarios'], ['editar', 'Editar usuarios'], ['eliminar', 'Eliminar usuarios'], ['roles', 'Gestionar roles y permisos']]},
];

export const TOTAL_PERMISSIONS = PERMISSION_MODULES.reduce((n, m) => n + m.actions.length, 0);

const all = () => Object.fromEntries(PERMISSION_MODULES.map(m => [m.id, m.actions.map(a => a[0])]));
const only = spec => spec;

export function defaultPermissions(){return structuredClone({"Gerente": {"inicio": ["ver"], "clientes": ["ver", "crear", "editar", "eliminar"], "vehiculos": ["ver", "crear", "editar", "eliminar"], "recepciones": ["ver", "crear"], "diagnosticos": ["ver", "crear", "asignar", "supervisar"], "cotizaciones": ["ver", "crear", "editar"], "ordenes": ["ver", "crear", "editar", "eliminar", "asignar"], "auxilio": ["ver", "crear"], "pagos": ["ver", "registrar", "entregar", "rectificar"], "mecanicos": ["ver"], "proveedores": ["ver", "crear", "editar", "eliminar"], "reportes": ["ver"], "usuarios": ["ver", "crear", "editar", "eliminar", "roles"]}, "Administrador": {"inicio": ["ver"], "clientes": ["ver", "crear", "editar", "eliminar"], "vehiculos": ["ver", "crear", "editar", "eliminar"], "recepciones": ["ver", "crear"], "diagnosticos": ["ver", "crear", "asignar", "supervisar"], "cotizaciones": ["ver", "crear", "editar"], "ordenes": ["ver", "crear", "editar", "eliminar", "asignar"], "auxilio": ["ver", "crear"], "pagos": ["ver", "registrar", "entregar"], "mecanicos": ["ver"], "proveedores": ["ver", "crear", "editar", "eliminar"], "reportes": ["ver"]}, "Mecánico": {"inicio": ["ver"], "vehiculos": ["ver"], "recepciones": ["ver"], "diagnosticos": ["ver", "crear"], "ordenes": ["ver", "editar"], "auxilio": ["ver"]}});}

// Reglas que nunca se pueden romper para no dejar el sistema sin gestión.
export const LOCKED_ROLE = 'Gerente';
export const REQUIRED_FOR_ADMIN = {usuarios: ['ver', 'editar', 'roles']};

// Limpia una matriz: solo módulos y acciones válidas, y sin acciones sueltas si falta "ver".
export function sanitizePermissions(input = {}) {
  const clean = {};
  for (const module of PERMISSION_MODULES) {
    const given = Array.isArray(input[module.id]) ? input[module.id] : [];
    const valid = module.actions.map(a => a[0]).filter(a => given.includes(a));
    if (valid.includes('ver')) clean[module.id] = valid;
  }
  return clean;
}
