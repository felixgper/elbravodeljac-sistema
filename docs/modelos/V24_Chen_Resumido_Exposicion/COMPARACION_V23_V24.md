# Comparación V23 → V24

## Decisión de presentación

| Aspecto | V23 | V24 |
| --- | --- | --- |
| Entidades | 27 | 13, todas principales del taller |
| Relaciones | 42 | 15, con proyecciones documentadas |
| Lienzo | 9400×5800 de diseño | 3900×2850 de diseño |
| Uso | Auditoría detallada y propuestas | Exposición académica general |
| Atributos | Omitidos | Omitidos |
| Formato editable | .drawio | .drawio nativo |
| Impresión | Lienzo amplio | PDF vectorial, una página A3 horizontal |
| Sistema / bases | Sin modificación | Sin modificación |

Se reduce aproximadamente un 52% el número de entidades y un 64% el de relaciones dibujadas. No se elimina ninguna tabla, colección, dato ni control. La simplicidad se consigue mediante selección y abstracción documentada.

## Correspondencia de relaciones

| V24 | Relación | Origen V23 | Tratamiento |
| --- | --- | --- | --- |
| S01 | CLIENTES / TIENE TITULAR REGISTRADO / VEHÍCULOS | R01 | Conservada; nombre simplificado si procede |
| S02 | VEHÍCULOS / INGRESA / RECEPCIONES | R02 | Conservada; nombre simplificado si procede |
| S03 | RECEPCIONES / REQUIERE / DIAGNÓSTICOS | R03 | Conservada; nombre simplificado si procede |
| S04 | RECEPCIONES / GENERA / COTIZACIONES | R27 | Conservada; nombre simplificado si procede |
| S05 | DIAGNÓSTICOS / SUSTENTA / COTIZACIONES | R04 | Conservada; nombre simplificado si procede |
| S06 | COTIZACIONES / AUTORIZA / ÓRDENES DE TRABAJO | R05 | Conservada; nombre simplificado si procede |
| S07 | ÓRDENES DE TRABAJO / RECIBE / PAGOS | R06 | Conservada; nombre simplificado si procede |
| S08 | ÓRDENES DE TRABAJO / CULMINA EN / ENTREGAS | R07 | Conservada; nombre simplificado si procede |
| S09 | CLIENTES / SOLICITA / AUXILIOS MECÁNICOS | R36 | Conservada; nombre simplificado si procede |
| S10 | EMPLEADOS / ATIENDE * / AUXILIOS MECÁNICOS | R44 | Proyección de intermediarios existentes |
| S11 | EMPLEADOS / REALIZA * / ÓRDENES DE TRABAJO | R17+R22 | Proyección de intermediarios existentes |
| S12 | EMPLEADOS / REALIZA EVALUACIÓN * / DIAGNÓSTICOS | R31 | Proyección de intermediarios existentes |
| S13 | COTIZACIONES / INCLUYE SERVICIO * / SERVICIOS | R21+R18 | Proyección de intermediarios existentes |
| S14 | COTIZACIONES / INCLUYE REPUESTO * / REPUESTOS | R21+R28 | Proyección de intermediarios existentes |
| S15 | PROVEEDORES / SUMINISTRA / REPUESTOS | R12 | Conservada; nombre simplificado si procede |

## Propuestas de V23 que no pasan a ser implementación

- **P1** (cuenta interna obligatoriamente de un empleado) no se dibuja al omitir USUARIOS y sigue pendiente de resolver cuentas legadas/técnicas. 007 tiene vínculo nullable y único. No se aplica NOT NULL ni se crean empleados ficticios.
- **P2** (empleado responsable frente a usuario autor) permanece como abstracción de negocio marcada con asterisco. Los responsables físicos continúan referenciando usuarios; la persona se deriva solo cuando hay vínculo laboral. S11 resume además las asignaciones. No se sustituyen las FK operativas.
- **P3** (auxilio–recepción (0,1)/(0,1)) se omite. El modelo físico actual admite varios auxilios por recepción, con coincidencia de cliente/vehículo validada. No se impone la propuesta pendiente ni se pierde el vínculo existente.

## Diferencias reales que no deben ocultarse

| Regla conceptual | Implementación contrastada | Consecuencia |
| --- | --- | --- |
| Auxilio nuevo con cliente y técnico | Columnas históricas nullable; SaveAuxilio exige cliente/vehículo/técnico/unidad | Los (1,1) son reglas de alta; no certifican el contenido histórico. |
| Orden con trabajador responsable | Alta atómica de asignación a usuario mecánico; empleado de ese usuario puede ser NULL | S11 expresa responsabilidad laboral objetivo, no una garantía física completa. Regularización sigue pendiente. |
| Diagnóstico con empleado responsable | Técnico usuario obligatorio; vínculo a empleado nullable | S12 proyecta negocio, sin afirmar que ya se migraron cuentas antiguas. |
| Catálogos en cotización | Relación mediante detalle, referencias opcionales/excluyentes y conceptos libres | S13/S14 son (0,N) en ambos extremos; la cotización sí exige al menos una línea, no al menos un servicio/repuesto de catálogo. |
| Titularidad actual | Cliente registrado e historial; no prueba registral externa | Se conserva TIENE TITULAR REGISTRADO en vez de afirmar propiedad legal probada mediante POSEE. |
| Roles y tipo de empleado | 007/009 separan vínculo laboral y acceso N:M | Omisión académica no elimina roles, permisos ni cargo. |
| TIPOS_DOCUMENTO | 008 crea tabla y dos referencias obligatorias | Continúa implementada; se omite únicamente del dibujo V24. |

La exportación SQL original no contiene por sí sola toda la evolución 001–009; no se compara V23 solamente con el respaldo base. Los informes históricos documentan pruebas aisladas de esa evolución; no se conectó esta tarea a bases para verificar despliegues actuales. No se encontró migración posterior a 009; no se inventa que una inexistente haya sido aplicada.

## Relaciones omitidas y conservación de lógica

Se omiten del lienzo contactos, titularidad temporal, catálogos documentales, autorrelación de ampliaciones, clasificación técnica, estados, actores de captura, RBAC y referencias MongoDB. También vínculos secundarios auxilio–vehículo/unidad/recepción y cliente–recepción. Sus reglas se explican en EXPLICACION_MODELO_CONCEPTUAL.md y siguen en V23/lógico/físico/código. No se reemplazan por conexiones directas falsas entre pagos y entregas ni entre clientes y órdenes.

El detalle de omisiones de las 14 entidades está en la explicación. No se crea un módulo de compras, ni se equipara una cotización con un trabajo ya realizado. El mecanismo del visor y las huellas de fuentes están documentados en VALIDACION_RENDER_LOCAL.json y FUENTES_REVISADAS.json.
