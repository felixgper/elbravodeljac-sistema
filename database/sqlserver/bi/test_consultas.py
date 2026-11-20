"""Guardas estáticas del instalador y de la prueba SELECT, sin conexión SQL."""
import re
import unittest
from pathlib import Path
import probar_consultas as bi
import preparar_qa as qa


class ConsultasTests(unittest.TestCase):
    def test_rechazo_antes_de_cambiar_opciones(self):
        sql = Path(__file__).with_name("vistas_analiticas.sql").read_text(encoding="utf-8")
        executable = re.sub(r"--[^\n]*", "", sql).strip()
        guard, rest = executable.split("SET ANSI_NULLS", 1)
        self.assertRegex(guard, r"^IF @@TRANCOUNT <> 0\s+BEGIN")
        self.assertIn("RAISERROR(N'Utilice una sesión sin transacciones abiertas.', 16, 1);", guard)
        self.assertRegex(guard, r"RETURN;\s+END;\s*$")
        self.assertNotRegex(guard, r"\b(?:SET|THROW|ROLLBACK|COMMIT|EXEC)\b")
        self.assertIn("SET XACT_ABORT ON", rest)

    def test_qa_transaccion_externa_en_ambos_modos(self):
        sql = qa.build_sql()
        for option, ident in (("OFF", 1), ("ON", 2)):
            self.assertIn(f"SET XACT_ABORT {option};\nBEGIN TRANSACTION;", sql)
            self.assertIn(f"@error{ident}<>50000 OR @@TRANCOUNT<>1 OR XACT_STATE()<>1", sql)
            self.assertIn(f"@@OPTIONS<>@options{ident}", sql)
            self.assertIn(f"UPDATE #transaccion_ajena SET valor=2 WHERE id={ident};\nCOMMIT", sql)

    def test_qa_destino_y_rollback_preparados_no_ejecutados(self):
        sql = qa.build_sql()
        self.assertIn("DB_NAME() <> N'elbravodeljac_sqlserver_qa'", sql)
        self.assertIn("name=N'jac.bi_qa'", sql)
        self.assertNotIn("elbravodeljac_sqlserver_dev", sql)
        self.assertEqual(sql.count("PASS: rollback ante fallo DDL en posición"), 6)
        self.assertIn("CONVERT(varbinary(max),OBJECT_DEFINITION(a.objeto))<>a.definicion", sql)
        self.assertIn("CREATE USER powerbi_reader WITHOUT LOGIN", sql)
        self.assertNotIn("CREATE LOGIN", sql)
        self.assertNotIn("GRANT SELECT ON dbo.", sql)

    def test_instalacion_atomica(self):
        sql = Path(__file__).with_name("vistas_analiticas.sql").read_text(encoding="utf-8")
        self.assertEqual(sql.count("BEGIN TRANSACTION;"), 1)
        self.assertEqual(sql.count("COMMIT TRANSACTION;"), 1)
        self.assertLess(sql.index("BEGIN TRY"), sql.index("BEGIN TRANSACTION;"))
        self.assertLess(sql.rindex("EXEC sys.sp_executesql"), sql.index("COMMIT TRANSACTION;"))
        self.assertIn("IF XACT_STATE() <> 0 ROLLBACK TRANSACTION;\n    THROW;", sql)
        self.assertEqual(len(bi.view_queries()), 6)

    def test_no_ddl_ni_tablas_en_prueba(self):
        sql = bi.build_sql()
        self.assertNotRegex(sql, r"\b(?:dbo|bi)\.")
        self.assertNotRegex(sql.upper(), r"\b(?:INSERT|UPDATE|DELETE|MERGE|CREATE|ALTER|DROP|EXEC|INTO|GRANT)\b")
        self.assertEqual(sql.count("AS resultado"), 28)

    def test_calculos_extraidos_de_produccion(self):
        queries = bi.backend_queries()
        self.assertIn("UNION ALL", queries[0])
        self.assertIn("-monto", queries[0])
        self.assertIn("SUM(d.cantidad)", queries[1])
        self.assertIn("performedItems", queries[1])

    def test_fuentes_sin_datos_personales(self):
        views = "\n".join(bi.view_queries())
        self.assertNotRegex(views, r"\b(?:password|contrasena|telefono|email|dni|descripcion)\b")
        self.assertIn("Servicio manual", views)

    def test_parametro_deshabilitado(self):
        expressions = (bi.ROOT / "powerbi/ElBravoDelJAC.SemanticModel/definition/expressions.tmdl").read_text(encoding="utf-8-sig")
        self.assertRegex(expressions, r"HabilitarAnalitica = false\b")

    def test_medidas_y_relaciones(self):
        model = bi.ROOT / "powerbi/ElBravoDelJAC.SemanticModel/definition"
        measures = (model / "tables/bi vw_resumen_taller.tmdl").read_text(encoding="utf-8-sig")
        for expression in ("SUM(Cobros[monto])", "DISTINCTCOUNT(Ordenes[id_orden])", "DISTINCTCOUNT(Cotizaciones[id_cotizacion])", "DISTINCTCOUNT(Atenciones[id_vehiculo])", "DISTINCTCOUNT(Atenciones[id_cliente])", "SUM(Servicios[cantidad])"):
            self.assertIn(expression, measures)
        relations = (model / "relationships.tmdl").read_text(encoding="utf-8-sig")
        self.assertEqual(len(re.findall(r"toColumn: Calendario.Fecha", relations)), 5)
        self.assertNotIn("bothDirections", relations)


if __name__ == "__main__":
    unittest.main()
