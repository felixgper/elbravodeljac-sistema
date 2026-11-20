"""Emite SELECT con CTE sintéticos para sqlcmd; no instala vistas ni lee tablas.

Uso: python -X utf8 database/sqlserver/bi/probar_consultas.py > consultas.sql
Ejecutar con sqlcmd -f 65001 -b -h -1 -w 65535; todas las filas deben ser PASS.
Solo requiere el motor SQL Server (JSON/fechas); no permisos sobre tablas.
"""
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[3]


def view_queries():
    sql = (Path(__file__).with_name("vistas_analiticas.sql")).read_text(encoding="utf-8")
    assert "IF @@TRANCOUNT <> 0" in sql
    assert "IF XACT_STATE() <> 0 ROLLBACK TRANSACTION;" in sql
    assert "BEGIN CATCH" in sql and "THROW;" in sql
    assert not re.search(r"^\s*GO\s*$", sql, re.M)
    batches = re.findall(r"EXEC sys\.sp_executesql N'((?:''|[^'])*)';", sql)
    assert len(batches) == 6
    queries = []
    for batch in batches:
        batch = batch.replace("''", "'")
        name, query = re.fullmatch(r"CREATE OR ALTER VIEW bi\.(\w+) AS\s*(.*);", batch, re.S).groups()
        query = re.sub(r"\b(?:dbo|bi)\.", "", query)
        queries.append(f"{name} AS ({query})")
    return queries


FIXTURES = """
pagos AS (
 SELECT id_pago,id_orden,CAST(monto AS decimal(12,2)) monto,
 CAST(fecha_pago AS datetimeoffset) fecha_pago,CAST(revertido_en AS datetimeoffset) revertido_en
 FROM (VALUES
 (1,1,100.10,'2026-10-31T12:00:00+00:00',NULL),
 (2,1,75.00,'2026-10-20T12:00:00+00:00','2026-11-10T12:00:00+00:00'),
 (3,2,49.90,'2026-11-01T04:59:59+00:00',NULL),
 (4,2,50.00,'2026-11-01T05:00:00+00:00',NULL)
 ) p(id_pago,id_orden,monto,fecha_pago,revertido_en)
),
cotizaciones AS (
 SELECT id_cotizacion,id_recepcion,id_cotizacion_base,
 CAST('2026-10-01T12:00:00+00:00' AS datetimeoffset) fecha_emision,estado,CAST(total AS decimal(12,2)) total
 FROM (VALUES (10,1,NULL,N'rechazada',250.00),(11,1,10,N'aprobada',300.00),
 (12,2,NULL,N'aprobada',100.00),(13,3,NULL,N'pendiente',120.00),
 (14,3,NULL,N'aprobada',150.00),(15,3,NULL,N'aprobada',180.00),
 (16,3,NULL,N'aprobada',100.00)
 ) q(id_cotizacion,id_recepcion,id_cotizacion_base,estado,total)
),
ordenes_trabajo AS (
 SELECT id_orden,id_cotizacion,estado,CAST(fecha_creacion AS datetimeoffset) fecha_creacion,
 CAST(fecha_fin AS datetimeoffset) fecha_fin,control_tecnico
 FROM (VALUES
 (1,11,N'finalizada','2026-09-01T12:00:00+00:00','2026-11-01T05:00:00+00:00',
 N'{"execution":{"101":{"status":"Realizado"},"102":{"status":"No realizado"},"103":{"status":"Realizado"},"104":{"status":"Realizado"},"105":{"status":"Realizado"}},"performedItems":[101,101,105]}'),
 (2,12,N'finalizada','2026-10-01T12:00:00+00:00','2026-11-01T04:59:59+00:00',N'{"performedItems":[201,201]}'),
 (3,13,N'en_reparacion','2026-10-02T12:00:00+00:00',NULL,N'{"performedItems":[301]}'),
 (4,14,N'cancelada','2026-10-03T12:00:00+00:00','2026-11-01T12:00:00+00:00',N'{"performedItems":[401]}'),
 (5,15,N'en_pruebas','2026-10-04T12:00:00+00:00',NULL,N'{}'),
 (6,16,N'creada','2026-10-05T12:00:00+00:00',NULL,N'{}')
 ) o(id_orden,id_cotizacion,estado,fecha_creacion,fecha_fin,control_tecnico)
),
detalle_cotizacion AS (
 SELECT id_detalle,id_cotizacion,id_servicio,tipo_concepto,CAST(cantidad AS decimal(10,2)) cantidad,descripcion FROM (VALUES
 (101,11,1,N'Servicio',2,N'Descripción privada'),
 (102,11,1,N'Servicio',8,N'No ejecutado'),
 (103,11,NULL,N'Repuesto',9,N'Pieza'),
 (104,11,NULL,N'Servicio',3,N'Libre con datos privados'),
 (105,11,1,NULL,1,N'Concepto antiguo'),
 (106,11,1,N'Servicio',7,N'Sin confirmar'),
 (201,12,1,N'Servicio',4,N'Legado'),
 (301,13,1,N'Servicio',6,N'Abierta'),
 (401,14,1,N'Servicio',5,N'Cancelada'),
 (501,10,1,N'Servicio',99,N'Revisión antigua')
 ) d(id_detalle,id_cotizacion,id_servicio,tipo_concepto,cantidad,descripcion)
),
servicios AS (SELECT 1 AS id_servicio,N'Afinamiento' AS nombre),
recepciones AS (
 SELECT id_recepcion,id_vehiculo,id_cliente_atencion,
 CAST('2026-10-01T12:00:00+00:00' AS datetimeoffset) fecha_ingreso,tipo_atencion
 FROM (VALUES (1,1,1,N'diagnostico'),(2,1,1,N'directa'),(3,2,2,N'directa'))
 r(id_recepcion,id_vehiculo,id_cliente_atencion,tipo_atencion)
)
"""


