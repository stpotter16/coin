package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/stpotter16/coin/internal/store"
	"github.com/stpotter16/coin/internal/types"
)

func (s Store) GetUserByUsername(ctx context.Context, username string) (types.User, error) {
	row := s.db.QueryRow(ctx,
		`SELECT id, username, password, is_admin, created_time, last_modified_time FROM user WHERE username = ?`,
		username,
	)

	var u types.User
	var isAdmin int
	var createdTime, lastModifiedTime string

	err := row.Scan(&u.ID, &u.Username, &u.Password, &isAdmin, &createdTime, &lastModifiedTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return types.User{}, store.ErrUserNotFound
		}
		return types.User{}, err
	}

	u.IsAdmin = isAdmin != 0

	u.CreatedTime, err = parseTime(createdTime)
	if err != nil {
		return types.User{}, err
	}

	u.LastModifiedTime, err = parseTime(lastModifiedTime)
	if err != nil {
		return types.User{}, err
	}

	return u, nil
}
