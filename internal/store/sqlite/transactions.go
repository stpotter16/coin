package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/stpotter16/coin/internal/parse"
	"github.com/stpotter16/coin/internal/store"
	"github.com/stpotter16/coin/internal/types"
)

func (s Store) GetFlexibleSpending(ctx context.Context, year, month int) (float64, error) {
	prefix := fmt.Sprintf("%04d-%02d-", year, month)
	var total float64
	err := s.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0)
		 FROM transactions
		 WHERE plan_item_id IS NULL
		   AND amount > 0
		   AND transaction_date LIKE ?`,
		prefix+"%",
	).Scan(&total)
	return total, err
}

func (s Store) UpdateTransactionPlanItem(ctx context.Context, transactionID int, planItemID *int) error {
	now := formatTime(time.Now().UTC())
	_, err := s.db.Exec(ctx,
		`UPDATE transactions SET plan_item_id = ?, last_modified_time = ? WHERE id = ?`,
		planItemID, now, transactionID,
	)
	return err
}

func (s Store) CreateTransaction(ctx context.Context, req types.TransactionRequest, userID int) (int, error) {
	now := formatTime(time.Now().UTC())
	result, err := s.db.Exec(ctx,
		`INSERT INTO transactions
			(account_id, amount, transaction_date, description, merchant_name,
			 pending, created_by, created_time, last_modified_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.AccountID,
		req.Amount,
		req.Date,
		req.Description,
		req.MerchantName,
		req.Pending,
		userID,
		now,
		now,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (s Store) UpdateTransaction(ctx context.Context, id int, req types.TransactionRequest) error {
	now := formatTime(time.Now().UTC())
	_, err := s.db.Exec(ctx,
		`UPDATE transactions
		SET account_id = ?, amount = ?, transaction_date = ?, description = ?,
		    merchant_name = ?, pending = ?, last_modified_time = ?
		WHERE id = ?`,
		req.AccountID,
		req.Amount,
		req.Date,
		req.Description,
		req.MerchantName,
		req.Pending,
		now,
		id,
	)
	return err
}

func (s Store) DeleteTransaction(ctx context.Context, id int) error {
	_, err := s.db.Exec(ctx, `DELETE FROM transactions WHERE id = ?`, id)
	return err
}

func (s Store) GetTransactionByID(ctx context.Context, id int) (types.Transaction, error) {
	row := s.db.QueryRow(ctx,
		`SELECT t.id, t.account_id, COALESCE(a.name, ''),
		        t.amount, t.transaction_date, t.description, t.merchant_name,
		        t.pending, t.plan_item_id, pi.name,
		        t.created_time, t.last_modified_time
		FROM transactions t
		LEFT JOIN account a ON t.account_id = a.id
		LEFT JOIN plan_items pi ON t.plan_item_id = pi.id
		WHERE t.id = ?`,
		id,
	)

	var dto types.TransactionDTO
	var createdTime, lastModifiedTime string
	err := row.Scan(
		&dto.ID,
		&dto.AccountID,
		&dto.AccountName,
		&dto.Amount,
		&dto.TransactionDate,
		&dto.Description,
		&dto.MerchantName,
		&dto.Pending,
		&dto.PlanItemID,
		&dto.PlanItemName,
		&createdTime,
		&lastModifiedTime,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return types.Transaction{}, store.ErrTransactionNotFound
	}
	if err != nil {
		return types.Transaction{}, err
	}

	dto.CreatedTime, err = parseTime(createdTime)
	if err != nil {
		return types.Transaction{}, err
	}
	dto.LastModifiedTime, err = parseTime(lastModifiedTime)
	if err != nil {
		return types.Transaction{}, err
	}

	return parse.ParseTransactionDTO(dto)
}

func (s Store) GetTransactions(ctx context.Context, filter types.TransactionFilter) (types.TransactionPage, error) {
	prefix := fmt.Sprintf("%04d-%02d-", filter.Year, filter.Month)

	page := filter.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * types.TransactionPageSize

	query := `
		SELECT t.id, t.account_id, COALESCE(a.name, ''),
		       t.amount, t.transaction_date, t.description, t.merchant_name,
		       t.pending, t.plan_item_id, pi.name,
		       t.created_time, t.last_modified_time
		FROM transactions t
		LEFT JOIN account a ON t.account_id = a.id
		LEFT JOIN plan_items pi ON t.plan_item_id = pi.id
		WHERE t.transaction_date LIKE ?`
	args := []any{prefix + "%"}

	if filter.AccountID != nil {
		query += " AND t.account_id = ?"
		args = append(args, *filter.AccountID)
	}

	query += " ORDER BY t.transaction_date DESC, t.id DESC LIMIT ? OFFSET ?"
	args = append(args, types.TransactionPageSize+1, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return types.TransactionPage{}, err
	}
	defer rows.Close()

	var txs []types.Transaction
	for rows.Next() {
		var dto types.TransactionDTO
		var createdTime, lastModifiedTime string
		if err := rows.Scan(
			&dto.ID,
			&dto.AccountID,
			&dto.AccountName,
			&dto.Amount,
			&dto.TransactionDate,
			&dto.Description,
			&dto.MerchantName,
			&dto.Pending,
			&dto.PlanItemID,
			&dto.PlanItemName,
			&createdTime,
			&lastModifiedTime,
		); err != nil {
			return types.TransactionPage{}, err
		}

		dto.CreatedTime, err = parseTime(createdTime)
		if err != nil {
			return types.TransactionPage{}, err
		}
		dto.LastModifiedTime, err = parseTime(lastModifiedTime)
		if err != nil {
			return types.TransactionPage{}, err
		}

		transaction, err := parse.ParseTransactionDTO(dto)
		if err != nil {
			return types.TransactionPage{}, err
		}

		txs = append(txs, transaction)
	}
	if err := rows.Err(); err != nil {
		return types.TransactionPage{}, err
	}

	hasMore := len(txs) > types.TransactionPageSize
	if hasMore {
		txs = txs[:types.TransactionPageSize]
	}
	return types.TransactionPage{Transactions: txs, HasMore: hasMore}, nil
}
