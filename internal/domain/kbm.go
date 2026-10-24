package domain

import (
	"time"

	"github.com/google/uuid"
)

// KBMSyncLog mencatat setiap hitung ulang KBM berbasis pemetaan periode:
// pemicu (webhook/manual/rekonsiliasi/perbaikan), berapa pengajar berubah,
// dan ringkasan perubahannya — jejak audit untuk bendahara.
type KBMSyncLog struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Month            int       `gorm:"not null" json:"month"`
	Year             int       `gorm:"not null" json:"year"`
	DufahName        string    `gorm:"type:varchar(100)" json:"dufah_name"`
	Trigger          string    `gorm:"type:varchar(30);not null" json:"trigger"`
	TeachersUpdated  int       `gorm:"not null;default:0" json:"teachers_updated"`
	TeachersSkipped  int       `gorm:"not null;default:0" json:"teachers_skipped"`
	ChangesJSON      string    `gorm:"type:text" json:"changes_json"`
	CreatedAt        time.Time `json:"created_at"`
}

type KBMSyncLogRepository interface {
	Create(log *KBMSyncLog) error
	Latest(month, year int) (*KBMSyncLog, error)
}
