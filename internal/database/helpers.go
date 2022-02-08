package database

import (
	"database/sql"
	"fmt"
	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"io/fs"
)

// version defines the current migration version. This ensures the app
// is always compatible with the version of the database.
const version = 1

// Migrate migrates the Postgres schema to the current version.
func ValidateSchema(db *sql.DB, scheme string, fSystem fs.FS) error {
	sourceInstance, err := iofs.New(fSystem, "migrations")
	if err != nil {
		return err
	}
	var driverInstance database.Driver
	switch scheme {
	case "postgres", "postgresql":
		driverInstance, err = postgres.WithInstance(db, new(postgres.Config))
	default:
		return fmt.Errorf("unknown scheme: %q", scheme)
	}
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", sourceInstance, scheme, driverInstance)
	if err != nil {
		return err
	}
	err = m.Migrate(version) // current version
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return sourceInstance.Close()
}
