# Dominio

Fase A: `models.go` contiene DTO/valores, validación de longitudes y precisión,
normalización de documentos/placas y errores de negocio. No trunca textos.

Implementar entidades, valores decimales exactos, estados, errores de negocio y
puertos de persistencia requeridos por cada caso de uso. No copiar IDs demo ni
passwordDigest del navegador. Los ID SQL son bigint; su representación HTTP será
string decimal. El código visible codigo_ot es distinto de id_orden.

Reglas: kilometraje no decreciente, transferencia verificada, presupuesto aprobado,
asignaciones vigentes, prueba satisfactoria posterior a reparaciones, saldo sin
sobrepago, retiro autorizado y conservación del historial. El servidor las verifica.
