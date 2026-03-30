package types

type PlanItemCreateRequest struct {
	Name           string
	Type           string
	ExpectedAmount float64
}

type PlanItemUpdateRequest struct {
	Name           string
	ExpectedAmount float64
}

type TransactionPlanItemRequest struct {
	PlanItemID *int
}
