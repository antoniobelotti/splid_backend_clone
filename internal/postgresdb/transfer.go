package postgresdb

import (
	"context"
	"fmt"
	"github.com/antoniobelotti/splid_backend_clone/internal/transfer"
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

func (pg *PostgresDatabase) GetTransfersByGroupId(ctx context.Context, groupId int) ([]transfer.Transfer, error) {
	var transfers []transfer.Transfer
	err := pg.SelectContext(ctx, &transfers, `SELECT * FROM transfer WHERE group_id=$1`, groupId)
	if err != nil {
		return nil, err
	}
	return transfers, nil
}
