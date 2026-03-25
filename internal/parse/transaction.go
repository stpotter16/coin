package parse

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/plaid/plaid-go/v21/plaid"
	"github.com/stpotter16/coin/internal/types"
)

func ParseTransactionCategoryPost(r *http.Request) (types.TransactionCategoryRequest, error) {
	var body struct {
		CategoryID string `json:"category_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return types.TransactionCategoryRequest{}, err
	}

	var categoryID *int
	if body.CategoryID != "" {
		parsed, err := strconv.Atoi(body.CategoryID)
		if err != nil {
			return types.TransactionCategoryRequest{}, errors.New("invalid category_id")
		}
		categoryID = &parsed
	}

	return types.TransactionCategoryRequest{CategoryID: categoryID}, nil
}

const MaxNoteLength = 2000

func ParseTransactionNotePost(r *http.Request) (types.TransactionNoteRequest, error) {
	var body struct {
		TransactionID int    `json:"transaction_id"`
		Note          string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return types.TransactionNoteRequest{}, err
	}

	if body.TransactionID == 0 {
		return types.TransactionNoteRequest{}, errors.New("transaction_id is required")
	}
	if body.Note == "" {
		return types.TransactionNoteRequest{}, errors.New("note is required")
	}
	if len([]rune(body.Note)) > MaxNoteLength {
		return types.TransactionNoteRequest{}, errors.New("note exceeds maximum length")
	}

	return types.TransactionNoteRequest{
		TransactionID: body.TransactionID,
		Note:          body.Note,
	}, nil
}

func ParsePlaidTransaction(pt plaid.Transaction, accountID int) (types.Transaction, error) {
	date, err := time.Parse("2006-01-02", pt.GetDate())
	if err != nil {
		return types.Transaction{}, fmt.Errorf("invalid transaction date %q: %w", pt.GetDate(), err)
	}

	t := types.Transaction{
		PlaidTransactionID: pt.GetTransactionId(),
		AccountID:          accountID,
		Amount:             pt.GetAmount(),
		TransactionDate:    date,
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
		Amount:             dto.Amount,
		TransactionDate:    date,
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
