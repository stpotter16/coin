package types

type TransactionDetailDisplay struct {
	ID             int
	DisplayName    string // merchant name or description
	Description    string // always the full description
	MerchantName   string // empty if nil
	Date           string // formatted, e.g. "Mon, Jan 2, 2006"
	Amount         string // formatted, e.g. "$12.34"
	AmountClass    string // "text-error" or "text-success"
	PaymentChannel string
	PlaidCategory  string // formatted plaid_category_primary, empty if nil
	CategoryID     int    // 0 if no user override
	Pending        bool
}
