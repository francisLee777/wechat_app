package util

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Define a secret key (In production, use environment variables)
var jwtSecret = []byte("lihaoyu_francis") // 不重要，直接写在代码里面

type Claims struct {
	OpenID string `json:"openid"`
	jwt.RegisteredClaims
}

const JWTExpireTime = 24 * time.Hour

// GenerateToken generates a JWT token for a given OpenID
func GenerateToken(openID string) (string, error) {
	expirationTime := time.Now().Add(JWTExpireTime) // Token valid for 24 hours
	claims := &Claims{
		OpenID: openID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "wx_mini_program",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	return tokenString, err
}

// ParseToken parses the JWT token and returns the claims
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
