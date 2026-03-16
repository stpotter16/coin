package types

import "time"

type Category struct {
	ID               int
	Name             string
	CreatedBy        int
	LastModifiedBy   int
	CreatedTime      time.Time
	LastModifiedTime time.Time
}
