package postgresdb

import (
	"context"
	"fmt"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
)

func (pg *PostgresDatabase) CreateGroup(ctx context.Context, g group.Group) error {
	transaction, err := pg.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	fmt.Println("test")
	_, err = transaction.NamedExecContext(
		ctx,
		`	INSERT INTO "group"(name, owner_id, balance, invitation_code) 
				VALUES (:name, :ownerId, :balance, :invitationCode);`,
		g,
	)

	_, err = transaction.ExecContext(
		ctx,
		`	INSERT INTO group_person(person_id, group_id) 
				VALUES ( ? , ? );`,
		g.OwnerId,
		g.Id,
	)
	if err != nil {
		return err
	}

	err = transaction.Commit()
	if err != nil {
		return err
	}
	return nil
}
