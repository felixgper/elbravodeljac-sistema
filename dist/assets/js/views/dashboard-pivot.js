const extension='xmlns:x14="http://schemas.microsoft.com/office/spreadsheetml/2009/9/main"';
const dateISO=value=>new Date((value-25569)*86400000).toISOString().slice(0,19);

export function nativePivot({pkg,source,rows,index,spreadsheet,officeRel,xml,put,ref}){
 const {id,name,head,date,money}=source;
 const labels=[...new Set(source.rows.map(row=>row[0]))];
 const pivot=`pivotTables/pivotTable${id}.xml`;
 const dates='containsSemiMixedTypes="0" containsNonDate="0" containsDate="1" containsString="0"';
 const numeric='containsSemiMixedTypes="0" containsString="0" containsNumber="1"';
 const sharedLabels=labels.map(label=>date?`<d v="${dateISO(label)}"/>`:`<s v="${xml(label)}"/>`).join('');
 pkg.add(`pivotCache/pivotCacheDefinition${id}.xml`,'pivotCacheDefinition',`<pivotCacheDefinition xmlns="${spreadsheet}" xmlns:r="${officeRel}" r:id="rId1" createdVersion="5" refreshedVersion="7" minRefreshableVersion="3" recordCount="${source.rows.length}"><cacheSource type="worksheet"><worksheetSource name="${name}"/></cacheSource><cacheFields count="3"><cacheField name="${xml(head[0])}" numFmtId="${date?165:0}"><sharedItems ${date?dates:''} count="${labels.length}">${sharedLabels}</sharedItems></cacheField><cacheField name="${xml(head[1])}" numFmtId="${money?164:0}"><sharedItems ${numeric}/></cacheField><cacheField name="Referencia" numFmtId="165"><sharedItems ${dates}/></cacheField></cacheFields><extLst><ext ${extension} uri="{725AE2AE-9491-48be-B2B4-4EB974FC3084}"><x14:pivotCacheDefinition/></ext></extLst></pivotCacheDefinition>`);
 pkg.add(`pivotCache/pivotCacheRecords${id}.xml`,'pivotCacheRecords',`<pivotCacheRecords xmlns="${spreadsheet}" count="${source.rows.length}">${source.rows.map(([label,value,reference])=>`<r><x v="${labels.indexOf(label)}"/><n v="${value}"/><d v="${dateISO(reference)}"/></r>`).join('')}</pivotCacheRecords>`);
 pkg.rels('pivotCache',`pivotCacheDefinition${id}.xml`,[[1,'pivotCacheRecords',`pivotCacheRecords${id}.xml`]]);
 const col=index*4,start=4,end=start+labels.length+1;
 const dataCaption=`Total · ${head[1]}`;
 pkg.add(pivot,'pivotTable',`<pivotTableDefinition xmlns="${spreadsheet}" name="Pivot${name}" cacheId="${id}" dataCaption="Valores" updatedVersion="7" minRefreshableVersion="3" useAutoFormatting="1" createdVersion="5" outline="1" outlineData="1"><location ref="${ref(col,start)}:${ref(col+1,end)}" firstHeaderRow="1" firstDataRow="1" firstDataCol="1"/><pivotFields count="3"><pivotField axis="axisRow" showAll="0"${date?' numFmtId="165"':''}><items count="${labels.length+1}">${labels.map((_,i)=>`<item x="${i}"/>`).join('')}<item t="default"/></items></pivotField><pivotField dataField="1" showAll="0"/><pivotField numFmtId="165" showAll="0"/></pivotFields><rowFields count="1"><field x="0"/></rowFields><rowItems count="${labels.length+1}">${labels.map((_,i)=>`<i><x v="${i}"/></i>`).join('')}<i t="grand"><x/></i></rowItems><colItems count="1"><i/></colItems><dataFields count="1"><dataField name="${xml(dataCaption)}" fld="1" subtotal="sum" baseField="0" baseItem="0"${money?' numFmtId="164"':''}/></dataFields><pivotTableStyleInfo name="PivotStyleMedium3" showRowHeaders="1" showColHeaders="1"/><extLst><ext ${extension} uri="{962EF5D1-5CA2-4c93-8EF4-DBF5C05439D2}"><x14:pivotTableDefinition hideValuesRow="1"/></ext></extLst></pivotTableDefinition>`);
 pkg.rels('pivotTables',`pivotTable${id}.xml`,[[1,'pivotCacheDefinition',`../pivotCache/pivotCacheDefinition${id}.xml`]]);
 put(rows,2,col,source.title,1);
 put(rows,start,col,head[0],1);
 put(rows,start,col+1,dataCaption,1);
 labels.forEach((label,i)=>{
  put(rows,start+i+1,col,label,date?3:0);
  put(rows,start+i+1,col+1,source.rows.filter(row=>row[0]===label).reduce((sum,row)=>sum+row[1],0),money?2:0);
 });
 put(rows,end,col,'Total general',1);
 put(rows,end,col+1,source.rows.reduce((sum,row)=>sum+row[1],0),money?2:0);
 return id;
}
