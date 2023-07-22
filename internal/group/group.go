package group

import (
	"context"
	"errors"
	"fmt"
	"github.com/antoniobelotti/splid_backend_clone/internal/expense"
	"github.com/antoniobelotti/splid_backend_clone/internal/transfer"
	"hash/fnv"
	"sort"
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
	// TODO: should probably accept a context and cancel operation if timeout exceeded

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

func calculateOpsToEvenBalance(currentBalance map[int]int) []transfer.Transfer {
	type pair struct {
		pId, amount int
	}
	var (
		creditors []pair
		debtors   []pair
	)
	for person, balance := range currentBalance {
		if balance < 0 {
			debtors = append(debtors, pair{pId: person, amount: -balance})
		} else if balance > 0 {
			creditors = append(creditors, pair{pId: person, amount: balance})
		}
	}

	sort.Slice(creditors, func(i, j int) bool {
		return creditors[i].amount < creditors[j].amount
	})
	sort.Slice(debtors, func(i, j int) bool {
		return debtors[i].amount < debtors[j].amount
	})

	// extinguish the smallest debt first.
	// TODO: it would be awesome to minimize the total number of transfers... optimal solution instead of this greedy garbage
	var transfers []transfer.Transfer
	for len(creditors) > 0 || len(debtors) > 0 {

		var transferAmount int
		if debtors[0].amount <= creditors[0].amount {
			// whole debt can be extinguished in one transfer. Credit cannot
			transferAmount = debtors[0].amount
		} else {
			// whole credit can be extinguished in one transfer. Debt cannot
			transferAmount = creditors[0].amount
		}

		transfers = append(transfers, transfer.Transfer{
			AmountInCents: transferAmount,
			SenderId:      debtors[0].pId,
			ReceiverId:    creditors[0].pId,
		})
		creditors[0].amount -= transferAmount
		debtors[0].amount -= transferAmount

		if debtors[0].amount == 0 {
			debtors = debtors[1:]
		}
		if creditors[0].amount == 0 {
			creditors = creditors[1:]
		}
	}
	return transfers
}

func (s *Service) GetOpsEvenBalance(ctx context.Context, groupId int) ([]transfer.Transfer, error) {
	currentBalance, err := s.GetGroupBalance(ctx, groupId)
	if err != nil {
		//todo
		return []transfer.Transfer{}, nil
	}
	return calculateOpsToEvenBalance(currentBalance), nil
}
