package parse

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/stpotter16/coin/internal/types"
)

func ParsePlanItemCreate(r *http.Request) (types.PlanItemCreateRequest, error) {
	var body struct {
		Name           string  `json:"name"`
		Type           string  `json:"type"`
		ExpectedAmount float64 `json:"expected_amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return types.PlanItemCreateRequest{}, err
	}
	if body.Name == "" {
		return types.PlanItemCreateRequest{}, errors.New("name is required")
	}
	if body.Type != "income" && body.Type != "fixed_expense" {
		return types.PlanItemCreateRequest{}, errors.New("type must be income or fixed_expense")
	}
	if body.ExpectedAmount <= 0 {
		return types.PlanItemCreateRequest{}, errors.New("expected_amount must be greater than zero")
	}
	return types.PlanItemCreateRequest{
		Name:           body.Name,
		Type:           body.Type,
		ExpectedAmount: body.ExpectedAmount,
	}, nil
}

func ParsePlanItemUpdate(r *http.Request) (types.PlanItemUpdateRequest, error) {
	var body struct {
		Name           string  `json:"name"`
		ExpectedAmount float64 `json:"expected_amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return types.PlanItemUpdateRequest{}, err
	}
	if body.Name == "" {
		return types.PlanItemUpdateRequest{}, errors.New("name is required")
	}
	if body.ExpectedAmount <= 0 {
		return types.PlanItemUpdateRequest{}, errors.New("expected_amount must be greater than zero")
	}
	return types.PlanItemUpdateRequest{
		Name:           body.Name,
		ExpectedAmount: body.ExpectedAmount,
	}, nil
}

func ParseTransactionPlanItem(r *http.Request) (types.TransactionPlanItemRequest, error) {
	var body struct {
		PlanItemID string `json:"plan_item_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return types.TransactionPlanItemRequest{}, err
	}
	var planItemID *int
	if body.PlanItemID != "" {
		id, err := strconv.Atoi(body.PlanItemID)
		if err != nil {
			return types.TransactionPlanItemRequest{}, errors.New("invalid plan_item_id")
		}
		planItemID = &id
	}
	return types.TransactionPlanItemRequest{PlanItemID: planItemID}, nil
}
