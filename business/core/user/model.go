package user

import (
	"net/mail"
	"time"

	"github.com/google/uuid"
)

// User represents an individual user.
type User struct {
	ID           uuid.UUID
	Name         string
	Email        mail.Address
	Roles        []Role
	PasswordHash []byte
	Department   string
	Enabled      bool
	DateCreated  time.Time
	DateUpdated  time.Time
}

// CreateUser contains information needed to create a new user.
type CreateUser struct {
	Name            string
	Email           mail.Address
	Roles           []Role
	Department      string
	Password        string
	PasswordConfirm string
}

// UpdateUser contains information needed to update a user.
type UpdateUser struct {
	Name            *string
	Email           *mail.Address
	Roles           []Role
	Department      *string
	Password        *string
	PasswordConfirm *string
	Enabled         *bool
}
