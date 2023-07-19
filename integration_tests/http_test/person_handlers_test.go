//go:build integration

package http_test

import (
	"context"
	"github.com/antoniobelotti/splid_backend_clone/integration_tests"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	internalHttp "github.com/antoniobelotti/splid_backend_clone/internal/http"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/antoniobelotti/splid_backend_clone/internal/transfer"
	"github.com/stretchr/testify/suite"
	psqlcont "github.com/testcontainers/testcontainers-go/modules/postgres"
	"net/http"
	"testing"
)

type PersonHandlerTestSuite struct {
	testSuiteHttp
	psqlContainer *psqlcont.PostgresContainer
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

	suite.server = internalHttp.NewRESTServer(suite.personService, suite.groupService, transfer.Service{})
}

func (suite *PersonHandlerTestSuite) TestCreatePersonChecksValidation() {
	table := []struct {
		requestBody    internalHttp.CreatePersonRequestBody
		respHttpStatus int
		respBody       string
	}{
		{
			requestBody: internalHttp.CreatePersonRequestBody{
				Name:            "",
				Email:           "cds@mail.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: internalHttp.CreatePersonRequestBody{
				Name:            "name",
				Email:           "@mail.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: internalHttp.CreatePersonRequestBody{
				Name:            "name",
				Email:           "cdsmail.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: internalHttp.CreatePersonRequestBody{
				Name:            "name",
				Email:           "cds@mail.com",
				Password:        "pass",
				ConfirmPassword: "pass",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: internalHttp.CreatePersonRequestBody{
				Name:            "name",
				Email:           "cds@mail.com",
				Password:        "password123",
				ConfirmPassword: "password13",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: internalHttp.CreatePersonRequestBody{
				Name:            "test",
				Email:           "cds@mail.com",
				Password:        "",
				ConfirmPassword: "",
			},
			respHttpStatus: http.StatusBadRequest,
		},
		{
			requestBody: internalHttp.CreatePersonRequestBody{
				Name:            "name",
				Email:           "",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			respHttpStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range table {
		w := suite.POST("/api/v1/person/signup", tc.requestBody)
		suite.Equal(tc.respHttpStatus, w.Code)
	}
}

func (suite *PersonHandlerTestSuite) TestCreatePerson() {
	table := []struct {
		requestBody    internalHttp.CreatePersonRequestBody
		respHttpStatus int
		respBody       person.Person
	}{
		{
			requestBody: internalHttp.CreatePersonRequestBody{
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
			requestBody: internalHttp.CreatePersonRequestBody{
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

	for _, tc := range table {
		response := suite.POST("/api/v1/person/signup", tc.requestBody)
		suite.Equal(tc.respHttpStatus, response.Code)

		got := ExtractBody[person.Person](response)

		// Id is set by api
		tc.respBody.Id = got.Id

		suite.Equal(tc.respBody, got)
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
	loginReqBody := internalHttp.LoginRequestBody{
		Email:    p.Email,
		Password: "password123",
	}
	response := suite.POST("/api/v1/person/login", loginReqBody)
	lr := ExtractBody[internalHttp.LoginResponseBody](response)
	suite.NotEmpty(lr.SignedToken)

	// try to retrieve person
	response = suite.GETWithJwt("/api/v1/person", lr.SignedToken)
	suite.Equal(http.StatusOK, response.Code)

	want := person.Person{
		Id:       p.Id,
		Name:     p.Name,
		Password: "",
		Email:    p.Email,
	}

	got := ExtractBody[person.Person](response)
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

	rb := internalHttp.LoginRequestBody{
		Email:    p.Email,
		Password: "password123",
	}
	response := suite.POST("/api/v1/person/login", rb)
	suite.Equal(http.StatusOK, response.Code)

	got := ExtractBody[internalHttp.LoginResponseBody](response)
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

	rb := internalHttp.LoginRequestBody{
		Email:    p.Email,
		Password: "WrongPassword",
	}
	response := suite.POST("/api/v1/person/login", rb)
	suite.Equal(http.StatusUnauthorized, response.Code)
}
