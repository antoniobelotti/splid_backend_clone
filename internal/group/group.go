package group

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"
)

type Group struct {
	Id             int     `json:"id"`
	Name           string  `json:"name"`
	OwnerId        int     `json:"owner-id"`
	ComponentIds   []int   `json:"components"`
	Balance        float64 `json:"balance"`
	InvitationCode string  `json:"invitation-code"`
}

type Store interface {
	CreateGroup(ctx context.Context, group Group) error
}

type Service struct {
	store Store
}

func NewService(store Store) Service {
	return Service{store: store}
}

var (
	//ErrGroupNotFound = errors.New("person_test does not exist")
	ErrUnexpected = errors.New("unexpected error")
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

	err = s.store.CreateGroup(ctx, g)
	if err != nil {
		return Group{}, ErrUnexpected
	}
	return g, nil
}
