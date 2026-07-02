package postgres

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ibnualwan/bisyaroh/internal/domain"
)

type payrollRepository struct {
	db *gorm.DB
}

func NewPayrollRepository(db *gorm.DB) domain.PayrollRepository {
	return &payrollRepository{db: db}
}

func (r *payrollRepository) CreateTransaction(tx *domain.PayrollTransaction) error {
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	return r.db.Create(tx).Error
}

func (r *payrollRepository) GetTransactionByID(id uuid.UUID) (*domain.PayrollTransaction, error) {
	var tx domain.PayrollTransaction
	err := r.db.Preload("Employee").Preload("Details").Preload("Details.Activity").First(&tx, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *payrollRepository) ListTransactions(month, year int) ([]domain.PayrollTransaction, error) {
	var txs []domain.PayrollTransaction
	query := r.db.Preload("Employee")
	
	if month > 0 {
		query = query.Where("month = ?", month)
	}
	if year > 0 {
		query = query.Where("year = ?", year)
	}

	err := query.Find(&txs).Error
	return txs, err
}

func (r *payrollRepository) UpdateTransaction(tx *domain.PayrollTransaction) error {
	return r.db.Save(tx).Error
}

func (r *payrollRepository) DeleteTransaction(id uuid.UUID) error {
	return r.db.Delete(&domain.PayrollTransaction{}, "id = ?", id).Error
}

func (r *payrollRepository) CreateDetail(detail *domain.PayrollDetail) error {
	if detail.ID == uuid.Nil {
		detail.ID = uuid.New()
	}
	return r.db.Create(detail).Error
}

func (r *payrollRepository) UpdateDetail(detail *domain.PayrollDetail) error {
	return r.db.Save(detail).Error
}

func (r *payrollRepository) DeleteDetail(id uuid.UUID) error {
	return r.db.Delete(&domain.PayrollDetail{}, "id = ?", id).Error
}

func (r *payrollRepository) GetDetailsByTransactionID(txID uuid.UUID) ([]domain.PayrollDetail, error) {
	var details []domain.PayrollDetail
	err := r.db.Preload("Activity").Where("payroll_transaction_id = ?", txID).Find(&details).Error
	return details, err
}

func (r *payrollRepository) GetDetailByID(id uuid.UUID) (*domain.PayrollDetail, error) {
	var detail domain.PayrollDetail
	err := r.db.First(&detail, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &detail, nil
}
