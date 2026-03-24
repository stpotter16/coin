package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/stpotter16/coin/internal/parse"
	"github.com/stpotter16/coin/internal/types"
)

func (s Store) GetAllAccounts(ctx context.Context) ([]types.Account, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, plaid_account_id, plaid_item_id, name, official_name, type, subtype,
			current_balance, available_balance, iso_currency_code,
			created_time, last_modified_time
		FROM account ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []types.Account
	for rows.Next() {
		var a types.AccountDTO
		var createdTime, lastModifiedTime string

		if err := rows.Scan(
			&a.ID, &a.PlaidAccountID, &a.PlaidItemID, &a.Name, &a.OfficialName,
			&a.Type, &a.Subtype, &a.CurrentBalance, &a.AvailableBalance,
			&a.IsoCurrencyCode, &createdTime, &lastModifiedTime,
		); err != nil {
			return nil, err
		}

		a.CreatedTime, err = parseTime(createdTime)
		if err != nil {
			return nil, err
		}
		a.LastModifiedTime, err = parseTime(lastModifiedTime)
		if err != nil {
			return nil, err
		}
		account, err := parse.ParseAccountDTO(a)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}

func (s Store) GetAccountsByItemID(ctx context.Context, plaidItemID int) ([]types.Account, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, plaid_account_id, plaid_item_id, name, official_name, type, subtype,
			current_balance, available_balance, iso_currency_code,
			created_time, last_modified_time
		FROM account WHERE plaid_item_id = ?`,
		plaidItemID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []types.Account
	for rows.Next() {
		var a types.AccountDTO
		var createdTime, lastModifiedTime string

		if err := rows.Scan(
			&a.ID, &a.PlaidAccountID, &a.PlaidItemID, &a.Name, &a.OfficialName,
			&a.Type, &a.Subtype, &a.CurrentBalance, &a.AvailableBalance,
			&a.IsoCurrencyCode, &createdTime, &lastModifiedTime,
		); err != nil {
			return nil, err
		}

		a.CreatedTime, err = parseTime(createdTime)
		if err != nil {
			return nil, err
		}
		a.LastModifiedTime, err = parseTime(lastModifiedTime)
		if err != nil {
			return nil, err
		}
		account, err := parse.ParseAccountDTO(a)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}

func (s Store) UpsertAccount(ctx context.Context, account types.Account) error {
	var officialName sql.NullString
	if account.OfficialName.Valid() {
		officialName = sql.NullString{
			String: *account.OfficialName.Value,
			Valid:  true,
		}
	} else {
		officialName = sql.NullString{
			String: "",
			Valid:  false,
		}
	}

	var currentBalance sql.NullFloat64
	if account.CurrentBalance.Valid() {
		currentBalance = sql.NullFloat64{
			Float64: *account.CurrentBalance.Value,
			Valid:   true,
		}
	} else {
		currentBalance = sql.NullFloat64{
			Float64: 0,
			Valid:   false,
		}
	}

	var availableBalance sql.NullFloat64
	if account.AvailableBalance.Valid() {
		availableBalance = sql.NullFloat64{
			Float64: *account.AvailableBalance.Value,
			Valid:   true,
		}
	} else {
		availableBalance = sql.NullFloat64{
			Float64: 0,
			Valid:   false,
		}
	}

	now := formatTime(time.Now().UTC())
	_, err := s.db.Exec(ctx,
		`INSERT INTO account
			(plaid_account_id, plaid_item_id, name, official_name, type, subtype,
			 current_balance, available_balance, iso_currency_code,
			 created_time, last_modified_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(plaid_account_id) DO UPDATE SET
			name               = excluded.name,
			official_name      = excluded.official_name,
			current_balance    = excluded.current_balance,
			available_balance  = excluded.available_balance,
			last_modified_time = excluded.last_modified_time`,
		account.PlaidAccountID,
		account.PlaidItemID,
		account.Name,
		officialName,
		account.Type,
		account.Subtype,
		currentBalance,
		availableBalance,
		account.IsoCurrencyCode,
		now,
		now,
	)
	return err
}
