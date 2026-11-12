const chartNS='http://schemas.openxmlformats.org/drawingml/2006/chart';
const drawingNS='http://schemas.openxmlformats.org/drawingml/2006/spreadsheetDrawing';
const shapeNS='http://schemas.openxmlformats.org/drawingml/2006/main';
const relationshipNS='http://schemas.openxmlformats.org/officeDocument/2006/relationships';
const palette=['BF5270','628FC3','5FA996','D2A566','9786BB'];
const fill=color=>`<a:solidFill><a:srgbClr val="${color}"/></a:solidFill>`;
const style=(color,border='DFE2E8')=>`<c:spPr>${fill(color)}<a:ln>${fill(border)}</a:ln></c:spPr>`;
const textColor=(size=1000)=>`<c:txPr><a:bodyPr/><a:lstStyle/><a:p><a:pPr><a:defRPr sz="${size}">${fill('454853')}</a:defRPr></a:pPr><a:endParaRPr/></a:p></c:txPr>`;

function chart(source,xml){
 const {rows,kind,color,head,date,money,title,id}=source;
 const horizontal=kind==='horizontal';
 const range=index=>`Datos!Serie${id}_${index}`;
 const points=(index)=>rows.map((row,i)=>`<c:pt idx="${i}"><c:v>${xml(row[index])}</c:v></c:pt>`).join('');
 const dateFormat=source.monthly?'[$-es-PE]mmm':'dd/mm';
 const categories=date?`<c:numRef><c:f>${range(0)}</c:f><c:numCache><c:formatCode>${dateFormat}</c:formatCode><c:ptCount val="${rows.length}"/>${points(0)}</c:numCache></c:numRef>`:
  `<c:strRef><c:f>${range(0)}</c:f><c:strCache><c:ptCount val="${rows.length}"/>${points(0)}</c:strCache></c:strRef>`;
 const valueFormat=money?'&quot;S/ &quot;#,##0.00':'General';
 const values=`<c:numRef><c:f>${range(1)}</c:f><c:numCache><c:formatCode>${valueFormat}</c:formatCode><c:ptCount val="${rows.length}"/>${points(1)}</c:numCache></c:numRef>`;
 const colors=source.colors??palette;
 const pointsStyle=kind==='doughnut'||horizontal||source.multicolor?rows.map((_,i)=>{
  const pointColor=colors[i%colors.length];
  return `<c:dPt><c:idx val="${i}"/>${style(pointColor,pointColor)}</c:dPt>`;
 }).join(''):'';
 const series=`<c:ser><c:idx val="0"/><c:order val="0"/><c:tx><c:v>${xml(head[1])}</c:v></c:tx>${kind==='line'?`<c:spPr><a:ln w="25000">${fill(color)}</a:ln></c:spPr><c:marker><c:symbol val="circle"/><c:size val="5"/>${style(color,color)}</c:marker>`:style(color,color)}${pointsStyle}<c:cat>${categories}</c:cat><c:val>${values}</c:val>${kind==='line'?'<c:smooth val="0"/>':''}</c:ser>`;
 const axisIds='<c:axId val="400"/><c:axId val="500"/>';
 const axis=(tag,id,position,cross,extra,labels='')=>`<c:${tag}><c:axId val="${id}"/><c:scaling><c:orientation val="${horizontal&&tag==='catAx'?'maxMin':'minMax'}"/></c:scaling><c:delete val="0"/><c:axPos val="${position}"/>${extra}<c:majorTickMark val="none"/><c:minorTickMark val="none"/><c:tickLblPos val="nextTo"/>${textColor(850)}<c:crossAx val="${cross}"/><c:crosses val="autoZero"/>${tag==='valAx'?'<c:crossBetween val="between"/>':labels}</c:${tag}>`;
 const axes=axis('catAx',400,horizontal?'l':'b',500,date?`<c:numFmt formatCode="${dateFormat}" sourceLinked="0"/>`:'',date&&rows.length>14?`<c:tickLblSkip val="${Math.ceil(rows.length/9)}"/>`:'')+
  axis('valAx',500,horizontal?'b':'l',400,`<c:majorGridlines><c:spPr><a:ln>${fill('E6E8ED')}</a:ln></c:spPr></c:majorGridlines><c:numFmt formatCode="${valueFormat}" sourceLinked="0"/>`);
 const plot=kind==='line'?`<c:lineChart><c:grouping val="standard"/><c:varyColors val="0"/>${series}${axisIds}</c:lineChart>${axes}`:kind==='bar'||horizontal?
  `<c:barChart><c:barDir val="${horizontal?'bar':'col'}"/><c:grouping val="clustered"/>${series}${horizontal?'<c:dLbls><c:dLblPos val="outEnd"/><c:showLegendKey val="0"/><c:showVal val="1"/><c:showCatName val="0"/><c:showSerName val="0"/><c:showPercent val="0"/><c:showBubbleSize val="0"/></c:dLbls>':''}<c:gapWidth val="65"/>${axisIds}</c:barChart>${axes}`:
  `<c:doughnutChart><c:varyColors val="1"/>${series}<c:holeSize val="58"/></c:doughnutChart>`;
 return `<c:chartSpace xmlns:c="${chartNS}" xmlns:a="${shapeNS}"><c:lang val="es-PE"/><c:chart><c:title><c:tx><c:rich><a:bodyPr/><a:lstStyle/><a:p><a:r><a:rPr sz="1300" b="1">${fill('292B32')}</a:rPr><a:t>${xml(title)}</a:t></a:r></a:p></c:rich></c:tx><c:overlay val="0"/></c:title><c:plotArea><c:layout/>${plot}${style('FFFFFF','FFFFFF')}</c:plotArea>${kind==='doughnut'?`<c:legend><c:legendPos val="b"/><c:overlay val="0"/>${textColor(800)}</c:legend>`:''}<c:plotVisOnly val="1"/><c:dispBlanksAs val="gap"/></c:chart>${style('FFFFFF')}</c:chartSpace>`;
}

