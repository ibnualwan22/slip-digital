package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ibnualwan/bisyaroh/internal/config"
	"github.com/ibnualwan/bisyaroh/internal/domain"
	"github.com/ibnualwan/bisyaroh/pkg/response"
	"github.com/labstack/echo/v4"
)

// MeHandler melayani halaman "Rekap Saya" milik pengajar.
// Identitas (employee_id, siakad_id) SELALU diambil dari klaim token —
// tidak pernah dari parameter klien — sehingga pengajar hanya dapat
// melihat datanya sendiri.
type MeHandler struct {
	periodRepo domain.PeriodRepository
	payRepo    domain.PayrollRepository
	cfg        *config.Config
}

func NewMeHandler(periodRepo domain.PeriodRepository, payRepo domain.PayrollRepository, cfg *config.Config) *MeHandler {
	return &MeHandler{periodRepo: periodRepo, payRepo: payRepo, cfg: cfg}
}

func (h *MeHandler) Profile(c echo.Context) error {
	return response.Success(c, http.StatusOK, "success", map[string]interface{}{
		"name":        c.Get("authName"),
		"username":    c.Get("authUsername"),
		"role":        "pengajar",
		"employee_id": c.Get("authEmployeeID"),
	})
}

func (h *MeHandler) Periods(c echo.Context) error {
	periods, err := h.periodRepo.List()
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list periods")
	}
	views := make([]periodView, 0, len(periods))
	for i := range periods {
		views = append(views, toPeriodView(&periods[i]))
	}
	return response.Success(c, http.StatusOK, "success", views)
}

type siakadPengajarResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		ID          string                   `json:"id"`
		Nama        string                   `json:"nama"`
		AbsenDetail []map[string]interface{} `json:"absenDetail"`
	} `json:"data"`
}

// fetchOwnAbsen mengambil rincian absen pengajar ini dari Siakad untuk
// rentang tanggal periode (memakai endpoint eksternal yang sama dengan
// halaman admin Data SIAKAD, ditambah parameter from/to).
func (h *MeHandler) fetchOwnAbsen(siakadID string, period *domain.PeriodMapping) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s?from=%s&to=%s",
		h.cfg.SiakadAPIURL,
		period.StartDate.Format("2006-01-02"),
		period.EndDate.Format("2006-01-02"))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", h.cfg.SiakadAPIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tidak dapat menghubungi Siakad: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Siakad membalas status %d", resp.StatusCode)
	}

	var result siakadPengajarResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("respons Siakad tidak valid")
	}
	for _, p := range result.Data {
		if p.ID == siakadID {
			if p.AbsenDetail == nil {
				return []map[string]interface{}{}, nil
			}
			return p.AbsenDetail, nil
		}
	}
	return []map[string]interface{}{}, nil
}

func (h *MeHandler) Recap(c echo.Context) error {
	periodID, err := uuid.Parse(c.QueryParam("period_id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "period_id tidak valid")
	}
	period, err := h.periodRepo.GetByID(periodID)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "periode tidak ditemukan")
	}

	siakadID, _ := c.Get("authSiakadID").(string)
	employeeID, _ := c.Get("authEmployeeID").(string)

	sessions, err := h.fetchOwnAbsen(siakadID, period)
	if err != nil {
		return response.Error(c, http.StatusBadGateway, "Gagal mengambil data kehadiran dari Siakad")
	}

	totalTerlambat := 0
	for _, s := range sessions {
		if v, ok := s["terlambatMenit"].(float64); ok {
			totalTerlambat += int(v)
		}
	}

	// Payroll bulan yang terhubung lewat pemetaan periode ini.
	// Hanya transaksi milik pengajar yang login (klaim employee_id).
	findTx := func(month, year int) *domain.PayrollTransaction {
		txs, err := h.payRepo.ListTransactions(month, year)
		if err != nil {
			return nil
		}
		for i := range txs {
			if txs[i].EmployeeID.String() == employeeID {
				return &txs[i]
			}
		}
		return nil
	}

	var payrollTx *domain.PayrollTransaction
	if tx := findTx(period.Month, period.Year); tx != nil {
		payrollTx = tx
	} else {
		// Cadangan (arahan user): bila payroll bulan pemetaan belum dibuat,
		// tampilkan payroll bulan berjalannya periode (bulan mulai, lalu
		// bulan selesai) — walau statusnya masih DRAFT/sedang diproses.
		type monthYear struct{ m, y int }
		seen := map[monthYear]bool{{period.Month, period.Year}: true}
		candidates := []monthYear{
			{int(period.StartDate.Month()), period.StartDate.Year()},
			{int(period.EndDate.Month()), period.EndDate.Year()},
		}
		for _, cand := range candidates {
			if seen[cand] {
				continue
			}
			seen[cand] = true
			if tx := findTx(cand.m, cand.y); tx != nil {
				payrollTx = tx
				break
			}
		}
	}

	// Rincian payroll: HANYA komponen yang benar-benar ada pada transaksi
	// ini (tidak ada baris template). Dipakai halaman Rincian Bisyaroh.
	var payrollView interface{}
	if payrollTx != nil {
		detailViews := []map[string]interface{}{}
		if details, err := h.payRepo.GetDetailsByTransactionID(payrollTx.ID); err == nil {
			for _, d := range details {
				name := d.Description
				code := ""
				if d.Activity != nil {
					code = d.Activity.Code
					if name == "" {
						name = d.Activity.ActivityName
					}
				}
				if name == "" {
					name = "Item"
				}
				detailViews = append(detailViews, map[string]interface{}{
					"activity_name": name,
					"code":          code,
					"quantity":      d.Quantity.String(),
					"rate":          d.Rate.String(),
					"total_amount":  d.TotalAmount.String(),
					"type":          string(d.Type),
					"source":        d.Source,
				})
			}
		}
		payrollView = map[string]interface{}{
			"id":               payrollTx.ID.String(),
			"month":            payrollTx.Month,
			"year":             payrollTx.Year,
			"gross_income":     payrollTx.GrossIncome.String(),
			"total_deductions": payrollTx.TotalDeductions.String(),
			"take_home_pay":    payrollTx.TakeHomePay.String(),
			"status":           string(payrollTx.Status),
			"details":          detailViews,
		}
	}

	return response.Success(c, http.StatusOK, "success", map[string]interface{}{
		"period":   toPeriodView(period),
		"sessions": sessions,
		"summary": map[string]interface{}{
			"total_sesi":           len(sessions),
			"total_terlambat_menit": totalTerlambat,
			"total_denda":           totalTerlambat * 1000,
		},
		"payroll": payrollView,
	})
}
