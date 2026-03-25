package parse

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/stpotter16/coin/internal/types"
)

func ParsePlaidExchangePost(r *http.Request) (types.PlaidExchangeRequest, error) {
	var req types.PlaidExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return types.PlaidExchangeRequest{}, err
	}

	if req.PublicToken == "" || req.InstitutionID == "" || req.InstitutionName == "" {
		return types.PlaidExchangeRequest{}, errors.New("public_token, institution_id, and institution_name are required")
	}

	return req, nil
}
