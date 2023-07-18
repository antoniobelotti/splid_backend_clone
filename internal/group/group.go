package group

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"
)

type Group struct {
	Id             int     `json:"id" db:"id"`
	Name           string  `json:"name" db:"name"`
	OwnerId        int     `json:"owner-id" db:"owner_id"`
	ComponentIds   []int   `json:"components"`
	Balance        float64 `json:"balance" db:"balance"`
	InvitationCode string  `json:"invitation-code" db:"invitation_code"`
}

type Expense struct {
	Id            int `json:"id" db:"id"`
	AmountInCents int `json:"amount-in-cents" db:"amount_in_cents"`
	PersonId      int `json:"person-id" db:"person_id"`
	GroupId       int `json:"group-id" db:"group_id"`
}

type Store interface {
	CreateGroup(ctx context.Context, group Group) (int, error)
	GetGroupById(ctx context.Context, groupId int) (Group, error)
	AddPersonToGroup(ctx context.Context, g Group, personId int) error
	GetGroupComponentsById(ctx context.Context, groupId int) ([]int, error)

	// CreateExpense : even if Expense is another entity, I think it belongs here
	// because an expense does not exist outside the context of a group
	CreateExpense(ctx context.Context, AmountInCents int, PersonId int, GroupId int) (int, error)
	IsPersonInGroup(ctx context.Context, groupId int, personId int) (bool, error)
}

type Service struct {
	store Store
}

func NewService(store Store) Service {
	return Service{store: store}
}

var (
	ErrGroupNotFound = errors.New("group not found")
	ErrUnexpected    = errors.New("unexpected error")
)

func getHopefullyUniqueInvitationCode(groupName string, ownerId int) (string, error) {
	h := fnv.New32a()
	_, err := h.Write([]byte(fmt.Sprintf("%s%d", groupName, ownerId)))
	if err != nil {
		return "", err
	}
	asStr := strconv.Itoa(int(h.Sum32()))
	return asStr[:6], nil
}

func (s *Service) CreateGroup(ctx context.Context, name string, ownerId int) (Group, error) {
	invitationCode, err := getHopefullyUniqueInvitationCode(name, ownerId)
	if err != nil {
		return Group{}, ErrUnexpected
	}

	var g = Group{
		Name:           name,
		OwnerId:        ownerId,
		ComponentIds:   nil,
		Balance:        0,
		InvitationCode: invitationCode,
	}

	g.Id, err = s.store.CreateGroup(ctx, g)
	if err != nil {
		return Group{}, ErrUnexpected
	}
	return g, nil
}

func (s *Service) GetGroupById(ctx context.Context, groupId int) (Group, error) {
	return s.store.GetGroupById(ctx, groupId)
}

func (s *Service) AddPersonToGroup(ctx context.Context, g Group, personId int) error {
	return s.store.AddPersonToGroup(ctx, g, personId)
}

func (s *Service) GetGroupComponentsById(ctx context.Context, groupId int) ([]int, error) {
	return s.store.GetGroupComponentsById(ctx, groupId)
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
