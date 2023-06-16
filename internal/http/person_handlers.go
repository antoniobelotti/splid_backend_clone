package http

import (
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
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

	ctx.JSON(http.StatusCreated, p)
	return
}

func (h *PersonHandlers) handleGetPerson(ctx *gin.Context) {
	personIdStr := ctx.Param("personId")
	personId, err := strconv.Atoi(strings.TrimSpace(personIdStr))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "malformed request url"})
		return
	}

	p, err := h.service.GetPerson(ctx, personId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "person not found"})
		return
	}

	ctx.JSON(http.StatusOK, p)
	return
}
