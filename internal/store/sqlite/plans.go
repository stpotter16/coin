package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/stpotter16/coin/internal/store"
	"github.com/stpotter16/coin/internal/types"
)

func (s Store) GetOrCreatePlan(ctx context.Context, year, month, userID int) (types.Plan, error) {
	plan, found, err := s.GetPlanByMonth(ctx, year, month)
	if err != nil {
		return types.Plan{}, err
	}
	if found {
		return plan, nil
	}

	now := formatTime(time.Now().UTC())
	err = s.db.WithTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx,
			`INSERT INTO plans (year, month, locked, created_by, created_time, last_modified_time)
			 VALUES (?, ?, 0, ?, ?, ?)`,
			year, month, userID, now, now)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		plan = types.Plan{ID: int(id), Year: year, Month: month}

		// Find the most recent prior plan to copy items from.
		var priorID int
		err = tx.QueryRowContext(ctx,
			`SELECT id FROM plans
			 WHERE (year < ? OR (year = ? AND month < ?))
			 ORDER BY year DESC, month DESC
			 LIMIT 1`,
			year, year, month).Scan(&priorID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil // no prior plan — start blank
		}
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx,
			`INSERT INTO plan_items
			     (plan_id, name, type, expected_amount, created_time, last_modified_time)
			 SELECT ?, name, type, expected_amount, ?, ?
			 FROM plan_items
			 WHERE plan_id = ?`,
			plan.ID, now, now, priorID)
		return err
	})
	if err != nil {
		return types.Plan{}, err
	}
	return plan, nil
}

func (s Store) GetPlanByMonth(ctx context.Context, year, month int) (types.Plan, bool, error) {
	var plan types.Plan
	var createdTime, lastModifiedTime string
	var locked int
	err := s.db.QueryRow(ctx,
		`SELECT id, year, month, locked, created_time, last_modified_time
		 FROM plans WHERE year = ? AND month = ?`,
		year, month,
	).Scan(&plan.ID, &plan.Year, &plan.Month, &locked, &createdTime, &lastModifiedTime)
	if errors.Is(err, sql.ErrNoRows) {
		return types.Plan{}, false, nil
	}
	if err != nil {
		return types.Plan{}, false, err
	}
	plan.Locked = locked == 1
	plan.CreatedTime, err = parseTime(createdTime)
	if err != nil {
		return types.Plan{}, false, err
	}
	plan.LastModifiedTime, err = parseTime(lastModifiedTime)
	if err != nil {
		return types.Plan{}, false, err
	}
	return plan, true, nil
}

func (s Store) LockPlan(ctx context.Context, id int) error {
	now := formatTime(time.Now().UTC())
	_, err := s.db.Exec(ctx,
		`UPDATE plans SET locked = 1, last_modified_time = ? WHERE id = ?`,
		now, id,
	)
	return err
}

func (s Store) GetPlanItems(ctx context.Context, planID int) ([]types.PlanItem, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, plan_id, name, type, expected_amount, created_time, last_modified_time
		 FROM plan_items WHERE plan_id = ?
		 ORDER BY type, id`,
		planID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []types.PlanItem
	for rows.Next() {
		var item types.PlanItem
		var createdTime, lastModifiedTime string
		if err := rows.Scan(
			&item.ID, &item.PlanID, &item.Name, &item.Type,
			&item.ExpectedAmount, &createdTime, &lastModifiedTime,
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

func (s Store) GetPlanItemSummaries(ctx context.Context, planID int) ([]types.PlanItemSummary, error) {
	rows, err := s.db.Query(ctx,
		`SELECT pi.id, pi.plan_id, pi.name, pi.type, pi.expected_amount,
		        COALESCE(SUM(t.amount), 0),
		        pi.created_time, pi.last_modified_time
		 FROM plan_items pi
		 LEFT JOIN transactions t ON t.plan_item_id = pi.id AND t.excluded = 0
		 WHERE pi.plan_id = ?
		 GROUP BY pi.id
		 ORDER BY pi.type, pi.id`,
		planID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []types.PlanItemSummary
	for rows.Next() {
		var s types.PlanItemSummary
		var createdTime, lastModifiedTime string
		if err := rows.Scan(
			&s.ID, &s.PlanID, &s.Name, &s.Type,
			&s.ExpectedAmount, &s.ActualAmount,
			&createdTime, &lastModifiedTime,
		); err != nil {
			return nil, err
		}
		s.CreatedTime, err = parseTime(createdTime)
		if err != nil {
			return nil, err
		}
		s.LastModifiedTime, err = parseTime(lastModifiedTime)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, rows.Err()
}

func (s Store) CreatePlanItem(ctx context.Context, item types.PlanItem) (int, error) {
	// Refuse to add items to a locked plan.
	if locked, err := s.isPlanLocked(ctx, item.PlanID); err != nil {
		return 0, err
	} else if locked {
		return 0, store.ErrPlanLocked
	}

	now := formatTime(time.Now().UTC())
	result, err := s.db.Exec(ctx,
		`INSERT INTO plan_items (plan_id, name, type, expected_amount, created_time, last_modified_time)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		item.PlanID, item.Name, item.Type, item.ExpectedAmount, now, now,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (s Store) UpdatePlanItem(ctx context.Context, item types.PlanItem) error {
	if locked, err := s.isPlanLockedByItemID(ctx, item.ID); err != nil {
		return err
	} else if locked {
		return store.ErrPlanLocked
	}

	now := formatTime(time.Now().UTC())
	result, err := s.db.Exec(ctx,
		`UPDATE plan_items SET name = ?, expected_amount = ?, last_modified_time = ?
		 WHERE id = ?`,
		item.Name, item.ExpectedAmount, now, item.ID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return store.ErrPlanItemNotFound
	}
	return nil
}

func (s Store) DeletePlanItem(ctx context.Context, id int) error {
	if locked, err := s.isPlanLockedByItemID(ctx, id); err != nil {
		return err
	} else if locked {
		return store.ErrPlanLocked
	}

	// Unassign any transactions referencing this item before deleting.
	now := formatTime(time.Now().UTC())
	if _, err := s.db.Exec(ctx,
		`UPDATE transactions SET plan_item_id = NULL, last_modified_time = ?
		 WHERE plan_item_id = ?`, now, id,
	); err != nil {
		return err
	}

	result, err := s.db.Exec(ctx, `DELETE FROM plan_items WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return store.ErrPlanItemNotFound
	}
	return nil
}

func (s Store) isPlanLocked(ctx context.Context, planID int) (bool, error) {
	var locked int
	err := s.db.QueryRow(ctx, `SELECT locked FROM plans WHERE id = ?`, planID).Scan(&locked)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return locked == 1, err
}

func (s Store) isPlanLockedByItemID(ctx context.Context, itemID int) (bool, error) {
	var locked int
	err := s.db.QueryRow(ctx,
		`SELECT p.locked FROM plan_items pi
		 JOIN plans p ON pi.plan_id = p.id
		 WHERE pi.id = ?`, itemID,
	).Scan(&locked)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return locked == 1, err
}
