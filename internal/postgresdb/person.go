package postgresdb

import (
	"context"
	"database/sql"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
)

func (pg *PostgresDatabase) GetPersonById(ctx context.Context, personId int) (person.Person, error) {
	var p person.Person
	err := pg.GetContext(ctx, &p, `SELECT * FROM person WHERE id=$1`, personId)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return p, person.ErrPersonNotFound
		default:
			return p, person.ErrUnexpected
		}
	}
	return p, err
}

func (pg *PostgresDatabase) GetPersonByEmail(ctx context.Context, email string) (person.Person, error) {
	var p person.Person
	err := pg.GetContext(ctx, &p, `SELECT * FROM person WHERE email=$1`, email)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return p, person.ErrPersonNotFound
		default:
			return p, person.ErrUnexpected
		}
	}
	return p, err
}

func (pg *PostgresDatabase) CreatePerson(ctx context.Context, p person.Person) (int, error) {
	var personId int
	err := pg.QueryRowContext(
		ctx,
		`INSERT INTO person(name, email, password) VALUES ($1, $2, $3)  RETURNING id`,
		p.Name, p.Email, p.Password,
	).Scan(&personId)

	if err != nil {
		return 0, person.ErrUnexpected
	}

	return personId, nil
}
