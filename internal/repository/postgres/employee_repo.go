package postgres

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ibnualwan/bisyaroh/internal/domain"
)

type employeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) domain.EmployeeRepository {
	return &employeeRepository{db: db}
}

func (r *employeeRepository) Create(employee *domain.Employee) error {
	if employee.ID == uuid.Nil {
		employee.ID = uuid.New()
	}
	return r.db.Create(employee).Error
}

func (r *employeeRepository) GetByID(id uuid.UUID) (*domain.Employee, error) {
	var emp domain.Employee
	err := r.db.Preload("Category").First(&emp, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

func (r *employeeRepository) List(categoryID *uuid.UUID, activeOnly bool) ([]domain.Employee, error) {
	var employees []domain.Employee
	query := r.db.Model(&domain.Employee{}).Preload("Category")
	
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Find(&employees).Error
	return employees, err
}

func (r *employeeRepository) Update(employee *domain.Employee) error {
	return r.db.Save(employee).Error
}

func (r *employeeRepository) Delete(id uuid.UUID) error {
	// Soft delete logically by setting is_active = false
	return r.db.Model(&domain.Employee{}).Where("id = ?", id).Update("is_active", false).Error
}
