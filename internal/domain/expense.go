package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ExpenseReport adalah laporan pengeluaran per bulan
type ExpenseReport struct {
	ID                uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	Month             int             `gorm:"not null;uniqueIndex:idx_expense_month_year" json:"month"`
	Year              int             `gorm:"not null;uniqueIndex:idx_expense_month_year" json:"year"`
	Keterangan        string          `gorm:"type:text" json:"keterangan"`
	Status            string          `gorm:"type:varchar(20);default:'DRAFT'" json:"status"` // DRAFT, CONFIRMED
	TotalHargaBarang  decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"total_harga_barang"`
	TotalPengeluaran  decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"total_pengeluaran"`
	TotalKembalian    decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"total_kembalian"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`

	Items []ExpenseItem `gorm:"foreignKey:ExpenseReportID;constraint:OnDelete:CASCADE" json:"items"`
}

// ExpenseItem adalah baris item pengeluaran dalam laporan
type ExpenseItem struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	ExpenseReportID uuid.UUID       `gorm:"type:uuid;not null" json:"expense_report_id"`
	NamaBarang      string          `gorm:"type:varchar(255);not null" json:"nama_barang"`
	HargaSatuan     decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"harga_satuan"`
	Jumlah          int             `gorm:"not null;default:1" json:"jumlah"`
	TotalHarga      decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"total_harga"`
	Kredit          decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"kredit"`
	Debit           decimal.Decimal `gorm:"type:numeric(19,4);default:0" json:"debit"`
	BuktiPembayaran string          `gorm:"type:text" json:"bukti_pembayaran"`
	GroupID         string          `gorm:"type:varchar(50);index" json:"group_id"` // used to group items from the same receipt
	IsKonfirmasi    bool            `gorm:"default:false" json:"is_konfirmasi"`
	IsTashih        bool            `gorm:"default:false" json:"is_tashih"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// ExpenseRepository adalah interface untuk operasi data pengeluaran
type ExpenseRepository interface {
	// Report
	CreateReport(report *ExpenseReport) error
	GetReportByID(id uuid.UUID) (*ExpenseReport, error)
	ListReports(month, year int) ([]ExpenseReport, error)
	UpdateReport(report *ExpenseReport) error
	DeleteReport(id uuid.UUID) error

	// Items
	CreateItem(item *ExpenseItem) error
	GetItemByID(id uuid.UUID) (*ExpenseItem, error)
	GetItemsByReportID(reportID uuid.UUID) ([]ExpenseItem, error)
	UpdateItem(item *ExpenseItem) error
	DeleteItem(id uuid.UUID) error
}
