package parse

import (
	"encoding/json"
	"net/http"

	"github.com/stpotter16/coin/internal/types"
)

func ParseLoginPost(r *http.Request) (types.LoginRequest, error) {
	body := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		return types.LoginRequest{}, err
	}

	request := types.LoginRequest{
		Username: body.Username,
		Password: body.Password,
	}
	return request, nil
}
