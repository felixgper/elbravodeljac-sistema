// Explicit first-account provisioning for a NEW isolated development database.
// Never runs from API startup and never overwrites accounts or permissions.
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func developmentTarget(dsn, confirmation string) (*pgxpool.Config, error) {
	if confirmation != "elbravodeljac_dev" {
		return nil, errors.New("Confirma explícitamente DEV_BOOTSTRAP_CONFIRM=elbravodeljac_dev")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil || cfg.ConnConfig.Database != "elbravodeljac_dev" {
		return nil, errors.New("Bootstrap solo permite la base aislada exacta elbravodeljac_dev")
	}
	return cfg, nil
}

func provision(ctx context.Context) error {
	cfg, err := developmentTarget(os.Getenv("POSTGRES_DSN"), os.Getenv("DEV_BOOTSTRAP_CONFIRM"))
	if err != nil {
		return err
	}
	username, password := os.Getenv("DEV_ADMIN_USERNAME"), os.Getenv("DEV_ADMIN_PASSWORD")
	if len(password) < 12 || len(password) > 72 || len(username) < 3 || len(username) > 60 {
		return errors.New("Credenciales de desarrollo ausentes o fuera de longitud permitida")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("No se pudo preparar bcrypt")
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return errors.New("No se pudo preparar PostgreSQL de desarrollo")
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return errors.New("PostgreSQL de desarrollo no disponible")
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('jac:dev-bootstrap:v1',0))`); err != nil {
		return errors.New("No se pudo bloquear la inicialización")
	}
	var tables, fks, users int
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'`).Scan(&tables); err != nil {
		return err
	}
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM pg_constraint WHERE contype='f' AND connamespace='public'::regnamespace`).Scan(&fks); err != nil {
		return err
	}
	if tables != 26 || fks != 46 {
		return errors.New("El esquema de desarrollo no coincide con las 26 tablas y 46 FK (42 originales y cuatro referencias aditivas)")
	}
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM usuarios`).Scan(&users); err != nil {
		return err
	}
	if users != 0 {
		log.Print("La base de desarrollo ya tiene usuarios; no se modifica ninguna cuenta ni permisos")
		return nil
	}
	policy := domain.PermissionPolicy()
	for module, actions := range policy.Modules {
		for _, action := range actions {
			if _, err = tx.Exec(ctx, `INSERT INTO permisos(codigo) VALUES($1) ON CONFLICT(codigo) DO NOTHING`, module+"."+action); err != nil {
				return err
			}
		}
	}
	for name, matrix := range policy.Defaults {
		var roleID string
		if err = tx.QueryRow(ctx, `INSERT INTO roles(nombre_rol) VALUES($1) ON CONFLICT(nombre_rol) DO UPDATE SET nombre_rol=excluded.nombre_rol RETURNING id_rol::text`, name).Scan(&roleID); err != nil {
			return err
		}
		codes, err := matrix.Codes()
		if err != nil {
			return err
		}
		for _, code := range codes {
			if _, err = tx.Exec(ctx, `INSERT INTO rol_permisos(id_rol,id_permiso) SELECT $1,id_permiso FROM permisos WHERE codigo=$2 ON CONFLICT DO NOTHING`, roleID, code); err != nil {
				return err
			}
		}
	}
	var roleID, userID string
	if err = tx.QueryRow(ctx, `SELECT id_rol::text FROM roles WHERE nombre_rol='Administrador'`).Scan(&roleID); err != nil {
		return err
	}
	name := "Administración de pruebas"
	if err = (domain.UserInput{Name: &name, Username: &username, Password: &password, RoleID: &roleID}).Validate(true); err != nil {
		return err
	}
	if err = tx.QueryRow(ctx, `INSERT INTO usuarios(nombre,usuario,password_hash) VALUES($1,$2,$3) RETURNING id_usuario::text`, name, username, string(hash)).Scan(&userID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO usuario_roles(id_usuario,id_rol) VALUES($1,$2)`, userID, roleID); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	log.Print("Cuenta inicial y política aprobada creadas en la base aislada de desarrollo; credenciales omitidas")
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := provision(ctx); err != nil {
		log.Fatal("Inicialización de desarrollo rechazada o fallida (sin publicar credenciales): ", err)
	}
}
