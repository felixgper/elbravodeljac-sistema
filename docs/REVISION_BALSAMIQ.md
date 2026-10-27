# Revisión del Balsamiq V14.1

Fuente: `Balsamiq_V14_1_REPARADO.bmpr`. Se revisaron 23 pantallas y ventanas directamente desde los controles del archivo. Las notas internas indican V14.1.1 revisada.

## Decisiones de esta entrega

- Se conserva la apariencia cobre y carbón aprobada: navegación oscura, superficies claras y marca roja.
- El Balsamiq determina los módulos, las operaciones, los campos y los indicadores. La imagen elegida determina el estilo.
- El menú conserva sus 13 opciones: Inicio, Clientes, Vehículos, Recepciones, Diagnósticos, Cotizaciones, Órdenes de trabajo, Auxilio mecánico, Pagos y entrega, Mecánicos, Proveedores, Reportes, Usuarios y roles.
- La imagen de estilo no incluía Recepciones y mostraba Configuración: se respeta el menú del Balsamiq, sin agregar Configuración ni quitar Recepciones y Usuarios y roles.
- El Dashboard mantiene los cuatro indicadores y los cinco gráficos del Balsamiq. Se añade únicamente el banner visual aprobado; no es un módulo.
- Mostrar u ocultar la contraseña mejora la revisión de escritura; no modifica los datos del Login.
- En celular el menú lateral se convierte en un panel desplegable desde la izquierda. No se agrega una segunda barra inferior de navegación.
- En la versión 1.0, las 23 pantallas y ventanas del inventario están implementadas. Se incorporan listados de registros guardados para volver a abrir recepciones, diagnósticos y cotizaciones; son vistas auxiliares del mismo módulo.
- Se agregan únicamente acciones necesarias para operar el prototipo: quitar conceptos de cotización, consultar historial, paginar listados y descargar/reiniciar datos demo. No son módulos nuevos.
- Se incluyen Propietario y Administrativo en la selección de roles porque aparecen en el listado de usuarios del propio Balsamiq.
- Finalizar requiere prueba satisfactoria; pago y entrega siguen separados. Se impide eliminar órdenes con pagos o entregas para conservar las relaciones y la trazabilidad.

## Compatibilidad y reglas confirmadas

- Nuevo cliente: tipo de documento + número; nombre o razón social; teléfono; correo opcional; dirección. Unicidad futura por (tipo_documento, numero_documento).
- Vehículo: sin VIN; kilometraje decimal compatible con PostgreSQL NUMERIC(10,2). Al integrar la API, intercambiarlo como cadena decimal o representación exacta acordada.
- Recepción y diagnóstico son pantallas y registros separados. Recepción no registra al entregante como autorizado al retiro.
- Crear OT: requiere cotización aprobada, asigna mecánico y auxiliar opcional; estado inicial Pendiente.
- Actualizar OT: conserva identidad de la OT, agrega avances y pruebas. Solo finalizar con pruebas exigidas satisfactorias.
- Eliminar OT exige confirmación independiente; queda para la etapa de órdenes.
- Pago y entrega son operaciones distintas; verificar identidad y autorización del receptor. Contacto nuevo es una ventana, no un módulo nuevo.
- Mecánicos consulta asignaciones; personal se administra en Usuarios y roles.
- PostgreSQL alojará datos estructurados y transaccionales; MongoDB documentos flexibles y evidencias técnicas. La primera etapa no conecta ninguna BD.
- No existe autenticación real: acceso local de demostración. La autorización y contraseñas con hash serán responsabilidad de Go.

## Inventario completo de pantallas

### 01 - Inicio de sesión

- **Title:** EL BRAVO DEL JAC
- **SubTitle:** Acceso exclusivo para trabajadores del taller
- **Label:** Correo o usuario
- **TextInput:** Ingrese su usuario
- **Label:** Contraseña
- **TextInput:** ••••••••
- **CheckBox:** Recordarme
- **Button:** Iniciar sesión
- **Label:** El menú del sistema se muestra después de iniciar sesión.

