package parse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/stpotter16/coin/internal/types"
)

func ParseTransactionRequest(r *http.Request) (types.TransactionRequest, error) {
	var body struct {
		AccountID    *int    `json:"account_id"`
		Amount       float64 `json:"amount"` // signed: positive = expense, negative = income
		Date         string  `json:"date"`
		Description  string  `json:"description"`
		MerchantName *string `json:"merchant_name"`
		Pending      bool    `json:"pending"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return types.TransactionRequest{}, err
	}
	if body.Description == "" {
		return types.TransactionRequest{}, fmt.Errorf("description is required")
	}
	if _, err := time.Parse("2006-01-02", body.Date); err != nil {
		return types.TransactionRequest{}, fmt.Errorf("invalid date %q", body.Date)
	}
	return types.TransactionRequest{
		AccountID:    body.AccountID,
		Amount:       body.Amount,
		Date:         body.Date,
		Description:  body.Description,
		MerchantName: body.MerchantName,
		Pending:      body.Pending,
	}, nil
}


func ParseTransactionDTO(dto types.TransactionDTO) (types.Transaction, error) {
	date, err := time.Parse("2006-01-02", dto.TransactionDate)
	if err != nil {
		return types.Transaction{}, fmt.Errorf("invalid transaction date %q: %w", dto.TransactionDate, err)
	}

	t := types.Transaction{
		ID:               dto.ID,
		AccountName:      dto.AccountName,
		Amount:           dto.Amount,
		TransactionDate:  date,
		Description:      dto.Description,
		Pending:          dto.Pending,
		CreatedTime:      dto.CreatedTime,
		LastModifiedTime: dto.LastModifiedTime,
	}

	if dto.AccountID.Valid {
		id := int(dto.AccountID.Int64)
		t.AccountID = &id
	}

	if dto.MerchantName.Valid {
		t.MerchantName = types.MerchantName{Value: &dto.MerchantName.String}
	}

	if dto.PlanItemID.Valid {
		t.PlanItem = &types.AssignedPlanItem{
			ID:   int(dto.PlanItemID.Int64),
			Name: dto.PlanItemName.String,
		}
	}

	return t, nil
}
