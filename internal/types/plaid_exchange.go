package types

type PlaidExchangeRequest struct {
	PublicToken     string `json:"public_token"`
	InstitutionID   string `json:"institution_id"`
	InstitutionName string `json:"institution_name"`
}
