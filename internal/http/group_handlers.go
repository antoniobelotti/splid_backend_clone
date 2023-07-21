package http

import (
	"errors"
	"github.com/antoniobelotti/splid_backend_clone/internal/group"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
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

func (h *GroupHandlers) handleJoinGroup(ctx *gin.Context) {
	personId := ctx.GetInt("PersonId")

	groupId, err := strconv.Atoi(ctx.Param("groupId"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "malformed group id"})
		return
	}
	requestInvitationCode := ctx.Query("invitationCode")

	g, err := h.service.GetGroupById(ctx, groupId)
	if err != nil {
		if errors.Is(err, group.ErrGroupNotFound) {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "the group you are trying to join does not exist"})
			return
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// request has wrong invitation code
	if g.InvitationCode != requestInvitationCode {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err = h.service.AddPersonToGroup(ctx, g, personId)
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "successfully joined group"})
	return
}

func (h *GroupHandlers) handleGetBalance(ctx *gin.Context) {
	groupId, err := strconv.Atoi(ctx.Param("groupId"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "malformed group id"})
		return
	}
	balance, err := h.service.GetGroupBalance(ctx, groupId)
	if err != nil {
		//todo
	}
	ctx.JSON(http.StatusOK, balance)
}
