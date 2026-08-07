package postgres

import (
	"github.com/ibnualwan/bisyaroh/internal/domain"
	"gorm.io/gorm"
)

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) domain.DashboardRepository {
	return &dashboardRepository{db}
}

func (r *dashboardRepository) GetDashboardStats(year int) (*domain.DashboardStats, error) {
	stats := &domain.DashboardStats{
		MonthlyStats: make([]domain.MonthlyExpenseStat, 12),
	}

	for i := 0; i < 12; i++ {
		stats.MonthlyStats[i] = domain.MonthlyExpenseStat{
			Month: i + 1,
			Year:  year,
		}
	}

	// 1. Employees Stats
	r.db.Model(&domain.Employee{}).Count(&stats.TotalEmployees)
	r.db.Model(&domain.Employee{}).Where("is_active = ?", true).Count(&stats.ActiveEmployees)

	// 2. Regular Expenses
	type expenseResult struct {
		Month int
		Total float64
	}
	var expenses []expenseResult
	r.db.Model(&domain.ExpenseReport{}).
		Select("month, SUM(total_pengeluaran) as total").
		Where("year = ?", year).
		Group("month").
		Scan(&expenses)

	for _, e := range expenses {
		if e.Month >= 1 && e.Month <= 12 {
			stats.MonthlyStats[e.Month-1].MonthlyExpense = e.Total
		}
	}

	// 3. Payroll
	type payrollResult struct {
		Month int
		Year  int
		Total float64
		Count int64
	}
	var payrolls []payrollResult
	r.db.Table("payroll_transactions").
		Select("month, year, SUM(take_home_pay) as total, COUNT(id) as count").
		Where("year = ?", year).
		Group("year, month").
		Scan(&payrolls)

	for _, p := range payrolls {
		if p.Month >= 1 && p.Month <= 12 {
			stats.MonthlyStats[p.Month-1].PayrollExpense += p.Total
			stats.MonthlyStats[p.Month-1].PayrollCount += p.Count
		}
	}

	// 4. Calculate Global Expenses
	for i := range stats.MonthlyStats {
		stats.MonthlyStats[i].GlobalExpense = stats.MonthlyStats[i].PayrollExpense + stats.MonthlyStats[i].MonthlyExpense
	}

	return stats, nil
}
