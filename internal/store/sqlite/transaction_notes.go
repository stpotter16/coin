package sqlite

import (
	"context"
	"time"

	"github.com/stpotter16/coin/internal/types"
)

func (s Store) CreateTransactionNote(ctx context.Context, note types.TransactionNote) error {
	now := formatTime(time.Now().UTC())
	_, err := s.db.Exec(ctx,
		`INSERT INTO transaction_notes (transaction_id, user_id, note, created_time)
		VALUES (?, ?, ?, ?)`,
		note.TransactionID, note.UserID, note.Note, now,
	)
	return err
}

func (s Store) GetNotesByTransactionID(ctx context.Context, transactionID int) ([]types.TransactionNote, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, transaction_id, user_id, note, created_time
		FROM transaction_notes WHERE transaction_id = ? ORDER BY created_time ASC`,
		transactionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []types.TransactionNote
	for rows.Next() {
		var n types.TransactionNote
		var createdTime string

		if err := rows.Scan(
			&n.ID, &n.TransactionID, &n.UserID, &n.Note, &createdTime,
		); err != nil {
			return nil, err
		}

		n.CreatedTime, err = parseTime(createdTime)
		if err != nil {
			return nil, err
		}

		notes = append(notes, n)
	}

	return notes, rows.Err()
}
