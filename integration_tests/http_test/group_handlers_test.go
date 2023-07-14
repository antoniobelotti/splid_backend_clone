//go:build integration

package http_test

import (
	"context"
	"fmt"
	"github.com/antoniobelotti/splid_backend_clone/integration_tests"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	internal_http "github.com/antoniobelotti/splid_backend_clone/internal/http"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
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
	suite.groupService = group.NewService(db)

	suite.server = internal_http.NewRESTServer(suite.personService, suite.groupService)
}

func (suite *GroupHandlerTestSuite) TestCreateGroupSuccess() {
	p, err := suite.personService.CreatePerson(context.Background(), "testPerson", "mail@email.com", "password123")
	suite.Require().NoError(err)

	// only logged-in users can create groups
	requestBody := internal_http.LoginRequestBody{
		Email:    p.Email,
		Password: "password123",
	}

	response := suite.POST("/api/v1/person/login", requestBody)

	suite.Equal(http.StatusOK, response.Code, "login failed")

	loginResp := ExtractBody[internal_http.LoginResponseBody](response)

	// create group request
	want := group.Group{
		Name:    "testGroup",
		OwnerId: p.Id,
	}
	createGroupRequestBody := internal_http.CreateGroupRequestBody{Name: want.Name}
	response = suite.POSTWithJwt("/api/v1/group", createGroupRequestBody, loginResp.SignedToken)
	suite.Equal(http.StatusCreated, response.Code)

	got := ExtractBody[group.Group](response)

	// Id is set by api
	want.Id = got.Id

	suite.Equal(want.Name, got.Name)
	suite.Equal(want.OwnerId, got.OwnerId)
	suite.NotEmpty(got.InvitationCode)
}

func (suite *GroupHandlerTestSuite) TestJoinGroupSuccess() {
	groupOwner, err := suite.personService.CreatePerson(context.Background(), "testPerson", "mail@email.com", "password123")
	suite.Require().NoError(err)
	g, err := suite.groupService.CreateGroup(context.Background(), "testGroup", groupOwner.Id)
	suite.Require().NoError(err)

	p, err := suite.personService.CreatePerson(context.Background(), "testPerson num2", "mail2@email.com", "password123")
	suite.Require().NoError(err)

	rb := internal_http.LoginRequestBody{
		Email:    p.Email,
		Password: "password123",
	}
	response := suite.POST("/api/v1/person/login", rb)
	suite.Equal(http.StatusOK, response.Code)
	loginResponseBody := ExtractBody[internal_http.LoginResponseBody](response)
	suite.NotEmpty(loginResponseBody.SignedToken)

	// perform request to join group g as user p
	joinGroupResponse := suite.POSTWithJwt(
		fmt.Sprintf("/api/v1/group/%d/join?invitationCode=%s", g.Id, g.InvitationCode),
		nil,
		loginResponseBody.SignedToken,
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

	p, err := suite.personService.CreatePerson(context.Background(), "testPerson num2", "mail2@email.com", "password123")
	suite.Require().NoError(err)

	rb := internal_http.LoginRequestBody{
		Email:    p.Email,
		Password: "password123",
	}
	response := suite.POST("/api/v1/person/login", rb)
	suite.Equal(http.StatusOK, response.Code)
	loginResponseBody := ExtractBody[internal_http.LoginResponseBody](response)
	suite.NotEmpty(loginResponseBody.SignedToken)

	// perform request to join group g as user p
	joinGroupResponse := suite.POSTWithJwt(
		fmt.Sprintf("/api/v1/group/%d/join?invitationCode=%s", g.Id, "INVALID_INVITATION_CODE"),
		nil,
		loginResponseBody.SignedToken,
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

	p, err := suite.personService.CreatePerson(context.Background(), "testPerson num2", "mail2@email.com", "password123")
	suite.Require().NoError(err)

	rb := internal_http.LoginRequestBody{
		Email:    p.Email,
		Password: "password123",
	}
	response := suite.POST("/api/v1/person/login", rb)
	suite.Equal(http.StatusOK, response.Code)
	loginResponseBody := ExtractBody[internal_http.LoginResponseBody](response)
	suite.NotEmpty(loginResponseBody.SignedToken)

	inexistentGroupId := 999
	// perform request to join group g as user p
	joinGroupResponse := suite.POSTWithJwt(
		fmt.Sprintf("/api/v1/group/%d/join?invitationCode=%s", inexistentGroupId, g.InvitationCode),
		nil,
		loginResponseBody.SignedToken,
	)
	suite.Equal(http.StatusBadRequest, joinGroupResponse.Code)

}
