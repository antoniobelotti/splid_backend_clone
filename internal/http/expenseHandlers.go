package http

import (
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ExpenseHandlers struct {
	service group.Service
}

func NewExpenseHandlers(gs group.Service) ExpenseHandlers {
	return ExpenseHandlers{service: gs}
}

type CreateExpenseRequestBody struct {
	AmountInCents int `json:"amount-in-cents"`
	GroupId       int `json:"group-id"`
}

func (h *ExpenseHandlers) handleCreateExpense(ctx *gin.Context) {
	requestBody := CreateExpenseRequestBody{}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "malformed request body"})
		return
	}

	personId := ctx.GetInt("PersonId")

	e, err := h.service.CreateExpense(ctx, requestBody.AmountInCents, personId, requestBody.GroupId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "malformed request body"})
		return
	}

	ctx.JSON(http.StatusCreated, e)
	return
}
