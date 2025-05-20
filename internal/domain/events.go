package domain

import (
	"context"
	"time"
)

type UserRegisteredEvent struct {
	UserID   string
	Email    string
	Role     Role
	Occurred time.Time
}

type UserLoggedInEvent struct {
	UserID    string
	IP        string
	UserAgent string
	Occurred  time.Time
}

type UserDeletedEvent struct {
	UserID   string
	Occurred time.Time
}

type PasswordChangedEvent struct {
	UserID   string
	Occurred time.Time
}

type UserEventPublisher interface {
	PublishUserRegistered(ctx context.Context, e UserRegisteredEvent) error

	// PublishPasswordChanged(ctx context.Context, e PasswordChangedEvent) error
	// PublishUserDeleted(ctx context.Context, e UserDeletedEvent) error
	// PublishUserLoggedIn(ctx context.Context, e UserLoggedInEvent) error
}
