# V24 actualizado — modelo lógico Crow's Foot sin tipos

Se conserva **la misma disposición y relaciones del V24 original**, sin inventar entidades ni alterar cardinalidades: 28 tablas, 208 atributos y 50 relaciones (FK).

[Editar modelo lógico original](El_Bravo_del_JAC_Logico_CrowsFoot.drawio)

- `clientes.id_cliente` es PK; `vehiculos.id_cliente` es FK.
- `recepciones.id_cliente_atencion` y `recepciones.id_vehiculo` apuntan a tablas diferentes para documentar **quién es atendido** y **qué vehículo ingresa**.
- El diagnóstico es opcional por tipo de atención. `cotizaciones.id_diagnostico` es nullable; `cotizaciones.id_recepcion` es obligatorio.
- La autorrelación de `cotizaciones.id_cotizacion_base` ya existe y permite revisar cotizaciones; no se elimina solo por simplificar la exposición.
- **Pendiente de implementar:** renombrar `id_recepcionista` por `id_usuario_registro`. El modelo lógico vigente conserva el nombre físico mientras SQL Server y el backend todavía lo usan.

[Modelo físico de SQL Server 2025, sobre el mismo V24](../V24_Logico_CrowsFoot/El_Bravo_del_JAC_V24_Logico_Completo.drawio)

> Las imágenes y PDF V24 antiguos se retiraron de esta rama para evitar confundirlos con el modelo actual; siguen recuperables desde el historial Git. El archivo `.drawio` editable es la fuente vigente de esta revisión. No se cambiaron la base, los datos, ni el sistema.
