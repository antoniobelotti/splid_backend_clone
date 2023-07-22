//go:build integration

package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	internal_http "github.com/antoniobelotti/splid_backend_clone/internal/http"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"time"
)

type testSuiteHttp struct {
	suite.Suite
	server internal_http.RESTServer
}

func (suite *testSuiteHttp) POST(endpoint string, requestBody any) *httptest.ResponseRecorder {
	return suite.post(endpoint, requestBody, "")
}

func (suite *testSuiteHttp) POSTWithJwt(endpoint string, requestBody any, jwtToken string) *httptest.ResponseRecorder {
	return suite.post(endpoint, requestBody, jwtToken)
}

func (suite *testSuiteHttp) post(endpoint string, requestBody any, jwtToken string) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	if jwtToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	}

	responseRecorder := httptest.NewRecorder()
	suite.server.ServeHTTP(responseRecorder, req)
	return responseRecorder
}

func (suite *testSuiteHttp) GET(endpoint string) *httptest.ResponseRecorder {
	return suite.get(endpoint, "")
}

func (suite *testSuiteHttp) GETWithJwt(endpoint string, jwtToken string) *httptest.ResponseRecorder {
	return suite.get(endpoint, jwtToken)
}

func (suite *testSuiteHttp) get(endpoint string, jwtToken string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, endpoint, nil)
	req.Header.Set("Content-Type", "application/json")
	if jwtToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	}

	responseRecorder := httptest.NewRecorder()
	suite.server.ServeHTTP(responseRecorder, req)
	return responseRecorder
}

func ExtractBody[T any](rr *httptest.ResponseRecorder) T {
	var body T
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	return body
}

// GetLoggedInPerson performs an authentication flow and returns the Person struct plus valid jwt token
func (suite *testSuiteHttp) GetLoggedInPerson() (person.Person, string) {

	// try to prevent PK and unique constraints violations
	rs := time.Now().Unix()

	signupResponse := suite.POST("/api/v1/person/signup", internal_http.CreatePersonRequestBody{
		Name:            fmt.Sprintf("test person %d", rs),
		Email:           fmt.Sprintf("main%d@mail.com", rs),
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	p := ExtractBody[person.Person](signupResponse)

	loginResponse := suite.POST("/api/v1/person/login", internal_http.LoginRequestBody{
		Email:    p.Email,
		Password: "password123",
	})
	loginRespBody := ExtractBody[internal_http.LoginResponseBody](loginResponse)

	return p, loginRespBody.SignedToken
}
