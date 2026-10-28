import {request,realMode} from '../services/api-service.js';
import {e,hint,showModal} from './ui.js';

export function quotePDFButton(quote){
 if(!realMode)return '';
 return `<button type="button" class="secondary-button" data-quote-pdf="${e(quote.id)}" aria-label="Descargar PDF de cotización ${e(quote.id)}"><svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true"><path d="M12 3v12m-5-5 5 5 5-5M4 16v5h16v-5"/></svg> PDF</button>`;
}
export function bindQuoteDownloads(root){
 root.querySelectorAll('[data-quote-pdf]').forEach(button=>button.onclick=async()=>{
  button.disabled=true;button.setAttribute('aria-busy','true');
  try{
   const id=button.dataset.quotePdf,blob=await request('/cotizaciones/'+encodeURIComponent(id)+'/pdf',{responseType:'blob'}),url=URL.createObjectURL(blob),a=document.createElement('a');
   a.href=url;a.download='cotizacion-'+id+'.pdf';document.body.append(a);a.click();a.remove();setTimeout(()=>URL.revokeObjectURL(url),1000);
  }catch(error){showModal('No se pudo descargar la cotización',hint(error.message));}
  finally{button.disabled=false;button.removeAttribute('aria-busy');}
 });
}
