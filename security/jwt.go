package security

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID string, userRoles []string, secretKey string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"roles": userRoles,
		"exp":   time.Now().Add(ttl).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}