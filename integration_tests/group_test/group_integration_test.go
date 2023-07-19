package group_test

import (
	"context"
	"errors"
	"github.com/antoniobelotti/splid_backend_clone/integration_tests"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/stretchr/testify/suite"
	psqlcont "github.com/testcontainers/testcontainers-go/modules/postgres"
	"testing"
)

type GroupTestSuite struct {
	suite.Suite
	psqlContainer *psqlcont.PostgresContainer
	groupService  group.Service
	personService person.Service
}

func (suite *GroupTestSuite) SetupTest() {
	db, cont := integration_tests.GetCleanContainerizedPsqlDb()
	suite.psqlContainer = cont
	suite.groupService = group.NewService(db)
	suite.personService = person.NewService(db)
}

func (suite *GroupTestSuite) TearDownTest() {
	_ = suite.psqlContainer.Terminate(context.Background())
}

func TestPersonTestSuite(t *testing.T) {
	suite.Run(t, new(GroupTestSuite))
}

func (suite *GroupTestSuite) TestGetGroupByIdFail() {
	g, err := suite.groupService.GetGroupById(context.Background(), 999)
	suite.Assert().True(errors.Is(err, group.ErrGroupNotFound))
	suite.Assert().Equal(group.Group{}, g)
}

func (suite *GroupTestSuite) TestCreateExpenseSuccess() {
	p, err := suite.personService.CreatePerson(context.Background(), "person", "email@email.com", "testtest123")
	suite.Require().NoError(err)

	g, err := suite.groupService.CreateGroup(context.Background(), "testgroup", p.Id)
	suite.Require().NoError(err)

	e, err := suite.groupService.CreateExpense(context.Background(), 42, p.Id, g.Id)
	suite.Require().NoError(err)

	suite.Assert().Equal(42, e.AmountInCents)
	suite.Assert().Equal(p.Id, e.PersonId)
	suite.Assert().Equal(g.Id, e.GroupId)
}

func (suite *GroupTestSuite) TestCreateExpenseFail() {
	p, err := suite.personService.CreatePerson(context.Background(), "person", "email@email.com", "testtest123")
	suite.Require().NoError(err)

	g, err := suite.groupService.CreateGroup(context.Background(), "testgroup", p.Id)
	suite.Require().NoError(err)

	notInGroupPerson, err := suite.personService.CreatePerson(context.Background(), "person 2", "email2@email.com", "testtest123")
	suite.Require().NoError(err)

	e, err := suite.groupService.CreateExpense(context.Background(), 42, notInGroupPerson.Id, g.Id)
	suite.Require().NotNil(err)
	suite.Require().Empty(e)
}
