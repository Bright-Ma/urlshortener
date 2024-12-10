package jwt

import (
	"fmt"
	"time"

	"github.com/aeilang/urlshortener/config"
	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	secret   []byte
	duration time.Duration
}

func NewJWT(cfg config.JWTConfig) *JWT {
	return &JWT{
		secret:   []byte(cfg.Secret),
		duration: cfg.Duration,
	}
}

type UserCliams struct {
	Email  string `json:"email"`
	UserID int    `json:"user_id"`
	jwt.RegisteredClaims
}

func (j *JWT) Generate(email string, useID int) (string, error) {
	claims := UserCliams{
		Email:  email,
		UserID: useID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWT) ParseToken(tokenString string) (*UserCliams, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserCliams{}, func(t *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserCliams); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("failed to parseToken")
}
