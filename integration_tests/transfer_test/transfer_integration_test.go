package transfer_test

import (
	"context"
	"github.com/antoniobelotti/splid_backend_clone/integration_tests"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/antoniobelotti/splid_backend_clone/internal/transfer"
	"github.com/stretchr/testify/suite"
	psqlcont "github.com/testcontainers/testcontainers-go/modules/postgres"
	"testing"
)

type TransferTestSuite struct {
	suite.Suite
	psqlContainer   *psqlcont.PostgresContainer
	transferService transfer.Service
	groupService    group.Service
	personService   person.Service
}

func (suite *TransferTestSuite) SetupTest() {
	db, cont := integration_tests.GetCleanContainerizedPsqlDb()
	suite.psqlContainer = cont
	suite.transferService = transfer.NewService(db)
	suite.groupService = group.NewService(db)
	suite.personService = person.NewService(db)
}

func (suite *TransferTestSuite) TearDownTest() {
	_ = suite.psqlContainer.Terminate(context.Background())
}

func TestTransferTestSuite(t *testing.T) {
	suite.Run(t, new(TransferTestSuite))
}

func (suite *TransferTestSuite) TestCreateTransferSuccess() {
	sender, err := suite.personService.CreatePerson(context.Background(), "person", "email@email.com", "testtest123")
	suite.Require().NoError(err)

	receiver, err := suite.personService.CreatePerson(context.Background(), "person", "sknvnkvsjnvd@email.com", "testtest123")
	suite.Require().NoError(err)

	g, err := suite.groupService.CreateGroup(context.Background(), "testgroup", sender.Id)
	suite.Require().NoError(err)

	err = suite.groupService.AddPersonToGroup(context.Background(), g, receiver.Id)
	suite.Require().NoError(err)

	e, err := suite.transferService.CreateTransfer(context.Background(), 42, g.Id, sender.Id, receiver.Id)
	suite.Require().NoError(err)

	suite.Assert().Equal(42, e.AmountInCents)
	suite.Assert().Equal(sender.Id, e.SenderId)
	suite.Assert().Equal(receiver.Id, e.ReceiverId)
	suite.Assert().Equal(g.Id, e.GroupId)
}

func (suite *TransferTestSuite) TestCreateTransferFail() {
	p, err := suite.personService.CreatePerson(context.Background(), "person", "email@email.com", "testtest123")
	suite.Require().NoError(err)

	g, err := suite.groupService.CreateGroup(context.Background(), "testgroup", p.Id)
	suite.Require().NoError(err)

	notInGroupPerson, err := suite.personService.CreatePerson(context.Background(), "person 2", "email2@email.com", "testtest123")
	suite.Require().NoError(err)

	e, err := suite.transferService.CreateTransfer(context.Background(), 42, g.Id, p.Id, notInGroupPerson.Id)
	suite.Require().NotNil(err)
	suite.Require().Empty(e)
}
