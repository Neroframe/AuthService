package domain

import (
	"context"
	"time"
)

type UserRegisteredEvent struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"` // UTC
}

type UserLoggedInEvent struct {
	UserID     string
	IP         string
	UserAgent  string
	Created_at time.Time
}

type UserDeletedEvent struct {
	UserID     string
	Created_at time.Time
}

type PasswordChangedEvent struct {
	UserID     string
	Created_at time.Time
}

type UserEventPublisher interface {
	PublishUserRegistered(ctx context.Context, e *UserRegisteredEvent) error
	// PublishUserLoggedIn(ctx context.Context, e *UserLoggedInEvent) error

	// PublishPasswordChanged(ctx context.Context, e *PasswordChangedEvent) error
	// PublishUserDeleted(ctx context.Context, e *UserDeletedEvent) error
}
