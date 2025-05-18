package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtManager struct {
	secret    string
	expiresIn time.Duration
}

type jwtClaims struct {
	Role string `json: role`
	jwt.RegisteredClaims // sub-userID
}


