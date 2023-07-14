//go:build integration

package person_test

import (
	"context"
	"fmt"
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

func (suite *PersonTestSuite) TestCreatePersonReturnsCorrectId() {
	p, err := suite.personService.CreatePerson(context.Background(), "test person", "mail@gmail.com", "test123pwd")
	fmt.Println(err)
	suite.Require().NoError(err)
	suite.Equal(1, p.Id)

	p2, err := suite.personService.CreatePerson(context.Background(), "test person 2", "mail2@gmail.com", "test123pwd")
	suite.Require().NoError(err)
	suite.Equal(2, p2.Id)
}