### 02 - Dashboard

- **Title:** Dashboard general
- **Label:** Resumen de actividad del taller | Datos ilustrativos
- **Label:** Ingresos del mes
- **Title:** S/ 12,450
- **Label:** Órdenes activas
- **Title:** 10
- **Label:** Servicios finalizados
- **Title:** 24
- **Label:** Auxilios de hoy
- **Title:** 3
- **SubTitle:** Ingresos de los últimos 6 meses
- **Label:** GRÁFICO DE LÍNEA  /  Últimos 6 meses
- **Label:** Ene
- **Label:** Feb
- **Label:** Mar
- **Label:** Abr
- **Label:** May
- **Label:** Jun
- **Label:** 6.2k
- **Label:** 7.8k
- **Label:** 8.5k
- **Label:** 10.2k
- **Label:** 11.3k
- **Label:** 12.4k
- **SubTitle:** Estados del servicio
- **Label:** Tipo de gráfico: circular (anillo)
- **Label:** Pendientes                 2
- **Label:** En reparación            5
- **Label:** En pruebas                3
- **Label:** Finalizadas               24
- **Label:** Canceladas                 1
- **SubTitle:** Órdenes por estado (En proceso = reparación + pruebas)
- **SubTitle:** Trabajo por mecánico
- **Label:** Pedro
- **Label:** 10
- **Label:** Luis
- **Label:** 8
- **Label:** María
- **Label:** 7
- **Label:** Kevin
- **Label:** 4
- **SubTitle:** Servicios más solicitados
- **Label:** Aceite
- **Label:** 22
- **Label:** Frenos
- **Label:** 18
- **Label:** Diagn.
- **Label:** 14
- **Label:** Susp.
- **Label:** 10
- **Label:** Pendientes   2
- **Label:** En proceso  8
- **Label:** Finalizadas 24
- **Label:** Canceladas  1

### 03 - Clientes

- **Title:** Clientes
- **Label:** Consultar y administrar clientes registrados
- **SearchBox:** Buscar por nombre o documento...
- **Button:** + Nuevo cliente
- **DataGrid:** ID,Tipo,N° documento,Nombre / razón social,Teléfono,Acción / 1,DNI,12345678,Juan Pérez,987654321,Editar / 2,RUC,20123456789,Transportes SAC,987000111,Editar / 3,CE,001294555,María López,988111222,Editar / 4,Pasaporte,AB567889,Martin Reyes,988222333,Editar
- **Label:** Editar reutiliza el formulario de cliente con datos cargados. Los contactos se crean al retirar.

### 04 - Nuevo cliente

- **Title:** Nuevo cliente
- **Label:** Complete los datos y seleccione el documento correspondiente.
- **SubTitle:** DATOS DEL CLIENTE
- **Label:** Nombre completo o razón social
- **TextInput:** Ej.: Juan Pérez / Transportes SAC
- **Label:** Tipo de documento
- **ComboBox:** DNI / RUC / Carné de extranjería / Pasaporte
- **Label:** Número de documento
- **TextInput:** Ingrese el número del documento
- **Label:** Opciones disponibles: DNI, RUC, carné de extranjería y pasaporte
- **Label:** Teléfono
- **TextInput:** Ej.: 987654321
- **Label:** Correo electrónico (opcional)
- **TextInput:** correo@ejemplo.com
- **Label:** Dirección
- **TextInput:** Calle, distrito y referencia
- **Label:** Al editar, este formulario aparece precargado y guarda cambios sobre el mismo id_cliente.
- **Button:** Guardar cliente
- **Button:** Volver a clientes

### 05 - Vehículos