def backend_queries():
    """Extrae las expresiones reales, no una copia reimplementada en Python."""
    schema = (ROOT / "database/sqlserver/esquema_entrega.sql").read_text(encoding="utf-8-sig")
    movements = re.search(r"CREATE VIEW dbo.jac_movimientos_pagos AS\s*(.*?)\nGO", schema, re.S)[1].rstrip(";\n ")
    go = (ROOT / "backend/internal/platform/sqlserver/statistics.go").read_text(encoding="utf-8-sig")
    service = go.split("), services AS (", 1)[1].split("), rescues AS (", 1)[0]
    service = service.replace("@p1", "CAST('2026-01-01T00:00:00-05:00' AS datetimeoffset)")
    service = service.replace("@p2", "CAST('2027-01-01T00:00:00-05:00' AS datetimeoffset)")
    service = re.sub(r"\b(?:dbo|bi)\.", "", service)
    movements = movements.replace("dbo.", "")
    return [f"movimientos AS ({movements})", f"backend_servicios AS ({service})"]


CHECKS = [
 ("pagos: excluye revertido", "(SELECT COUNT(*) FROM vw_pagos_vigentes)=3 AND NOT EXISTS(SELECT 1 FROM vw_pagos_vigentes WHERE id_pago=2)"),
 ("pagos: suma exacta y sin duplicados", "(SELECT SUM(monto) FROM vw_pagos_vigentes)=200 AND (SELECT COUNT(DISTINCT id_pago) FROM vw_pagos_vigentes)=3"),
 ("meses: octubre Lima", "(SELECT ingresos_cobrados FROM vw_ingresos_mensuales WHERE mes='20261001')=150"),
 ("meses: noviembre Lima", "(SELECT ingresos_cobrados FROM vw_ingresos_mensuales WHERE mes='20261101')=50"),
 ("meses: conciliación", "(SELECT SUM(pagos) FROM vw_ingresos_mensuales)=3 AND (SELECT SUM(ingresos_cobrados) FROM vw_ingresos_mensuales)=200"),
 ("cotizaciones: emisiones distintas", "(SELECT COUNT(*) FROM vw_cotizaciones_analiticas)=7 AND (SELECT COUNT(DISTINCT id_cotizacion) FROM vw_cotizaciones_analiticas)=7"),
 ("cotizaciones: revisión preservada", "EXISTS(SELECT 1 FROM vw_cotizaciones_analiticas WHERE id_cotizacion=11 AND id_cotizacion_base=10 AND total=300)"),
 ("cotizaciones: no suma revisiones como ingreso", "(SELECT SUM(total) FROM vw_cotizaciones_analiticas)=1200 AND (SELECT SUM(monto) FROM vw_pagos_vigentes)=200"),
 ("ordenes: una fila por orden", "(SELECT COUNT(*) FROM vw_ordenes_analiticas)=6 AND (SELECT COUNT(DISTINCT id_orden) FROM vw_ordenes_analiticas)=6"),
 ("ordenes: cinco estados", "(SELECT COUNT(DISTINCT estado) FROM vw_ordenes_analiticas)=5"),
 ("atenciones: visitas sin duplicar vehiculos", "(SELECT COUNT(*) FROM vw_vehiculos_atendidos)=3 AND (SELECT COUNT(DISTINCT id_vehiculo) FROM vw_vehiculos_atendidos)=2 AND (SELECT COUNT(DISTINCT id_cliente) FROM vw_vehiculos_atendidos)=2"),
 ("servicios: ejecución y cantidades", "(SELECT COUNT(*) FROM vw_servicios_realizados)=4 AND (SELECT SUM(cantidad) FROM vw_servicios_realizados)=10"),
 ("servicios: exclusiones", "NOT EXISTS(SELECT 1 FROM vw_servicios_realizados WHERE id_detalle IN(102,103,106,301,401,501))"),
 ("servicios: JSON legado sin multiplicación", "(SELECT COUNT(*) FROM vw_servicios_realizados WHERE id_detalle=201)=1 AND (SELECT COUNT(*) FROM vw_servicios_realizados WHERE id_detalle=101)=1"),
 ("servicios: fecha finalización Lima", "EXISTS(SELECT 1 FROM vw_servicios_realizados WHERE id_detalle=201 AND fecha='20261031') AND EXISTS(SELECT 1 FROM vw_servicios_realizados WHERE id_detalle=101 AND fecha='20261101')"),
 ("privacidad: descripción libre omitida", "EXISTS(SELECT 1 FROM vw_servicios_realizados WHERE id_detalle=104 AND servicio=N'Servicio manual')"),
 ("backend: cobros históricos equivalentes", "(SELECT SUM(monto) FROM movimientos)=(SELECT SUM(monto) FROM vw_pagos_vigentes)"),
 ("backend: diferencia mensual explícita", "(SELECT SUM(monto) FROM movimientos WHERE fecha>='2026-10-01T00:00:00-05:00' AND fecha<'2026-11-01T00:00:00-05:00')=225 AND (SELECT ingresos_cobrados FROM vw_ingresos_mensuales WHERE mes='20261001')=150"),
 ("backend: reversión en mes posterior", "(SELECT SUM(monto) FROM movimientos WHERE fecha>='2026-11-01T00:00:00-05:00' AND fecha<'2026-12-01T00:00:00-05:00')=-25 AND (SELECT ingresos_cobrados FROM vw_ingresos_mensuales WHERE mes='20261101')=50"),
 ("backend: servicios coinciden", "(SELECT SUM(count) FROM backend_servicios)=(SELECT SUM(cantidad) FROM vw_servicios_realizados)"),
 ("backend: órdenes creadas vs finalizadas", "(SELECT COUNT(*) FROM vw_ordenes_analiticas WHERE fecha>='20261001' AND fecha<'20261101')=5 AND (SELECT COUNT(*) FROM ordenes_trabajo WHERE estado=N'finalizada' AND fecha_fin>='2026-10-01T00:00:00-05:00' AND fecha_fin<'2026-11-01T00:00:00-05:00')=1"),
 ("vacío: cobros sin filas", "COALESCE((SELECT SUM(monto) FROM vw_pagos_vigentes WHERE 1=0),0)=0"),
]


