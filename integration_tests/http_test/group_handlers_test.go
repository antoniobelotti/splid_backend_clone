//go:build integration

package http_test

import (
	"context"
	"fmt"
	"github.com/antoniobelotti/splid_backend_clone/integration_tests"
	"github.com/antoniobelotti/splid_backend_clone/internal/expense"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	internal_http "github.com/antoniobelotti/splid_backend_clone/internal/http"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/antoniobelotti/splid_backend_clone/internal/transfer"
	"github.com/stretchr/testify/suite"
	psqlcont "github.com/testcontainers/testcontainers-go/modules/postgres"
	"net/http"
	"testing"
)

type GroupHandlerTestSuite struct {
	testSuiteHttp
	psqlContainer *psqlcont.PostgresContainer
	personService person.Service
	groupService  group.Service
}

func TestGroupHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(GroupHandlerTestSuite))
}

func (suite *GroupHandlerTestSuite) TearDownTest() {
	_ = suite.psqlContainer.Terminate(context.Background())
}

func (suite *GroupHandlerTestSuite) SetupTest() {
	db, cont := integration_tests.GetCleanContainerizedPsqlDb()

	suite.psqlContainer = cont
	suite.personService = person.NewService(db)
	suite.groupService = group.NewService(db, expense.NewService(db), transfer.NewService(db))

	suite.server = internal_http.NewRESTServer(suite.personService, suite.groupService, expense.Service{}, transfer.Service{})
}

func (suite *GroupHandlerTestSuite) TestCreateGroupSuccess() {
	p, signedToken := suite.GetLoggedInPerson()

	// create group request
	want := group.Group{
		Name:    "testGroup",
		OwnerId: p.Id,
	}
	createGroupRequestBody := internal_http.CreateGroupRequestBody{Name: want.Name}
	response := suite.POSTWithJwt("/api/v1/group", createGroupRequestBody, signedToken)
	suite.Equal(http.StatusCreated, response.Code)

	got := ExtractBody[group.Group](response)

	suite.Equal(want.Name, got.Name)
	suite.Equal(want.OwnerId, got.OwnerId)
	suite.NotEmpty(got.InvitationCode)
}

func (suite *GroupHandlerTestSuite) TestJoinGroupSuccess() {
	groupOwner, err := suite.personService.CreatePerson(context.Background(), "testPerson", "mail@email.com", "password123")
	suite.Require().NoError(err)
	g, err := suite.groupService.CreateGroup(context.Background(), "testGroup", groupOwner.Id)
	suite.Require().NoError(err)

	p, signedToken := suite.GetLoggedInPerson()

	// perform request to join group g as user p
	joinGroupResponse := suite.POSTWithJwt(
		fmt.Sprintf("/api/v1/group/%d/join?invitationCode=%s", g.Id, g.InvitationCode),
		nil,
		signedToken,
	)
	suite.Equal(http.StatusOK, joinGroupResponse.Code)

	components, err := suite.groupService.GetGroupComponentsById(context.Background(), g.Id)
	suite.Require().Contains(components, p.Id)
}

func (suite *GroupHandlerTestSuite) TestJoinGroupFailGivenWrongInvitationCode() {
	groupOwner, err := suite.personService.CreatePerson(context.Background(), "testPerson", "mail@email.com", "password123")
	suite.Require().NoError(err)
	g, err := suite.groupService.CreateGroup(context.Background(), "testGroup", groupOwner.Id)
	suite.Require().NoError(err)

	p, signedToken := suite.GetLoggedInPerson()

	// perform request to join group g as user p
	joinGroupResponse := suite.POSTWithJwt(
		fmt.Sprintf("/api/v1/group/%d/join?invitationCode=%s", g.Id, "INVALID_INVITATION_CODE"),
		nil,
		signedToken,
	)
	suite.Equal(http.StatusUnauthorized, joinGroupResponse.Code)

	components, err := suite.groupService.GetGroupComponentsById(context.Background(), g.Id)
	suite.Require().NotContains(components, p.Id)
}

func (suite *GroupHandlerTestSuite) TestJoinGroupFailIfGroupDoesNotExist() {
	groupOwner, err := suite.personService.CreatePerson(context.Background(), "testPerson", "mail@email.com", "password123")
	suite.Require().NoError(err)
	g, err := suite.groupService.CreateGroup(context.Background(), "testGroup", groupOwner.Id)
	suite.Require().NoError(err)

	_, signedToken := suite.GetLoggedInPerson()

	nonexistentGroupId := 999
	// perform request to join group g as user p
	joinGroupResponse := suite.POSTWithJwt(
		fmt.Sprintf("/api/v1/group/%d/join?invitationCode=%s", nonexistentGroupId, g.InvitationCode),
		nil,
		signedToken,
	)
	suite.Equal(http.StatusBadRequest, joinGroupResponse.Code)
}
