package postgresdb

import (
	"context"
	"database/sql"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
)

func (pg *PostgresDatabase) GetById(ctx context.Context, personId int) (person.Person, error) {
	var p person.Person
	err := pg.GetContext(ctx, &p, `SELECT * FROM person WHERE id=$1`, personId)
	return p, err
}

func (pg *PostgresDatabase) GetByEmail(ctx context.Context, email string) (person.Person, error) {
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

func (pg *PostgresDatabase) GetAll(ctx context.Context) ([]person.Person, error) {
	var res []person.Person
	err := pg.SelectContext(ctx, &res, `SELECT * FROM person`)
	return res, err
}

func (pg *PostgresDatabase) Create(ctx context.Context, p person.Person) (int, error) {
	res, err := pg.ExecContext(
		ctx,
		`INSERT INTO person(name, email, password) VALUES ($1, $2, $3)  RETURNING id`,
		p.Name, p.Email, p.Password,
	)

	if err != nil {
		return 0, person.ErrUnexpected
	}

	// It's not rowsaffected but the `RETURNING id` from the query.
	id, err := res.RowsAffected()
	if err != nil {
		return 0, person.ErrUnexpected
	}

	return int(id), nil
}
