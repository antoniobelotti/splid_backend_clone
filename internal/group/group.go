package group

import (
	"context"
	"errors"
	"fmt"
	"github.com/antoniobelotti/splid_backend_clone/internal/expense"
	"github.com/antoniobelotti/splid_backend_clone/internal/transfer"
	"hash/fnv"
	"strconv"
)

type Group struct {
	Id             int    `json:"id" db:"id"`
	Name           string `json:"name" db:"name"`
	OwnerId        int    `json:"owner-id" db:"owner_id"`
	ComponentIds   []int  `json:"components"`
	InvitationCode string `json:"invitation-code" db:"invitation_code"`
}

type Store interface {
	CreateGroup(ctx context.Context, group Group) (int, error)
	GetGroupById(ctx context.Context, groupId int) (Group, error)
	AddPersonToGroup(ctx context.Context, g Group, personId int) error
	GetGroupComponentsById(ctx context.Context, groupId int) ([]int, error)
}

type Service struct {
	store           Store
	expenseService  expense.Service
	transferService transfer.Service
}

func NewService(store Store, es expense.Service, ts transfer.Service) Service {
	return Service{store: store, expenseService: es, transferService: ts}
}

var (
	ErrGroupNotFound = errors.New("group not found")
	ErrUnexpected    = errors.New("unexpected error")
)

func getHopefullyUniqueInvitationCode(groupName string, ownerId int) (string, error) {
	h := fnv.New32a()
	_, err := h.Write([]byte(fmt.Sprintf("%s %v", groupName, ownerId)))
	if err != nil {
		return "", err
	}
	asStr := strconv.Itoa(int(h.Sum32()))
	for len(asStr) < 6 {
		asStr = "0" + asStr
	}
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

func calculateGroupBalance(componentIds []int, expenses []expense.Expense, transfers []transfer.Transfer) map[int]int {
	// 	b_person1 = ((sum expenses person1) - avg all expenses) - (sum transfers from person1 -> any) + (sum transfers any -> person1)
	balance := make(map[int]int, len(componentIds))
	expensesSumByPerson := make(map[int]int)

	expensesSum := 0
	for _, e := range expenses {
		expensesSum += e.AmountInCents
		expensesSumByPerson[e.PersonId] += e.AmountInCents
	}
	expensesAverage := expensesSum / len(expenses)

	for _, personId := range componentIds {
		balance[personId] = expensesSumByPerson[personId] - expensesAverage
	}

	// now apply transfers
	for _, t := range transfers {
		// withdraw from sender
		balance[t.SenderId] -= t.AmountInCents
		// deposit to receiver
		balance[t.ReceiverId] += t.AmountInCents
	}

	return balance
}

func (s *Service) GetGroupBalance(ctx context.Context, groupId int) (map[int]int, error) {
	componentIds, err := s.GetGroupComponentsById(ctx, groupId)
	if err != nil {
		//TODO
		return nil, err
	}

	expenses, err := s.expenseService.GetExpenseByGroupId(ctx, groupId)
	if err != nil {
		//TODO
		return nil, err
	}

	transfers, err := s.transferService.GetTransfersByGroupId(ctx, groupId)
	if err != nil {
		//TODO
		return nil, err
	}

	return calculateGroupBalance(componentIds, expenses, transfers), nil
}
