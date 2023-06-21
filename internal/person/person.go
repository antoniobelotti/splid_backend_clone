package person

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

// Person - models users of the application
type Person struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"-"` // never exported in json
	Email    string `json:"email"`
}

// Store - this interface defines all methods the service needs to work
type Store interface {
	GetById(ctx context.Context, id int) (Person, error)
	Create(ctx context.Context, person Person) (int, error)
	GetByEmail(ctx context.Context, email string) (Person, error)
}

var (
	ErrPersonNotFound = errors.New("person does not exist")
	ErrUnexpected     = errors.New("unexpected error")
)

// Service - will handle all logic related to Person types
type Service struct {
	store Store
}

func NewService(store Store) Service {
	return Service{store: store}
}

func (s *Service) GetPersonById(ctx context.Context, id int) (Person, error) {
	return s.store.GetById(ctx, id)
}

func (s *Service) GetPersonByEmail(ctx context.Context, email string) (Person, error) {
	return s.store.GetByEmail(ctx, email)
}

func (s *Service) CreatePerson(ctx context.Context, name string, email string, clearPassword string) (Person, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(clearPassword), bcrypt.DefaultCost)
	if err != nil {
		return Person{}, ErrUnexpected
	}

	p := Person{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	id, err := s.store.Create(ctx, p)
	if err != nil {
		return Person{}, err
	}
	p.Id = id

	return p, nil
}
