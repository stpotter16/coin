package parse

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type AccountRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func ParseAccountRequest(r *http.Request) (AccountRequest, error) {
	var req AccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return AccountRequest{}, err
	}
	if req.Name == "" {
		return AccountRequest{}, fmt.Errorf("name is required")
	}
	if req.Type == "" {
		return AccountRequest{}, fmt.Errorf("type is required")
	}
	return req, nil
}
