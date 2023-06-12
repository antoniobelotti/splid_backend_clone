package person

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// Person - models users of the application
type Person struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"-"` // never exported in json
}

// Store - this interface defines all methods the service needs to work
type Store interface {
	GetById(ctx context.Context, id int) (Person, error)
	GetAll(ctx context.Context) ([]Person, error)
	Create(ctx context.Context, person Person) error
}

// Service - will handle all logic related to Person types
type Service struct {
	store Store
}

func NewService(store Store) Service {
	return Service{store: store}
}

func (s *Service) GetPerson(ctx context.Context, id int) (Person, error) {
	p, err := s.store.GetById(ctx, id)
	if err != nil {
		return Person{}, err
	}
	return p, nil
}

func (s *Service) GetAllPerson(ctx context.Context) ([]Person, error) {
	p, err := s.store.GetAll(ctx)
	if err != nil {
		fmt.Println(err)
		return []Person{}, err
	}
	return p, nil
}

func (s *Service) CreatePerson(ctx context.Context, name string, clearPassword string) (Person, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(clearPassword), bcrypt.DefaultCost)
	if err != nil {
		return Person{}, err
	}

	p := Person{
		Name:     name,
		Password: string(hashedPassword),
	}

	err = s.store.Create(ctx, p)
	if err != nil {
		return Person{}, err
	}
	return p, nil
}
