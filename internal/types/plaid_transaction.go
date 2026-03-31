package types

// PlaidTransaction is the raw Plaid transaction data written to the
// plaid_transactions table by the sync job. Fields map 1-to-1 with the
// Plaid API response; nullable fields use *string.
type PlaidTransaction struct {
	PlaidTransactionID      string
	AccountID               int
	Amount                  float64
	TransactionDate         string // YYYY-MM-DD, stored as-is from Plaid
	Description             string
	MerchantName            *string
	Pending                 bool
	PaymentChannel          string
	PlaidCategoryPrimary    *string
	PlaidCategoryDetailed   *string
}
