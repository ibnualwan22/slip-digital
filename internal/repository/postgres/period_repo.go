package postgres

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ibnualwan/bisyaroh/internal/domain"
)

type periodRepository struct {
	db *gorm.DB
}

func NewPeriodRepository(db *gorm.DB) domain.PeriodRepository {
	return &periodRepository{db: db}
}

func (r *periodRepository) Create(period *domain.PeriodMapping) error {
	if period.ID == uuid.Nil {
		period.ID = uuid.New()
	}
	return r.db.Create(period).Error
}

func (r *periodRepository) GetByID(id uuid.UUID) (*domain.PeriodMapping, error) {
	var p domain.PeriodMapping
	err := r.db.First(&p, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *periodRepository) GetByMonthYear(month, year int) (*domain.PeriodMapping, error) {
	var p domain.PeriodMapping
	err := r.db.First(&p, "month = ? AND year = ?", month, year).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *periodRepository) List() ([]domain.PeriodMapping, error) {
	var periods []domain.PeriodMapping
	err := r.db.Order("year DESC, month DESC, start_date DESC").Find(&periods).Error
	return periods, err
}

func (r *periodRepository) Update(period *domain.PeriodMapping) error {
	return r.db.Save(period).Error
}

func (r *periodRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.PeriodMapping{}, "id = ?", id).Error
}