- **Title:** Vehículos
- **Label:** Consulta los vehículos registrados y actualiza sus datos desde una sola opción.
- **SearchBox:** Buscar por placa, modelo o cliente...
- **Button:** + Nuevo vehículo
- **DataGrid:** Placa,Marca,Modelo,Año,Propietario actual,Historial / ABC-123,JAC,T8,2022,Juan Pérez,Ver historial / DEF-456,JAC,S3,2021,María Torres,Ver historial / GHI-789,JAC,T6,2023,Transportes SAC,Ver historial
- **Button:** Actualizar vehículo
- **Label:** Para modificar un vehículo o registrar un cambio de propietario, selecciona «Actualizar vehículo».

### 06 - Nuevo vehículo

- **Title:** Nuevo vehículo
- **Label:** Registro inicial: asociamos una placa única con un cliente existente. Para transferencias use Vehículos → Editar.
- **SubTitle:** DATOS DEL VEHÍCULO · PRIMER REGISTRO
- **Label:** Cliente propietario
- **ComboBox:** Seleccione cliente...
- **Label:** Placa
- **TextInput:** ABC-123
- **Label:** Marca
- **ComboBox:** JAC
- **Label:** Modelo
- **TextInput:** T8, S3, T6...
- **Label:** Año
- **TextInput:** 2024
- **Label:** Kilometraje actual (km)
- **TextInput:** 25300.50
- **Label:** Observaciones (opcional)
- **TextInput:** Color, versión o detalle relevante
- **Button:** Guardar vehículo
- **Button:** Volver a vehículos

### 06B - Actualizar vehículo

- **Title:** Actualizar vehículo
- **Label:** Busca la placa y actualiza los datos. Si cambió de dueño, registra aquí la transferencia verificada.
- **SubTitle:** DATOS DEL VEHÍCULO
- **Label:** Seleccionar vehículo registrado (búsqueda por placa)
- **ComboBox:** ABC-123  ·  JAC T8  ·  Juan Pérez
- **Label:** Placa (se conserva)
- **TextInput:** ABC-123
- **Label:** Propietario actual (solo lectura)
- **TextInput:** Juan Pérez  ·  DNI 45879632
- **Label:** Marca
- **ComboBox:** JAC
- **Label:** Modelo
- **TextInput:** T8
- **Label:** Año
- **TextInput:** 2022
- **Label:** Kilometraje actual (km)
- **TextInput:** 25300.50
- **Label:** Observaciones (opcional)
- **TextInput:** Color, versión o información adicional
- **SubTitle:** CAMBIO DE PROPIETARIO  ·  OPCIONAL
- **CheckBox:** El vehículo cambió de propietario (activar solo cuando corresponda)
- **Label:** Nuevo propietario: buscar cliente ya registrado por nombre o documento
- **ComboBox:** Pedro Ramírez  ·  DNI 43659874
- **CheckBox:** Transferencia verificada por usuario autorizado (se guarda usuario y fecha)
- **Label:** Se mantiene el mismo id_vehiculo, la misma placa y todo el historial técnico del vehículo.
- **Button:** Guardar cambios
- **Button:** Volver a vehículos
- **Label:** Los datos del nuevo propietario solo se aplicarán si se confirma la transferencia.

### 07 - Nueva recepción

- **Title:** Nueva recepción
- **Label:** Registrar cliente de esta atención, vehículo, kilometraje y síntomas.
- **SubTitle:** DATOS DE LA RECEPCIÓN
- **Label:** Cliente responsable de esta atención
- **ComboBox:** Buscar por nombre, RUC o documento...
- **Button:** + Cliente
- **Label:** Vehículos del cliente (placa y modelo)
- **ComboBox:** Seleccione placa y modelo...
- **Button:** + Nuevo vehículo
- **Label:** Kilometraje de ingreso (km)
- **TextInput:** 25300.50
- **Label:** Fecha de ingreso
- **TextInput:** Hoy: fecha y hora al guardar
- **Label:** Motivo de ingreso / síntomas
- **TextArea:** Ej.: Cambio de aceite, ruido en frenos
- **Label:** Recepcionista: usuario autenticado (automático)   ·   Fecha y hora: hoy, al guardar.
- **Button:** Guardar recepción
- **Button:** Ir a diagnóstico

