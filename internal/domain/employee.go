package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Employee struct {
	ID                  uuid.UUID         `gorm:"type:uuid;primaryKey" json:"id"`
	Name                string            `gorm:"type:varchar(255);not null" json:"name"`
	CategoryID          *uuid.UUID        `gorm:"type:uuid" json:"category_id"`
	Category            *EmployeeCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Role                string            `gorm:"type:varchar(100)" json:"role"`
	// Per-employee overrides (null = use category default)
	StructuralAllowance *decimal.Decimal  `gorm:"type:numeric(19,4)" json:"structural_allowance"`
	HourlyRate          *decimal.Decimal  `gorm:"type:numeric(19,4)" json:"hourly_rate"`
	PhoneWA             string            `gorm:"type:varchar(20)" json:"phone_wa"`
	IsActive            bool              `gorm:"default:true" json:"is_active"`
	SiakadID            *string           `gorm:"type:varchar(50);index" json:"siakad_id"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

type EmployeeRepository interface {
	Create(employee *Employee) error
	GetByID(id uuid.UUID) (*Employee, error)
	List(categoryID *uuid.UUID, activeOnly bool) ([]Employee, error)
	Update(employee *Employee) error
	Delete(id uuid.UUID) error
}
