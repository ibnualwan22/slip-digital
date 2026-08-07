package domain

type DashboardStats struct {
	TotalEmployees    int64                `json:"total_employees"`
	ActiveEmployees   int64                `json:"active_employees"`
	MonthlyStats      []MonthlyExpenseStat `json:"monthly_stats"`
}

type MonthlyExpenseStat struct {
	Month          int     `json:"month"`
	Year           int     `json:"year"`
	PayrollCount   int64   `json:"payroll_count"`
	PayrollExpense float64 `json:"payroll_expense"`
	MonthlyExpense float64 `json:"monthly_expense"`
	GlobalExpense  float64 `json:"global_expense"`
}

type DashboardRepository interface {
	GetDashboardStats(year int) (*DashboardStats, error)
}

type DashboardService interface {
	GetDashboardStats(year int) (*DashboardStats, error)
}
