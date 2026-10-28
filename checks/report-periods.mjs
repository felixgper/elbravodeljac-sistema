import assert from 'node:assert/strict';
import {reportRange,getReport} from '../dist/assets/js/services/dashboard-service.js';
import {resetDemo} from '../dist/assets/js/services/store-service.js';
const storage=new Map();globalThis.localStorage={getItem:k=>storage.get(k)??null,setItem:(k,v)=>storage.set(k,String(v)),removeItem:k=>storage.delete(k)};resetDemo();
for(const [period,date,start,end,previous]of [
 ['Semana','2026-01-01','2025-12-29','2026-01-05','2025-12-22'],
 ['Mes','2024-02-29','2024-02-01','2024-03-01','2024-01-01'],
 ['Mes','2026-01-31','2026-01-01','2026-02-01','2025-12-01'],
 ['Año','2026-10-27','2026-01-01','2027-01-01','2025-01-01']
]){
 const range=reportRange(period,date);assert.equal(range.start.toISOString(),start+'T05:00:00.000Z');assert.equal(range.end.toISOString(),end+'T05:00:00.000Z');assert.equal(range.previous.toISOString(),previous+'T05:00:00.000Z');
 const report=await getReport(period,date);assert.equal(report.months.reduce((sum,m)=>sum+m.cents,0),report.income);assert.equal(report.comparison.months.reduce((sum,m)=>sum+m.cents,0),report.comparison.income);
}
assert.throws(()=>reportRange('Mes','2025-02-29'),/inválida/);assert.throws(()=>reportRange('Otro','2026-01-01'),/inválidos/);
console.log('Reportes: períodos históricos, cambio de año, febrero bisiesto, comparación y conciliación de totales: OK.');
