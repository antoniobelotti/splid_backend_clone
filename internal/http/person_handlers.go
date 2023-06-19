package http

import (
	"github.com/antoniobelotti/splid_backend_clone/internal/person"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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

	p, err := h.service.GetPersonById(ctx, personId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "person not found"})
		return
	}

	ctx.JSON(http.StatusOK, p)
	return
}

type LoginRequestBody struct {
	Email    string `json:"email" binding:"email,required"`
	Password string `json:"password" binding:"min=8,required"`
}

type LoginResponseBody struct {
	SignedToken string `json:"signed-token"`
}

func (h *PersonHandlers) handleLogin(ctx *gin.Context) {
	requestBody := LoginRequestBody{}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "malformed request body"})
		return
	}

	p, err := h.service.GetPersonByEmail(ctx, requestBody.Email)
	if err != nil {
		// TODO: better error handling
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "internal server error"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(p.Password), []byte(requestBody.Password))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// crate signedJwtToken
	var (
		signingKey     []byte
		token          *jwt.Token
		signedJwtToken string
	)

	expirationTime := time.Now().Add(1440 * time.Minute)
	claims := jwt.MapClaims{
		"UserID":    p.Id,
		"ExpiresAt": expirationTime.Unix(),
	}

	strKey := os.Getenv("JWT_SIGNING_KEY")
	if strKey == "" {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "unable to login: internal server error"})
		return
	}

	signingKey = []byte(strKey)
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedJwtToken, err = token.SignedString(signingKey)

	ctx.JSON(http.StatusOK, LoginResponseBody{
		SignedToken: signedJwtToken,
	})
}
