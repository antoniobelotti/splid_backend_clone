package postgresdb

import (
	"context"
	"fmt"
)

func (pg *PostgresDatabase) CreateTransfer(ctx context.Context, amountInCents int, groupId int, senderId int, receiverId int) (int, error) {
	var transferId int
	err := pg.QueryRowContext(
		ctx,
		`INSERT INTO transfer(amount_in_cents, sender_id, receiver_id, group_id)
				VALUES ($1, $2, $3, $4)
				RETURNING id`,
		amountInCents, senderId, receiverId, groupId,
	).Scan(&transferId)
	if err != nil {
		return 0, fmt.Errorf("CreateTransfer unable to insert: %w", err)
	}

	return transferId, nil
}
