# Puerto de aplicación

`store.go` y las interfaces de esta carpeta conectan el transporte HTTP con los adaptadores. Los DTO y validadores están en `internal/domain`; PostgreSQL delimita las transacciones de negocio y MongoDB almacena documentos técnicos autorizados.

No dividir operaciones atómicas (cotización/detalles, OT/asignaciones, pagos/entrega) en escrituras independientes para eludir sus reglas. [Contratos REST](../../../docs/backend/ENDPOINTS.md) y [organización vigente](../../README.md).
