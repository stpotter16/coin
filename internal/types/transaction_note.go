package types

import "time"

type TransactionNote struct {
	ID            int
	TransactionID int
	UserID        int
	Note          string
	CreatedTime   time.Time
}
