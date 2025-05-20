package bcrypt

import (
	"context"

	"github.com/Neroframe/AuthService/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type Hasher struct{}

func NewHasher() domain.PasswordHasher {
	return &Hasher{}
}

func (h *Hasher) Hash(ctx context.Context, plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (h *Hasher) Verify(ctx context.Context, hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}
