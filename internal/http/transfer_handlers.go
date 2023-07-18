package http

import (
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TransferHandlers struct {
	service group.Service
}

func NewTransferHandlers(gs group.Service) TransferHandlers {
	return TransferHandlers{service: gs}
}

type CreateTransferRequestBody struct {
	AmountInCents int `json:"amount-in-cents"`
	GroupId       int `json:"group-id"`
	ReceiverId    int `json:"receiver-id"`
}

func (h *TransferHandlers) handleCreateTransfer(ctx *gin.Context) {
	requestBody := CreateTransferRequestBody{}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "malformed request body"})
		return
	}

	senderId := ctx.GetInt("PersonId")

	e, err := h.service.CreateTransfer(ctx, requestBody.AmountInCents, requestBody.GroupId, senderId, requestBody.ReceiverId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "malformed request body"})
		return
	}

	ctx.JSON(http.StatusCreated, e)
	return
}
