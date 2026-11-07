# Trazabilidad del prototipo al aplicativo

El [prototipo Balsamiq V14.1](Balsamiq_V14_1_REPARADO.bmpr) conserva 23 pantallas/ventanas. Su [JSON](balsamiq-v14.1.json) y checks/prototype.mjs verifican navegación histórica; no acreditan persistencia, permisos ni indicadores actuales.

El aplicativo vigente utiliza dist/assets/js/app.js y sus vistas/servicios, con modo real separado de /demo/. Los recorridos actuales se prueban con HTTP/Chrome y bases temporales: [pruebas](../checks/README.md). [Login y Dashboard actuales](frontend/LOGIN_DASHBOARD_CARBON.md).
