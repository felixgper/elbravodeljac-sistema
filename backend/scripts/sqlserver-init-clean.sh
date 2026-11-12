#!/bin/sh
set -eu

case "${JAC_SQL_APP_PASSWORD:-}" in
  ''|*[!A-Za-z0-9]*) echo "Contraseña SQL de aplicación inválida" >&2; exit 1 ;;
esac

export SQLCMDPASSWORD="$MSSQL_SA_PASSWORD"
sqlcmd() {
  /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -C -b -l 5 "$@"
}

# A previous initialization, even a failed one, requires operator review.
sqlcmd -Q "
IF DB_ID(N'elbravodeljac_sqlserver_dev') IS NOT NULL
    OR SUSER_ID(N'jac_app') IS NOT NULL
    THROW 51000, N'La base o el login ya existe. Revise su estado antes de continuar.', 1;"
sqlcmd -Q 'CREATE DATABASE [elbravodeljac_sqlserver_dev];'
sqlcmd -Q 'ALTER DATABASE [elbravodeljac_sqlserver_dev] SET COMPATIBILITY_LEVEL = 170;'

sqlcmd -f 65001 -d elbravodeljac_sqlserver_dev -i /setup/esquema_entrega.sql

sqlcmd -x -d elbravodeljac_sqlserver_dev -i /dev/stdin <<SQL
BEGIN TRANSACTION;
CREATE LOGIN [jac_app] WITH PASSWORD = '$JAC_SQL_APP_PASSWORD', CHECK_POLICY = ON,
    DEFAULT_DATABASE = [elbravodeljac_sqlserver_dev];
CREATE USER [jac_app] FOR LOGIN [jac_app];
GRANT SELECT, INSERT, UPDATE, DELETE, EXECUTE ON SCHEMA::dbo TO [jac_app];
COMMIT TRANSACTION;
GO
SQL
