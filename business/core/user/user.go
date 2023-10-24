// Package user provides an example of a core business API. Right now these
// calls are just wrapping the data/data layer. But at some point you will
// want auditing or something that isn't specific to the data/store layer.
package user

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/farmani/service/business/data/order"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("user not found")
	ErrUniqueEmail           = errors.New("email is not unique")
	ErrAuthenticationFailure = errors.New("authentication.rego failed")
)

// =============================================================================

// Storer interface declares the behavior this package needs to persists and
// retrieve data.
type Storer interface {
	Create(ctx context.Context, usr User) error
	Update(ctx context.Context, usr User) error
	Delete(ctx context.Context, usr User) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]User, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, userID uuid.UUID) (User, error)
	QueryByIDs(ctx context.Context, userID []uuid.UUID) ([]User, error)
	QueryByEmail(ctx context.Context, email mail.Address) (User, error)
}

// =============================================================================

// Core manages the set of APIs for user access.
type Core struct {
	storer Storer
	log    *zap.SugaredLogger
}

// NewCore constructs a core for user api access.
func NewCore(log *zap.SugaredLogger, storer Storer) *Core {
	return &Core{
		storer: storer,
		log:    log,
	}
}

// Create adds a new user to the system.
func (c *Core) Create(ctx context.Context, cu CreateUser) (User, error) {
	password := Password{}
	if err := password.set(cu.Password); err != nil {
		return User{}, fmt.Errorf("generating password hash: %w", err)
	}

	now := time.Now()

	usr := User{
		ID:           uuid.New(),
		Name:         cu.Name,
		Email:        cu.Email,
		PasswordHash: password,
		Roles:        cu.Roles,
		Department:   cu.Department,
		Enabled:      true,
		DateCreated:  now,
		DateUpdated:  now,
	}

	if err := c.storer.Create(ctx, usr); err != nil {
		return User{}, fmt.Errorf("creating user: %w", err)
	}

	return usr, nil
}

// Update modifies information about a user.
func (c *Core) Update(ctx context.Context, user User, updateUser UpdateUser) (User, error) {
	if updateUser.Name != nil {
		user.Name = *updateUser.Name
	}

	if updateUser.Email != nil {
		user.Email = *updateUser.Email
	}

	if updateUser.Roles != nil {
		user.Roles = updateUser.Roles
	}

	if updateUser.Department != nil {
		user.Department = *updateUser.Department
	}

	if updateUser.Enabled != nil {
		user.Enabled = *updateUser.Enabled
	}

	if updateUser.Password != nil {
		pw := Password{}
		if err := pw.set(*updateUser.Password); err != nil {
			return User{}, fmt.Errorf("generating password hash: %w", err)
		}
		user.PasswordHash = pw
	}

	user.DateUpdated = time.Now()

	if err := c.storer.Update(ctx, user); err != nil {
		return User{}, fmt.Errorf("update: %w", err)
	}

	return user, nil
}

// Delete removes the specified user.
func (c *Core) Delete(ctx context.Context, usr User) error {
	if err := c.storer.Delete(ctx, usr); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of existing users.
// TODO: Add pagination and Search struct beside so many parameters.
func (c *Core) Query(ctx context.Context, filter QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]User, error) {
	users, err := c.storer.Query(ctx, filter, orderBy, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return users, nil
}

// Count returns the total number of users.
func (c *Core) Count(ctx context.Context, filter QueryFilter) (int, error) {
	return c.storer.Count(ctx, filter)
}

// QueryByID finds the user by the specified ID.
func (c *Core) QueryByID(ctx context.Context, userID uuid.UUID) (User, error) {
	user, err := c.storer.QueryByID(ctx, userID)
	if err != nil {
		return User{}, fmt.Errorf("query: userID[%s]: %w", userID, err)
	}

	return user, nil
}

// QueryByIDs finds the users by a specified User IDs.
func (c *Core) QueryByIDs(ctx context.Context, userIDs []uuid.UUID) ([]User, error) {
	user, err := c.storer.QueryByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("query: userIDs[%s]: %w", userIDs, err)
	}

	return user, nil
}

// QueryByEmail finds the user by a specified user email.
func (c *Core) QueryByEmail(ctx context.Context, email mail.Address) (User, error) {
	user, err := c.storer.QueryByEmail(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("query: email[%s]: %w", email, err)
	}

	return user, nil
}

// =============================================================================

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Claims User representing this user. The claims can be
// used to generate a token for future authentication.
func (c *Core) Authenticate(ctx context.Context, email mail.Address, password string) (User, error) {
	usr, err := c.QueryByEmail(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("query: email[%s]: %w", email, err)
	}

	pw := usr.PasswordHash
	res, err := pw.Matches(password)
	if err != nil || !res {
		return User{}, fmt.Errorf("comparehashpassword: %w", ErrAuthenticationFailure)
	}

	return usr, nil
}
