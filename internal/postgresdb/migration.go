package postgresdb

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"

	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5"
)

var (
	ErrFailedToReadMigrations = errors.New("unable to read migrations")
	ErrApplyMigrationFail     = errors.New("unable to apply migrations")
)

// MigrateDB - runs all migrations in the migrations
func (pg *PostgresDatabase) MigrateDB() error {
	driver, err := postgres.WithInstance(pg.DB.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create the postgres driver: %w", err) // TODO
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file:///migrations",
		"postgres", driver,
	)
	if err != nil {
		fmt.Println(err.Error())
		return ErrFailedToReadMigrations
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return ErrApplyMigrationFail
	}

	return nil
}
