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
  const name=encoder.encode(path),data=encoder.encode(declaration+source),checksum=crc32(data);
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

const moneyCell=value=>({value,style:2});
const dateCell=value=>({value:Date.parse(value)/86400000+25569,style:3});
const row=(label,value)=>[label,value];
const sheet=(name,heading,rows)=>({name,heading,rows});

export function dashboardWorkbook(data,generated){
 const sheets=[
  sheet('Resumen',['Indicador','Valor'],[
   row('Período',data.period),row('Fecha de referencia',dateCell(data.reference)),row('Generado (hora de Lima)',generated),
   ...(data.demo?[row('Origen','Demostración')]:[]),
   ...data.cards.map(card=>row(card.label,card.amount===undefined?card.value:moneyCell(card.amount))),
   ...data.activity,
   ...(data.recent?.orders?[row('Órdenes recientes','Últimas disponibles, sin filtro de período')]:[])
  ]),
  sheet('Estados de órdenes',['Estado','Órdenes'],data.states.map(item=>row(item.label,item.count))),
  sheet('Trabajo por mecánico',['Mecánico','Órdenes activas'],data.mechanics.map(item=>row(item.label,item.count)))
 ];
 if(data.income) sheets.push(sheet('Evolución de ingresos',['Intervalo','Ingresos netos (PEN)'],data.income.map(item=>row(item.month,moneyCell(item.cents/100)))));
 sheets.push(sheet('Servicios más realizados',['Servicio','Cantidad'],data.services.map(item=>row(item.label,item.count))));
 if(data.recent?.orders) sheets.push(sheet('Órdenes recientes',['N.º OT','Cliente','Placa','Estado'],data.recent.orders.map(order=>[
  order.code||order.id,order.clientName||'—',order.vehiclePlate||'—',order.state
 ])));
 const rel=(id,type,target)=>`<Relationship Id="rId${id}" Type="${officeRel}/${type}" Target="${target}"/>`;
 const files=[
  ['[Content_Types].xml',`<Types xmlns="${contentTypes}"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>${sheets.map((_,i)=>`<Override PartName="/xl/worksheets/sheet${i+1}.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`).join('')}</Types>`],
  ['_rels/.rels',`<Relationships xmlns="${packageRel}">${rel(1,'officeDocument','xl/workbook.xml')}</Relationships>`],
  ['xl/workbook.xml',`<workbook xmlns="${spreadsheet}" xmlns:r="${officeRel}"><bookViews><workbookView/></bookViews><sheets>${sheets.map(({name},i)=>`<sheet name="${xml(name)}" sheetId="${i+1}" r:id="rId${i+1}"/>`).join('')}</sheets></workbook>`],
  ['xl/_rels/workbook.xml.rels',`<Relationships xmlns="${packageRel}">${sheets.map((_,i)=>rel(i+1,'worksheet',`worksheets/sheet${i+1}.xml`)).join('')}${rel(sheets.length+1,'styles','styles.xml')}</Relationships>`],
  ['xl/styles.xml',`<styleSheet xmlns="${spreadsheet}"><numFmts count="2"><numFmt numFmtId="164" formatCode="&quot;S/ &quot;#,##0.00"/><numFmt numFmtId="165" formatCode="dd/mm/yyyy"/></numFmts><fonts count="2"><font><sz val="11"/><name val="Calibri"/></font><font><b/><color rgb="FFFFFFFF"/><sz val="11"/><name val="Calibri"/></font></fonts><fills count="3"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill><fill><patternFill patternType="solid"><fgColor rgb="FF79223A"/><bgColor indexed="64"/></patternFill></fill></fills><borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders><cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs><cellXfs count="4"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/><xf numFmtId="0" fontId="1" fillId="2" borderId="0" xfId="0" applyFont="1" applyFill="1"/><xf numFmtId="164" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1"/><xf numFmtId="165" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1"/></cellXfs><cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles></styleSheet>`]
 ];
 sheets.forEach(({heading,rows},index)=>{
  const dataRows=[heading,...rows].map((values,i)=>`<row r="${i+1}">${values.map((cell,j)=>{
   const value=cell?.value??cell,style=i===0?1:cell?.style??0,ref=String.fromCharCode(65+j)+(i+1);
   return typeof value==='number'
    ?`<c r="${ref}" s="${style}"><v>${value}</v></c>`
    :`<c r="${ref}" s="${style}" t="inlineStr"><is><t xml:space="preserve">${xml(value)}</t></is></c>`;
  }).join('')}</row>`).join('');
  files.push([`xl/worksheets/sheet${index+1}.xml`,`<worksheet xmlns="${spreadsheet}"><sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews><cols><col min="1" max="1" width="40" customWidth="1"/><col min="2" max="2" width="36" customWidth="1"/><col min="3" max="4" width="19" customWidth="1"/></cols><sheetData>${dataRows}</sheetData></worksheet>`]);
 });
 return workbookZip(files);
}
