package postgresdb

import (
	"context"
	"errors"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
)

func (pg *PostgresDatabase) GetById(ctx context.Context, personId int) (person.Person, error) {
	var p person.Person
	err := pg.Get(&p, `SELECT * FROM person WHERE id=$1`, personId)
	if err != nil {
		return person.Person{}, err
	}
	return p, nil
}

func (pg *PostgresDatabase) GetByEmail(ctx context.Context, email string) (person.Person, error) {
	var pp []person.Person
	err := pg.SelectContext(ctx, &pp, `SELECT * FROM person WHERE email=$1`, email)
	if err != nil {
		return person.Person{}, err
	}
	if len(pp) != 1 {
		// should never happen
		return person.Person{}, errors.New("multiple person returned with same email. This should not happen as email is unique in the db")
	}

	return pp[0], nil
}

func (pg *PostgresDatabase) GetAll(ctx context.Context) ([]person.Person, error) {
	var res []person.Person
	err := pg.Select(&res, `SELECT * FROM person`)
	if err != nil {
		return []person.Person{}, err
	}
	return res, nil
}

func (pg *PostgresDatabase) Create(ctx context.Context, p person.Person) (int, error) {
	res, err := pg.ExecContext(
		ctx,
		`INSERT INTO person(name, email, password) VALUES ($1, $2, $3)  RETURNING id`,
		p.Name, p.Email, p.Password,
	)

	if err != nil {
		return 0, err
	}

	// It's not rowsaffected but the `RETURNING id` from the query.
	id, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}
