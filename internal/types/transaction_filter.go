package types

type TransactionFilter struct {
	Year      int
	Month     int  // 1–12
	AccountID *int // nil = all accounts
}
