// Datos ilustrativos del Dashboard del Balsamiq V14.1. Importes en céntimos.
export const dashboardDemo = {
  monthlyIncomeCents: 1245000,
  completedServices: 24,
  rescuesToday: 3,
  income: [
    { month: 'Ene', cents: 620000 }, { month: 'Feb', cents: 780000 },
    { month: 'Mar', cents: 850000 }, { month: 'Abr', cents: 1020000 },
    { month: 'May', cents: 1130000 }, { month: 'Jun', cents: 1245000 },
  ],
  states: [
    { key: 'pending', label: 'Pendientes', count: 2, color: '#d89a2a' },
    { key: 'repair', label: 'En reparación', count: 5, color: '#bc622e' },
    { key: 'testing', label: 'En pruebas', count: 3, color: '#467bbb' },
    { key: 'completed', label: 'Finalizadas', count: 24, color: '#24906d' },
    { key: 'cancelled', label: 'Canceladas', count: 1, color: '#b85858' },
  ],
  mechanics: [{ label: 'Pedro', count: 10 }, { label: 'Luis', count: 8 }, { label: 'María', count: 7 }, { label: 'Kevin', count: 4 }],
  services: [{ label: 'Aceite', count: 22 }, { label: 'Frenos', count: 18 }, { label: 'Diagnóstico', count: 14 }, { label: 'Suspensión', count: 10 }],
};
