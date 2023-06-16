package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/antoniobelotti/splid_backend_clone/internal/postgresdb"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type PersonHandlerTestSuite struct {
	suite.Suite
	server        RESTServer
	personService person.Service
	groupService  group.Service
}

func TestPersonHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PersonHandlerTestSuite))
}

func (suite *PersonHandlerTestSuite) SetupSuite() {
	db, _ := postgresdb.NewDatabase()

	suite.personService = person.NewService(db)
	suite.groupService = group.NewService(db)

	suite.server = NewRESTServer(suite.personService, suite.groupService)
}

func (suite *PersonHandlerTestSuite) TestCreatePersonChecksValidation() {
	table := []struct {
		requestBody    CreatePersonRequestBody
		respHttpStatus int
		respBody       string
	}{
		{
			requestBody: CreatePersonRequestBody{
				Name:            "",
				Email:           "cds@mail.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: CreatePersonRequestBody{
				Name:            "name",
				Email:           "@mail.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: CreatePersonRequestBody{
				Name:            "name",
				Email:           "cdsmail.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: CreatePersonRequestBody{
				Name:            "name",
				Email:           "cds@mail.com",
				Password:        "pass",
				ConfirmPassword: "pass",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: CreatePersonRequestBody{
				Name:            "name",
				Email:           "cds@mail.com",
				Password:        "password123",
				ConfirmPassword: "password13",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: CreatePersonRequestBody{
				Name:            "test",
				Email:           "cds@mail.com",
				Password:        "",
				ConfirmPassword: "",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: CreatePersonRequestBody{
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
		requestBody    CreatePersonRequestBody
		respHttpStatus int
		respBody       person.Person
	}{
		{
			requestBody: CreatePersonRequestBody{
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
		if err != nil {
			suite.Fail(err.Error())
		}
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
	if err != nil {
		if !strings.Contains(err.Error(), "duplicate") {
			suite.Fail(err.Error())
		}
	}

	table := []struct {
		personId       int
		respHttpStatus int
		respBody       person.Person
	}{
		{
			personId:       0,
			respHttpStatus: http.StatusOK,
			respBody:       p,
		},
	}

	for _, testCase := range table {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/person/%d", testCase.personId), nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		suite.server.ServeHTTP(w, req)

		suite.Equal(testCase.respHttpStatus, w.Code)

		var got person.Person
		err := json.Unmarshal(w.Body.Bytes(), &got)
		if err != nil {
			suite.Fail(err.Error())
		}
		suite.Equal(testCase.respBody, got)
	}

}