def select_checks(sources, checks):
    cases = []
    for name, predicate in checks:
        cases.append("SELECT N'" + name.replace("'", "''") + "' AS prueba, CASE WHEN " + predicate + " THEN 'PASS' ELSE 'FAIL' END AS resultado")
    return ";WITH " + ",\n".join(sources) + "\n" + "\nUNION ALL\n".join(cases) + ";\n"


def build_sql():
    result = select_checks([FIXTURES.strip(), *view_queries(), *backend_queries()], CHECKS)
    # Segundo escenario: las seis fuentes vacías, conservando sus tipos reales.
    empty_views = view_queries()
    for table in ("pagos", "ordenes_trabajo", "cotizaciones", "recepciones", "detalle_cotizacion", "servicios"):
        empty_views = [re.sub(r"\b(FROM|JOIN) " + table + r"\b", r"\1 (SELECT * FROM " + table + " WHERE 1=0)", query) for query in empty_views]
    empty_checks = [("vacío: " + query.split(" AS", 1)[0], "(SELECT COUNT(*) FROM " + query.split(" AS", 1)[0] + ")=0") for query in empty_views]
    result += select_checks([FIXTURES.strip(), *empty_views], empty_checks)
    # Garantiza que la salida es un único SELECT sin objetos persistentes.
    assert not re.search(r"\b(?:dbo|bi)\.", result)
    assert not re.search(r"\b(?:INSERT|UPDATE|DELETE|MERGE|CREATE|ALTER|DROP|EXEC|INTO|GRANT)\b", result, re.I)
    return result


if __name__ == "__main__":
    print(build_sql())
