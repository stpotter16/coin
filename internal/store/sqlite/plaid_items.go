package sqlite

import (
	"context"
	"time"

	"github.com/stpotter16/coin/internal/types"
)

func (s Store) CreatePlaidItem(ctx context.Context, item types.PlaidItem) error {
	now := formatTime(time.Now().UTC())
	_, err := s.db.Exec(ctx,
		`INSERT INTO plaid_item
			(plaid_item_id, plaid_access_token, institution_id, institution_name,
			 transaction_cursor, created_time, last_modified_time)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		item.PlaidItemID,
		item.PlaidAccessToken,
		item.InstitutionID,
		item.InstitutionName,
		item.TransactionCursor,
		now,
		now,
	)
	return err
}

func (s Store) GetPlaidItems(ctx context.Context) ([]types.PlaidItem, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, plaid_item_id, plaid_access_token, institution_id,
			institution_name, transaction_cursor, created_time, last_modified_time
		FROM plaid_item`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []types.PlaidItem
	for rows.Next() {
		var item types.PlaidItem
		var createdTime, lastModifiedTime string

		if err := rows.Scan(
			&item.ID,
			&item.PlaidItemID,
			&item.PlaidAccessToken,
			&item.InstitutionID,
			&item.InstitutionName,
			&item.TransactionCursor,
			&createdTime,
			&lastModifiedTime,
		); err != nil {
			return nil, err
		}

		item.CreatedTime, err = parseTime(createdTime)
		if err != nil {
			return nil, err
		}
		item.LastModifiedTime, err = parseTime(lastModifiedTime)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (s Store) UpdatePlaidItemCursor(ctx context.Context, id int, cursor string) error {
	now := formatTime(time.Now().UTC())
	_, err := s.db.Exec(ctx,
		`UPDATE plaid_item SET transaction_cursor = ?, last_modified_time = ? WHERE id = ?`,
		cursor, now, id,
	)
	return err
}
