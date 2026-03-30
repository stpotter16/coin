package types

import (
	"fmt"
	"time"
)

type Plan struct {
	ID               int
	Year             int
	Month            int
	Locked           bool
	CreatedTime      time.Time
	LastModifiedTime time.Time
}

func (p Plan) MonthLabel() string {
	return time.Date(p.Year, time.Month(p.Month), 1, 0, 0, 0, 0, time.UTC).Format("January 2006")
}

type PlanItem struct {
	ID               int
	PlanID           int
	Name             string
	Type             string // "income" or "fixed_expense"
	ExpectedAmount   float64
	CreatedTime      time.Time
	LastModifiedTime time.Time
}

// PlanItemSummary extends PlanItem with the actual amount derived from
// assigned transactions. ActualAmount follows Plaid's sign convention
// (negative = money in, positive = money out).
type PlanItemSummary struct {
	PlanItem
	ActualAmount float64
}

func (p PlanItemSummary) FormattedExpected() string {
	return fmt.Sprintf("$%.2f", p.ExpectedAmount)
}

// ActualDisplay returns the actual amount as a positive number for display.
// Income transactions carry negative Plaid amounts; we negate them here.
func (p PlanItemSummary) ActualDisplay() float64 {
	if p.Type == "income" {
		return -p.ActualAmount
	}
	return p.ActualAmount
}

func (p PlanItemSummary) FormattedActual() string {
	return fmt.Sprintf("$%.2f", p.ActualDisplay())
}
