package sqlserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	mssql "github.com/microsoft/go-mssqldb"
)

type Store struct{ db *sql.DB }

// Applied to every new and reused connection. Filtered indexes require these settings.
const sessionOptions = `SET ANSI_NULLS ON; SET ANSI_PADDING ON; SET ANSI_WARNINGS ON;
SET ARITHABORT ON; SET CONCAT_NULL_YIELDS_NULL ON; SET QUOTED_IDENTIFIER ON;
SET NUMERIC_ROUNDABORT OFF; SET XACT_ABORT ON;`

func Open(dsn string) (*Store, error) {
	connector, err := mssql.NewConnector(dsn)
	if err != nil {
		return nil, errors.New("SQLSERVER_DSN inválido")
	}
	connector.SessionInitSQL = sessionOptions
	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxIdleTime(5 * time.Minute)
	return &Store{db: db}, nil
}

func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
func (s *Store) Close() error                   { return s.db.Close() }

func translate(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	var dbErr mssql.Error
	if errors.As(err, &dbErr) {
		switch dbErr.Number {
		case 2601, 2627, 1205, 1222, 3960:
			return domain.ErrConflict
		case 547:
			// SQL Server reports both CHECK and FK violations as 547.
			if strings.Contains(dbErr.Message, "CHECK constraint") {
				return domain.Invalid("request", "Los valores no cumplen las restricciones de SQL Server.")
			}
			return domain.ErrConflict
		case 515, 8114, 8115, 8152, 2628, 51001, 51002, 51003, 51004, 51005, 51006, 51007, 51008, 51009, 51010, 51011, 51012:
			return domain.Invalid("request", "Los valores no cumplen las restricciones de SQL Server.")
		}
	}
	return err
}

func acquireLock(ctx context.Context, tx *sql.Tx, name string) error {
	var result int
	err := tx.QueryRowContext(ctx, `
DECLARE @result int;
EXEC @result = sys.sp_getapplock @Resource=@p1, @LockMode=N'Exclusive', @LockOwner=N'Transaction', @LockTimeout=5000;
SELECT @result;`, name).Scan(&result)
	if err != nil {
		return translate(err)
	}
	if result == -2 && ctx.Err() != nil {
		return ctx.Err()
	}
	if result < 0 {
		return domain.ErrConflict
	}
	return nil
}

func (s *Store) User(ctx context.Context, id string) (domain.User, error) {
	return findUser(ctx, s.db, "u.id_usuario=@p1", id)
}
func (s *Store) FindLogin(ctx context.Context, login string) (domain.User, error) {
	return findUser(ctx, s.db, "(u.usuario=@p1 OR u.correo=@p1)", login)
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func findUser(ctx context.Context, q queryer, where string, value string) (domain.User, error) {
	rows, err := q.QueryContext(ctx, `
SELECT TOP (2) CONVERT(varchar(20),u.id_usuario),u.nombre,u.usuario,COALESCE(u.correo,N''),u.activo,u.password_hash,
    (SELECT r.nombre_rol AS name FROM dbo.roles r JOIN dbo.usuario_roles ur ON ur.id_rol=r.id_rol
     WHERE ur.id_usuario=u.id_usuario ORDER BY r.nombre_rol FOR JSON PATH),
    (SELECT DISTINCT p.codigo AS code FROM dbo.permisos p JOIN dbo.rol_permisos rp ON rp.id_permiso=p.id_permiso
     JOIN dbo.usuario_roles ur ON ur.id_rol=rp.id_rol WHERE ur.id_usuario=u.id_usuario ORDER BY p.codigo FOR JSON PATH)
FROM dbo.usuarios u WHERE `+where, value)
	if err != nil {
		return domain.User{}, translate(err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return domain.User{}, translate(err)
		}
		return domain.User{}, domain.ErrNotFound
	}
	u := domain.User{Roles: []string{}, Permissions: []string{}}
	var roles, permissions []byte
	if err := rows.Scan(&u.ID, &u.Name, &u.Username, &u.Email, &u.Active, &u.PasswordHash, &roles, &permissions); err != nil {
		return domain.User{}, translate(err)
	}
	if rows.Next() {
		return domain.User{}, domain.ErrNotFound
	}
	if err := rows.Err(); err != nil {
		return domain.User{}, translate(err)
	}
	var roleList []struct{ Name string }
	var permissionList []struct{ Code string }
	if err := json.Unmarshal(roles, &roleList); err != nil {
		return domain.User{}, err
	}
	if err := json.Unmarshal(permissions, &permissionList); err != nil {
		return domain.User{}, err
	}
	for _, role := range roleList {
		u.Roles = append(u.Roles, role.Name)
	}
	for _, permission := range permissionList {
		u.Permissions = append(u.Permissions, permission.Code)
	}
	return u, nil
}
