package user

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	Plaintext *string
	Hash      []byte
}

// Set calculates the bcrypt hash of a plaintext password, and stores both the has and the
// plaintext versions in the password struct.
func (p *Password) set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.Plaintext = &plaintextPassword
	p.Hash = hash
	return nil
}

// Matches checks whether the provided plaintext password matches the hashed password stored in
// the password struct, returning true if it matches and false otherwise.
func (p *Password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}
