import {realMode} from './services/api-service.js';
import {loadModule} from './services/real-workshop.js';
import {clearRealStore} from './services/real-state.js';
import { icon } from './components/icons.js';
import { modules } from './components/navigation.js';
import { incomeChart, donutChart, verticalBars, horizontalBars } from './components/charts.js';
import { refreshSession, getSession, getRememberedUser, login, logout, demoCredentials } from './services/auth-service.js';
import { getDashboard, getActiveOrders } from './services/dashboard-service.js';
import { renderClients, renderVehicles } from './views/clients-vehicles.js';
import { renderReceptions, renderDiagnoses, renderQuotes, renderOrders, renderRescues } from './views/operations.js';
import { renderPayments } from './views/payments.js';
import { renderMechanics, renderSuppliers, renderUsers, renderReports } from './views/administration.js';
import { getStore, resetDemo } from './services/store-service.js';
import { permissionChecker, homePath } from './services/permissions.js';
import { notify, showModal, form, submit, button, bindForm, go } from './components/ui.js';

const root = document.getElementById('app');
const escapeHtml = value => String(value).replace(/[&<>"']/g, char => ({'&':'&amp;', '<':'&lt;', '>':'&gt;', '"':'&quot;', "'":'&#39;'}[char]));
const currency = cents => new Intl.NumberFormat('es-PE', {style:'currency',currency:'PEN',maximumFractionDigits:0}).format(cents / 100);
let renderId = 0;
let chartObserver;
let disposeShell;
let allowed = modules;

function brand() {
  return `<span class="atelier-brand" role="img" aria-label="El Bravo del JAC · Taller automotriz"><img class="brand-bull" src="./assets/images/bravo-borgona.svg" alt="" width="72" height="62"><span class="brand-wordmark">EL BRAVO DEL JAC<small>TALLER AUTOMOTRIZ</small></span></span>`;
}
function renderLogin() {
  root.setAttribute('aria-busy','false');root.dataset.route='#/login';
  const remembered = getRememberedUser();
  document.title = 'Iniciar sesión · El Bravo del JAC';
  root.innerHTML = `<main class="login-page" id="main-content"><div class="login-glow" aria-hidden="true"></div>
    <section class="login-card" aria-labelledby="login-title">
      <div class="login-brand">${brand()}</div>
      <div class="login-heading"><span class="eyebrow">GESTIÓN DEL TALLER</span><h1 id="login-title">Iniciar sesión</h1><p>Acceso exclusivo para trabajadores del taller</p></div>
      <form id="login-form">
        <label class="field-label" for="username">Correo o usuario</label><div class="input-wrap">${icon('user')}<input id="username" name="username" autocomplete="username" placeholder="Ingrese su usuario" value="${escapeHtml(remembered)}" required maxlength="${realMode?150:120}" autocapitalize="none" spellcheck="false"></div>
        <label class="field-label" for="password">Contraseña</label><div class="input-wrap">${icon('lock')}<input id="password" name="password" type="password" autocomplete="current-password" placeholder="Ingrese su contraseña" required maxlength="${realMode?72:128}"><button type="button" class="password-toggle" id="password-toggle" aria-label="Mostrar contraseña" aria-pressed="false">${icon('eye')}</button></div>
        <label class="remember"><input type="checkbox" name="remember" ${remembered ? 'checked' : ''}> Recordarme</label>
        <p id="login-error" class="form-error" role="alert" hidden></p>
        <button class="primary-button login-submit" type="submit">Iniciar sesión ${icon('arrow')}</button>
      </form>
      <div class="demo-access">${icon('info')}<div><strong>${realMode?'Acceso al taller':'Acceso de demostración'}</strong>${realMode?'<p>Modo real · Usa tu cuenta del taller. Los cambios se guardan en el servidor. <a href="/demo/">Abrir demostración separada</a></p>':`<p>Usuario: <code>${demoCredentials.username}</code> <span>·</span> Contraseña: <code>${demoCredentials.password}</code></p><p>Auxiliar: <code>kevin</code> · <code>Auxiliar2026!</code></p>`}</div></div>
      <p class="login-hint">El menú del sistema se muestra después de iniciar sesión.</p>
    </section><footer class="login-footer">EL BRAVO DEL JAC <span>·</span> Proyecto universitario <span>·</span> Etapa 1</footer>
  </main>`;
  document.getElementById('password-toggle').addEventListener('click', event => {
    const input = document.getElementById('password'), button = event.currentTarget;
    const visible = input.type === 'password'; input.type = visible ? 'text' : 'password';
    button.setAttribute('aria-label', visible ? 'Ocultar contraseña' : 'Mostrar contraseña');
    button.setAttribute('aria-pressed', String(visible)); button.innerHTML = icon(visible ? 'eyeOff' : 'eye');
  });
  document.getElementById('login-form').addEventListener('submit', async event => {
    event.preventDefault(); const form = event.currentTarget;
    const error = document.getElementById('login-error'), button = form.querySelector('[type="submit"]');
    error.hidden = true; button.disabled = true; button.textContent = 'Ingresando…';
    try {
      const session = await login({username:form.elements.username.value,password:form.elements.password.value,remember:form.elements.remember.checked});
      go(homePath(session.role));
    } catch (reason) {
      error.textContent = reason.message; error.hidden = false; button.disabled = false; button.innerHTML = `Iniciar sesión ${icon('arrow')}`;
    }
  });
}
function shell(session, module) {
  const date = new Intl.DateTimeFormat('es-PE', {day:'numeric',month:'long',year:'numeric',timeZone:'America/Lima'}).format(new Date());
  root.innerHTML = `<div class="app-shell">
    <header class="topbar"><div class="topbar-brand"><button class="icon-button menu-button" id="menu-toggle" aria-label="Abrir o contraer menú" aria-expanded="false" aria-controls="sidebar">${icon('menu')}</button><a class="brand-link" href="#/inicio" aria-label="Inicio de El Bravo del JAC">${brand()}</a></div>
      <div class="module-search"><label class="sr-only" for="module-search">Buscar módulo</label>${icon('search')}<input id="module-search" role="combobox" aria-autocomplete="list" aria-haspopup="listbox" placeholder="Buscar módulo…" autocomplete="off" aria-controls="search-results" aria-expanded="false"><kbd>/</kbd><div id="search-results" class="search-results" role="listbox" aria-label="Módulos encontrados" hidden></div></div>
      <span class="topbar-demo"><span></span>${realMode?'Modo real':'Demostración'}</span><div class="profile-wrap"><button class="profile-button" id="profile-toggle" aria-label="Opciones de ${escapeHtml(session.name)}" aria-expanded="false" aria-controls="profile-menu"><span class="profile-avatar">${escapeHtml(session.name.trim().charAt(0).toUpperCase())}</span><span>${escapeHtml(session.name)}</span>${icon('down')}</button><div class="profile-menu" id="profile-menu" hidden><span>${escapeHtml(session.role)}</span><button id="export-demo" ${realMode?'hidden':''}>${icon('quote')} Descargar datos demo</button><button id="reset-demo" ${realMode?'hidden':''}>${icon('info')} Reiniciar datos demo</button><button id="logout">${icon('logout')} Cerrar sesión</button></div></div>
    </header>
    <button class="sidebar-scrim" id="sidebar-scrim" aria-label="Cerrar menú" hidden></button>
    <aside class="sidebar" id="sidebar" aria-label="Menú principal"><a class="sidebar-brand" href="#/${homePath(session.role)}" aria-label="Inicio de El Bravo del JAC">${brand()}</a><div class="sidebar-caption">GESTIÓN DEL TALLER</div><nav>${allowed.map(m => `${m.id === 'mecanicos' && allowed.indexOf(m) > 0 ? '<div class="nav-divider"></div>' : ''}<a class="nav-item ${m.id === module.id ? 'active' : ''}" href="#/${m.id}" ${m.id === module.id ? 'aria-current="page"' : ''} title="${m.label}">${icon(m.icon)}<span>${m.label}</span>${!m.ready ? '<span class="nav-pending" aria-label="Próxima etapa"></span>' : ''}</a>`).join('')}</nav><div class="sidebar-footer"><span>EL BRAVO DEL JAC</span><small>Cuidado experto. En cada detalle.</small></div></aside>
    <main class="workspace" id="main-content"><div class="page-heading"><div><div class="breadcrumb">El Bravo del JAC ${icon('chevron')} <span>${module.id === 'inicio' ? 'Dashboard' : module.label}</span></div><h1>${module.id === 'inicio' ? `Bienvenido, ${escapeHtml(session.name)}` : module.label}</h1><p>${module.id === 'inicio' ? 'Panel de control del taller' : 'Sistema de gestión del taller automotriz'}</p></div><div class="page-date">${icon('clock')}<span>${date}</span></div></div><div id="page-content"></div><footer class="workspace-footer"><span>El Bravo del JAC · Gestión del taller</span><span>${realMode?'Modo real':'Frontend demo'} <span class="footer-dot">•</span> ${realMode?'Datos del taller':'Datos ilustrativos'}</span></footer></main>
  </div>`;
  bindShell();
}
function bindShell() {
  const app = root.querySelector('.app-shell'), toggle = document.getElementById('menu-toggle'), scrim = document.getElementById('sidebar-scrim');
  const sidebar = document.getElementById('sidebar');
  const menuMedia = matchMedia('(max-width: 900px)');
  sidebar.inert = menuMedia.matches;
  toggle.setAttribute('aria-expanded', String(!menuMedia.matches));
  function adjustMenu() { closeMenu(); toggle.setAttribute('aria-expanded',String(!menuMedia.matches)); }
  menuMedia.addEventListener('change',adjustMenu);
  disposeShell = () => menuMedia.removeEventListener('change',adjustMenu);
  const profileButton = document.getElementById('profile-toggle'), profileMenu = document.getElementById('profile-menu');
  const search = document.getElementById('module-search'), results = document.getElementById('search-results');
  function closeMenu() { app.classList.remove('sidebar-open'); scrim.hidden = true; toggle.setAttribute('aria-expanded','false'); sidebar.inert = matchMedia('(max-width: 900px)').matches; }
  toggle.addEventListener('click', () => {
    if (matchMedia('(max-width: 900px)').matches) {
      const open = app.classList.toggle('sidebar-open'); scrim.hidden = !open; sidebar.inert = !open; toggle.setAttribute('aria-expanded',String(open));
      if (open) document.querySelector('.nav-item.active').focus();
    } else {
      const compact = app.classList.toggle('sidebar-compact'); toggle.setAttribute('aria-expanded', String(!compact));
    }
  });
  scrim.addEventListener('click', () => { closeMenu(); toggle.focus(); });
  profileButton.addEventListener('click', () => { const open = profileMenu.hidden; profileMenu.hidden = !open; profileButton.setAttribute('aria-expanded', String(open)); });
  document.getElementById('export-demo').addEventListener('click', () => {
    if(realMode)return;
    const data = getStore(); data.users = data.users.map(({passwordDigest,...user}) => user);
    const url = URL.createObjectURL(new Blob([JSON.stringify(data,null,2)],{type:'application/json'}));
    const a = document.createElement('a'); a.href=url; a.download='El_Bravo_del_JAC_Datos_Demo.json'; a.click(); setTimeout(()=>URL.revokeObjectURL(url),1000); notify('Datos de demostración descargados.');
  });
  document.getElementById('reset-demo').addEventListener('click', () => {
    if(realMode)return;
    const dialog=showModal('Reiniciar datos de demostración',form('CONFIRMAR REINICIO','<p class="full-width">Se reemplazarán los registros de este navegador por los ejemplos iniciales. Puedes descargar tus datos antes de continuar.</p>',submit('Reiniciar datos')+button('Cancelar','cancel'),'reset-form'));
    dialog.querySelector('[data-action="cancel"]').onclick=()=>dialog.close();
    bindForm('reset-form',()=>{resetDemo();dialog.close();notify('Datos de demostración reiniciados.');go('inicio');});
  });
  document.getElementById('logout').addEventListener('click', async () => { try {await logout(); location.hash = '#/login';}catch(error){notify(error.message);} });
  search.addEventListener('input', () => {
    const query = search.value.trim().normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase();
    const matches = allowed.filter(m => m.label.normalize('NFD').replace(/[\u0300-\u036f]/g,'').toLowerCase().includes(query));
    results.innerHTML = matches.length ? matches.map(m => `<a role="option" href="#/${m.id}">${icon(m.icon)}<span>${m.label}</span>${!m.ready ? '<small>Próxima etapa</small>' : ''}</a>`).join('') : '<p role="option" aria-disabled="true">No se encontraron módulos.</p>';
    results.hidden = !query; search.setAttribute('aria-expanded',String(Boolean(query)));
  });
  search.addEventListener('keydown', event => { if (event.key === 'ArrowDown' && !results.hidden) { event.preventDefault(); results.querySelector('a')?.focus(); } });
  root.onkeydown = event => {
    if (event.key === 'Escape') { closeMenu(); results.hidden = true; search.setAttribute('aria-expanded','false'); profileMenu.hidden = true; profileButton.setAttribute('aria-expanded','false'); toggle.focus(); }
    if (event.key === 'Tab' && app.classList.contains('sidebar-open')) {
      const links = [...document.querySelectorAll('.nav-item')];
      if (!event.shiftKey && document.activeElement === links.at(-1)) { event.preventDefault(); toggle.focus(); }
      if (event.shiftKey && document.activeElement === toggle) { event.preventDefault(); links.at(-1).focus(); }
    }
    if (event.key === '/' && !['INPUT','TEXTAREA'].includes(event.target.tagName)) { event.preventDefault(); search.focus(); }
  };
  root.onclick = event => {
    if (!event.target.closest('.profile-wrap')) { profileMenu.hidden = true; profileButton.setAttribute('aria-expanded','false'); }
    if (!event.target.closest('.module-search')) { results.hidden = true; search.setAttribute('aria-expanded','false'); }
  };
}
async function renderDashboard(currentRender) {
  const content = document.getElementById('page-content');
  content.innerHTML = '<div class="loading-state" role="status">Cargando resumen…</div>';
  try {
    const data = await getDashboard(); if (currentRender !== renderId) return;
    const cards = [
      {label:'Ingresos del mes',value:currency(data.monthlyIncomeCents),icon:'wallet',class:'copper',note:data.currentMonth},
      {label:'Órdenes activas',value:getActiveOrders(data),icon:'wrench',class:'copper',note:'Pendientes, reparación y pruebas'},
      {label:'Servicios finalizados',value:data.completedServices,icon:'check',class:'green',note:'Atenciones concluidas'},
      {label:'Auxilios de hoy',value:data.rescuesToday,icon:'rescue',class:'blue',note:realMode?'Solicitudes registradas':'Solicitudes de demostración'},
    ];
    content.innerHTML = `<div class="dashboard-note"><span class="live-dot"></span>Dashboard general <span class="note-divider">/</span><span>${realMode?'Datos del servidor':'Datos de este navegador'} · ${escapeHtml(data.period)}</span></div>
      <section class="kpi-grid" aria-label="Resumen de actividad">${cards.map(card => `<article class="kpi-card"><span class="kpi-icon ${card.class}">${icon(card.icon)}</span><div><strong class="kpi-value">${card.value}</strong><h2>${card.label}</h2><p>${card.note}</p></div></article>`).join('')}</section>
      <div class="main-charts"><section class="panel income-panel"><div class="panel-heading"><div><h2>Ingresos de los últimos 6 meses</h2><p>Evolución de ingresos registrados</p></div><span class="panel-chip">${escapeHtml(String(new Date().getFullYear()))}</span></div>${incomeChart(data.income)}<div class="chart-caption"><span>${escapeHtml(data.period)}</span><strong>Último mes: ${currency(data.monthlyIncomeCents)}</strong></div></section>
      <section class="panel states-panel"><div class="panel-heading"><div><h2>Estados del servicio</h2><p>Distribución de las órdenes de trabajo</p></div>${icon('orders')}</div>${donutChart(data.states)}</section></div>
      <div class="secondary-charts"><section class="panel"><div class="panel-heading"><div><h2>Órdenes por estado</h2><p>En proceso = reparación + pruebas</p></div></div>${verticalBars(data.states)}</section><section class="panel"><div class="panel-heading"><div><h2>Trabajo por mecánico</h2><p>Distribución del trabajo registrado</p></div></div>${horizontalBars(data.mechanics,'mechanics')}</section><section class="panel"><div class="panel-heading"><div><h2>Servicios más solicitados</h2><p>Atenciones por tipo de servicio</p></div></div>${horizontalBars(data.services,'services')}</section></div>
      <section class="workshop-banner" aria-label="El Bravo del JAC, servicio especializado en vehículos JAC"><div class="banner-crop"><img src="./assets/images/referencia-cobre-carbon.png" alt="Camioneta JAC y mecánico del taller. Diagnóstico, mantenimiento, repuestos y auxilio mecánico." loading="lazy"></div></section>`;
    const panel = content.querySelector('.income-panel');
    let lastWidth = 0;
    function fitIncomeChart() {
      const chart = panel.querySelector('.income-chart');
      const width = Math.floor(chart.getBoundingClientRect().width);
      if (width !== lastWidth && width > 0) { lastWidth = width; chart.outerHTML = incomeChart(data.income,width); }
    }
    fitIncomeChart();
    chartObserver = new ResizeObserver(fitIncomeChart); chartObserver.observe(panel);
  } catch {
    if (currentRender === renderId) content.innerHTML = '<section class="panel error-state" role="alert"><h2>No se pudo cargar el resumen</h2><p>Vuelve a Inicio para intentarlo nuevamente.</p><a class="primary-button" href="#/inicio">Volver a Inicio</a></section>';
  }
}
// Módulos cuyos formularios guardan datos: sin permiso, el formulario queda en solo lectura.
const formGates = {recepciones:['crear'], diagnosticos:['crear'], cotizaciones:['crear','editar'], auxilio:['crear']};
function gateForms(id, path, can, content) {
  const needed = formGates[id];
  if (!needed || (id === 'auxilio' && path[1] === 'tareas') || needed.some(action => can(id, action))) return;
  content.querySelectorAll('form button[type="submit"], form #add-line').forEach(button => { button.disabled = true; });
  content.insertAdjacentHTML('afterbegin', `<div class="inline-note">${icon('lock')}<span>Tu rol puede consultar este módulo, pero no registrar ni modificar datos.</span></div>`);
}
function contentSubtitle(){return document.querySelector('.module-heading p');}
async function render() {
  root.setAttribute('aria-busy','true');
  chartObserver?.disconnect();
  disposeShell?.();
  const currentRender = ++renderId;
  root.onkeydown = null; root.onclick = null;
  const route = location.hash.replace(/^#\//,'') || 'login';
  const [pathname, query = ''] = route.split('?');
  const path = pathname.split('/');
  const params = new URLSearchParams(query);
  let session;
  try {session=await refreshSession();}catch(error){if(currentRender===renderId)root.innerHTML='<section class="panel error-state" role="alert"><h2>No se pudo comprobar la sesión</h2><p>'+escapeHtml(error.message)+'</p><a href="">Reintentar</a></section>';return;}
  if(currentRender!==renderId)return;
  if (!session) { if (route !== 'login') { history.replaceState(null,'','#/login'); } renderLogin(); return; }
  if (route === 'login') { history.replaceState(null,'','#/inicio'); }
  const can = permissionChecker(session.role);
  allowed = modules.filter(m => can(m.id, 'ver'));
  let module = modules.find(m => m.id === path[0]);
  const denied = Boolean(module) && !can(module.id, 'ver');
  if (!module) { module = allowed[0] || modules[0]; if (allowed[0]) history.replaceState(null,'',`#/${homePath(session.role)}`); }
  if (module.id === 'auxilio' && !path[1] && !denied && !can('auxilio', 'crear')) { history.replaceState(null,'','#/auxilio/tareas'); path[1] = 'tareas'; }
  document.title = `${module.id === 'inicio' ? 'Dashboard' : module.label} · El Bravo del JAC`;
  shell(session,module);
  try {
    if (denied || !allowed.length) {
      const target = homePath(session.role);
      document.getElementById('page-content').innerHTML = `<section class="panel error-state no-access" role="alert"><span class="no-access-icon">${icon('lock')}</span><h2>No tienes permiso para entrar aquí</h2><p>${allowed.length ? 'Tu rol no incluye este módulo. Si lo necesitas, pide a un administrador que revise tus permisos.' : 'Tu rol no tiene módulos habilitados. Pide a un administrador que revise tus permisos.'}</p>${allowed.length ? `<a class="primary-button" href="#/${target}">Ir a mi pantalla de inicio</a>` : ''}</section>`;
    }
    else if (module.id === 'inicio') await renderDashboard(currentRender);
    else {
      const views={clientes:renderClients,vehiculos:renderVehicles,recepciones:renderReceptions,diagnosticos:renderDiagnoses,cotizaciones:renderQuotes,ordenes:renderOrders,auxilio:renderRescues,pagos:renderPayments,mecanicos:renderMechanics,proveedores:renderSuppliers,usuarios:renderUsers,reportes:renderReports};
      if(realMode){document.getElementById('page-content').innerHTML='<div class="loading-state" role="status">Cargando datos del servidor…</div>';if(!await loadModule(module.id,params)||currentRender!==renderId)return;}
      await views[module.id]({root:document.getElementById('page-content'),path,params,session});
      gateForms(module.id, path, can, document.getElementById('page-content'));
      if(realMode){const missing=getStore().unavailable||[];if(missing.length){document.getElementById('page-content').insertAdjacentHTML('afterbegin','<div class="inline-note" role="status">Tu acceso no incluye las consultas auxiliares: '+escapeHtml(missing.join(', '))+'. Los datos no autorizados no se cargan.</div>');}}
      const viewTitle=document.querySelector('.module-heading h2');
      if(viewTitle){
        const heading=document.querySelector('.page-heading');
        heading.querySelector('h1').textContent=viewTitle.textContent;
        const subtitle=contentSubtitle();
        heading.querySelector('p').textContent=subtitle?.textContent||'';
        const actions=document.querySelector('.module-heading .toolbar-actions');
        if(actions?.children.length)heading.appendChild(actions);
        document.querySelector('.module-heading').hidden=true;
      }
    }
  } catch(reason) {
    if(currentRender!==renderId)return;
    document.getElementById('page-content').innerHTML='<section class="panel error-state" role="alert"><h2>No se pudo abrir esta vista</h2><p>'+escapeHtml(reason.message)+'</p><a class="primary-button" href="#/'+module.id+'">Volver a '+escapeHtml(module.label)+'</a></section>';
  }
  if(currentRender!==renderId)return;
  document.querySelectorAll('#page-content a').forEach(a=>a.addEventListener('click',()=>document.getElementById('workshop-dialog')?.close()));
  root.dataset.route=location.hash;root.setAttribute('aria-busy','false');
  window.scrollTo({top:0,behavior:'instant'});
  const pageTitle = document.querySelector('.page-heading h1');
  if (pageTitle) { pageTitle.tabIndex = -1; pageTitle.focus({preventScroll:true}); }
}
document.querySelector('.skip-link').addEventListener('click', event => {
  event.preventDefault();
  const main = document.getElementById('main-content');
  main.setAttribute('tabindex','-1'); main.focus(); main.scrollIntoView();
});
window.addEventListener('bravo:unauthorized',()=>{clearRealStore();if(location.hash!=='#/login')go('login');});
window.addEventListener('hashchange', render);
render();
