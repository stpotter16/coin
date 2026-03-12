package sqlite

import (
	"context"
	"time"

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

func (s Store) DeleteTransaction(ctx context.Context, plaidTransactionID string) error {
	_, err := s.db.Exec(ctx,
		`DELETE FROM transactions WHERE plaid_transaction_id = ?`,
		plaidTransactionID,
	)
	return err
}
