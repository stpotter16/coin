package types

type AccountDisplay struct {
	Name             string
	Subtype          string
	CurrentBalance   string // formatted, e.g. "$1,234.56" or "—"
	AvailableBalance string // formatted, e.g. "$1,234.56" or "" if nil
	LastSynced       string // formatted date/time
}

type InstitutionWithAccounts struct {
	InstitutionName string
	Accounts        []AccountDisplay
}
