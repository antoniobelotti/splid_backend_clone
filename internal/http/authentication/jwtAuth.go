package authentication

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

type CustomJwtClaims struct {
	PersonId int `json:"person-id"`
	jwt.RegisteredClaims
}

func GetJwtToken(personId int) (string, error) {
	expirationTime := time.Now().Add(1440 * time.Minute)
	claims := CustomJwtClaims{
		PersonId: personId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	strKey := os.Getenv("JWT_SIGNING_KEY")
	if strKey == "" {
		return "", errors.New("JWT_SIGNING_KEY not set")
	}

	signingKey := []byte(strKey)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey)
}

var (
	ErrSigningKeyNotSet = errors.New("jwt sign key not set")
	ErrInvalidToken     = errors.New("invalid token")
	ErrInternalError    = errors.New("unable to verify token")
)

func GetPersonIdFromValidJwtToken(jwtTokenString string) (int, error) {
	jwtToken, _ := jwt.ParseWithClaims(jwtTokenString, &CustomJwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		keyStr := os.Getenv("JWT_SIGNING_KEY")
		if keyStr == "" {
			return nil, ErrSigningKeyNotSet
		}
		return []byte(keyStr), nil
	})

	if claims, ok := jwtToken.Claims.(*CustomJwtClaims); ok {
		if !jwtToken.Valid {
			return 0, ErrInvalidToken
		}

		return claims.PersonId, nil
	}
	return 0, ErrInternalError
}
