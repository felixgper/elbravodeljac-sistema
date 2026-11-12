import {money} from '../components/ui.js';

const encoder=new TextEncoder();
const burgundy='.46 .12 .22', ink='.13 .14 .17', muted='.35 .36 .40', grid='.86 .86 .88';
const width=595,height=842,margin=42,right=553;

function literal(value){
 // Helvetica uses WinAnsi in the report. Escape PDF delimiters and encode Latin-1 as octal.
 return Array.from(String(value).replace(/[–—]/g,'-').replace(/\u00a0/g,' '),char=>{
  if('\\()'.includes(char))return '\\'+char;
  const code=char.codePointAt(0);
  if(code>=32&&code<127)return char;
  return code>=160&&code<=255?'\\'+code.toString(8).padStart(3,'0'):'?';
 }).join('');
}

async function logoJpeg(){
 const image=new Image();
 image.src=new URL('../../images/diseno19/emblema.svg',import.meta.url).href;
 await image.decode();
 const canvas=document.createElement('canvas');
 canvas.width=canvas.height=160;
 const ctx=canvas.getContext('2d');
 ctx.fillStyle='#751f38';ctx.fillRect(0,0,160,160);
 const scale=150/Math.max(image.naturalWidth,image.naturalHeight),w=image.naturalWidth*scale,h=image.naturalHeight*scale;
 ctx.drawImage(image,(160-w)/2,(160-h)/2,w,h);
 const blob=await new Promise(resolve=>canvas.toBlob(resolve,'image/jpeg',.92));
 return new Uint8Array(await blob.arrayBuffer());
}

function assemblePdf(pages,logo){
 const objects=[];
 const stream=(dict,body)=>[encoder.encode(`<< ${dict} /Length ${body.length} >>\nstream\n`),body,encoder.encode('\nendstream')];
 const pageIds=pages.map((_,i)=>6+i*2);
 objects.push(['<< /Type /Catalog /Pages 2 0 R >>'],
  [`<< /Type /Pages /Count ${pages.length} /Kids [${pageIds.map(id=>id+' 0 R').join(' ')}] >>`],
  ['<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>'],
  ['<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>'],
  stream('/Type /XObject /Subtype /Image /Width 160 /Height 160 /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode',logo));
 pages.forEach((commands,i)=>{
  objects.push([`<< /Type /Page /Parent 2 0 R /MediaBox [0 0 ${width} ${height}] /Resources << /Font << /F1 3 0 R /F2 4 0 R >> /XObject << /Logo 5 0 R >> >> /Contents ${pageIds[i]+1} 0 R >>`],
   stream('',encoder.encode(commands.join('\n'))));
 });
 const chunks=[],offsets=[];
 let length=0;
 const append=value=>{const bytes=typeof value==='string'?encoder.encode(value):value;chunks.push(bytes);length+=bytes.length;};
 append('%PDF-1.4\n');append(Uint8Array.of(37,226,227,207,211,10));
 objects.forEach((parts,i)=>{offsets.push(length);append(`${i+1} 0 obj\n`);parts.forEach(append);append('\nendobj\n');});
 const xref=length;
 append(`xref\n0 ${objects.length+1}\n0000000000 65535 f \n${offsets.map(offset=>String(offset).padStart(10,'0')+' 00000 n \n').join('')}trailer\n<< /Size ${objects.length+1} /Root 1 0 R >>\nstartxref\n${xref}\n%%EOF\n`);
 return new Blob(chunks,{type:'application/pdf'});
}

