package parse

import (
	"github.com/plaid/plaid-go/v21/plaid"
	"github.com/stpotter16/coin/internal/types"
)

func ParsePlaidTransaction(pt plaid.Transaction, accountID int) types.Transaction {
	t := types.Transaction{
		PlaidTransactionID: pt.GetTransactionId(),
		AccountID:          accountID,
		Amount:             pt.GetAmount(),
		TransactionDate:    pt.GetDate(),
		Description:        pt.GetName(),
		Pending:            pt.GetPending(),
		PaymentChannel:     pt.GetPaymentChannel(),
	}

	if name, ok := pt.GetMerchantNameOk(); ok && name != nil {
		t.MerchantName = types.MerchantName{Value: name}
	}

	if pt.HasPersonalFinanceCategory() {
		pfc := pt.GetPersonalFinanceCategory()
		primary := pfc.GetPrimary()
		detailed := pfc.GetDetailed()
		t.PlaidCategoryPrimary = types.PlaidCategory{Value: &primary}
		t.PlaidCategoryDetailed = types.PlaidCategory{Value: &detailed}
	}

	return t
}

func ParseTransactionDTO(dto types.TransactionDTO) (types.Transaction, error) {
	t := types.Transaction{
		ID:                 dto.ID,
		PlaidTransactionID: dto.PlaidTransactionID,
		AccountID:          dto.AccountID,
		Amount:             dto.Amount,
		TransactionDate:    dto.TransactionDate,
		Description:        dto.Description,
		Pending:            dto.Pending,
		PaymentChannel:     dto.PaymentChannel,
		CreatedTime:        dto.CreatedTime,
		LastModifiedTime:   dto.LastModifiedTime,
	}

	if dto.MerchantName.Valid {
		t.MerchantName = types.MerchantName{Value: &dto.MerchantName.String}
	}

	if dto.PlaidCategoryPrimary.Valid {
		t.PlaidCategoryPrimary = types.PlaidCategory{Value: &dto.PlaidCategoryPrimary.String}
	}

	if dto.PlaidCategoryDetailed.Valid {
		t.PlaidCategoryDetailed = types.PlaidCategory{Value: &dto.PlaidCategoryDetailed.String}
	}

	if dto.CategoryID.Valid {
		id := int(dto.CategoryID.Int64)
		t.CategoryID = types.CategoryID{Value: &id}
	}

	if dto.LastModifiedByID.Valid {
		t.LastModifiedBy = types.NullableUser{
			Value: &types.User{
				ID:       int(dto.LastModifiedByID.Int64),
				Username: dto.LastModifiedByUsername.String,
			},
		}
	}

	return t, nil
}
