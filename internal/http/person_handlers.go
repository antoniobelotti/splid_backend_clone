package http

import (
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/gin-gonic/gin"
	"net/http"
)

type PersonHandlers struct {
	service person.Service
}

func NewPersonHandlers(service person.Service) PersonHandlers {
	return PersonHandlers{
		service: service,
	}
}

type CreatePersonRequestBody struct {
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email" binding:"email,required"`
	Password        string `json:"password" binding:"min=8,required"`
	ConfirmPassword string `json:"confirm-password" binding:"min=8,required,eqfield=Password"`
}

type CreatePersonResponseBody struct {
	PersonName string `json:"name"`
}

func (h *PersonHandlers) handleCreatePerson(ctx *gin.Context) {
	requestBody := CreatePersonRequestBody{}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "malformed request body"})
		return
	}

	p, err := h.service.CreatePerson(ctx, requestBody.Name, requestBody.Email, requestBody.Password)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "cannot create person"})
		return
	}

	ctx.JSON(http.StatusCreated, CreatePersonResponseBody{
		PersonName: p.Name,
	})
	return
}
