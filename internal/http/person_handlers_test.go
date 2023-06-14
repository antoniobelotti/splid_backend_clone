package http

import (
	"bytes"
	"encoding/json"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/antoniobelotti/splid_backend_clone/internal/postgresdb"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreatePersonChecksValidation(t *testing.T) {
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

	db, _ := postgresdb.NewDatabase()

	ps := person.NewService(db)
	gs := group.NewService(db)

	server := NewRESTServer(ps, gs)

	for _, testCase := range table {
		jsonBody, _ := json.Marshal(testCase.requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/person", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		server.engine.ServeHTTP(w, req)

		assert.Equal(t, testCase.respHttpStatus, w.Code)
	}

}