### 08 - Diagnóstico técnico

- **Title:** Registrar diagnóstico técnico
- **Label:** Seleccionar recepción guardada y registrar evaluación técnica por separado.
- **SubTitle:** EVALUACIÓN DEL MECÁNICO
- **Label:** Recepción registrada (seleccionar pendiente de diagnóstico)
- **ComboBox:** RX-100  ·  ABC-123  ·  Juan Pérez
- **Label:** Cliente — solo lectura
- **TextInput:** Juan Pérez
- **Label:** Vehículo — solo lectura
- **TextInput:** ABC-123 — JAC T8
- **Label:** Síntomas registrados por recepción (solo lectura)
- **TextArea:** Ruido en el motor al acelerar
- **Label:** Diagnóstico técnico / evaluación del mecánico
- **TextArea:** Escriba hallazgos, pruebas y conclusión técnica...
- **Label:** Técnico autenticado y fecha automáticos. Guardar diagnóstico antes de cotizar.
- **Button:** Guardar diagnóstico
- **Button:** Continuar a cotización

### 09 - Cotización

- **Title:** Nueva cotización
- **Label:** Presupuesto de servicios y repuestos, sujeto a aprobación del cliente
- **Label:** Diagnóstico guardado / Recepción
- **ComboBox:** RX-025 / Juan Pérez
- **Label:** Vehículo
- **TextInput:** JAC T8 - ABC-123
- **DataGrid:** Concepto,Tipo,Origen repuesto,Cant.,P. unit.,Subtotal / Cambio de aceite,Servicio,-,1,S/ 120,S/ 120 / Filtro,Repuesto,Taller,1,S/ 35,S/ 35
- **Button:** + Agregar servicio o repuesto
- **Label:** TOTAL PROPUESTO:  S/ 155.00
- **SubTitle:** AUTORIZACIÓN DEL CLIENTE
- **Label:** ¿El cliente acepta el servicio?
- **ComboBox:** Pendiente / Aprobada / Rechazada
- **Button:** Guardar cotización
- **Label:** Origen Taller/Cliente se define en cada fila de repuesto.

### 10 - Órdenes de trabajo

- **Title:** Órdenes de trabajo
- **Label:** Crea desde cotizaciones aprobadas. Lápiz: editar. Papelera: eliminar con confirmación.
- **SearchBox:** Buscar por OT, cliente o placa...
- **Button:** + Nueva orden
- **DataGrid:** OT,Fecha,Cliente,Placa,Estado,Responsable,Acciones / OT-025,20/10,Juan Pérez,ABC-123,En reparación,Pedro,✎ / 🗑 / OT-026,20/10,María Torres,DEF-456,Pendiente,Luis,✎ / 🗑 / OT-027,19/10,Transportes SAC,GHI-789,Finalizada,Carlos,✎ / 🗑
- **Label:** Acciones por fila: ✎ Actualizar  ·  🗑 Eliminar con confirmación.

### 10B - Confirmar eliminación de orden

- **Title:** Órdenes de trabajo
- **Label:** Crea desde cotizaciones aprobadas. Lápiz: editar. Papelera: eliminar con confirmación.
- **SearchBox:** Buscar por OT, cliente o placa...
- **Button:** + Nueva orden
- **DataGrid:** OT,Fecha,Cliente,Placa,Estado,Responsable,Acciones / OT-025,20/10,Juan Pérez,ABC-123,En reparación,Pedro,✎ / 🗑 / OT-026,20/10,María Torres,DEF-456,Pendiente,Luis,✎ / 🗑 / OT-027,19/10,Transportes SAC,GHI-789,Finalizada,Carlos,✎ / 🗑
- **Label:** Acciones por fila: ✎ Actualizar  ·  🗑 Eliminar con confirmación.
- **SubTitle:** ¿Eliminar orden OT-026?
- **Paragraph:** ¿Estás seguro de que deseas eliminar definitivamente la orden OT-026? Esta acción no se puede deshacer.
- **Button:** Cancelar
- **Button:** Eliminar

