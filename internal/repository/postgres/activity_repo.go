package postgres

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ibnualwan/bisyaroh/internal/domain"
)

type activityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) domain.ActivityRepository {
	return &activityRepository{db: db}
}

func (r *activityRepository) Create(activity *domain.MasterActivity) error {
	if activity.ID == uuid.Nil {
		activity.ID = uuid.New()
	}
	return r.db.Create(activity).Error
}

func (r *activityRepository) GetByID(id uuid.UUID) (*domain.MasterActivity, error) {
	var act domain.MasterActivity
	err := r.db.First(&act, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &act, nil
}

func (r *activityRepository) GetByCode(code string) (*domain.MasterActivity, error) {
	var act domain.MasterActivity
	err := r.db.First(&act, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &act, nil
}

func (r *activityRepository) List(activeOnly bool) ([]domain.MasterActivity, error) {
	var activities []domain.MasterActivity
	query := r.db.Model(&domain.MasterActivity{})
	
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Find(&activities).Error
	return activities, err
}

func (r *activityRepository) Update(activity *domain.MasterActivity) error {
	return r.db.Save(activity).Error
}
