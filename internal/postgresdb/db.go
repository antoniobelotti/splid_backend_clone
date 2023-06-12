package postgresdb

import (
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
	"os"
)

var (
	ErrDBConnectionError = errors.New("unable to connect to postgres database")
)

type PostgresDatabase struct {
	*sqlx.DB
}

func NewDatabase() (*PostgresDatabase, error) {
	connectionStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_SSLMODE"),
	)
	db, err := sqlx.Connect("postgres", connectionStr)
	if err != nil {
		return &PostgresDatabase{}, ErrDBConnectionError
	}

	return &PostgresDatabase{db}, nil
}
