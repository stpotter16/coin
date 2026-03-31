package types

const TransactionPageSize = 100

type TransactionFilter struct {
	Year      int
	Month     int  // 1–12
	AccountID *int // nil = all accounts
	Page      int  // 1-indexed; 0 or 1 both mean the first page
}
