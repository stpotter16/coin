package types

type TransactionRequest struct {
	AccountID    *int
	Amount       float64 // signed: positive = expense, negative = income
	Date         string  // YYYY-MM-DD
	Description  string
	MerchantName *string
	Pending      bool
}
