//go:build integration
// +build integration

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	internal_http "github.com/antoniobelotti/splid_backend_clone/internal/http"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/antoniobelotti/splid_backend_clone/internal/postgresdb"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	postgrestc "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type GroupHandlerTestSuite struct {
	suite.Suite
	psqlContainer *postgrestc.PostgresContainer
	server        internal_http.RESTServer
	personService person.Service
	groupService  group.Service
}

func TestGroupHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(GroupHandlerTestSuite))
}

func (suite *GroupHandlerTestSuite) TearDownTest() {
	ctx := context.Background()

	suite.Require().NoError(suite.psqlContainer.Terminate(ctx))
}

func (suite *GroupHandlerTestSuite) SetupTest() {
	ctx := context.Background()

	files, err := os.ReadDir("/home/anto/GolandProjects/splid_backend_clone/migrations/")
	suite.Require().NoError(err)

	/* I was unable to make golang-migrate work in this setting.
	as a workaround, pass all *.up.sql migration files to the container. They will be executed after start up,
	hopefully in the correct order
	*/
	var migrationFiles []string
	for _, file := range files {
		if strings.Contains(file.Name(), "up") {
			migrationFiles = append(migrationFiles, filepath.Join("/home/anto/GolandProjects/splid_backend_clone/migrations/", file.Name()))
		}
	}

	container, err := postgrestc.RunContainer(ctx,
		testcontainers.WithImage("postgres:12-alpine"),
		postgrestc.WithInitScripts(migrationFiles...),
		postgrestc.WithDatabase("postgres"),
		postgrestc.WithUsername("postgres"),
		postgrestc.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	suite.Require().NoError(err)

	connStr, err := container.ConnectionString(ctx, "sslmode=disable", "application_name=test")

	suite.psqlContainer = container

	db, err := postgresdb.NewDatabase(connStr)
	suite.Require().NoError(err)

	// this is where golang-migrate should apply migrations

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
