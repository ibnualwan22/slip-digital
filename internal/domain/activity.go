package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MasterActivity struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	ActivityName string          `gorm:"type:varchar(255);not null" json:"activity_name"`
	Code         string          `gorm:"type:varchar(50);unique;not null" json:"code"`
	DefaultRate  decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"default_rate"`
	Type         ActivityType    `gorm:"type:varchar(20);not null" json:"type"`
	Description  string          `gorm:"type:text" json:"description"`
	IsActive     bool            `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time       `json:"created_at"`
}

type ActivityRepository interface {
	Create(activity *MasterActivity) error
	GetByID(id uuid.UUID) (*MasterActivity, error)
	GetByCode(code string) (*MasterActivity, error)
	List(activeOnly bool) ([]MasterActivity, error)
	Update(activity *MasterActivity) error
}
