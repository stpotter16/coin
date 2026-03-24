package parse

import (
	"github.com/plaid/plaid-go/v21/plaid"
	"github.com/stpotter16/coin/internal/types"
)

func ParsePlaidAccount(pa plaid.AccountBase, plaidItemID int) (types.Account, error) {
	balances := pa.GetBalances()
	a := types.Account{
		PlaidAccountID:  pa.GetAccountId(),
		PlaidItemID:     plaidItemID,
		Name:            pa.GetName(),
		Type:            string(pa.GetType()),
		Subtype:         string(*pa.GetSubtype().Ptr()),
		IsoCurrencyCode: balances.GetIsoCurrencyCode(),
	}

	var accountName types.AccountName
	if name := pa.GetOfficialName(); name != "" {
		accountName.Value = &name
	} else {
		accountName.Value = nil
	}
	a.OfficialName = accountName

	var currentBalance types.Balance
	if balance, ok := balances.GetCurrentOk(); ok && balance != nil {
		currentBalance.Value = balance
	} else {
		currentBalance.Value = nil
	}
	a.CurrentBalance = currentBalance

	var availableBalance types.Balance
	if balance, ok := balances.GetAvailableOk(); ok && balance != nil {
		availableBalance.Value = balance
	} else {
		availableBalance.Value = nil
	}
	a.AvailableBalance = availableBalance

	return a, nil
}

func ParseAccountDTO(a types.AccountDTO) (types.Account, error) {
	var accountName types.AccountName
	if a.OfficialName.Valid {
		accountName.Value = &a.OfficialName.String
	} else {
		accountName.Value = nil
	}

	var currentBalance types.Balance
	if a.CurrentBalance.Valid {
		currentBalance.Value = &a.CurrentBalance.Float64
	} else {
		currentBalance.Value = nil
	}

	var availableBalance types.Balance
	if a.AvailableBalance.Valid {
		availableBalance.Value = &a.AvailableBalance.Float64
	} else {
		availableBalance.Value = nil
	}

	account := types.Account{
		ID:               a.ID,
		PlaidAccountID:   a.PlaidAccountID,
		PlaidItemID:      a.PlaidItemID,
		Name:             a.Name,
		OfficialName:     accountName,
		Type:             a.Type,
		Subtype:          a.Subtype,
		CurrentBalance:   currentBalance,
		AvailableBalance: availableBalance,
		IsoCurrencyCode:  a.IsoCurrencyCode,
		CreatedTime:      a.CreatedTime,
		LastModifiedTime: a.LastModifiedTime,
	}

	return account, nil
}
