//go:build integration
// +build integration

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/antoniobelotti/splid_backend_clone/integration_tests"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	internal_http "github.com/antoniobelotti/splid_backend_clone/internal/http"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/stretchr/testify/suite"
	psqlcont "github.com/testcontainers/testcontainers-go/modules/postgres"
	"net/http"
	"net/http/httptest"
	"testing"
)

type GroupHandlerTestSuite struct {
	suite.Suite
	psqlContainer *psqlcont.PostgresContainer
	server        internal_http.RESTServer
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
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/person/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.server.ServeHTTP(w, req)
	suite.Equal(http.StatusOK, w.Code, "login failed")

	var loginResp internal_http.LoginResponseBody
	err = json.Unmarshal(w.Body.Bytes(), &loginResp)
	suite.Require().NoError(err)

	// create group request
	want := group.Group{
		Name:    "testGroup",
		OwnerId: p.Id,
	}

	createGroupRequestBody := internal_http.CreateGroupRequestBody{Name: want.Name}
	jsonBody, _ = json.Marshal(createGroupRequestBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/group", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", loginResp.SignedToken))

	w = httptest.NewRecorder()
	suite.server.ServeHTTP(w, req)
	suite.Equal(http.StatusCreated, w.Code)

	var got group.Group
	err = json.Unmarshal(w.Body.Bytes(), &got)
	suite.Require().NoError(err)

	// Id is set by api
	want.Id = got.Id

	suite.Equal(want.Name, got.Name)
	suite.Equal(want.OwnerId, got.OwnerId)
	suite.NotEmpty(got.InvitationCode)
}
