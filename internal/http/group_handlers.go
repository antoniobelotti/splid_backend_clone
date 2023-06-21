package http

import (
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	"github.com/gin-gonic/gin"
	"net/http"
)

type GroupHandlers struct {
	service group.Service
}

func NewGroupHandlers(service group.Service) GroupHandlers {
	return GroupHandlers{
		service: service,
	}
}

type CreateGroupRequestBody struct {
	Name string `json:"name" binding:"required"`
}

func (h *GroupHandlers) handleCreateGroup(ctx *gin.Context) {
	requestBody := CreateGroupRequestBody{}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "malformed request body"})
		return
	}

	ownerIdStr := ctx.GetInt("PersonId")
	g, err := h.service.CreateGroup(ctx, requestBody.Name, ownerIdStr)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "cannot create group"})
		return
	}

	ctx.JSON(http.StatusCreated, g)
	return
}
