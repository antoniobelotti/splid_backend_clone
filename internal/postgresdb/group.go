package postgresdb

import (
	"context"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
)

func (pg *PostgresDatabase) CreateGroup(ctx context.Context, g group.Group) (int, error) {
	transaction, err := pg.DB.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}

	var groupId int
	err = transaction.QueryRowContext(
		ctx,
		`	INSERT INTO "group"(name, owner_id, balance, invitation_code) 
				VALUES ($1, $2, $3, $4)
				RETURNING id;`,
		g.Name, g.OwnerId, g.Balance, g.InvitationCode,
	).Scan(&groupId)

	if err != nil {
		return 0, transaction.Rollback()
	}

	if _, err = transaction.ExecContext(
		ctx,
		`	INSERT INTO group_person(person_id, group_id) 
				VALUES ($1, $2);`,
		g.OwnerId,
		groupId,
	); err != nil {
		return 0, transaction.Rollback()
	}

	return groupId, transaction.Commit()
}

func (pg *PostgresDatabase) GetGroupById(ctx context.Context, groupId int) (group.Group, error) {
	var g group.Group
	err := pg.GetContext(ctx, &g, `SELECT * FROM "group" WHERE id=$1`, groupId)
	if err != nil {
		//TODO
		return g, err
	}

	return g, nil
}

func (pg *PostgresDatabase) AddPersonToGroup(ctx context.Context, g group.Group, personId int) error {
	res, err := pg.ExecContext(ctx, `INSERT INTO group_person(group_id, person_id) VALUES ($1, $2)`, g.Id, personId)
	if err != nil {
		//TODO
		return err
	}

	if ra, err := res.RowsAffected(); err != nil && ra != 1 {
		//todo
	}
	return nil
}

func (pg *PostgresDatabase) GetGroupComponentsById(ctx context.Context, groupId int) ([]int, error) {
	var componentIds []int
	err := pg.Select(&componentIds, `SELECT person_id FROM group_person WHERE group_id=$1`, groupId)
	if err != nil {
		// TODO
		return nil, err
	}
	return componentIds, nil
}
