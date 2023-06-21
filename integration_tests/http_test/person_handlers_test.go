//go:build integration
// +build integration

package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/antoniobelotti/splid_backend_clone/integration_tests"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	internal_http "github.com/antoniobelotti/splid_backend_clone/internal/http"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	psqlcont "github.com/testcontainers/testcontainers-go/modules/postgres"
	"net/http"
	"net/http/httptest"
	"testing"
)

type PersonHandlerTestSuite struct {
	suite.Suite
	psqlContainer *psqlcont.PostgresContainer
	server        internal_http.RESTServer
	personService person.Service
	groupService  group.Service
}

func TestPersonHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PersonHandlerTestSuite))
}

func (suite *PersonHandlerTestSuite) TearDownTest() {
	_ = suite.psqlContainer.Terminate(context.Background())
}

func (suite *PersonHandlerTestSuite) SetupTest() {
	db, cont := integration_tests.GetCleanContainerizedPsqlDb()

	suite.psqlContainer = cont

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
		req := httptest.NewRequest(http.MethodPost, "/api/v1/person/signup", bytes.NewBuffer(jsonBody))
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
		req := httptest.NewRequest(http.MethodPost, "/api/v1/person/signup", bytes.NewBuffer(jsonBody))
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
