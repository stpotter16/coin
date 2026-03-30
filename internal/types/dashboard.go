package types

import (
	"fmt"
	"math"
)

// DashboardSummary holds the pre-computed values shown on the dashboard.
// All amount fields are positive absolute values ready for display.
type DashboardSummary struct {
	HasPlan          bool
	ExpectedIncome   float64
	ActualIncome     float64 // negated from Plaid convention — positive = received
	ExpectedFixed    float64
	ActualFixed      float64 // positive = paid
	FlexibleSpending float64 // sum of unassigned positive transactions
}

// Remaining is the core metric: income received minus fixed paid minus flexible spent.
func (d DashboardSummary) Remaining() float64 {
	return d.ActualIncome - d.ActualFixed - d.FlexibleSpending
}

func (d DashboardSummary) FormattedRemaining() string {
	return fmt.Sprintf("$%.2f", math.Abs(d.Remaining()))
}

func (d DashboardSummary) RemainingClass() string {
	if d.Remaining() < 0 {
		return "text-error"
	}
	return "text-success"
}

func (d DashboardSummary) FormattedExpectedIncome() string {
	return fmt.Sprintf("$%.2f", d.ExpectedIncome)
}

func (d DashboardSummary) FormattedActualIncome() string {
	return fmt.Sprintf("$%.2f", d.ActualIncome)
}

func (d DashboardSummary) FormattedExpectedFixed() string {
	return fmt.Sprintf("$%.2f", d.ExpectedFixed)
}

func (d DashboardSummary) FormattedActualFixed() string {
	return fmt.Sprintf("$%.2f", d.ActualFixed)
}

func (d DashboardSummary) FormattedFlexible() string {
	return fmt.Sprintf("$%.2f", d.FlexibleSpending)
}
