const money = value => new Intl.NumberFormat('es-PE', { style: 'currency', currency: 'PEN', maximumFractionDigits: 0 }).format(value / 100);
export function incomeChart(points, width = 520) {
  const max = Math.max(1500000, Math.ceil(Math.max(...points.map(p => p.cents), 0) / 500000) * 500000), w = Math.max(240, Math.floor(width)), h = 190, left = 43, right = 15, top = 16, bottom = 36;
  const x = index => left + index * ((w - left - right) / (points.length - 1));
  const y = value => top + (1 - value / max) * (h - top - bottom);
  const line = points.map((p, i) => `${x(i)},${y(p.cents)}`).join(' ');
  const area = `${left},${h - bottom} ${line} ${w - right},${h - bottom}`;
  return `<svg class="income-chart" viewBox="0 0 ${w} ${h}" role="img" aria-label="Ingresos: ${points.map(p => `${p.month}, ${money(p.cents)}`).join('; ')}">
    <defs><linearGradient id="income-fill" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#791c2b" stop-opacity=".2"/><stop offset="1" stop-color="#791c2b" stop-opacity="0"/></linearGradient></defs>
    ${[0, max / 3, max * 2 / 3, max].map(n => `<line x1="${left}" x2="${w - right}" y1="${y(n)}" y2="${y(n)}" class="chart-grid"/><text x="${left - 10}" y="${y(n) + 4}" text-anchor="end" class="chart-label">${n === 0 ? '0' : `${Number((n / 100000).toFixed(1))}k`}</text>`).join('')}
    <polygon points="${area}" fill="url(#income-fill)"/><polyline points="${line}" fill="none" stroke="#791c2b" stroke-width="3" stroke-linejoin="round"/>
    ${points.map((p, i) => `<circle cx="${x(i)}" cy="${y(p.cents)}" r="4" fill="#fff" stroke="#791c2b" stroke-width="2"><title>${p.month}: ${money(p.cents)}</title></circle><text x="${x(i)}" y="${h - 12}" text-anchor="middle" class="chart-label">${p.month}</text>`).join('')}
    <text x="0" y="12" class="chart-unit">S/</text>
  </svg>`;
}
export function donutChart(states) {
  const total = states.reduce((sum, s) => sum + s.count, 0);
  let angle = 0;
  const stops = states.map(s => { const start = angle; angle += s.count / (total || 1) * 360; return `${s.color} ${start}deg ${angle}deg`; }).join(', ');
  return `<div class="donut-layout"><div class="donut" style="background:conic-gradient(${stops})" role="img" aria-label="Estados del servicio: ${states.map(s => `${s.label}: ${s.count}`).join(', ')}"><div><strong>${total}</strong><span>órdenes</span></div></div><ul class="chart-legend">${states.map(s => `<li><span class="legend-dot" style="background:${s.color}"></span><span>${s.label}</span><strong>${s.count}</strong></li>`).join('')}</ul></div>`;
}
export function verticalBars(states) {
  const bars = [states[0], {label: 'En proceso', count: states[1].count + states[2].count, color: '#791c2b'}, states[3], states[4]];
  const max = Math.max(1,...bars.map(s => s.count));
  return `<div class="bar-chart" role="img" aria-label="Órdenes por estado. ${bars.map(b => `${b.label}: ${b.count}`).join(', ')}"><div class="bar-columns">${bars.map(b => `<div class="bar-column"><div class="bar-space"><div class="bar" style="height:${b.count / max * 100}%;background:${b.color}"><strong>${b.count}</strong></div></div><span>${b.label}</span></div>`).join('')}</div></div>`;
}
export function horizontalBars(items, kind) {
  const max = Math.max(1,...items.map(i => i.count));
  return `<div class="horizontal-bars" role="img" aria-label="${items.map(i => `${i.label}: ${i.count}`).join(', ')}">${items.map((i, index) => `<div class="hbar-row"><span class="hbar-label">${kind === 'mechanics' ? `<span class="initial-avatar">${i.label[0]}</span>` : ''}${i.label}</span><div class="hbar-track"><span class="hbar-fill ${kind === 'services' ? 'service-fill' : ''}" style="width:${i.count / max * 100}%;opacity:${1 - index * .1}"></span></div><strong>${i.count}</strong></div>`).join('')}</div>`;
}
