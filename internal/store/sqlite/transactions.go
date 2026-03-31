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
	var merchantName, categoryPrimary, categoryDetailed *string
	if tx.MerchantName.Valid() {
		v := tx.MerchantName.String()
		merchantName = &v
	}
	if tx.PlaidCategoryPrimary.Valid() {
		v := tx.PlaidCategoryPrimary.Value
		categoryPrimary = v
	}
	if tx.PlaidCategoryDetailed.Valid() {
		v := tx.PlaidCategoryDetailed.Value
		categoryDetailed = v
	}
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
		merchantName,
		tx.Pending,
		tx.PaymentChannel,
		categoryPrimary,
		categoryDetailed,
		now,
		now,
	)
	return err
}

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
		        t.plaid_category_detailed, t.plan_item_id, pi.name,
		        t.created_time, t.last_modified_time
		FROM transactions t
		LEFT JOIN plaid_transactions pt ON t.plaid_transaction_id = pt.id
		LEFT JOIN plan_items pi ON t.plan_item_id = pi.id
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
		&tx.PlanItemID,
		&tx.PlanItemName,
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

func (s Store) GetTransactions(ctx context.Context, filter types.TransactionFilter) (types.TransactionPage, error) {
	prefix := fmt.Sprintf("%04d-%02d-", filter.Year, filter.Month)

	page := filter.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * types.TransactionPageSize

	query := `
		SELECT t.id, COALESCE(pt.plaid_transaction_id, ''), t.account_id,
		       t.amount, t.transaction_date, t.description, t.merchant_name,
		       t.pending, t.payment_channel, t.plaid_category_primary,
		       t.plaid_category_detailed, t.plan_item_id, pi.name,
		       t.created_time, t.last_modified_time
		FROM transactions t
		LEFT JOIN plaid_transactions pt ON t.plaid_transaction_id = pt.id
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
			&tx.PlanItemID,
			&tx.PlanItemName,
			&createdTime,
			&lastModifiedTime,
		); err != nil {
			return types.TransactionPage{}, err
		}

		tx.CreatedTime, err = parseTime(createdTime)
		if err != nil {
			return types.TransactionPage{}, err
		}
		tx.LastModifiedTime, err = parseTime(lastModifiedTime)
		if err != nil {
			return types.TransactionPage{}, err
		}

		transaction, err := parse.ParseTransactionDTO(tx)
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
