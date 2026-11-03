// Read legacy pairs losslessly; equal values are displayed only once.
export function diagnosisText(record={}) {
 const findings=String(record.findings??record.resumen??''),evaluation=String(record.evaluation??record.evaluacion??'');
 if(!findings.trim())return evaluation;
 if(!evaluation.trim()||findings.trim()===evaluation.trim())return findings;
 return 'Hallazgos registrados:\n'+findings+'\n\nEvaluación técnica registrada:\n'+evaluation;
}
