export async function downloadDashboard(format,data,summary,isCurrent){
 const generated=new Intl.DateTimeFormat('es-PE',{dateStyle:'medium',timeStyle:'short',timeZone:'America/Lima'}).format(new Date());
 let blob;
 if(format==='xlsx'){
  const {dashboardWorkbook}=await import('./dashboard-workbook.js');
  blob=await dashboardWorkbook(data,generated);
 }else{
  const {dashboardPdf}=await import('./dashboard-pdf.js');
  if(!isCurrent())throw Error('La selección cambió durante la exportación.');
  blob=await dashboardPdf(summary,data,generated);
 }
 const url=URL.createObjectURL(blob);
 const link=document.createElement('a');
 link.href=url;link.download=`el-bravo-del-jac-resumen-${data.reference}.${format}`;
 document.body.append(link);
 link.click();
 link.remove();
 setTimeout(()=>URL.revokeObjectURL(url),30000);
}
