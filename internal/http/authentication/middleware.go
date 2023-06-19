package authentication

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// AuthenticateMiddleware is a middleware that fetches user details from token
func AuthenticateMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(c.Request.Header["Authorization"]) == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "no Authorization header provided"})
			return
		}

		tokenParts := strings.Split(c.Request.Header["Authorization"][0], " ")
		if len(tokenParts) != 2 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "invalid token format"})
			return
		}

		if !strings.EqualFold(tokenParts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "invalid token type"})
			return
		}

		jwtTokenString := tokenParts[1]

		if len(jwtTokenString) == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Please login to your account"})
			return
		}

		personId, err := GetPersonIdFromValidJwtToken(jwtTokenString)
		if err != nil {
			switch err {
			case ErrInvalidToken:
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
			case ErrSigningKeyNotSet:
				fmt.Println("JWT SIGNING KEY NOT SET - unable to authenticate")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			default:
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			}
			return
		}
		c.Set("PersonId", personId)
		c.Next()
	}
}
