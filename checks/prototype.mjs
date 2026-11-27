import assert from 'node:assert/strict';
import fs from 'node:fs';
import {dashboardDemo} from './fixtures/dashboard-prototype.js';
import {getActiveOrders} from '../dist/assets/js/services/dashboard-service.js';
import {modules} from '../dist/assets/js/components/navigation.js';
// Historical prototype contract; NOT a live Login/Dashboard or database test.
assert.equal(getActiveOrders(dashboardDemo),10);
assert.equal(dashboardDemo.states.reduce((n,s)=>n+s.count,0),35);
assert.equal(dashboardDemo.income.at(-1).cents,dashboardDemo.monthlyIncomeCents);
const screens=JSON.parse(fs.readFileSync('docs/balsamiq-v14.1.json','utf8'));
const modals=screens.filter(s=>/08[AB] - .*\(modal\)/.test(s.name));
assert.equal(screens.length,25);assert.equal(modals.length,2);
const dashboard=screens.find(s=>s.name==='02 - Dashboard');
const sourceNav=dashboard.controls.filter(c=>c.typeID==='Label'&&Number(c.x)<220).map(c=>c.properties.text.trim().replace(/^>\s*/,'')).filter(t=>t!=='TALLER AUTOMOTRIZ');
assert.deepEqual(modules.map(m=>m.label),sourceNav.filter(label=>label!=='Reportes'),'El menú conserva los módulos operativos; Reportes está integrado en Inicio');
const mapped=['01 - Inicio de sesión','02 - Dashboard',...modules.flatMap(m=>m.screens??[])];
assert.deepEqual([...new Set(mapped)].sort(),screens.filter(s=>!modals.includes(s)).map(s=>s.name).sort(),'Las 23 pantallas principales deben estar contempladas');
const texts=dashboard.controls.map(c=>c.properties?.text);
assert(texts.includes('S/ 12,450'));assert(texts.includes('10'));assert(texts.includes('24'));assert(texts.includes('3'));
const create=screens.find(s=>s.name==='11 - Nueva orden de trabajo');
const update=screens.find(s=>s.name==='11B - Actualizar orden');
assert.notEqual(create.id,update.id);assert(create.controls.some(c=>c.properties?.text==='Cotización aprobada (obligatoria)'));
assert(update.controls.some(c=>c.properties?.text==='Observación / avance de esta actualización'));
console.log('Prototipo histórico Balsamiq V14.1: 23 pantallas principales y dos estados del modal, navegación y valores ilustrativos conservados; no acredita el Dashboard real.');

assert(screens.find(s=>s.name.startsWith('08 -')).controls.some(c=>c.properties?.text==='Evidencias y escáner'));
