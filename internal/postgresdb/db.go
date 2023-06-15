package postgresdb

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"os"
)

var (
	ErrDBConnectionError = errors.New("unable to connect to postgres database")
)

type PostgresDatabase struct {
	*sqlx.DB
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("%s env variable is not set", key))
	}
	return value
}

func NewDatabase() (*PostgresDatabase, error) {
	connectionStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		mustGetEnv("DB_HOST"),
		mustGetEnv("DB_PORT"),
		mustGetEnv("DB_NAME"),
		mustGetEnv("DB_USERNAME"),
		mustGetEnv("DB_PASSWORD"),
		mustGetEnv("DB_SSLMODE"),
	)
	db, err := sqlx.Connect("postgres", connectionStr)
	if err != nil {
		return &PostgresDatabase{}, ErrDBConnectionError
	}

	return &PostgresDatabase{db}, nil
}
