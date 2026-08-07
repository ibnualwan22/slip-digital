package main

import (
	"fmt"
	"log"

	"github.com/ibnualwan/bisyaroh/internal/config"
)

func main() {
	cfg := config.LoadConfig()
	db := cfg.DB

	// Drop ALL unique constraints temporarily
	db.Exec(`ALTER TABLE payroll_transactions DROP CONSTRAINT IF EXISTS payroll_transactions_employee_id_month_year_key`)
	db.Exec(`ALTER TABLE payroll_transactions DROP CONSTRAINT IF EXISTS idx_emp_month_year`)
	db.Exec(`DROP INDEX IF EXISTS idx_emp_month_year`)
	db.Exec(`DROP INDEX IF EXISTS payroll_transactions_employee_id_month_year_key`)

	// Perform Shift
	err := db.Exec(`
		UPDATE payroll_transactions 
		SET 
			month = CASE WHEN month = 1 THEN 12 ELSE month - 1 END,
			year = CASE WHEN month = 1 THEN year - 1 ELSE year END
	`).Error

	if err != nil {
		log.Fatalf("FAILED UPDATE: %v", err)
	}

	// Restore one clean Unique Index
	err = db.Exec(`CREATE UNIQUE INDEX idx_emp_month_year ON payroll_transactions(employee_id, month, year)`).Error
	if err != nil {
		fmt.Printf("Warning restoring index: %v\n", err)
	}

	fmt.Println("SHIFT SUCCESSFUL!")
}
