//go:build integration
// +build integration

package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	internal_http "github.com/antoniobelotti/splid_backend_clone/internal/http"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/antoniobelotti/splid_backend_clone/internal/postgresdb"
	_ "github.com/lib/pq"
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

type PersonHandlerTestSuite struct {
	suite.Suite
	psqlContainer *postgrestc.PostgresContainer
	server        internal_http.RESTServer
	personService person.Service
	groupService  group.Service
}

func TestPersonHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PersonHandlerTestSuite))
}

func (suite *PersonHandlerTestSuite) TearDownTest() {
	ctx := context.Background()

	suite.Require().NoError(suite.psqlContainer.Terminate(ctx))
}

func (suite *PersonHandlerTestSuite) SetupTest() {
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

func (suite *PersonHandlerTestSuite) TestCreatePersonChecksValidation() {
	table := []struct {
		requestBody    internal_http.CreatePersonRequestBody
		respHttpStatus int
		respBody       string
	}{
		{
			requestBody: internal_http.CreatePersonRequestBody{
				Name:            "",
				Email:           "cds@mail.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: internal_http.CreatePersonRequestBody{
				Name:            "name",
				Email:           "@mail.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: internal_http.CreatePersonRequestBody{
				Name:            "name",
				Email:           "cdsmail.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: internal_http.CreatePersonRequestBody{
				Name:            "name",
				Email:           "cds@mail.com",
				Password:        "pass",
				ConfirmPassword: "pass",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: internal_http.CreatePersonRequestBody{
				Name:            "name",
				Email:           "cds@mail.com",
				Password:        "password123",
				ConfirmPassword: "password13",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: internal_http.CreatePersonRequestBody{
				Name:            "test",
				Email:           "cds@mail.com",
				Password:        "",
				ConfirmPassword: "",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: internal_http.CreatePersonRequestBody{
				Name:            "name",
				Email:           "",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			respHttpStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range table {
		jsonBody, _ := json.Marshal(testCase.requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/person", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.server.ServeHTTP(w, req)

		suite.Equal(testCase.respHttpStatus, w.Code)
	}
}

func (suite *PersonHandlerTestSuite) TestCreatePerson() {
	table := []struct {
		requestBody    internal_http.CreatePersonRequestBody
		respHttpStatus int
		respBody       person.Person
	}{
		{
			requestBody: internal_http.CreatePersonRequestBody{
				Name:            "name",
				Email:           "cds@mail.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			respHttpStatus: http.StatusCreated,
			respBody: person.Person{
				Name:  "name",
				Email: "cds@mail.com",
			},
		},
		{
			requestBody: internal_http.CreatePersonRequestBody{
				Name:            "sdfhaskdjgbasdfg",
				Email:           "uniquejhbkjhabfds@mail.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			respHttpStatus: http.StatusCreated,
			respBody: person.Person{
				Name:  "sdfhaskdjgbasdfg",
				Email: "uniquejhbkjhabfds@mail.com",
			},
		},
	}

	for _, testCase := range table {
		jsonBody, _ := json.Marshal(testCase.requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/person", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		suite.server.ServeHTTP(w, req)

		suite.Equal(testCase.respHttpStatus, w.Code)

		var got person.Person
		err := json.Unmarshal(w.Body.Bytes(), &got)
		suite.Require().NoError(err)

		// Id is set by api
		testCase.respBody.Id = got.Id

		suite.Equal(testCase.respBody, got)
	}

}

func (suite *PersonHandlerTestSuite) TestGetPerson() {
	// make sure there's a person to retrieve
	p, err := suite.personService.CreatePerson(
		context.Background(),
		"person_to_retrieve",
		"email@mail.com",
		"password123",
	)
	suite.Require().NoError(err)

	// perform login
	j, err := json.Marshal(internal_http.LoginRequestBody{
		Email:    p.Email,
		Password: "password123",
	})
	suite.Require().NoError(err)

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/person/login", bytes.NewBuffer(j))
	loginReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.server.ServeHTTP(w, loginReq)
	var lr internal_http.LoginResponseBody
	err = json.Unmarshal(w.Body.Bytes(), &lr)
	suite.Require().NoError(err)
	suite.NotEmpty(lr.SignedToken)

	// try to retrieve person
	req := httptest.NewRequest(http.MethodGet, "/api/v1/person", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", lr.SignedToken))

	w = httptest.NewRecorder()
	suite.server.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	want := person.Person{
		Id:       p.Id,
		Name:     p.Name,
		Password: "",
		Email:    p.Email,
	}

	var got person.Person
	err = json.Unmarshal(w.Body.Bytes(), &got)
	suite.Require().NoError(err)
	suite.Equal(want, got)
}

func (suite *PersonHandlerTestSuite) TestPersonLoginSuccess() {
	p, err := suite.personService.CreatePerson(
		context.Background(),
		"person_to_retrieve",
		"email@mail.com",
		"password123",
	)
	suite.Require().NoError(err)

	rb := internal_http.LoginRequestBody{
		Email:    p.Email,
		Password: "password123",
	}
	jsonBody, err := json.Marshal(rb)
	suite.Require().NoError(err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/person/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	suite.server.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var got internal_http.LoginResponseBody
	err = json.Unmarshal(w.Body.Bytes(), &got)
	suite.Require().NoError(err)
	suite.NotEmpty(got.SignedToken)
}

func (suite *PersonHandlerTestSuite) TestPersonLoginFail() {
	p, err := suite.personService.CreatePerson(
		context.Background(),
		"person_to_retrieve",
		"email@mail.com",
		"password123",
	)
	suite.Require().NoError(err)

	rb := internal_http.LoginRequestBody{
		Email:    p.Email,
		Password: "WrongPassword",
	}
	jsonBody, err := json.Marshal(rb)
	suite.Require().NoError(err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/person/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	suite.server.ServeHTTP(w, req)

	suite.Equal(http.StatusUnauthorized, w.Code)
}
