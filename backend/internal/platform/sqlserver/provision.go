package sqlserver

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type initialAccount struct {
	name, username, password, role string
}

// Provision creates catalogs and the initial manager in an empty development or QA
// database. API startup never calls it.
func (s *Store) Provision(ctx context.Context, target, username, password string) error {
	return s.provision(ctx, target, []initialAccount{{"Administración de pruebas", username, password, "Gerente"}})
}

// ProvisionInstallation creates the three installation accounts and the catalogs atomically.
func (s *Store) ProvisionInstallation(ctx context.Context, target, owner, ownerPassword, adminPassword, mechanicPassword string) error {
	return s.provision(ctx, target, []initialAccount{
		{"Dueño", owner, ownerPassword, "Gerente"},
		{"Administradora", "administradora", adminPassword, "Administrador"},
		{"Mecánico", "mecanico", mechanicPassword, "Mecánico"},
	})
}

func (s *Store) provision(ctx context.Context, target string, accounts []initialAccount) error {
	if target != "elbravodeljac_sqlserver_dev" && target != "elbravodeljac_sqlserver_qa" {
		return errors.New("Confirma la base SQL Server de desarrollo o QA por su nombre exacto")
	}
	placeholderRole := "1"
	hashes := make([][]byte, len(accounts))
	usernames := make(map[string]bool, len(accounts))
	for i, account := range accounts {
		account.username = strings.TrimSpace(account.username)
		if err := (domain.UserInput{Name: &account.name, Username: &account.username, Password: &account.password, RoleID: &placeholderRole}).Validate(true); err != nil {
			return fmt.Errorf("cuenta inicial %s: %w", account.role, err)
		}
		username := strings.ToLower(account.username)
		if usernames[username] {
			return errors.New("Las cuentas iniciales deben tener usuarios distintos")
		}
		usernames[username] = true
		hash, err := bcrypt.GenerateFromPassword([]byte(account.password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		hashes[i] = hash
		accounts[i].username = account.username
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return translate(err)
	}
	defer tx.Rollback()
	if err := acquireLock(ctx, tx, "jac:sqlserver:provision"); err != nil {
		return err
	}

	var nameDB string
	var tables, columns, foreignKeys int
	err = tx.QueryRowContext(ctx, `SELECT DB_NAME(),
      (SELECT COUNT(*) FROM sys.tables WHERE schema_id=SCHEMA_ID(N'dbo')),
      (SELECT COUNT(*) FROM sys.columns c JOIN sys.tables t ON t.object_id=c.object_id WHERE t.schema_id=SCHEMA_ID(N'dbo')),
      (SELECT COUNT(*) FROM sys.foreign_keys WHERE schema_id=SCHEMA_ID(N'dbo') AND is_disabled=0 AND is_not_trusted=0)`).
		Scan(&nameDB, &tables, &columns, &foreignKeys)
	if err != nil {
		return translate(err)
	}
	if nameDB != target || tables != 28 || columns != 208 || foreignKeys != 50 {
		return errors.New("La base y el esquema de instalación no coinciden")
	}
	// A separate, explicit provisioning command must never augment an existing installation.
	var occupied int
	err = tx.QueryRowContext(ctx, `
SELECT CASE WHEN EXISTS (
    SELECT 1 FROM sys.partitions p JOIN sys.tables t ON t.object_id=p.object_id
    WHERE t.schema_id=SCHEMA_ID(N'dbo') AND p.index_id IN(0,1) AND p.rows>0
) THEN 1 ELSE 0 END`).Scan(&occupied)
	if err != nil {
		return translate(err)
	}
	if occupied != 0 {
		return errors.New("La base SQL Server ya contiene datos")
	}

	_, err = tx.ExecContext(ctx, `INSERT dbo.tipos_documento(codigo,nombre) VALUES
(N'DNI',N'Documento Nacional de Identidad'),(N'RUC',N'Registro Único de Contribuyentes'),
(N'CE',N'Carné de Extranjería'),(N'PASAPORTE',N'Pasaporte')`)
	if err != nil {
		return translate(err)
	}
	policy := domain.PermissionPolicy()
	for module, actions := range policy.Modules {
		for _, action := range actions {
			if _, err := tx.ExecContext(ctx, `INSERT dbo.permisos(codigo) VALUES(@p1)`, module+"."+action); err != nil {
				return translate(err)
			}
		}
	}
	for role, matrix := range policy.Defaults {
		var id int64
		if err := tx.QueryRowContext(ctx, `INSERT dbo.roles(nombre_rol) OUTPUT inserted.id_rol VALUES(@p1)`, role).Scan(&id); err != nil {
			return translate(err)
		}
		codes, err := matrix.Codes()
		if err != nil {
			return err
		}
		for _, code := range codes {
			if _, err := tx.ExecContext(ctx, `INSERT dbo.rol_permisos(id_rol,id_permiso)
SELECT @p1,id_permiso FROM dbo.permisos WHERE codigo=@p2`, id, code); err != nil {
				return translate(err)
			}
		}
	}
	for i, account := range accounts {
		_, err = tx.ExecContext(ctx, `
DECLARE @employee bigint = NULL;
IF @p4=N'Mecánico'
BEGIN
    SET @employee = NEXT VALUE FOR dbo.empleados_id_empleado_seq;
    INSERT dbo.empleados(id_empleado,nombre,tipo_empleado,activo) VALUES(@employee,@p1,@p4,1);
END;
DECLARE @user bigint = NEXT VALUE FOR dbo.usuarios_id_usuario_seq;
INSERT dbo.usuarios(id_usuario,nombre,usuario,password_hash,id_empleado) VALUES(@user,@p1,@p2,@p3,@employee);
INSERT dbo.usuario_roles(id_usuario,id_rol) SELECT @user,id_rol FROM dbo.roles WHERE nombre_rol=@p4;
IF @@ROWCOUNT <> 1 THROW 51000, N'Rol inicial no disponible.', 1;`, account.name, account.username, string(hashes[i]), account.role)
		if err != nil {
			return translate(err)
		}
	}
	return translate(tx.Commit())
}
