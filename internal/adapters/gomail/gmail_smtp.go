package gomail

import (
	"github.com/Neroframe/AuthService/internal/domain"
	gomailpkg "github.com/Neroframe/AuthService/pkg/gomail"
)

func NewGomailService(sender *gomailpkg.Sender) domain.EmailSender {
	return sender
}
