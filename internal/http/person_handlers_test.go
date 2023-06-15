package http

import (
	"bytes"
	"encoding/json"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/antoniobelotti/splid_backend_clone/internal/postgresdb"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type PersonHandlerTestSuite struct {
	suite.Suite
	server RESTServer
}

func TestPersonHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PersonHandlerTestSuite))
}

func (suite *PersonHandlerTestSuite) SetupSuite() {
	db, _ := postgresdb.NewDatabase()

	ps := person.NewService(db)
	gs := group.NewService(db)

	suite.server = NewRESTServer(ps, gs)
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
