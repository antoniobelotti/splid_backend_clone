package postgresdb

import (
	"context"
	"fmt"
	"github.com/antoniobelotti/splid_backend_clone/internal/expense"
)

func (pg *PostgresDatabase) IsPersonInGroup(ctx context.Context, groupId int, personId int) (bool, error) {
	var personInGroup bool
	if err := pg.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM group_person WHERE person_id=$1 AND group_id=$2)`,
		personId, groupId,
	).Scan(&personInGroup); err != nil {
		return false, fmt.Errorf("IsPersonInGroup unexpected: %w", err)
	}
	return personInGroup, nil
}

func (pg *PostgresDatabase) CreateExpense(ctx context.Context, AmountInCents int, PersonId int, GroupId int) (int, error) {
	var expenseId int
	err := pg.QueryRowContext(
		ctx,
		`INSERT INTO expense(amount_in_cents, person_id, group_id)
				VALUES ($1, $2, $3)
				RETURNING id`,
		AmountInCents, PersonId, GroupId,
	).Scan(&expenseId)
	if err != nil {
		return 0, fmt.Errorf("CreateExpense unable to insert: %w", err)
	}

	return expenseId, nil
}

func (pg *PostgresDatabase) GetExpenseByGroupId(ctx context.Context, groupId int) ([]expense.Expense, error) {
	var expenses []expense.Expense
	err := pg.SelectContext(ctx, &expenses, `SELECT * FROM expense WHERE group_id=$1`, groupId)
	if err != nil {
		return nil, err
	}
	return expenses, nil
}