### 11 - Nueva orden de trabajo

- **Title:** Nueva orden de trabajo
- **Label:** Seleccionar una cotización aprobada y asignar el personal responsable.
- **SubTitle:** DATOS DE LA ORDEN
- **Label:** Cotización aprobada (obligatoria)
- **ComboBox:** COT-025 / Juan Pérez / ABC-123 — Aprobada
- **Label:** Cliente (solo lectura)
- **TextInput:** Juan Pérez
- **Label:** Vehículo (solo lectura)
- **TextInput:** ABC-123 — JAC T8
- **Label:** Trabajos autorizados (derivados de la cotización, solo lectura)
- **TextArea:** Revisión y reparación aprobada del sistema de frenos
- **Label:** Mecánico responsable
- **ComboBox:** Seleccionar mecánico...
- **Label:** Auxiliar (opcional)
- **ComboBox:** Sin auxiliar / Kevin
- **Label:** Estado inicial: Pendiente  ·  Código de OT y fecha generados automáticamente.
- **Button:** Crear orden
- **Button:** Volver al listado

### 11B - Actualizar orden

- **Title:** Actualizar orden de trabajo
- **Label:** Modificar personal, estado y registrar avances de una orden existente.
- **SubTitle:** DATOS DE LA ORDEN
- **Label:** N.º de OT (solo lectura)
- **TextInput:** OT-025
- **Label:** Cliente (solo lectura)
- **TextInput:** Juan Pérez
- **Label:** Vehículo / placa (solo lectura)
- **TextInput:** JAC T8 — ABC-123
- **Label:** Diagnóstico inicial (solo lectura)
- **TextInput:** Ruido en motor
- **Label:** Mecánico responsable
- **ComboBox:** Pedro / Carlos
- **Label:** Auxiliar asignado
- **ComboBox:** Kevin / Sin auxiliar
- **Label:** Estado de la orden
- **ComboBox:** Pendiente / En reparación / En pruebas / Finalizada / Cancelada
- **Label:** Fecha de creación (solo lectura)
- **TextInput:** 20/10/2026
- **Label:** Observación / avance de esta actualización
- **TextArea:** Describa únicamente el avance nuevo, sin sobrescribir registros previos...
- **Label:** Prueba realizada (opcional)
- **TextInput:** Ej. Prueba de ruta
- **Label:** Resultado de prueba
- **ComboBox:** No aplica / Satisfactoria / Fallida
- **Label:** Se guardará el usuario y fecha. Solo finalizar si las pruebas exigidas son satisfactorias.
- **Button:** Guardar cambios
- **Button:** Volver al listado
- **Button:** Finalizar orden

### 12 - Auxilio mecánico

- **Title:** Nuevo auxilio mecánico
- **Label:** Registro de atención fuera del taller
- **SubTitle:** SOLICITUD DE AUXILIO
- **Label:** Cliente
- **ComboBox:** Buscar cliente
- **Label:** Vehículo
- **ComboBox:** Seleccione vehículo
- **Label:** Ubicación reportada
- **TextInput:** Distrito, dirección o referencia
- **Label:** Tipo de incidente
- **ComboBox:** Falla mecánica / Vehículo inmovilizado / Otro
- **Label:** Unidad de auxilio
- **ComboBox:** Seleccionar unidad
- **Label:** Descripción del problema
- **TextArea:** Datos reportados por el cliente
- **Label:** Técnico asignado
- **ComboBox:** Seleccione técnico
- **Button:** Registrar auxilio

### 13 - Pagos y entrega

