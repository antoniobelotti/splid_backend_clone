package expense

import (
	"context"
	"fmt"
)

type Expense struct {
	Id            int `json:"id" db:"id"`
	AmountInCents int `json:"amount-in-cents" db:"amount_in_cents"`
	PersonId      int `json:"person-id" db:"person_id"`
	GroupId       int `json:"group-id" db:"group_id"`
}

type Store interface {
	CreateExpense(ctx context.Context, AmountInCents int, PersonId int, GroupId int) (int, error)
	IsPersonInGroup(ctx context.Context, groupId int, personId int) (bool, error)
	GetExpenseByGroupId(ctx context.Context, groupId int) ([]Expense, error)
}

type Service struct {
	store Store
}

func NewService(store Store) Service {
	return Service{store: store}
}

func (s *Service) CreateExpense(ctx context.Context, AmountInCents int, PersonId int, GroupId int) (Expense, error) {
	isPersonInGroup, err := s.store.IsPersonInGroup(ctx, GroupId, PersonId)
	if err != nil {
		return Expense{}, fmt.Errorf("unexpected error: %w", err)
	}
	if !isPersonInGroup {
		return Expense{}, fmt.Errorf("person id %d does not belong to group %d and as such cannot add an expense: %w", PersonId, GroupId, err)
	}

	e := Expense{
		AmountInCents: AmountInCents,
		PersonId:      PersonId,
		GroupId:       GroupId,
	}
	id, err := s.store.CreateExpense(ctx, AmountInCents, PersonId, GroupId)
	if err != nil {
		return Expense{}, fmt.Errorf("unable to create Expense: %w", err)
	}
	e.Id = id

	return e, nil
}

func (s *Service) GetExpenseByGroupId(ctx context.Context, groupId int) ([]Expense, error) {
	return s.store.GetExpenseByGroupId(ctx, groupId)
}
