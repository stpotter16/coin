package sqlite

import (
	"context"
	"time"
)

func (s Store) RunTransform(ctx context.Context) error {
	now := formatTime(time.Now().UTC())

	// Promote new plaid_transactions rows that have no domain transaction yet.
	_, err := s.db.Exec(ctx, `
		INSERT INTO transactions (
			plaid_transaction_id, account_id, amount, transaction_date,
			description, merchant_name, pending, payment_channel,
			plaid_category_primary, plaid_category_detailed,
			created_time, last_modified_time
		)
		SELECT
			pt.id, pt.account_id, pt.amount, pt.transaction_date,
			pt.description, pt.merchant_name, pt.pending, pt.payment_channel,
			pt.plaid_category_primary, pt.plaid_category_detailed,
			?, ?
		FROM plaid_transactions pt
		WHERE NOT EXISTS (
			SELECT 1 FROM transactions t WHERE t.plaid_transaction_id = pt.id
		)`, now, now)
	if err != nil {
		return err
	}

	// Update Plaid-sourced fields on domain transactions where the raw row
	// has been modified since the domain row was last written. User data
	// (plan_item_id) is untouched.
	_, err = s.db.Exec(ctx, `
		UPDATE transactions
		SET
			amount                  = pt.amount,
			transaction_date        = pt.transaction_date,
			description             = pt.description,
			merchant_name           = pt.merchant_name,
			pending                 = pt.pending,
			payment_channel         = pt.payment_channel,
			plaid_category_primary  = pt.plaid_category_primary,
			plaid_category_detailed = pt.plaid_category_detailed,
			last_modified_time      = ?
		FROM plaid_transactions pt
		WHERE transactions.plaid_transaction_id = pt.id
		  AND pt.last_modified_time > transactions.last_modified_time`,
		now)
	return err
}
