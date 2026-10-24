package domain

import (
	"time"

	"github.com/google/uuid"
)

// PeriodMapping memetakan satu periode bisyaroh (bulan gaji) ke dufah dan
// rentang tanggal cut-off-nya. Diatur admin di awal bulan; rekap pengajar
// dan tautan payroll mengikuti pemetaan ini.
type PeriodMapping struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DufahName string    `gorm:"type:varchar(100);not null" json:"dufah_name"`
	Usbu      int       `gorm:"not null;default:0" json:"usbu"`
	Month     int       `gorm:"not null" json:"month"`
	Year      int       `gorm:"not null" json:"year"`
	StartDate time.Time `gorm:"type:date;not null" json:"start_date"`
	EndDate   time.Time `gorm:"type:date;not null" json:"end_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PeriodRepository interface {
	Create(period *PeriodMapping) error
	GetByID(id uuid.UUID) (*PeriodMapping, error)
	GetByMonthYear(month, year int) (*PeriodMapping, error)
	List() ([]PeriodMapping, error)
	Update(period *PeriodMapping) error
	Delete(id uuid.UUID) error
}
