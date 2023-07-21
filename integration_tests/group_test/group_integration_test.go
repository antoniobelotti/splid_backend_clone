//go:build integration

package group_test

import (
	"context"
	"errors"
	"github.com/antoniobelotti/splid_backend_clone/integration_tests"
	"github.com/antoniobelotti/splid_backend_clone/internal/expense"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/antoniobelotti/splid_backend_clone/internal/transfer"
	"github.com/stretchr/testify/suite"
	psqlcont "github.com/testcontainers/testcontainers-go/modules/postgres"
	"testing"
)

type GroupTestSuite struct {
	suite.Suite
	psqlContainer   *psqlcont.PostgresContainer
	groupService    group.Service
	personService   person.Service
	expenseService  expense.Service
	transferService transfer.Service
}

func (suite *GroupTestSuite) SetupTest() {
	db, cont := integration_tests.GetCleanContainerizedPsqlDb()
	suite.psqlContainer = cont
	suite.personService = person.NewService(db)
	suite.expenseService = expense.NewService(db)
	suite.transferService = transfer.NewService(db)
	suite.groupService = group.NewService(db, suite.expenseService, suite.transferService)
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

func (suite *GroupTestSuite) TestGetBalanceSuccess() {
	c := context.Background()
	pwd := "passowrd123"
	p1, err := suite.personService.CreatePerson(c, "person 1", "email@email.com", pwd)
	suite.Require().NoError(err)
	p2, err := suite.personService.CreatePerson(c, "person 2", "email2@email.com", pwd)
	suite.Require().NoError(err)

	g, err := suite.groupService.CreateGroup(c, "testgroup", p1.Id)
	suite.Require().NoError(err)

	err = suite.groupService.AddPersonToGroup(c, g, p2.Id)
	suite.Require().NoError(err)

	// add some expenses
	// p1: 10€ + 5.30€
	// p2: 2.30€
	exp1, err := suite.expenseService.CreateExpense(c, 1000, p1.Id, g.Id)
	suite.Require().NoError(err)
	exp2, err := suite.expenseService.CreateExpense(c, 530, p1.Id, g.Id)
	suite.Require().NoError(err)
	exp3, err := suite.expenseService.CreateExpense(c, 230, p2.Id, g.Id)
	suite.Require().NoError(err)

	balance, err := suite.groupService.GetGroupBalance(c, g.Id)
	suite.Require().NoError(err)

	avg := (exp1.AmountInCents + exp2.AmountInCents + exp3.AmountInCents) / 3
	expected := map[int]int{
		p1.Id: (exp1.AmountInCents + exp2.AmountInCents) - avg,
		p2.Id: exp3.AmountInCents - avg,
	}

	suite.Assert().Equal(expected, balance)

	// p2 returns 5€ to p1
	t, err := suite.transferService.CreateTransfer(c, 500, g.Id, p2.Id, p1.Id)
	suite.Require().NoError(err)

	expected = map[int]int{
		p1.Id: (exp1.AmountInCents + exp2.AmountInCents) - avg + t.AmountInCents,
		p2.Id: exp3.AmountInCents - avg - t.AmountInCents,
	}
	balance, err = suite.groupService.GetGroupBalance(c, g.Id)
	suite.Require().NoError(err)

	suite.Assert().Equal(expected, balance)
}
