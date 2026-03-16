package types

type TransactionDisplay struct {
	ID          int
	DisplayName string // merchant name, falling back to description
	Amount      string // formatted, e.g. "$12.34"
	AmountClass string // "text-error" (positive/out) or "text-success" (negative/in)
	Category    string // user category name or formatted plaid primary category
	Pending     bool
}

type TransactionGroup struct {
	Date         string // formatted, e.g. "Mon, Mar 16"
	Transactions []TransactionDisplay
}
