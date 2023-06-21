package postgresdb

import (
	"context"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
)

func (pg *PostgresDatabase) CreateGroup(ctx context.Context, g group.Group) error {
	transaction, err := pg.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	if _, err = transaction.ExecContext(
		ctx,
		`	INSERT INTO "group"(name, owner_id, balance, invitation_code) 
				VALUES ($1, $2, $3, $4);`,
		g.Name, g.OwnerId, g.Balance, g.InvitationCode,
	); err != nil {
		return transaction.Rollback()
	}

	if _, err = transaction.ExecContext(
		ctx,
		`	INSERT INTO group_person(person_id, group_id) 
				VALUES ($1, $2);`,
		g.OwnerId,
		g.Id,
	); err != nil {
		return transaction.Rollback()
	}

	return transaction.Commit()
}
