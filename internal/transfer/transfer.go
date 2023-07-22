package transfer

import (
	"context"
	"fmt"
)

type Transfer struct {
	Id            int `json:"id,omitempty" db:"id"`
	AmountInCents int `json:"amount-in-cents" db:"amount_in_cents"`
	GroupId       int `json:"group-id,omitempty" db:"group_id"`
	SenderId      int `json:"sender-id" db:"sender_id"`
	ReceiverId    int `json:"receiver-id" db:"receiver_id"`
}

type Store interface {
	CreateTransfer(ctx context.Context, amountInCents int, groupId int, senderId int, receiverId int) (int, error)
	IsPersonInGroup(ctx context.Context, groupId int, personId int) (bool, error)
	GetTransfersByGroupId(ctx context.Context, groupId int) ([]Transfer, error)
}

type Service struct {
	store Store
}

func NewService(store Store) Service {
	return Service{store: store}
}

func (s *Service) CreateTransfer(ctx context.Context, amountInCents int, groupId int, senderId int, receiverId int) (Transfer, error) {
	isSenderInGroup, err := s.store.IsPersonInGroup(ctx, groupId, senderId)
	if err != nil {
		return Transfer{}, fmt.Errorf("unexpected error: %w", err)
	}
	isReceiverInGroup, err := s.store.IsPersonInGroup(ctx, groupId, receiverId)
	if err != nil {
		return Transfer{}, fmt.Errorf("unexpected error: %w", err)
	}
	if !isSenderInGroup || !isReceiverInGroup {
		return Transfer{}, fmt.Errorf("either sender of receiver do not belong to group %d: %w", groupId, err)
	}

	e := Transfer{
		AmountInCents: amountInCents,
		GroupId:       groupId,
		SenderId:      senderId,
		ReceiverId:    receiverId,
	}
	id, err := s.store.CreateTransfer(ctx, amountInCents, groupId, senderId, receiverId)
	if err != nil {
		return Transfer{}, fmt.Errorf("unable to create Transfer: %w", err)
	}
	e.Id = id

	return e, nil
}

func (s *Service) GetTransfersByGroupId(ctx context.Context, groupId int) ([]Transfer, error) {
	return s.store.GetTransfersByGroupId(ctx, groupId)
}
