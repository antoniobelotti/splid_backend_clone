//go:build integration

package integration_tests

import (
	"context"
	"github.com/antoniobelotti/splid_backend_clone/internal/postgresdb"
	"github.com/joho/godotenv"
	"github.com/testcontainers/testcontainers-go"
	postgrestc "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func GetCleanContainerizedPsqlDb() (*postgresdb.PostgresDatabase, *postgrestc.PostgresContainer) {
	err := godotenv.Load("../../.env.dev")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()

	container, err := postgrestc.RunContainer(ctx,
		testcontainers.WithImage("postgres:12-alpine"),
		postgrestc.WithDatabase("postgres"),
		postgrestc.WithUsername("postgres"),
		postgrestc.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(20*time.Second)),
	)
	if err != nil {
		panic(err.Error())
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable", "application_name=test")

	db, err := postgresdb.NewDatabase(connStr)
	if err != nil {
		panic(err.Error())
	}

	driver, err := postgres.WithInstance(db.DB.DB, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations/",
		"postgres", driver)
	if err != nil {
		panic(err.Error())
	}

	err = m.Up()
	if err != nil {
		panic(err.Error())
	}

	return db, container
}
