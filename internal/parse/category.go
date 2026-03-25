package parse

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/stpotter16/coin/internal/types"
)

func ParseCategoryCreatePost(r *http.Request) (types.CategoryCreateRequest, error) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return types.CategoryCreateRequest{}, err
	}
	if body.Name == "" {
		return types.CategoryCreateRequest{}, errors.New("name is required")
	}
	return types.CategoryCreateRequest{Name: body.Name}, nil
}

func ParseCategoryUpdatePost(r *http.Request) (types.CategoryUpdateRequest, error) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return types.CategoryUpdateRequest{}, err
	}
	if body.Name == "" {
		return types.CategoryUpdateRequest{}, errors.New("name is required")
	}
	return types.CategoryUpdateRequest{Name: body.Name}, nil
}
