package group

import (
	"context"
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
		return Group{}, err
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
		return Group{}, err
	}
	return g, nil
}
