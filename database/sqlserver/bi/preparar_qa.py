"""Genera pruebas SQL exclusivas de QA. No conecta ni ejecuta SQL o Docker.

python -X utf8 database/sqlserver/bi/preparar_qa.py > pruebas_qa.sql
Ejecución pendiente de autorización. Requiere esquema oficial en QA y marcador.
"""
from pathlib import Path
import re

QA_DATABASE = "elbravodeljac_sqlserver_qa"


def literal(value):
    return "N'" + value.replace("'", "''") + "'"


def installer():
    source = Path(__file__).with_name("vistas_analiticas.sql").read_text(encoding="utf-8")
    assert source.count("N'elbravodeljac_sqlserver_dev'") == 1
    return source.replace("N'elbravodeljac_sqlserver_dev'", literal(QA_DATABASE))


def build_sql():
    source = installer()
    batches = list(re.finditer(r"EXEC sys\.sp_executesql N'((?:''|[^'])*)';", source))
    assert len(batches) == 6
    sql = [f"""-- SOLO QA: contiene DDL y permisos de prueba. No ejecutar en la base oficial.
IF DB_NAME() <> N'{QA_DATABASE}' OR @@TRANCOUNT <> 0
BEGIN
    RAISERROR(N'Destino QA o estado de sesión no autorizado.',16,1);
    RETURN;
END;
IF NOT EXISTS(SELECT 1 FROM sys.extended_properties
              WHERE class=0 AND name=N'jac.bi_qa' AND CONVERT(nvarchar(40),value)=N'synthetic-only')
    THROW 51100,N'Falta marcador explícito de QA sintético.',1;
IF USER_ID(N'powerbi_reader') IS NOT NULL
    THROW 51100,N'El usuario QA ya existe; no modificar una identidad desconocida.',1;
SET NOCOUNT ON;
"""]
    # La primera prueba ocurre antes de crear bi o instalar una sola vista.
    sql.append("CREATE TABLE #transaccion_ajena (id int PRIMARY KEY, valor int);\n")
    for option, ident in (("OFF", 1), ("ON", 2)):
        sql.append(f"""
SET XACT_ABORT {option};
BEGIN TRANSACTION;
INSERT #transaccion_ajena VALUES ({ident},1);
DECLARE @options{ident} int=@@OPTIONS, @error{ident} int=0;
BEGIN TRY
    EXEC sys.sp_executesql {literal(source)};
END TRY
BEGIN CATCH
    SET @error{ident}=ERROR_NUMBER();
END CATCH;
IF @error{ident}<>50000 OR @@TRANCOUNT<>1 OR XACT_STATE()<>1 OR @@OPTIONS<>@options{ident}
   OR NOT EXISTS(SELECT 1 FROM #transaccion_ajena WHERE id={ident} AND valor=1)
BEGIN
    IF XACT_STATE()<>0 ROLLBACK TRANSACTION;
    THROW 51100,N'El rechazo alteró la transacción externa.',1;
END;
-- Demuestra que el llamador puede seguir escribiendo y confirmar SU trabajo.
UPDATE #transaccion_ajena SET valor=2 WHERE id={ident};
COMMIT TRANSACTION;
IF NOT EXISTS(SELECT 1 FROM #transaccion_ajena WHERE id={ident} AND valor=2)
    THROW 51100,N'No se conservó el trabajo del llamador.',1;
SELECT N'PASS: transacción externa con XACT_ABORT {option}' AS prueba;
""")
    sql.append("DROP TABLE #transaccion_ajena;\nSET XACT_ABORT ON;\n")
    sql.append("IF SCHEMA_ID(N'bi') IS NULL EXEC(N'CREATE SCHEMA bi AUTHORIZATION dbo');\n")
    sql.append(f"EXEC sys.sp_executesql {literal(source)};\n")
    names = [re.search(r"VIEW bi\.(\w+) AS", match[1])[1] for match in batches]
    names_sql = ",".join(literal(name) for name in names)
    sql.append(f"""
DECLARE @anteriores TABLE(nombre sysname PRIMARY KEY, objeto int, definicion varbinary(max));
INSERT @anteriores
SELECT v.name,v.object_id,CONVERT(varbinary(max),OBJECT_DEFINITION(v.object_id))
FROM sys.views v WHERE v.schema_id=SCHEMA_ID(N'bi') AND v.name IN ({names_sql});
IF (SELECT COUNT(*) FROM @anteriores)<>6 OR EXISTS(SELECT 1 FROM @anteriores WHERE definicion IS NULL)
    THROW 51100,N'No se pudieron conservar las seis definiciones previas.',1;
""")
    # Un fallo de compilación en cada posición; antes del fallo, las definiciones
    # válidas llevan un comentario diferente, para que revertir sea comprobable.
    for index in range(6):
        mutated = source
        for pos in range(5, -1, -1):
            match = batches[pos]
            if pos == index:
                replacement = "EXEC sys.sp_executesql N'CREATE OR ALTER VIEW bi." + names[pos] + " AS SELECT FROM;';"
            else:
                replacement = match[0].replace(" AS\nSELECT", " AS\n-- definición modificada por QA\nSELECT")
                assert replacement != match[0]
            mutated = mutated[:match.start()] + replacement + mutated[match.end():]
        sql.append(f"""
DECLARE @fallo{index} int=0;
BEGIN TRY
    EXEC sys.sp_executesql {literal(mutated)};
END TRY
BEGIN CATCH
    SET @fallo{index}=ERROR_NUMBER();
END CATCH;
IF @fallo{index} NOT IN (102,156) OR @@TRANCOUNT<>0 OR XACT_STATE()<>0
    THROW 51100,N'Fallo DDL sin propagación o con transacción pendiente.',1;
IF EXISTS(SELECT 1 FROM @anteriores a
          WHERE OBJECT_ID(N'bi.'+a.nombre) IS NULL
             OR OBJECT_ID(N'bi.'+a.nombre)<>a.objeto
             OR OBJECT_DEFINITION(a.objeto) IS NULL
             OR CONVERT(varbinary(max),OBJECT_DEFINITION(a.objeto))<>a.definicion)
    THROW 51100,N'El rollback no recuperó exactamente las definiciones anteriores.',1;
SELECT N'PASS: rollback ante fallo DDL en posición {index+1}' AS prueba;
""")
    # Usuario de base sin login: prueba la cadena de propiedad sin contraseñas.
    sql.append("CREATE USER powerbi_reader WITHOUT LOGIN;\nGRANT CONNECT TO powerbi_reader;\n")
    for name in names:
        sql.append(f"GRANT SELECT ON OBJECT::bi.{name} TO powerbi_reader;\n")
    sql.append("EXECUTE AS USER=N'powerbi_reader';\nBEGIN TRY\n")
    for name in names:
        sql.append(f"    IF HAS_PERMS_BY_NAME(N'bi.{name}',N'OBJECT',N'SELECT')<>1 THROW 51100,N'Falta lectura de vista aprobada.',1;\n")
        sql.append(f"    EXEC sys.sp_executesql N'SELECT COUNT_BIG(*) AS filas FROM bi.{name}';\n")
    sql.append("""
    IF EXISTS(SELECT 1 FROM fn_my_permissions(NULL,N'DATABASE')
              WHERE permission_name NOT IN(N'CONNECT',N'VIEW ANY COLUMN MASTER KEY DEFINITION',N'VIEW ANY COLUMN ENCRYPTION KEY DEFINITION'))
        THROW 51100,N'El lector tiene permisos de base inesperados.',1;
""")
    for name in ("pagos", "ordenes_trabajo", "cotizaciones", "recepciones", "detalle_cotizacion", "servicios"):
        for perm in ("SELECT", "INSERT", "UPDATE", "DELETE", "ALTER", "CONTROL"):
            sql.append(f"    IF COALESCE(HAS_PERMS_BY_NAME(N'dbo.{name}',N'OBJECT',N'{perm}'),0)<>0 THROW 51100,N'El lector tiene acceso directo inesperado.',1;\n")
    sql.append("""
    REVERT;
END TRY
BEGIN CATCH
    REVERT;
    THROW;
END CATCH;
SELECT N'PASS: SELECT por vistas sin acceso directo a tablas' AS prueba;
-- Ejecutar después verificar_vistas.sql como administrador QA, no como lector.
""")
    return "".join(sql)


if __name__ == "__main__":
    print(build_sql())
