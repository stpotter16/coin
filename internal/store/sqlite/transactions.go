package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/stpotter16/coin/internal/store"
	"github.com/stpotter16/coin/internal/types"
)

func (s Store) UpsertTransaction(ctx context.Context, tx types.Transaction) error {
	now := formatTime(time.Now().UTC())
	_, err := s.db.Exec(ctx,
		`INSERT INTO transactions
			(plaid_transaction_id, account_id, amount, transaction_date, description,
			 merchant_name, pending, payment_channel, plaid_category_primary,
			 plaid_category_detailed, category_id, last_modified_by,
			 created_time, last_modified_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(plaid_transaction_id) DO UPDATE SET
			amount                  = excluded.amount,
			transaction_date        = excluded.transaction_date,
			description             = excluded.description,
			merchant_name           = excluded.merchant_name,
			pending                 = excluded.pending,
			payment_channel         = excluded.payment_channel,
			plaid_category_primary  = excluded.plaid_category_primary,
			plaid_category_detailed = excluded.plaid_category_detailed,
			last_modified_time      = excluded.last_modified_time`,
		tx.PlaidTransactionID,
		tx.AccountID,
		tx.Amount,
		tx.TransactionDate,
		tx.Description,
		tx.MerchantName,
		tx.Pending,
		tx.PaymentChannel,
		tx.PlaidCategoryPrimary,
		tx.PlaidCategoryDetailed,
		tx.CategoryID,
		tx.LastModifiedBy,
		now,
		now,
	)
	return err
}

func (s Store) GetTransactionByID(ctx context.Context, id int) (types.Transaction, error) {
	row := s.db.QueryRow(ctx,
		`SELECT id, plaid_transaction_id, account_id, amount, transaction_date,
		        description, merchant_name, pending, payment_channel,
		        plaid_category_primary, plaid_category_detailed,
		        category_id, last_modified_by, created_time, last_modified_time
		FROM transactions WHERE id = ?`,
		id,
	)

	var tx types.Transaction
	var createdTime, lastModifiedTime string
	err := row.Scan(
		&tx.ID,
		&tx.PlaidTransactionID,
		&tx.AccountID,
		&tx.Amount,
		&tx.TransactionDate,
		&tx.Description,
		&tx.MerchantName,
		&tx.Pending,
		&tx.PaymentChannel,
		&tx.PlaidCategoryPrimary,
		&tx.PlaidCategoryDetailed,
		&tx.CategoryID,
		&tx.LastModifiedBy,
		&createdTime,
		&lastModifiedTime,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return types.Transaction{}, store.ErrTransactionNotFound
	}
	if err != nil {
		return types.Transaction{}, err
	}

	tx.CreatedTime, err = parseTime(createdTime)
	if err != nil {
		return types.Transaction{}, err
	}
	tx.LastModifiedTime, err = parseTime(lastModifiedTime)
	if err != nil {
		return types.Transaction{}, err
	}
	return tx, nil
}

func (s Store) UpdateTransactionCategory(ctx context.Context, id int, categoryID *int) error {
	now := formatTime(time.Now().UTC())
	_, err := s.db.Exec(ctx,
		`UPDATE transactions SET category_id = ?, last_modified_time = ? WHERE id = ?`,
		categoryID, now, id,
	)
	return err
}

func (s Store) DeleteTransaction(ctx context.Context, plaidTransactionID string) error {
	_, err := s.db.Exec(ctx,
		`DELETE FROM transactions WHERE plaid_transaction_id = ?`,
		plaidTransactionID,
	)
	return err
}

func (s Store) GetTransactions(ctx context.Context, filter types.TransactionFilter) ([]types.Transaction, error) {
	prefix := fmt.Sprintf("%04d-%02d-", filter.Year, filter.Month)

	query := `
		SELECT id, plaid_transaction_id, account_id, amount, transaction_date,
		       description, merchant_name, pending, payment_channel,
		       plaid_category_primary, plaid_category_detailed,
		       category_id, last_modified_by, created_time, last_modified_time
		FROM transactions
		WHERE transaction_date LIKE ?`
	args := []any{prefix + "%"}

	if filter.AccountID != nil {
		query += " AND account_id = ?"
		args = append(args, *filter.AccountID)
	}

	query += " ORDER BY transaction_date DESC, id DESC"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []types.Transaction
	for rows.Next() {
		var tx types.Transaction
		var createdTime, lastModifiedTime string
		if err := rows.Scan(
			&tx.ID,
			&tx.PlaidTransactionID,
			&tx.AccountID,
			&tx.Amount,
			&tx.TransactionDate,
			&tx.Description,
			&tx.MerchantName,
			&tx.Pending,
			&tx.PaymentChannel,
			&tx.PlaidCategoryPrimary,
			&tx.PlaidCategoryDetailed,
			&tx.CategoryID,
			&tx.LastModifiedBy,
			&createdTime,
			&lastModifiedTime,
		); err != nil {
			return nil, err
		}
		tx.CreatedTime, err = parseTime(createdTime)
		if err != nil {
			return nil, err
		}
		tx.LastModifiedTime, err = parseTime(lastModifiedTime)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, rows.Err()
}
