//go:build integration

package expense_test

import (
	"context"
	"github.com/antoniobelotti/splid_backend_clone/integration_tests"
	"github.com/antoniobelotti/splid_backend_clone/internal/expense"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/antoniobelotti/splid_backend_clone/internal/transfer"
	"github.com/stretchr/testify/suite"
	psqlcont "github.com/testcontainers/testcontainers-go/modules/postgres"
	"testing"
)

type ExpenseTestSuite struct {
	suite.Suite
	psqlContainer  *psqlcont.PostgresContainer
	expenseService expense.Service
	groupService   group.Service
	personService  person.Service
}

func (suite *ExpenseTestSuite) SetupTest() {
	db, cont := integration_tests.GetCleanContainerizedPsqlDb()
	suite.psqlContainer = cont
	suite.expenseService = expense.NewService(db)
	suite.personService = person.NewService(db)
	suite.groupService = group.NewService(db, suite.expenseService, transfer.NewService(db))
}

func (suite *ExpenseTestSuite) TearDownTest() {
	_ = suite.psqlContainer.Terminate(context.Background())
}

func TestExpenseTestSuite(t *testing.T) {
	suite.Run(t, new(ExpenseTestSuite))
}

func (suite *ExpenseTestSuite) TestCreateExpenseSuccess() {
	p, err := suite.personService.CreatePerson(context.Background(), "person", "email@email.com", "testtest123")
	suite.Require().NoError(err)

	g, err := suite.groupService.CreateGroup(context.Background(), "testgroup", p.Id)
	suite.Require().NoError(err)

	e, err := suite.expenseService.CreateExpense(context.Background(), 42, p.Id, g.Id)
	suite.Require().NoError(err)

	suite.Assert().Equal(42, e.AmountInCents)
	suite.Assert().Equal(p.Id, e.PersonId)
	suite.Assert().Equal(g.Id, e.GroupId)
}

func (suite *ExpenseTestSuite) TestCreateExpenseFail() {
	p, err := suite.personService.CreatePerson(context.Background(), "person", "email@email.com", "testtest123")
	suite.Require().NoError(err)

	g, err := suite.groupService.CreateGroup(context.Background(), "testgroup", p.Id)
	suite.Require().NoError(err)

	notInGroupPerson, err := suite.personService.CreatePerson(context.Background(), "person 2", "email2@email.com", "testtest123")
	suite.Require().NoError(err)

	e, err := suite.expenseService.CreateExpense(context.Background(), 42, notInGroupPerson.Id, g.Id)
	suite.Require().NotNil(err)
	suite.Require().Empty(e)
}
