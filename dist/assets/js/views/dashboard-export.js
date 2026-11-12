export async function downloadDashboard(format,data){
 const generated=new Intl.DateTimeFormat('es-PE',{dateStyle:'medium',timeStyle:'short',timeZone:'America/Lima'}).format(new Date());
 const blob=format==='xlsx'
  ?(await import('./dashboard-workbook.js')).dashboardWorkbook(data,generated)
  :await (await import('./dashboard-pdf.js')).dashboardPdf(data,generated);
 const url=URL.createObjectURL(blob);
 const link=document.createElement('a');
 link.href=url;link.download=`el-bravo-del-jac-resumen-${data.reference}.${format}`;
 document.body.append(link);
 link.click();
 link.remove();
 setTimeout(()=>URL.revokeObjectURL(url),30000);
}
