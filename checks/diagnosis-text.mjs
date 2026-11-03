import assert from 'node:assert/strict';
import {diagnosisText} from '../dist/assets/js/services/diagnosis-text.js';
assert.equal(diagnosisText({findings:'A',evaluation:'A'}),'A');assert.equal(diagnosisText({findings:'',evaluation:'B'}),'B');
const legacy=diagnosisText({findings:'Hallazgo original',evaluation:'Causa distinta'});assert(legacy.includes('Hallazgo original')&&legacy.includes('Causa distinta'));assert.equal(legacy.match(/Hallazgo original/g).length,1);assert.equal(diagnosisText({findings:legacy,evaluation:''}),legacy);
console.log('Diagnóstico unificado: legado distinto conservado, iguales sin duplicar y relectura estable.');