export async function dashboardPdf(data,generated){
 const pages=[];
 let page,y;
 const text=(value,x,top,size=11,bold=false,color=ink)=>page.push(`BT /${bold?'F2':'F1'} ${size} Tf ${color} rg 1 0 0 1 ${x} ${height-top} Tm (${literal(value)}) Tj ET`);
 const rect=(x,top,w,h,color)=>page.push(`${color} rg ${x} ${height-top-h} ${w} ${h} re f`);
 const line=(x1,y1,x2,y2,color=grid,stroke=1)=>page.push(`${color} RG ${stroke} w ${x1} ${height-y1} m ${x2} ${height-y2} l S`);
 const measure=document.createElement('canvas').getContext('2d');
 function wrap(value,maxWidth){
  measure.font='11px Arial';
  const rows=[];
  let current='';
  for(const char of String(value)){
   if(current&&measure.measureText(current+char).width>maxWidth){
    const lastSpace=current.lastIndexOf(' ');
    if(lastSpace>0){rows.push(current.slice(0,lastSpace));current=current.slice(lastSpace+1);}
    else{rows.push(current);current='';}
   }
   current+=char;
  }
  return [...rows,current];
 }
 function newPage(){
  page=[];pages.push(page);
  rect(0,0,width,94,burgundy);
  page.push('q 52 0 0 52 42 768 cm /Logo Do Q');
  text('EL BRAVO DEL JAC',108,40,12,true,'1 1 1');
  text('Resumen del taller',108,66,22,true,'1 1 1');
  text(`Período: ${data.period}   |   Referencia: ${data.reference}`,margin,118,11,true);
  text(`Generado: ${generated} (Lima)${data.demo?'   |   Demostración':''}`,margin,136,10,false,muted);
  y=160;
 }
 const section=title=>{text(title,margin,y,15,true,burgundy);y+=20;};
 function list(title,rows,{bars=false,empty='Sin registros en el período.'}={}){
  if(y+54>780)newPage();
  section(title);
  if(!rows.length){text(empty,margin,y,11,false,muted);y+=28;return;}
  const max=Math.max(1,...rows.map(row=>Number(row[1])||0));
  for(const [label,value] of rows){
   const labels=wrap(label,bars?245:350),rowHeight=Math.max(20,labels.length*14+6);
   if(y+rowHeight>780){newPage();section(title+' (continuación)');}
   labels.forEach((part,i)=>text(part,margin,y+i*14));
   if(bars){
    rect(302,y-8,145,8,grid);
    rect(302,y-8,145*value/max,8,burgundy);
   }
   measure.font='bold 11px Arial';
   text(value,right-measure.measureText(String(value)).width,y,11,true);
   line(margin,y+rowHeight-13,right,y+rowHeight-13);
   y+=rowHeight;
  }
  y+=10;
 }
 newPage();
 list('Indicadores',data.cards.map(card=>[card.label,card.value]));
 if(data.income){
  const points=data.income;
  if(y+180>780)newPage();
  section('Evolución de ingresos');
  if(points.some(point=>point.cents!==0)){
   const min=Math.min(0,...points.map(point=>point.cents)),max=Math.max(0,...points.map(point=>point.cents));
   const left=143,top=y,chartWidth=right-left-6,chartHeight=94;
   const x=i=>left+chartWidth*i/Math.max(1,points.length-1);
   const at=value=>top+chartHeight*(max-value)/(max-min);
   for(const value of new Set([min,0,max])){
    line(left,at(value),right,at(value));text(money(value),margin,at(value)+3,9,false,muted);
   }
   points.forEach((point,i)=>{
    const px=x(i),py=at(point.cents);
    if(i)line(x(i-1),at(points[i-1].cents),px,py,burgundy,1.5);
    rect(px-2,py-2,4,4,burgundy);
    if(i===0||i===points.length-1||i%Math.ceil(points.length/5)===0&&i<points.length-2)text(point.month,px-6,top+112,9,false,muted);
   });
   y+=135;
  }else{text('El ingreso neto del período es cero.',margin,y,11,false,muted);y+=28;}
 }
 list('Estados de órdenes',data.states.map(item=>[item.label,item.count]),{bars:true});
 list('Trabajo por mecánico',data.mechanics.map(item=>[item.label,item.count]),{bars:true,empty:'Sin órdenes activas asignadas.'});
 list('Servicios más realizados',data.services.map(item=>[item.label,item.count]),{bars:true});
 if(data.recent?.orders)list('Órdenes recientes (sin filtro de período)',data.recent.orders.map(order=>[
  [order.code||order.id,order.clientName||'—',order.vehiclePlate||'—'].join(' · '),order.state
 ]));
 list('Actividad del taller',data.activity);
 pages.forEach((commands,index)=>{page=commands;line(margin,798,right,798);text('El Bravo del JAC',margin,816,9,false,muted);text(`Página ${index+1} de ${pages.length}`,478,816,9,false,muted);});
 return assemblePdf(pages,await logoJpeg());
}
