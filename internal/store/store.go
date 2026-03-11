package store

import (
	"context"
	"errors"

	"github.com/stpotter16/coin/internal/types"
)

var ErrUserNotFound = errors.New("user not found")

type Store interface {
	GetUserByUsername(ctx context.Context, username string) (types.User, error)
	CreateUser(ctx context.Context, username, passwordHash string, isAdmin bool) error
}
