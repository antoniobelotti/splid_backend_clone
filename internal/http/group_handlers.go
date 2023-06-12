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

type CreateGroupResponseBody struct {
	GroupName      string `json:"group-name"`
	InvitationCode string `json:"invitation-code"`
}

func (h *GroupHandlers) handleCreateGroup(ctx *gin.Context) {
	requestBody := CreateGroupRequestBody{}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "malformed request body"})
		return
	}

	ownerId := 0 // todo: pull id of loggedId person from context
	g, err := h.service.CreateGroup(ctx, requestBody.Name, ownerId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "cannot create grup"})
		return
	}

	ctx.JSON(http.StatusCreated, CreateGroupResponseBody{
		GroupName:      g.Name,
		InvitationCode: g.InvitationCode,
	})
	return
}
