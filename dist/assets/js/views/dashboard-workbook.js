const encoder=new TextEncoder();
const xml=value=>String(value).replace(/[\x00-\x08\x0b\x0c\x0e-\x1f]/g,'').replace(/[&<>"']/g,char=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&apos;'})[char]);
const spreadsheet='http://schemas.openxmlformats.org/spreadsheetml/2006/main';
const officeRel='http://schemas.openxmlformats.org/officeDocument/2006/relationships';
const packageRel='http://schemas.openxmlformats.org/package/2006/relationships';
const contentTypes='http://schemas.openxmlformats.org/package/2006/content-types';
const declaration='<?xml version="1.0" encoding="UTF-8" standalone="yes"?>';

function crc32(bytes){
 let crc=0xffffffff;
 for(const byte of bytes){
  crc^=byte;
  for(let bit=0;bit<8;bit++)crc=(crc>>>1)^(0xedb88320&-(crc&1));
 }
 return (crc^0xffffffff)>>>0;
}

function header(size,fields){
 const bytes=new Uint8Array(size),view=new DataView(bytes.buffer);
 for(const [offset,value,width] of fields){
  if(width===4)view.setUint32(offset,value,true);
  else view.setUint16(offset,value,true);
 }
 return bytes;
}

// XLSX uses ZIP. Storing these small XML entries avoids a compression dependency.
function workbookZip(files){
 const entries=[],directory=[];
 let offset=0,directorySize=0;
 for(const [path,source] of files){
  const name=encoder.encode(path),data=typeof source==='string'?encoder.encode(declaration+source):source,checksum=crc32(data);
  const local=header(30,[[0,0x04034b50,4],[4,20],[12,33],[14,checksum,4],[18,data.length,4],[22,data.length,4],[26,name.length]]);
  const central=header(46,[[0,0x02014b50,4],[4,20],[6,20],[14,33],[16,checksum,4],[20,data.length,4],[24,data.length,4],[28,name.length],[42,offset,4]]);
  entries.push(local,name,data);
  directory.push(central,name);
  offset+=local.length+name.length+data.length;
  directorySize+=central.length+name.length;
 }
 const end=header(22,[[0,0x06054b50,4],[8,files.length],[10,files.length],[12,directorySize,4],[16,offset,4]]);
 return new Blob([...entries,...directory,end],{type:'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'});
}


import {reportRange} from '../services/dashboard-service.js';
import {nativePivot} from './dashboard-pivot.js';
import {nativeCharts} from './dashboard-excel-charts.js';

const serial=date=>Date.parse(date.slice(0,10))/86400000+25569;
const letter=index=>index<26?String.fromCharCode(65+index):letter(Math.floor(index/26)-1)+letter(index%26);
const ref=(col,row)=>letter(col)+row;
const moneyFormat='&quot;S/ &quot;#,##0.00';
function styles(){
 const fonts=[['292B32',11,false],['79223A',11,true],['79223A',24,true],['79223A',22,true],['51545F',11,false],['292B32',11,true]];
 const fills=['F5E9ED','F6F7F9','FFFFFF'];
 const formats=[[0,0,0],[1,2,0],[0,0,164],[0,0,165],[4,3,0],[2,3,0],[3,4,0,1],[3,4,164,1],[5,4,0,1],[0,3,165],[0,4,0,1],[0,4,0,1,true],[4,4,0],[1,4,0,1],[4,3,0,0,true]];
 return `<styleSheet xmlns="${spreadsheet}"><numFmts count="2"><numFmt numFmtId="164" formatCode="${moneyFormat}"/><numFmt numFmtId="165" formatCode="dd/mm/yyyy"/></numFmts><fonts count="${fonts.length}">${fonts.map(([color,size,bold])=>`<font>${bold?'<b/>':''}<sz val="${size}"/><color rgb="FF${color}"/><name val="Calibri"/></font>`).join('')}</fonts><fills count="5"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill>${fills.map(color=>`<fill><patternFill patternType="solid"><fgColor rgb="FF${color}"/><bgColor indexed="64"/></patternFill></fill>`).join('')}</fills><borders count="2"><border><left/><right/><top/><bottom/><diagonal/></border><border><left style="thin"><color rgb="FFE1E3E8"/></left><right style="thin"><color rgb="FFE1E3E8"/></right><top style="thin"><color rgb="FFE1E3E8"/></top><bottom style="thin"><color rgb="FFE1E3E8"/></bottom><diagonal/></border></borders><cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs><cellXfs count="${formats.length}">${formats.map(([f,b,n,border=0,wrap=false])=>`<xf numFmtId="${n}" fontId="${f}" fillId="${b}" borderId="${border}" xfId="0" applyNumberFormat="1" applyFont="1" applyFill="1" applyBorder="1" applyAlignment="1"><alignment vertical="center"${wrap?' wrapText="1"':''}/></xf>`).join('')}</cellXfs><cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles></styleSheet>`;
}
function put(rows,row,col,value,style=0,formula){
 if(!rows.has(row))rows.set(row,new Map());
 rows.get(row).set(col,{value,style,formula});
}
function sheet(rows,widths,{cover=false,extra='',freeze=false,merges=[],heights={},landscape=true}={}){
 const body=[...rows].sort(([a],[b])=>a-b).map(([row,entries])=>`<row r="${row}"${heights[row]?` ht="${heights[row]}" customHeight="1"`:''}>${[...entries].sort(([a],[b])=>a-b).map(([col,{value,style,formula}])=>{
  const opening=`<c r="${ref(col,row)}" s="${style}"`;
  if(value===null)return opening+'/>';
  if(typeof value==='number')return `${opening}>${formula?`<f>${xml(formula)}</f>`:''}<v>${value}</v></c>`;
  if(formula)return `${opening} t="str"><f>${xml(formula)}</f><v>${xml(value)}</v></c>`;
  return `${opening} t="inlineStr"><is><t xml:space="preserve">${xml(value)}</t></is></c>`;
 }).join('')}</row>`).join('');
 return `<worksheet xmlns="${spreadsheet}" xmlns:r="${officeRel}"><sheetPr><tabColor rgb="FF79223A"/><pageSetUpPr fitToPage="1"/></sheetPr><sheetViews><sheetView showGridLines="0" workbookViewId="0"${cover?' zoomScale="85" zoomScaleNormal="85"':''}>${freeze?'<pane ySplit="3" topLeftCell="A4" activePane="bottomLeft" state="frozen"/>':''}</sheetView></sheetViews><sheetFormatPr defaultRowHeight="18"/><cols>${widths.map((w,i)=>`<col min="${i+1}" max="${i+1}" width="${w}" customWidth="1"/>`).join('')}</cols><sheetData>${body}</sheetData>${merges.length?`<mergeCells count="${merges.length}">${merges.map(r=>`<mergeCell ref="${r}"/>`).join('')}</mergeCells>`:''}<pageMargins left="0.25" right="0.25" top="0.35" bottom="0.35" header="0.15" footer="0.15"/><pageSetup paperSize="9" orientation="${landscape?'landscape':'portrait'}" fitToWidth="1" fitToHeight="1"/>${extra}</worksheet>`;
}
function packageBuilder(){
 const files=[],types=[];
 const add=(path,type,content,family='spreadsheetml')=>{
  files.push(['xl/'+path,content]);
  types.push(`<Override PartName="/xl/${path}" ContentType="application/vnd.openxmlformats-officedocument.${family?`${family}.`: ''}${type}+xml"/>`);
 };
 const relationships=entries=>`<Relationships xmlns="${packageRel}">${entries.map(([id,type,target])=>`<Relationship Id="rId${id}" Type="${officeRel}/${type}" Target="${target}"/>`).join('')}</Relationships>`;
 return {add,files,rels(folder,name,entries){files.push([`xl/${folder?folder+'/':''}_rels/${name}.rels`,relationships(entries)]);},finish(){
  files.push(['_rels/.rels',relationships([[1,'officeDocument','xl/workbook.xml']])]);
  files.push(['[Content_Types].xml',`<Types xmlns="${contentTypes}"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Default Extension="png" ContentType="image/png"/>${types.join('')}</Types>`]);
  return workbookZip(files);
 }};
}

function sourcesFor(data){
 const reference=serial(data.reference),sources=[];
 if(data.income){
  const period=data.period.split(' · ')[0],{start}=reportRange(period,data.reference);
  const rows=data.income.map((point,index)=>{
   const date=new Date(start);
   if(period==='Año')date.setUTCMonth(date.getUTCMonth()+index);
   else date.setUTCDate(date.getUTCDate()+index);
   return [serial(date.toISOString()),point.cents/100,reference];
  });
  sources.push({name:'Ingresos',title:'Evolución de ingresos',head:['Intervalo','Ingreso neto','Referencia'],rows,kind:'line',color:'C6607D',date:true,money:true,monthly:period==='Año'});
 }
 sources.push(
  {name:'Estados',title:'Estados de órdenes',head:['Estado','Órdenes','Referencia'],rows:data.states.map(x=>[x.label,x.count,reference]),kind:'doughnut',color:'A55068',colors:data.states.map(x=>/^#[0-9a-f]{6}$/i.test(x.color)?x.color.slice(1):'A55068')},
  {name:'Mecanicos',title:'Trabajo por mecánico',head:['Mecánico','Órdenes activas','Referencia'],rows:data.mechanics.map(x=>[x.label,x.count,reference]),kind:'bar',color:'5B91BF'},
  {name:'Servicios',title:'Servicios más realizados',head:['Servicio','Cantidad','Referencia'],rows:data.services.map(x=>[x.label,x.count,reference]),kind:'bar',color:'4D9E88',multicolor:true}
 );
 sources.push(
  {name:'Indicadores',title:'Indicadores y actividad',head:['Indicador','Cantidad'],rows:[
   ...data.cards.filter(card=>card.amount===undefined).map(card=>[card.label,card.value]),...data.activity
  ]},
  {name:'Consulta',title:'Período consultado',head:['Período','Fecha de referencia'],rows:[[data.period,reference]],reference:true}
 );
 return sources.map((s,i)=>({...s,col:i*5,id:i+1}));
}

export async function dashboardWorkbook(data,generated){
 const pkg=packageBuilder(),sources=sourcesFor(data);
 const series=sources.filter(s=>s.kind&&s.rows.length);
 const cover=new Map(),details=new Map(),pivots=new Map(),graphs=new Map(),guide=new Map();
 const states=sources.find(source=>source.name==='Estados');
 const statusChart={...states,name:'Estatus',title:'Órdenes y estatus',kind:'horizontal'};
 const charts=[...series,...(states.rows.length?[statusChart]:[])];
 const names=['Dashboard','Datos','Tablas dinámicas','Gráficos','Instrucciones'];
 const putTitles=(rows,title)=>put(rows,1,0,title,1);
 const coverMerges=['B2:L3','B4:L4','B5:C5','D5:E5','G5:L5'];
 const coverHeights={2:24,3:13,6:9,8:28,9:9,19:32,20:23,32:9,34:24};
 const merge=(row,start,end,value,style,formula)=>{
  put(cover,row,start,value,style,formula);
  for(let col=start+1;col<=end;col++)put(cover,row,col,null,style);
  if(start!==end)coverMerges.push(`${ref(start,row)}:${ref(end,row)}`);
 };
 put(cover,2,1,'EL BRAVO DEL JAC',5);
 put(cover,4,1,data.period+(data.demo?' · Demostración':''),4,'INDEX(Consulta[Período],1)'+(data.demo?'&" · Demostración"':''));
 put(cover,5,1,'Fecha de referencia',4);
 put(cover,5,3,serial(data.reference),9,'INDEX(Consulta[Fecha de referencia],1)');
 put(cover,5,6,'Generado en Lima: '+generated,4);
 const metric=label=>`SUMIF(Indicadores[Indicador],"${label}",Indicadores[Cantidad])`;
 const active=['Pendientes','En reparación','En pruebas'].map(label=>`SUMIF(Estados[Estado],"${label}",Estados[Órdenes])`).join('+');
 data.cards.forEach((card,i)=>{
  const span=12/data.cards.length,col=i*span;
  merge(7,col,col+span-1,card.label,8);
  merge(8,col,col+span-1,card.amount??card.value,card.amount===undefined?6:7,card.amount!==undefined?'SUM(Ingresos[Ingreso neto])':card.label==='Órdenes activas'?active:metric(card.label));
 });
 const stateTotal=data.states.reduce((total,state)=>total+Number(state.count),0);
 merge(32,0,4,`Total: ${stateTotal} órdenes`,14,'"Total: "&SUM(Estados[Órdenes])&" órdenes"');
 data.activity.forEach(([label,value],i)=>{
  merge(33,i*3,i*3+2,label,8);
  merge(34,i*3,i*3+2,value,6,metric(label));
 });
 const tableRels=[];
 const widths=[];
 for(const source of sources){
  const {id,name,head,col,rows}=source;
  put(details,1,col,source.title,1);
  widths.push(source.reference?39:source.date?16:29,head.length===4?32:17,16,19,4);
  head.forEach((label,i)=>put(details,3,col+i,label,1));
  if(!rows.length)head.forEach((_,i)=>put(details,4,col+i,null));
  rows.forEach((row,i)=>row.forEach((v,j)=>put(details,4+i,col+j,v,j===0&&source.date||j===2&&source.kind||j===1&&source.reference?3:j===1&&source.money?2:0)));
  const range=`${ref(col,3)}:${ref(col+head.length-1,3+Math.max(1,rows.length))}`;
  pkg.add(`tables/table${id}.xml`,'table',`<table xmlns="${spreadsheet}" id="${id}" name="${name}" displayName="${name}" ref="${range}" totalsRowShown="0"><autoFilter ref="${range}"/><tableColumns count="${head.length}">${head.map((label,i)=>`<tableColumn id="${i+1}" name="${xml(label)}"/>`).join('')}</tableColumns><tableStyleInfo name="TableStyleMedium3" showRowStripes="1"/></table>`);
  tableRels.push([id,'table',`../tables/table${id}.xml`]);
 }
 const caches=series.map((s,i)=>nativePivot({pkg,source:s,rows:pivots,index:i,spreadsheet,officeRel,xml,put,ref}));
 if(caches.length)pkg.rels('worksheets','sheet3.xml',caches.map(id=>[id,'pivotTable',`../pivotTables/pivotTable${id}.xml`]));
 putTitles(pivots,'Análisis dinámico · datos agregados');
 putTitles(graphs,'Gráficos editables · vinculados a Datos');
 const positions={Ingresos:[0,9,8,18],Estados:[data.income?8:0,9,12,18],
  Estatus:[0,19,5,31],Mecanicos:[5,19,8,31],Servicios:[8,19,12,31]};
 for(const source of sources.filter(s=>s.kind&&!s.rows.length)){
  const [left,top,right]=positions[source.name];
  merge(top+1,left,right-1,source.title,1);
  merge(top+2,left,right-1,'Sin datos en el resumen',10);
 }
 const drawingCover=await nativeCharts({pkg,sources:charts,id:1,letter,xml,logo:true,positions});
 const drawingGraphs=await nativeCharts({pkg,sources:charts,id:2,firstRow:2,letter,xml});
 putTitles(guide,'Cómo utilizar el informe');
 const instructions=[
  'Usa los filtros de las tablas de Datos para ordenar y revisar los agregados autorizados.',
  'En Tablas dinámicas, selecciona una tabla y abre la lista de campos. Mueve los campos entre filas, columnas, filtros y valores.',
  'Para recalcular después de editar o ampliar la tabla de origen, usa Datos → Actualizar todo.',
  'En Configuración de campo de valor puedes cambiar la suma por promedio o recuento de filas agregadas.',
  'Si hay ingresos autorizados, agrupa el campo Intervalo por meses, trimestres o años en Excel.',
  'Los gráficos están vinculados a las columnas de sus tablas. Los cambios y las filas nuevas se reflejan al recalcular el libro.',
  'Modo real: los estados activos son actuales. Finalizadas y canceladas se cuentan por fecha de cierre dentro del período.',
  'Datos contiene los agregados autorizados de Inicio. El servicio no proporciona el detalle transaccional de esos agregados.',
  'El libro es una copia de los datos del momento de la descarga. Actualizar en Excel no vuelve a consultar el sistema.'
 ];
 instructions.forEach((line,i)=>put(guide,i+3,0,line));
 const chartWidths=[12,14,14,12,16,13,13,13,13,13,13,13];
 const chartRanges=series.flatMap(source=>source.head.slice(0,2).map((label,i)=>
  `<definedName name="Serie${source.id}_${i}" localSheetId="1">${source.name}[${xml(label)}]</definedName>`)).join('');
 const pivotWidths=series.flatMap(s=>[s.name==='Servicios'?40:25,23,3,3]);
 const contents=[
  sheet(cover,chartWidths,{cover:true,extra:drawingCover,merges:coverMerges,heights:coverHeights}),
  sheet(details,widths,{freeze:true,merges:sources.map(s=>`${ref(s.col,1)}:${ref(s.col+s.head.length-1,1)}`),extra:`<tableParts count="${tableRels.length}">${tableRels.map(([id])=>`<tablePart r:id="rId${id}"/>`).join('')}</tableParts>`}),
  sheet(pivots,pivotWidths,{merges:series.length?['A1:D1']:[]}),
  sheet(graphs,chartWidths,{extra:drawingGraphs}),
  sheet(guide,[132],{landscape:false})
 ];
 contents.forEach((content,i)=>pkg.add(`worksheets/sheet${i+1}.xml`,'worksheet',content));
 pkg.rels('worksheets','sheet2.xml',tableRels);
 pkg.add('styles.xml','styles',styles());
 pkg.add('workbook.xml','sheet.main',`<workbook xmlns="${spreadsheet}" xmlns:r="${officeRel}"><bookViews><workbookView/></bookViews><sheets>${names.map((name,i)=>`<sheet name="${xml(name)}" sheetId="${i+1}" r:id="rId${i+1}"/>`).join('')}</sheets><definedNames><definedName name="_xlnm.Print_Area" localSheetId="0">Dashboard!$A$1:$L$34</definedName>${chartRanges}</definedNames><calcPr calcId="191029"/>${caches.length?`<pivotCaches>${caches.map(id=>`<pivotCache cacheId="${id}" r:id="rId${10+id}"/>`).join('')}</pivotCaches>`:''}</workbook>`);
 pkg.rels('','workbook.xml',[...names.map((_,i)=>[i+1,'worksheet',`worksheets/sheet${i+1}.xml`]),[6,'styles','styles.xml'],...caches.map(id=>[10+id,'pivotCacheDefinition',`pivotCache/pivotCacheDefinition${id}.xml`])]);
 return pkg.finish();
}
