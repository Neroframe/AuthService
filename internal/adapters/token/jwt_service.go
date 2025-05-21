package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Neroframe/AuthService/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type service struct {
	secretKey string
	accessTTL time.Duration
}

func NewJWTService(secret string, ttl time.Duration) domain.JWTService {
	return &service{secretKey: secret, accessTTL: ttl}
}

// TODO: use domain errors and map in usecase 
func (s *service) Validate(ctx context.Context, tokenStr string) (*domain.TokenPayload, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return []byte(s.secretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims := token.Claims.(jwt.MapClaims)
	if err := validatePayload(claims); err != nil {
		return nil, err
	}

	return &domain.TokenPayload{
		UserID:    claims["sub"].(string),
		Role:      domain.Role(claims["role"].(string)),
		ExpiresAt: int64(claims["exp"].(float64)),
	}, nil
}

func (s *service) Generate(userID string, role domain.Role) (string, time.Time, int64, error) {
	exp := time.Now().Add(s.accessTTL).Unix()
	iat := time.Now().UTC()

	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  exp,
		"iat":  iat,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", time.Time{}, 0, err
	}
	
	return signed, iat, exp, nil
}

func validatePayload(claims jwt.MapClaims) error {
	_, ok := claims["sub"].(string)
	if !ok {
		return errors.New("missing sub claim")
	}

	_, ok = claims["role"].(string)
	if !ok {
		return errors.New("missing role claim")
	}

	_, ok = claims["exp"].(float64)
	if !ok {
		return errors.New("missing exp claim")
	}

	return nil
}
