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

type Store interface {
	CreateGroup(ctx context.Context, group Group) (int, error)
	GetGroupById(ctx context.Context, groupId int) (Group, error)
	AddPersonToGroup(ctx context.Context, g Group, personId int) error
	GetGroupComponentsById(ctx context.Context, groupId int) ([]int, error)
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
