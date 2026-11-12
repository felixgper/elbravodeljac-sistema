const encoder=new TextEncoder();

function cloneAppearance(source){
 const copy=source.cloneNode(true);
 const originals=[source,...source.querySelectorAll('*')];
 const targets=[copy,...copy.querySelectorAll('*')];
 originals.forEach((element,index)=>{
  const computed=getComputedStyle(element),target=targets[index];
  for(const key of computed){
   // SVG references must resolve inside the captured document.
   const value=computed.getPropertyValue(key).replace(/url\(["']?[^)"']*#([^)"']+)["']?\)/g,'url(#$1)');
   target.style.setProperty(key,value);
  }
 });
 const {width}=source.getBoundingClientRect();
 Object.assign(copy.style,{width:`${width}px`,maxWidth:'none',margin:'0',boxSizing:'border-box',position:'static'});
 return copy;
}

async function embedImages(root){
 for(const img of root.querySelectorAll('img')){
  const url=new URL(img.getAttribute('src'),document.baseURI);
  if(url.origin!==location.origin)throw Error('El reporte requiere imágenes locales.');
  const response=await fetch(url);
  if(!response.ok)throw Error('No se pudo cargar el logotipo.');
  img.removeAttribute('srcset');
  img.src=await new Promise((resolve,reject)=>{
   const reader=new FileReader();
   reader.onload=()=>resolve(reader.result);
   reader.onerror=reject;
   response.blob().then(blob=>reader.readAsDataURL(blob),reject);
  });
 }
}

async function captureSummary(summary,data,generated){
 const workspace=summary.closest('.workspace');
 const title=workspace.querySelector('.page-heading>div');
 const brand=workspace.closest('.app-shell').querySelector('.brand-emblem');
 const {width,height}=summary.getBoundingClientRect();
 if(!width||!height||summary.hidden)throw Error('Abre el resumen antes de exportar.');
 const titleHeight=title.getBoundingClientRect().height;
 const padding=24,headerHeight=Math.max(58,titleHeight)+22,footerHeight=width<520?68:40;
 const pageWidth=Math.ceil(width+padding*2),pageHeight=Math.ceil(padding+headerHeight+height+footerHeight);
 const sheet=document.createElement('div');
 sheet.setAttribute('xmlns','http://www.w3.org/1999/xhtml');
 Object.assign(sheet.style,{width:`${pageWidth}px`,height:`${pageHeight}px`,boxSizing:'border-box',
  padding:`${padding}px`,background:getComputedStyle(workspace).backgroundColor,
  color:getComputedStyle(workspace).color,fontFamily:getComputedStyle(workspace).fontFamily});
 const header=document.createElement('div');
 Object.assign(header.style,{height:`${headerHeight}px`,display:'flex',gap:'12px',alignItems:'flex-start'});
 const emblem=brand.cloneNode();
 Object.assign(emblem.style,{width:'56px',height:'58px',objectFit:'contain'});
 const titleCopy=cloneAppearance(title);
 Object.assign(titleCopy.style,{width:'auto',minWidth:'0'});
 header.append(emblem,titleCopy);
 const content=cloneAppearance(summary);
 const footer=document.createElement('p');
 footer.textContent=`${data.period} · Fecha de referencia: ${data.reference.split('-').reverse().join('/')} · ${data.demo?'Demostración · ':''}Generado: ${generated} (Lima)`;
 Object.assign(footer.style,{margin:'10px 0 0',fontSize:'12px',color:'#c2c3cc'});
 sheet.append(header,content,footer);
 await embedImages(sheet);
 const markup=new XMLSerializer().serializeToString(sheet);
 const svg=`<svg xmlns="http://www.w3.org/2000/svg" width="${pageWidth}" height="${pageHeight}"><foreignObject width="100%" height="100%">${markup}</foreignObject></svg>`;
 const image=new Image();
 image.src='data:image/svg+xml;charset=utf-8,'+encodeURIComponent(svg);
 await image.decode();
 const scale=Math.min(3,8192/Math.max(pageWidth,pageHeight));
 const canvas=document.createElement('canvas');
 canvas.width=Math.round(pageWidth*scale);
 canvas.height=Math.round(pageHeight*scale);
 const ctx=canvas.getContext('2d');
 ctx.fillStyle=sheet.style.background;
 ctx.fillRect(0,0,canvas.width,canvas.height);
 ctx.drawImage(image,0,0,canvas.width,canvas.height);
 const jpeg=await new Promise(resolve=>canvas.toBlob(resolve,'image/jpeg',.97));
 if(!jpeg)throw Error('No se pudo capturar el resumen.');
 return {jpeg,pixelsWide:canvas.width,pixelsHigh:canvas.height,width:pageWidth*.75,height:pageHeight*.75};
}

async function makePdf({jpeg,pixelsWide,pixelsHigh,width,height}){
 const stream=(dict,bytes)=>[`<< ${dict} /Length ${bytes.length} >>\nstream\n`,bytes,'\nendstream'];
 const commands=encoder.encode(`q ${width} 0 0 ${height} 0 0 cm /Summary Do Q`);
 const objects=[
  ['<< /Type /Catalog /Pages 2 0 R >>'],
  ['<< /Type /Pages /Count 1 /Kids [3 0 R] >>'],
  [`<< /Type /Page /Parent 2 0 R /MediaBox [0 0 ${width} ${height}] /Resources << /XObject << /Summary 4 0 R >> >> /Contents 5 0 R >>`],
  stream(`/Type /XObject /Subtype /Image /Width ${pixelsWide} /Height ${pixelsHigh} /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode`,new Uint8Array(await jpeg.arrayBuffer())),
  stream('',commands)
 ];
 const chunks=[],offsets=[];
 let length=0;
 const append=value=>{const bytes=typeof value==='string'?encoder.encode(value):value;chunks.push(bytes);length+=bytes.length;};
 append('%PDF-1.4\n');append(Uint8Array.of(37,226,227,207,211,10));
 objects.forEach((parts,i)=>{offsets.push(length);append(`${i+1} 0 obj\n`);parts.forEach(append);append('\nendobj\n');});
 const xref=length;
 append(`xref\n0 ${objects.length+1}\n0000000000 65535 f \n${offsets.map(offset=>String(offset).padStart(10,'0')+' 00000 n \n').join('')}trailer\n<< /Size ${objects.length+1} /Root 1 0 R >>\nstartxref\n${xref}\n%%EOF\n`);
 return new Blob(chunks,{type:'application/pdf'});
}

export async function dashboardPdf(summary,data,generated){
 return makePdf(await captureSummary(summary,data,generated));
}