- **Title:** Pago y entrega del vehículo
- **Label:** El pago y la identificación del receptor son acciones diferentes.
- **SubTitle:** 1. PAGO DE LA ORDEN
- **Label:** Orden de trabajo
- **ComboBox:** OT-025 / Transportes JAC SAC / JAC T8
- **Label:** Monto pagado
- **TextInput:** S/ 450.00
- **Label:** Medio de pago
- **ComboBox:** Transferencia / Efectivo / Tarjeta
- **Button:** Registrar pago
- **SubTitle:** 2. ¿QUIÉN RETIRA EL VEHÍCULO?
- **Label:** Identificar a quien recoge el vehículo y comprobar su autorización.
- **ButtonBar:** Titular / representante, Tercero
- **Label:** Identificar a la persona que retira; aunque haya traído el vehículo, no queda autorizada automáticamente.
- **Label:** SI SELECCIONA TERCERO: elegir contacto registrado o añadir uno nuevo
- **ComboBox:** Buscar persona vinculada al cliente...
- **Button:** + Nuevo contacto
- **Label:** Si es un tercero nuevo, registrar datos en una ventana breve (no otro módulo).
- **Label:** ¿Cómo se verificó la autorización para este retiro?
- **ComboBox:** Confirmación del titular / Documento / Representante facultado
- **CheckBox:** Verifiqué identidad y autorización específica antes de entregar
- **Label:** El sistema registra automáticamente fecha, hora y usuario responsable.
- **Button:** Confirmar entrega

### 13A - Nuevo contacto (ventana emergente)

- **Title:** Pago y entrega del vehículo
- **Label:** El pago y la identificación del receptor son acciones diferentes.
- **SubTitle:** 1. PAGO DE LA ORDEN
- **Label:** Orden de trabajo
- **ComboBox:** OT-025 / Transportes JAC SAC / JAC T8
- **Label:** Monto pagado
- **TextInput:** S/ 450.00
- **Label:** Medio de pago
- **ComboBox:** Transferencia / Efectivo / Tarjeta
- **Button:** Registrar pago
- **SubTitle:** 2. ¿QUIÉN RETIRA EL VEHÍCULO?
- **Label:** Identificar a quien recoge el vehículo y comprobar su autorización.
- **ButtonBar:** Titular / representante, Tercero
- **Label:** Identificar a la persona que retira; aunque haya traído el vehículo, no queda autorizada automáticamente.
- **Label:** SI SELECCIONA TERCERO: elegir contacto registrado o añadir uno nuevo
- **ComboBox:** Buscar persona vinculada al cliente...
- **Button:** + Nuevo contacto
- **Label:** Si es un tercero nuevo, registrar datos en una ventana breve (no otro módulo).
- **Label:** ¿Cómo se verificó la autorización para este retiro?
- **ComboBox:** Confirmación del titular / Documento / Representante facultado
- **CheckBox:** Verifiqué identidad y autorización específica antes de entregar
- **Label:** El sistema registra automáticamente fecha, hora y usuario responsable.
- **Button:** Confirmar entrega
- **SubTitle:** Nuevo contacto para retiro
- **Label:** Esta persona se vinculará al cliente seleccionado; NO queda autorizada automáticamente.
- **Label:** Nombre completo
- **TextInput:** Ej. Carlos Mendoza
- **Label:** Tipo de documento
- **ComboBox:** DNI / CE / Pasaporte
- **Label:** N.º documento
- **TextInput:** Documento
- **Label:** Teléfono internacional
- **TextInput:** Ej. +51987654321
- **Label:** Relación con el cliente
- **ComboBox:** Conductor / Representante / Familiar / Otro
- **Button:** Guardar contacto
- **Button:** Cancelar

### 13B - Mecánicos

- **Title:** Mecánicos y auxiliares
- **Label:** Consultar personal técnico registrado y sus asignaciones actuales.
- **SearchBox:** Buscar por nombre o rol...
- **Button:** Ver asignaciones
- **DataGrid:** Trabajador,Rol,Órdenes activas,Estado / Pedro,Mecánico,3,Activo / Luis,Mecánico,2,Activo / Kevin,Auxiliar,2,Activo / Carlos,Supervisor,1,Activo
- **Label:** El personal se administra desde Usuarios y roles; aquí no se crean usuarios duplicados.

