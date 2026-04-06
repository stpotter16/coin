package sqlite

import (
	"context"
	"time"

	"github.com/stpotter16/coin/internal/types"
)

func (s Store) GetAllAccounts(ctx context.Context) ([]types.Account, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, name, type, created_time, last_modified_time
		FROM account ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []types.Account
	for rows.Next() {
		var a types.AccountDTO
		var createdTime, lastModifiedTime string

		if err := rows.Scan(&a.ID, &a.Name, &a.Type, &createdTime, &lastModifiedTime); err != nil {
			return nil, err
		}

		a.CreatedTime, err = parseTime(createdTime)
		if err != nil {
			return nil, err
		}
		a.LastModifiedTime, err = parseTime(lastModifiedTime)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, types.Account{
			ID:               a.ID,
			Name:             a.Name,
			Type:             a.Type,
			CreatedTime:      a.CreatedTime,
			LastModifiedTime: a.LastModifiedTime,
		})
	}

	return accounts, rows.Err()
}

func (s Store) CreateAccount(ctx context.Context, name, accountType string) (int, error) {
	now := formatTime(time.Now().UTC())
	result, err := s.db.Exec(ctx,
		`INSERT INTO account (name, type, created_time, last_modified_time) VALUES (?, ?, ?, ?)`,
		name, accountType, now, now,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (s Store) DeleteAccount(ctx context.Context, id int) error {
	_, err := s.db.Exec(ctx, `DELETE FROM account WHERE id = ?`, id)
	return err
}
