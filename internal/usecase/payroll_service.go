package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ibnualwan/bisyaroh/internal/domain"
	"github.com/shopspring/decimal"
)

type PayrollService interface {
	CreateTransaction(tx *domain.PayrollTransaction) error
	GetTransaction(id uuid.UUID) (*domain.PayrollTransaction, error)
	ListTransactions(month, year int) ([]domain.PayrollTransaction, error)
	DeleteTransaction(id uuid.UUID) error

	AddDetail(detail *domain.PayrollDetail) error
	UpdateDetail(detail *domain.PayrollDetail) error
	RemoveDetail(id uuid.UUID) error

	CalculateTransactionTHP(txID uuid.UUID) error
	UpdateStatus(txID uuid.UUID, status domain.PayrollStatus) error
	EnsurePayrollForEmployee(empID uuid.UUID, month, year int) error
	UpsertDetail(detail *domain.PayrollDetail) error
	GetDetail(id uuid.UUID) (*domain.PayrollDetail, error)
}

type payrollService struct {
	repo    domain.PayrollRepository
	empRepo domain.EmployeeRepository
}

func NewPayrollService(repo domain.PayrollRepository, empRepo domain.EmployeeRepository) PayrollService {
	return &payrollService{
		repo:    repo,
		empRepo: empRepo,
	}
}

func (s *payrollService) CreateTransaction(tx *domain.PayrollTransaction) error {
	// Ensure employee exists
	emp, err := s.empRepo.GetByID(tx.EmployeeID)
	if err != nil {
		return errors.New("employee not found")
	}

	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	tx.Status = domain.StatusDraft

	var initialGross decimal.Decimal

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.Local
	}
	nowWIB := time.Now().In(loc).Format("02/01/2006 15:04:05")

	// 1. Snapshot: Gaji Pokok (Fixed Salary)
	if emp.Category != nil && !emp.Category.FixedSalary.IsZero() {
		tx.Details = append(tx.Details, domain.PayrollDetail{
			ID:                   uuid.New(),
			PayrollTransactionID: tx.ID,
			Quantity:             decimal.NewFromInt(1),
			Rate:                 emp.Category.FixedSalary,
			TotalAmount:          emp.Category.FixedSalary,
			Type:                 domain.TypeAddition,
			Description:          "Gaji Pokok",
			RecordedAt:           nowWIB,
		})
		initialGross = initialGross.Add(emp.Category.FixedSalary)
	}

	// 2. Snapshot: Tunjangan Struktural (Employee override > Category default)
	allowance := decimal.Zero
	if emp.StructuralAllowance != nil && !emp.StructuralAllowance.IsZero() {
		allowance = *emp.StructuralAllowance
	} else if emp.Category != nil && !emp.Category.StructuralAllowance.IsZero() {
		allowance = emp.Category.StructuralAllowance
	}
	if !allowance.IsZero() {
		tx.Details = append(tx.Details, domain.PayrollDetail{
			ID:                   uuid.New(),
			PayrollTransactionID: tx.ID,
			Quantity:             decimal.NewFromInt(1),
			Rate:                 allowance,
			TotalAmount:          allowance,
			Type:                 domain.TypeAddition,
			Description:          "Tunjangan Struktural",
			RecordedAt:           nowWIB,
		})
		initialGross = initialGross.Add(allowance)
	}

	tx.GrossIncome = initialGross
	tx.TakeHomePay = initialGross // Pada awal pembuatan belum ada potongan

	return s.repo.CreateTransaction(tx)
}

func (s *payrollService) GetTransaction(id uuid.UUID) (*domain.PayrollTransaction, error) {
	return s.repo.GetTransactionByID(id)
}

func (s *payrollService) ListTransactions(month, year int) ([]domain.PayrollTransaction, error) {
	return s.repo.ListTransactions(month, year)
}

func (s *payrollService) DeleteTransaction(id uuid.UUID) error {
	return s.repo.DeleteTransaction(id)
}

func (s *payrollService) AddDetail(detail *domain.PayrollDetail) error {
	// calculate total amount directly if not provided
	if detail.TotalAmount.IsZero() && !detail.Quantity.IsZero() && !detail.Rate.IsZero() {
		detail.TotalAmount = detail.Quantity.Mul(detail.Rate)
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.Local
	}
	detail.RecordedAt = time.Now().In(loc).Format("02/01/2006 15:04:05")

	err = s.repo.CreateDetail(detail)
	if err == nil {
		s.CalculateTransactionTHP(detail.PayrollTransactionID)
	}
	return err
}

func (s *payrollService) UpsertDetail(detail *domain.PayrollDetail) error {
	details, err := s.repo.GetDetailsByTransactionID(detail.PayrollTransactionID)
	if err == nil {
		for _, d := range details {
			if d.ActivityID != nil && detail.ActivityID != nil && *d.ActivityID == *detail.ActivityID {
				// Update existing instead
				d.Quantity = detail.Quantity
				d.Description = detail.Description
				d.Rate = detail.Rate
				d.TotalAmount = detail.TotalAmount
				return s.UpdateDetail(&d)
			}
		}
	}
	return s.AddDetail(detail)
}

