package domain

type User struct {
	ID       string
	Email    string
	Password string
	Role     string // "student", "teacher", "admin"
}

type UserRepository interface {
	Create(user *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id string) (*User, error)
	Update(user *User) (*User, error)
	Delete(user *User) error
}