async function logoPng(){
 const img=new Image();
 img.src=new URL('../../images/diseno19/emblema.svg',import.meta.url).href;
 await img.decode();
 const canvas=document.createElement('canvas');
 canvas.width=96;canvas.height=100;
 canvas.getContext('2d').drawImage(img,0,0,96,100);
 const image=await new Promise(resolve=>canvas.toBlob(resolve,'image/png'));
 if(!image)throw Error('No se pudo cargar el logo de Excel.');
 return new Uint8Array(await image.arrayBuffer());
}

const point=(tag,col,row,offset=12000)=>`<xdr:${tag}><xdr:col>${col}</xdr:col><xdr:colOff>${offset}</xdr:colOff><xdr:row>${row}</xdr:row><xdr:rowOff>12000</xdr:rowOff></xdr:${tag}>`;
export async function nativeCharts({pkg,sources,id,firstRow,xml,logo=false,positions}){
 const anchors=[],relationships=[];
 if(logo){
  pkg.files.push(['xl/media/emblema.png',await logoPng()]);
  relationships.push([1,'image','../media/emblema.png']);
  anchors.push(`<xdr:oneCellAnchor>${point('from',0,0)}<xdr:ext cx="520000" cy="545000"/><xdr:pic><xdr:nvPicPr><xdr:cNvPr id="1" name="Logo oficial"/><xdr:cNvPicPr/></xdr:nvPicPr><xdr:blipFill><a:blip r:embed="rId1"/><a:stretch><a:fillRect/></a:stretch></xdr:blipFill><xdr:spPr><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></xdr:spPr></xdr:pic><xdr:clientData/></xdr:oneCellAnchor>`);
 }
 sources.forEach((source,index)=>{
  const key=index+2,chartId=id*10+key,row=firstRow+Math.floor(index/2)*12;
  const wide=sources[0]?.kind==='line'&&index<2;
  const col=index%2*(wide?8:6),span=wide?(index===0?8:4):6;
  const [left,top,right,bottom]=positions?.[source.name]??[col,row,col+span,row+11];
  pkg.add(`charts/chart${chartId}.xml`,'chart',chart(source,xml),'drawingml');
  relationships.push([key,'chart',`../charts/chart${chartId}.xml`]);
  anchors.push(`<xdr:twoCellAnchor editAs="oneCell">${point('from',left,top,left?59000:12000)}${point('to',right,bottom,0)}<xdr:graphicFrame macro=""><xdr:nvGraphicFramePr><xdr:cNvPr id="${key}" name="${xml(source.title)}"/><xdr:cNvGraphicFramePr/></xdr:nvGraphicFramePr><xdr:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/></xdr:xfrm><a:graphic><a:graphicData uri="${chartNS}"><c:chart r:id="rId${key}"/></a:graphicData></a:graphic></xdr:graphicFrame><xdr:clientData/></xdr:twoCellAnchor>`);
 });
 if(!anchors.length)return '';
 pkg.add(`drawings/drawing${id}.xml`,'drawing',`<xdr:wsDr xmlns:xdr="${drawingNS}" xmlns:c="${chartNS}" xmlns:a="${shapeNS}" xmlns:r="${relationshipNS}">${anchors.join('')}</xdr:wsDr>`,'');
 pkg.rels('drawings',`drawing${id}.xml`,relationships);
 pkg.rels('worksheets',`sheet${id===1?1:4}.xml`,[[1,'drawing',`../drawings/drawing${id}.xml`]]);
 return '<drawing r:id="rId1"/>';
}