func (s *payrollService) UpdateDetail(detail *domain.PayrollDetail) error {
	if !detail.Quantity.IsZero() && !detail.Rate.IsZero() {
		detail.TotalAmount = detail.Quantity.Mul(detail.Rate)
	}
	err := s.repo.UpdateDetail(detail)
	if err == nil {
		s.CalculateTransactionTHP(detail.PayrollTransactionID)
	}
	return err
}

func (s *payrollService) GetDetail(id uuid.UUID) (*domain.PayrollDetail, error) {
	return s.repo.GetDetailByID(id)
}

func (s *payrollService) RemoveDetail(id uuid.UUID) error {
	// Need to get detail first to know which tx to calculate
	var txID uuid.UUID

	// Quick hack to get txID since we don't have GetDetailByID in repo
	// Let's modify the repo later if needed, but for now we can rely on
	// the fact that we can just pass the txID from the frontend?
	// Wait, the handler doesn't have it.
	// Oh well, we'll fix the handler to fetch it, or add GetDetailByID.

	// Actually let's add GetDetailByID to payrollRepository
	// I'll update it separately.
	detail, err := s.repo.GetDetailByID(id)
	if err != nil {
		return err
	}
	txID = detail.PayrollTransactionID

	err = s.repo.DeleteDetail(id)
	if err == nil {
		s.CalculateTransactionTHP(txID)
	}
	return err
}

func (s *payrollService) CalculateTransactionTHP(txID uuid.UUID) error {
	tx, err := s.repo.GetTransactionByID(txID)
	if err != nil {
		return err
	}

	if tx.Status != domain.StatusDraft {
		return errors.New("only DRAFT transactions can be calculated")
	}

	// Fetch fresh employee details just in case
	emp, err := s.empRepo.GetByID(tx.EmployeeID)
	if err != nil {
		return err
	}

	details, err := s.repo.GetDetailsByTransactionID(txID)
	if err != nil {
		return err
	}

	// 1. Recalculate JAM_AJAR items based on employee override > category rules
	seventy := decimal.NewFromInt(70)
	for i, detail := range details {
		if detail.Activity != nil && detail.Activity.Code == "JAM_AJAR" && emp.Category != nil {
			switch emp.Category.CalcMethod {
			case "PROPORTIONAL": // S2
				if !emp.Category.TargetIncentive.IsZero() {
					rate := emp.Category.TargetIncentive.DivRound(seventy, 4)
					detail.Rate = rate
					detail.TotalAmount = detail.Quantity.Mul(rate)
				}
			case "HOURLY": // BUK, KSU, Reguler (with hourly)
				// Employee override first, then category
				hourly := decimal.Zero
				if emp.HourlyRate != nil && !emp.HourlyRate.IsZero() {
					hourly = *emp.HourlyRate
				} else if !emp.Category.HourlyRate.IsZero() {
					hourly = emp.Category.HourlyRate
				}
				if !hourly.IsZero() {
					detail.Rate = hourly
					detail.TotalAmount = detail.Quantity.Mul(hourly)
				}
			}
			s.repo.UpdateDetail(&detail)
			details[i] = detail
		}
	}

	// 2. Sum up everything accurately
	gross := decimal.Zero
	deductions := decimal.Zero

	for _, detail := range details {
		if detail.Type == domain.TypeAddition {
			gross = gross.Add(detail.TotalAmount)
		} else if detail.Type == domain.TypeDeduction {
			deductions = deductions.Add(detail.TotalAmount)
		}
	}

	tx.GrossIncome = gross
	tx.TotalDeductions = deductions
	tx.TakeHomePay = gross.Sub(deductions)

	return s.repo.UpdateTransaction(tx)
}

func (s *payrollService) UpdateStatus(txID uuid.UUID, status domain.PayrollStatus) error {
	tx, err := s.repo.GetTransactionByID(txID)
	if err != nil {
		return err
	}

	// Can only transition forward (DRAFT -> CONFIRMED -> PAID)
	// For simplicity, just update it here
	tx.Status = status
	return s.repo.UpdateTransaction(tx)
}

func (s *payrollService) EnsurePayrollForEmployee(empID uuid.UUID, month, year int) error {
	// check if existing
	txs, err := s.repo.ListTransactions(month, year)
	if err == nil {
		for _, tx := range txs {
			if tx.EmployeeID == empID {
				return nil // already exists
			}
		}
	}
	// Create
	newTx := &domain.PayrollTransaction{
		EmployeeID: empID,
		Month:      month,
		Year:       year,
	}
	return s.CreateTransaction(newTx)
}
