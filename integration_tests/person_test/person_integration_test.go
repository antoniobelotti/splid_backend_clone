//go:build integration
// +build integration

package person_test

import (
	"context"
	"github.com/antoniobelotti/splid_backend_clone/integration_tests"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/stretchr/testify/suite"
	psqlcont "github.com/testcontainers/testcontainers-go/modules/postgres"
	"testing"
)

type PersonTestSuite struct {
	suite.Suite
	psqlContainer *psqlcont.PostgresContainer
	personService person.Service
}

func (suite *PersonTestSuite) SetupTest() {
	db, cont := integration_tests.GetCleanContainerizedPsqlDb()
	suite.psqlContainer = cont
	suite.personService = person.NewService(db)
}

func (suite *PersonTestSuite) TearDownTest() {
	_ = suite.psqlContainer.Terminate(context.Background())
}

func TestPersonTestSuite(t *testing.T) {
	suite.Run(t, new(PersonTestSuite))
}

func (suite *PersonTestSuite) TestGetPersonByEmailFails() {
	p, err := suite.personService.GetPersonByEmail(context.Background(), "nonExistentEmail")

	suite.Equal(person.ErrPersonNotFound, err)
	suite.Equal(person.Person{}, p)
}
