package usecase

import (
	"time"

	"github.com/google/uuid"
	"github.com/ibnualwan/bisyaroh/internal/domain"
)

type EmployeeService interface {
	CreateEmployee(emp *domain.Employee) error
	GetEmployee(id uuid.UUID) (*domain.Employee, error)
	ListEmployees(categoryID *uuid.UUID, activeOnly bool) ([]domain.Employee, error)
	UpdateEmployee(emp *domain.Employee) error
	DeleteEmployee(id uuid.UUID) error
}

type employeeService struct {
	repo       domain.EmployeeRepository
	payrollSvc PayrollService
}

func NewEmployeeService(repo domain.EmployeeRepository, payrollSvc PayrollService) EmployeeService {
	return &employeeService{repo: repo, payrollSvc: payrollSvc}
}

func (s *employeeService) CreateEmployee(emp *domain.Employee) error {
	err := s.repo.Create(emp)
	if err == nil {
		now := time.Now()
		// Try to create payroll for current month, ignore errors to not block employee creation
		s.payrollSvc.EnsurePayrollForEmployee(emp.ID, int(now.Month()), now.Year())
	}
	return err
}

func (s *employeeService) GetEmployee(id uuid.UUID) (*domain.Employee, error) {
	return s.repo.GetByID(id)
}

func (s *employeeService) ListEmployees(categoryID *uuid.UUID, activeOnly bool) ([]domain.Employee, error) {
	return s.repo.List(categoryID, activeOnly)
}

func (s *employeeService) UpdateEmployee(emp *domain.Employee) error {
	return s.repo.Update(emp)
}

func (s *employeeService) DeleteEmployee(id uuid.UUID) error {
	return s.repo.Delete(id)
}
