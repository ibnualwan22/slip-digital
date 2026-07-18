package postgres

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ibnualwan/bisyaroh/internal/domain"
)

type expenseRepository struct {
	db *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) domain.ExpenseRepository {
	return &expenseRepository{db: db}
}

// ----- Report -----

func (r *expenseRepository) CreateReport(report *domain.ExpenseReport) error {
	if report.ID == uuid.Nil {
		report.ID = uuid.New()
	}
	return r.db.Create(report).Error
}

func (r *expenseRepository) GetReportByID(id uuid.UUID) (*domain.ExpenseReport, error) {
	var report domain.ExpenseReport
	err := r.db.Preload("Items").First(&report, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *expenseRepository) ListReports(month, year int) ([]domain.ExpenseReport, error) {
	var reports []domain.ExpenseReport
	query := r.db.Model(&domain.ExpenseReport{})
	if month > 0 {
		query = query.Where("month = ?", month)
	}
	if year > 0 {
		query = query.Where("year = ?", year)
	}
	err := query.Order("year DESC, month DESC").Find(&reports).Error
	return reports, err
}

func (r *expenseRepository) UpdateReport(report *domain.ExpenseReport) error {
	return r.db.Save(report).Error
}

func (r *expenseRepository) DeleteReport(id uuid.UUID) error {
	return r.db.Delete(&domain.ExpenseReport{}, "id = ?", id).Error
}

// ----- Items -----

func (r *expenseRepository) CreateItem(item *domain.ExpenseItem) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	return r.db.Create(item).Error
}

func (r *expenseRepository) GetItemByID(id uuid.UUID) (*domain.ExpenseItem, error) {
	var item domain.ExpenseItem
	err := r.db.First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *expenseRepository) GetItemsByReportID(reportID uuid.UUID) ([]domain.ExpenseItem, error) {
	var items []domain.ExpenseItem
	err := r.db.Where("expense_report_id = ?", reportID).Order("created_at ASC").Find(&items).Error
	return items, err
}

func (r *expenseRepository) UpdateItem(item *domain.ExpenseItem) error {
	return r.db.Save(item).Error
}

func (r *expenseRepository) DeleteItem(id uuid.UUID) error {
	return r.db.Delete(&domain.ExpenseItem{}, "id = ?", id).Error
}
