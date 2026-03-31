package parse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/plaid/plaid-go/v21/plaid"
	"github.com/stpotter16/coin/internal/types"
)

func ParseTransactionExcluded(r *http.Request) (types.TransactionExcludedRequest, error) {
	var body struct {
		Excluded bool `json:"excluded"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return types.TransactionExcludedRequest{}, err
	}
	return types.TransactionExcludedRequest{Excluded: body.Excluded}, nil
}

func ParsePlaidTransaction(pt plaid.Transaction, accountID int) (types.PlaidTransaction, error) {
	t := types.PlaidTransaction{
		PlaidTransactionID: pt.GetTransactionId(),
		AccountID:          accountID,
		Amount:             pt.GetAmount(),
		TransactionDate:    pt.GetDate(),
		Description:        pt.GetName(),
		Pending:            pt.GetPending(),
		PaymentChannel:     pt.GetPaymentChannel(),
	}

	if name, ok := pt.GetMerchantNameOk(); ok && name != nil {
		t.MerchantName = name
	}

	if pt.HasPersonalFinanceCategory() {
		pfc := pt.GetPersonalFinanceCategory()
		primary := pfc.GetPrimary()
		detailed := pfc.GetDetailed()
		t.PlaidCategoryPrimary = &primary
		t.PlaidCategoryDetailed = &detailed
	}

	return t, nil
}

func ParseTransactionDTO(dto types.TransactionDTO) (types.Transaction, error) {
	date, err := time.Parse("2006-01-02", dto.TransactionDate)
	if err != nil {
		return types.Transaction{}, fmt.Errorf("invalid transaction date %q: %w", dto.TransactionDate, err)
	}

	t := types.Transaction{
		ID:                 dto.ID,
		PlaidTransactionID: dto.PlaidTransactionID,
		AccountID:          dto.AccountID,
		AccountName:        dto.AccountName,
		Amount:             dto.Amount,
		TransactionDate:    date,
		Description:        dto.Description,
		Pending:            dto.Pending,
		Excluded:           dto.Excluded,
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

	if dto.PlanItemID.Valid {
		t.PlanItem = &types.AssignedPlanItem{
			ID:   int(dto.PlanItemID.Int64),
			Name: dto.PlanItemName.String,
		}
	}

	return t, nil
}
