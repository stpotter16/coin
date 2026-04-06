package types

import "time"

type AccountDTO struct {
	ID               int
	Name             string
	Type             string
	CreatedTime      time.Time
	LastModifiedTime time.Time
}

type Account struct {
	ID               int
	Name             string
	Type             string
	CreatedTime      time.Time
	LastModifiedTime time.Time
}
