package types

type TransactionCategoryRequest struct {
	CategoryID *int
}

type TransactionNoteRequest struct {
	TransactionID int
	Note          string
}
