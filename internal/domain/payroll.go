package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PayrollTransaction struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	EmployeeID      uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_emp_month_year" json:"employee_id"`
	Employee        Employee        `gorm:"foreignKey:EmployeeID" json:"employee"`
	Month           int             `gorm:"not null;uniqueIndex:idx_emp_month_year" json:"month"`
	Year            int             `gorm:"not null;uniqueIndex:idx_emp_month_year" json:"year"`
	GrossIncome     decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"gross_income"`
	TotalDeductions decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"total_deductions"`
	TakeHomePay     decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"take_home_pay"`
	Status          PayrollStatus   `gorm:"type:varchar(20);default:'DRAFT'" json:"status"`
	Notes           string          `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`

	Details []PayrollDetail `gorm:"foreignKey:PayrollTransactionID;constraint:OnDelete:CASCADE" json:"details"`
}

type PayrollDetail struct {
	ID                   uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	PayrollTransactionID uuid.UUID       `gorm:"type:uuid;not null" json:"payroll_transaction_id"`
	ActivityID           *uuid.UUID      `gorm:"type:uuid" json:"activity_id"`
	Activity             *MasterActivity `gorm:"foreignKey:ActivityID" json:"activity"`
	Quantity             decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"quantity"`
	Rate                 decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"rate"`
	TotalAmount          decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"total_amount"`
	Type                 ActivityType    `gorm:"type:varchar(20);not null" json:"type"`
	Description          string          `gorm:"type:text" json:"description"`
	RecordedAt           string          `gorm:"type:varchar(50)" json:"recorded_at"`
	CreatedAt            time.Time       `json:"created_at"`
}

type PayrollRepository interface {
	CreateTransaction(tx *PayrollTransaction) error
	GetTransactionByID(id uuid.UUID) (*PayrollTransaction, error)
	ListTransactions(month, year int) ([]PayrollTransaction, error)
	UpdateTransaction(tx *PayrollTransaction) error
	DeleteTransaction(id uuid.UUID) error
	
	CreateDetail(detail *PayrollDetail) error
	UpdateDetail(detail *PayrollDetail) error
	DeleteDetail(id uuid.UUID) error
	GetDetailsByTransactionID(txID uuid.UUID) ([]PayrollDetail, error)
	GetDetailByID(id uuid.UUID) (*PayrollDetail, error)
}
