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

func (s Store) UpsertPlaidTransaction(ctx context.Context, tx types.Transaction) error {
	now := formatTime(time.Now().UTC())
	_, err := s.db.Exec(ctx,
		`INSERT INTO plaid_transactions
			(plaid_transaction_id, account_id, amount, transaction_date, description,
			 merchant_name, pending, payment_channel, plaid_category_primary,
			 plaid_category_detailed, created_time, last_modified_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		tx.TransactionDate.Format("2006-01-02"),
		tx.Description,
		tx.MerchantName,
		tx.Pending,
		tx.PaymentChannel,
		tx.PlaidCategoryPrimary,
		tx.PlaidCategoryDetailed,
		now,
		now,
	)
	return err
}

func (s Store) DeletePlaidTransaction(ctx context.Context, plaidTransactionID string) error {
	// Nullify the FK on any domain transaction that references this raw row,
	// preserving the domain row (and its plan assignment) for historical plans.
	_, err := s.db.Exec(ctx, `
		UPDATE transactions
		SET plaid_transaction_id = NULL
		WHERE plaid_transaction_id = (
			SELECT id FROM plaid_transactions WHERE plaid_transaction_id = ?
		)`, plaidTransactionID)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx,
		`DELETE FROM plaid_transactions WHERE plaid_transaction_id = ?`,
		plaidTransactionID,
	)
	return err
}

func (s Store) GetTransactionByID(ctx context.Context, id int) (types.Transaction, error) {
	row := s.db.QueryRow(ctx,
		`SELECT t.id, COALESCE(pt.plaid_transaction_id, ''), t.account_id,
		        t.amount, t.transaction_date, t.description, t.merchant_name,
		        t.pending, t.payment_channel, t.plaid_category_primary,
		        t.plaid_category_detailed, t.created_time, t.last_modified_time
		FROM transactions t
		LEFT JOIN plaid_transactions pt ON t.plaid_transaction_id = pt.id
		WHERE t.id = ?`,
		id,
	)

	var tx types.TransactionDTO
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

	return parse.ParseTransactionDTO(tx)
}

func (s Store) GetTransactions(ctx context.Context, filter types.TransactionFilter) ([]types.Transaction, error) {
	prefix := fmt.Sprintf("%04d-%02d-", filter.Year, filter.Month)

	query := `
		SELECT t.id, COALESCE(pt.plaid_transaction_id, ''), t.account_id,
		       t.amount, t.transaction_date, t.description, t.merchant_name,
		       t.pending, t.payment_channel, t.plaid_category_primary,
		       t.plaid_category_detailed, t.created_time, t.last_modified_time
		FROM transactions t
		LEFT JOIN plaid_transactions pt ON t.plaid_transaction_id = pt.id
		WHERE t.transaction_date LIKE ?`
	args := []any{prefix + "%"}

	if filter.AccountID != nil {
		query += " AND t.account_id = ?"
		args = append(args, *filter.AccountID)
	}

	query += " ORDER BY t.transaction_date DESC, t.id DESC"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []types.Transaction
	for rows.Next() {
		var tx types.TransactionDTO
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

		transaction, err := parse.ParseTransactionDTO(tx)
		if err != nil {
			return nil, err
		}

		txs = append(txs, transaction)
	}
	return txs, rows.Err()
}
