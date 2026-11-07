# Utilidades manuales conservadas

SIMULAR_POSTGRESQL.sql y MANIFIESTO_MONGODB.js son herramientas QA históricas, no inicializadores ni pasos de instalación. Su conservación no autoriza ejecutarlas sobre bases existentes. backend/scripts/backup-audit.ps1 también se conserva como herramienta operativa; no se ejecutó en esta etapa.

Para pruebas actuales usar [el runner aislado](../../../checks/README.md), que crea nuevas bases con UUID. Esta limpieza solo retira archivos rastreados justificados, sin ejecutar estas utilidades.
