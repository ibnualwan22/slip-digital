package postgres

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ibnualwan/bisyaroh/internal/domain"
)

type kbmSyncLogRepository struct {
	db *gorm.DB
}

func NewKBMSyncLogRepository(db *gorm.DB) domain.KBMSyncLogRepository {
	return &kbmSyncLogRepository{db: db}
}

func (r *kbmSyncLogRepository) Create(log *domain.KBMSyncLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	return r.db.Create(log).Error
}

func (r *kbmSyncLogRepository) Latest(month, year int) (*domain.KBMSyncLog, error) {
	var log domain.KBMSyncLog
	err := r.db.Where("month = ? AND year = ?", month, year).Order("created_at DESC").First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}
