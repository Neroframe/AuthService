package domain

type TokenManager interface {
	Generate(userID, role string) (string, error)
	Validate(token string) (*TokenPayload, error)
}

type TokenPayload struct {
	UserID    string
	Role      string
	ExpiresAt int64
}
