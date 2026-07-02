package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type EmployeeCategory struct {
	ID                  uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	Name                string          `gorm:"type:varchar(255);not null" json:"name"`
	Code                string          `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	FixedSalary         decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"fixed_salary"`
	StructuralAllowance decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"structural_allowance"`
	TargetIncentive     decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"target_incentive"`
	HourlyRate          decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"hourly_rate"`
	CalcMethod          string          `gorm:"type:varchar(50);not null;default:'FIXED'" json:"calc_method"` // FIXED, HOURLY, PROPORTIONAL
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type CategoryRepository interface {
	Create(category *EmployeeCategory) error
	GetByID(id uuid.UUID) (*EmployeeCategory, error)
	GetByCode(code string) (*EmployeeCategory, error)
	List() ([]EmployeeCategory, error)
	Update(category *EmployeeCategory) error
	Delete(id uuid.UUID) error
}
