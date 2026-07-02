package postgres

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ibnualwan/bisyaroh/internal/domain"
)

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) domain.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(category *domain.EmployeeCategory) error {
	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}
	return r.db.Create(category).Error
}

func (r *categoryRepository) GetByID(id uuid.UUID) (*domain.EmployeeCategory, error) {
	var cat domain.EmployeeCategory
	err := r.db.First(&cat, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *categoryRepository) GetByCode(code string) (*domain.EmployeeCategory, error) {
	var cat domain.EmployeeCategory
	err := r.db.First(&cat, "code = ?", code).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *categoryRepository) List() ([]domain.EmployeeCategory, error) {
	var categories []domain.EmployeeCategory
	err := r.db.Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) Update(category *domain.EmployeeCategory) error {
	return r.db.Save(category).Error
}

func (r *categoryRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.EmployeeCategory{}, "id = ?", id).Error
}