### 13C - Proveedores

- **Title:** Proveedores
- **Label:** Gestionar proveedores y su relación con repuestos y costos de referencia.
- **SearchBox:** Buscar proveedor por nombre o RUC...
- **Button:** + Nuevo proveedor
- **DataGrid:** Proveedor,RUC,Teléfono,Repuestos,Acción / Repuestos Lima,20112233445,+51987654321,Filtros,Editar / Autopartes Norte,20445566778,+51986543210,Frenos,Editar
- **Label:** Nuevo y Editar abren el mismo formulario emergente; datos guardados en proveedores.

### 14 - Reportes

- **Title:** Reportes e indicadores
- **Label:** Información para supervisión y análisis del taller
- **SubTitle:** FILTROS
- **Label:** Periodo:
- **ComboBox:** Este mes / Últimos 6 meses
- **Button:** Aplicar filtros
- **SubTitle:** SERVICIOS REALIZADOS
- **DataGrid:** Tipo,Cantidad / Mantenimiento,14 / Diagnóstico,10 / Frenos,8 / Auxilio,3
- **SubTitle:** INGRESOS REGISTRADOS
- **DataGrid:** Mes,Ingresos / Mayo,S/ 11,300 / Junio,S/ 12,450 / Julio,S/ 10,200
- **Label:** Los gráficos del dashboard usarán consultas a PostgreSQL; Power BI se integrará después.

### 15 - Usuarios y roles

- **Title:** Usuarios y permisos
- **Label:** Agregar/Editar usuario abre un mismo formulario; los permisos dependen del rol.
- **SearchBox:** Buscar usuario por nombre o rol...
- **Button:** + Agregar usuario
- **DataGrid:** Usuario,Rol,Acceso,Estado / Jimmy,Propietario,Completo,Activo / Carlos,Supervisor,Órdenes / Mecánicos,Activo / Kevin,Auxiliar,Órdenes asignadas,Activo / Administración,Administrativo,Pagos / Clientes,Activo
- **SubTitle:** REGLAS DE SEGURIDAD
- **Paragraph:** Crear y editar reutilizan un formulario con nombre, usuario, correo, rol, estado y clave inicial al crear. Go verifica permisos.

### 15A - Agregar / editar usuario

- **Title:** Agregar / editar usuario
- **Label:** Mismo formulario: al editar carga datos existentes; al crear solicita contraseña inicial.
- **SubTitle:** DATOS DE ACCESO Y ROL
- **Label:** Nombre del trabajador
- **Label:** Nombre de usuario (único)
- **Label:** Correo electrónico (único)
- **Label:** Rol asignado
- **Label:** Contraseña inicial (solo alta)
- **TextInput:** Ej.: Ana Torres
- **TextInput:** Ej.: admin_taller
- **TextInput:** Ej.: ana@taller.com
- **TextInput:** Se configura solo al crear
- **ComboBox:** Administrador / Supervisor / Mecánico / Auxiliar
- **CheckBox:** Usuario activo
- **Label:** Contraseña guardada como hash; el rol se vincula mediante usuario_roles.
- **Button:** Guardar usuario
- **Button:** Volver a usuarios

### 16 - Vista móvil del auxiliar

- **SubTitle:** EL BRAVO DEL JAC
- **Title:** Hola, Kevin
- **SubTitle:** OT-025 · JAC T8
- **Label:** Cambio de aceite
- **Label:** Estado: En reparación
- **Button:** Actualizar estado
- **SubTitle:** OT-029 · JAC S3
- **Label:** Revisión de frenos
- **Label:** Estado: Pendiente
- **Button:** Actualizar estado
- **SubTitle:** OT-030 · JAC T6
- **Label:** Revisión técnica
- **Label:** Estado: Pruebas
- **Button:** Actualizar estado
