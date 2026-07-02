package domain



type ActivityType string

const (
	TypeAddition   ActivityType = "ADDITION"
	TypeDeduction  ActivityType = "DEDUCTION"
)

type PayrollStatus string

const (
	StatusDraft     PayrollStatus = "DRAFT"
	StatusConfirmed PayrollStatus = "CONFIRMED"
	StatusPaid      PayrollStatus = "PAID"
)
