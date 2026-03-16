package sqlite

import (
	"context"

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
