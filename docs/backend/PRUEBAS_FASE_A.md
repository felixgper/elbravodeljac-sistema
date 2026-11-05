# Informe de pruebas de fase A

## Ejecutado y aprobado

Se ejecutó `backend/scripts/test-local-postgres.ps1` con PostgreSQL 18 instalado
localmente y Go 1.27.1. Creó un clúster independiente con puerto aleatorio y base
`elbravodeljac_test`, inicializó el esquema aprobado exclusivamente allí y ejecutó
`go test ./... -count=1 -v`. El proceso PostgreSQL temporal se detuvo al terminar.

| Grupo | Resultado verificado |
|---|---|
| Inventario vivo de la base de pruebas | 26 tablas y 42 FK, sin alterar el esquema original |
| Autenticación HTTP/PostgreSQL | Login bcrypt, cookie HttpOnly/SameSite, identidad, dos roles conservados y readiness real |
| Clientes | Alta, consulta, búsqueda/paginación, PATCH, DELETE, duplicados, longitudes Unicode y FK |
| Contactos | Alta y listado vinculados; bloqueo de borrado de cliente con referencias |
| Vehículos | Alta con propietario inicial, consulta/filtro, PATCH, duplicados y FK de cliente |
| Transferencia | Verificación obligatoria, historial, rollback completo ante propietario inexistente |
| Borrado de vehículo | Bloqueo con transferencias; rollback del historial cuando existe recepción; borrado permitido sin operación |
| Concurrencia PostgreSQL | Dos actualizaciones de kilometraje se serializan; valor final no disminuye |
| Permisos y CSRF | Token CSRF inválido→403; revocación de permiso en SQL afecta la siguiente petición |
| Logout | JWT anterior queda revocado inmediatamente |
| JWT unitario | Rechazo de firma ajena, algoritmo incorrecto/none, expiración, issuer/audience incorrectos y falta de exp |
| Validación unitaria | IDs bigint, precisión/rango decimal, nombres Unicode y longitudes sin truncamiento |
| Seguridad HTTP unitaria | Origen incorrecto, cookie Secure, ausencia de hashes en respuesta, rate limit y JSON inválido |
| Guardia de configuración | DSN de ElBravoDelJac rechazado antes de conectar |

`go vet ./...` terminó correctamente. También pasaron las tres pruebas existentes:
`node checks/verify.mjs`, `node checks/workflow.mjs`, `node checks/edge-cases.mjs`.
La comparación Git confirmó que `dist/` y el SQL aprobado permanecen sin cambios.

## Docker: validado, no ejecutado

`docker compose -f compose.test.yaml config --quiet` pasó con una variable de
validación temporal. No se ejecutaron contenedores ni se verificó la construcción
de Dockerfile.test: el cliente Docker está instalado, pero no se encuentra el
motor desktop-linux. `docker desktop start` también falló por falta de la clave
de instalación de Docker Desktop. No se afirma que Docker haya pasado pruebas.

La integración real sí se ejecutó mediante el clúster local aislado descrito arriba.
Cuando Docker Engine esté disponible, ejecutar `backend/scripts/test-compose.ps1`
para verificar además el camino de contenedores, descarga de imágenes y build.

## Qué no se probó ni se cambió

No se conectó a ElBravoDelJac, no se verificaron ni modificaron cuentas/datos reales.
No se cambiaron el frontend, sus datos demo, el esquema aprobado ni proveedores.
No hay pruebas de APIs de las fases siguientes ni conexión MongoDB en esta entrega.
La integración de navegador con el backend todavía no está activada.
