package sqlite

import (
	"context"
	"time"

	"github.com/stpotter16/coin/internal/store"
	"github.com/stpotter16/coin/internal/types"
)

func (s Store) GetCategories(ctx context.Context) ([]types.Category, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, name, created_by, last_modified_by, created_time, last_modified_time
		FROM category ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []types.Category
	for rows.Next() {
		var c types.Category
		var createdTime, lastModifiedTime string

		if err := rows.Scan(
			&c.ID, &c.Name, &c.CreatedBy, &c.LastModifiedBy,
			&createdTime, &lastModifiedTime,
		); err != nil {
			return nil, err
		}

		c.CreatedTime, err = parseTime(createdTime)
		if err != nil {
			return nil, err
		}
		c.LastModifiedTime, err = parseTime(lastModifiedTime)
		if err != nil {
			return nil, err
		}

		cats = append(cats, c)
	}

	return cats, rows.Err()
}

func (s Store) CreateCategory(ctx context.Context, name string, userID int) error {
	now := formatTime(time.Now().UTC())
	_, err := s.db.Exec(ctx,
		`INSERT INTO category (name, created_by, last_modified_by, created_time, last_modified_time)
		VALUES (?, ?, ?, ?, ?)`,
		name, userID, userID, now, now,
	)
	return err
}

func (s Store) UpdateCategory(ctx context.Context, id int, name string, userID int) error {
	now := formatTime(time.Now().UTC())
	result, err := s.db.Exec(ctx,
		`UPDATE category SET name = ?, last_modified_by = ?, last_modified_time = ? WHERE id = ?`,
		name, userID, now, id,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return store.ErrCategoryNotFound
	}
	return nil
}

func (s Store) DeleteCategory(ctx context.Context, id int) error {
	if _, err := s.db.Exec(ctx,
		`UPDATE transactions SET category_id = NULL WHERE category_id = ?`, id,
	); err != nil {
		return err
	}
	result, err := s.db.Exec(ctx, `DELETE FROM category WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return store.ErrCategoryNotFound
	}
	return nil
}
