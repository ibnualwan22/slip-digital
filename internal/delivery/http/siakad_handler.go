package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ibnualwan/bisyaroh/internal/config"
	"github.com/ibnualwan/bisyaroh/internal/domain"
	"github.com/ibnualwan/bisyaroh/internal/usecase"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
)

type SiakadHandler struct {
	cfg        *config.Config
	empRepo    domain.EmployeeRepository
	actRepo    domain.ActivityRepository
	payService usecase.PayrollService
}

func NewSiakadHandler(cfg *config.Config, empRepo domain.EmployeeRepository, actRepo domain.ActivityRepository, payService usecase.PayrollService) *SiakadHandler {
	return &SiakadHandler{
		cfg:        cfg,
		empRepo:    empRepo,
		actRepo:    actRepo,
		payService: payService,
	}
}

// 1. Proxy GET Pengajar
func (h *SiakadHandler) GetPengajar(c echo.Context) error {
	req, err := http.NewRequest("GET", h.cfg.SiakadAPIURL, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	req.Header.Set("x-api-key", h.cfg.SiakadAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "Failed to connect to SIAKAD API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return echo.NewHTTPError(resp.StatusCode, "SIAKAD API returned status: "+fmt.Sprint(resp.StatusCode))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Parse SIAKAD Response
	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to parse SIAKAD response: "+err.Error())
	}

	// Fetch all local employees to map
	employees, err := h.empRepo.List(nil, true)
	if err != nil {
		employees = []domain.Employee{} // ignore error, just no mapping
	}

	siakadMap := make(map[string]domain.Employee)
	for _, emp := range employees {
		if emp.SiakadID != nil && *emp.SiakadID != "" {
			siakadMap[*emp.SiakadID] = emp
		}
	}

	// Enrich the data
	if result["data"] != nil {
		dataList := result["data"].([]interface{})
		for _, item := range dataList {
			pengajar := item.(map[string]interface{})
			idStr := fmt.Sprintf("%v", pengajar["id"])
			
			// Total denda
			terlambatMenit := 0
			if absenDetailRaw, ok := pengajar["absenDetail"]; ok && absenDetailRaw != nil {
				if absenList, isArr := absenDetailRaw.([]interface{}); isArr {
					for _, a := range absenList {
						absen, isMap := a.(map[string]interface{})
						if isMap {
							if tVal, has := absen["terlambatMenit"]; has && tVal != nil {
								terlambatMenit += int(tVal.(float64))
							}
						}
					}
				}
			}
			pengajar["totalTerlambatMenit"] = terlambatMenit
			pengajar["totalDenda"] = terlambatMenit * 1000

			// Local mapping
			if localEmp, ok := siakadMap[idStr]; ok {
				pengajar["localEmployeeId"] = localEmp.ID.String()
				pengajar["localEmployeeName"] = localEmp.Name
			} else {
				pengajar["localEmployeeId"] = nil
			}
		}
	}

	return c.JSON(http.StatusOK, result)
}

// 2. Proxy PUT Terlambat
func (h *SiakadHandler) UpdateTerlambat(c echo.Context) error {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	req, err := http.NewRequest("PUT", h.cfg.SiakadAPIURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	req.Header.Set("x-api-key", h.cfg.SiakadAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "Failed to connect to SIAKAD API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return echo.NewHTTPError(resp.StatusCode, "Failed to update SIAKAD API")
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return c.JSON(http.StatusOK, result)
}

// 3. Sync to Payroll
type SyncRequest struct {
	TotalJamMengajar   int `json:"total_jam_mengajar"`
	TotalTerlambatMenit int `json:"total_terlambat_menit"`
	Bulan              int `json:"bulan"` // current month if not provided
	Tahun              int `json:"tahun"` // current year if not provided
}

func (h *SiakadHandler) SyncToPayroll(c echo.Context) error {
	siakadId := c.Param("siakadId")
	if siakadId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Siakad ID required")
	}

	var req SyncRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	now := time.Now()
	bulan := req.Bulan
	if bulan == 0 {
		bulan = int(now.Month())
	}
	tahun := req.Tahun
	if tahun == 0 {
		tahun = now.Year()
	}

	// Cari employee yang mappingnya pass
	employees, err := h.empRepo.List(nil, true)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch employees")
	}

	var localEmp *domain.Employee
	for _, emp := range employees {
		if emp.SiakadID != nil && *emp.SiakadID == siakadId {
			// create a copy to avoid pointer reuse issues
			empCopy := emp
			localEmp = &empCopy
			break
		}
	}

	if localEmp == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Pengajar ini belum tertaut ke profil Asatidz di E-Rekap")
	}

	// Pastikan payroll bulan ini sudah digenerate
	err = h.payService.EnsurePayrollForEmployee(localEmp.ID, bulan, tahun)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to ensure payroll: "+err.Error())
	}

	txs, err := h.payService.ListTransactions(bulan, tahun)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list payrolls: "+err.Error())
	}

	var tx *domain.PayrollTransaction
	for _, t := range txs {
		if t.EmployeeID == localEmp.ID {
			tx = &t
			break
		}
	}

	if tx == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch created payroll")
	}

	if tx.Status != domain.StatusDraft {
		return echo.NewHTTPError(http.StatusBadRequest, "Slip gaji bulan ini sudah tidak dalam status DRAFT")
	}

	// 1. Sync Jam Mengajar
	if req.TotalJamMengajar > 0 {
		// Try to find addition activity named "Jam Mengajar" or similar
		activities, err := h.actRepo.List(true)
		var mengajarAct *domain.MasterActivity
		for _, a := range activities {
			if a.Type == "ADDITION" && (a.Code == "MENGAJAR" || a.Code == "JAM_MENGAJAR") {
				mengajarAct = &a
				break
			}
		}

		// determine rate (hourly rate overrides category, then default activity rate)
		rate := decimal.Zero
		if localEmp.HourlyRate != nil && localEmp.HourlyRate.GreaterThan(decimal.Zero) {
			rate = *localEmp.HourlyRate
		} else if localEmp.Category != nil && localEmp.Category.HourlyRate.GreaterThan(decimal.Zero) {
			rate = localEmp.Category.HourlyRate
		} else if mengajarAct != nil {
			rate = mengajarAct.DefaultRate
		}

		if mengajarAct == nil {
			newAct := domain.MasterActivity{
				ID:           uuid.New(),
				ActivityName: "Jam Mengajar",
				Code:         "JAM_MENGAJAR",
				Type:         domain.TypeAddition,
				DefaultRate:  decimal.NewFromInt(0),
				IsActive:     true,
				CreatedAt:    time.Now(),
			}
			h.actRepo.Create(&newAct)
			mengajarAct = &newAct
		}

		if mengajarAct != nil {
			err = h.payService.UpsertDetail(&domain.PayrollDetail{
				ID: uuid.New(),
				PayrollTransactionID: tx.ID,
				ActivityID: &mengajarAct.ID,
				Quantity: decimal.NewFromFloat(float64(req.TotalJamMengajar)),
				Rate: rate,
				Type: domain.TypeAddition,
				Description: "Jam Mengajar",
			})
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Gagal sinkronisasi jam mengajar: "+err.Error())
			}
		}
	}

	// 2. Sync Denda Terlambat
	if req.TotalTerlambatMenit > 0 {
		activities, _ := h.actRepo.List(true)
		var dendaAct *domain.MasterActivity
		for _, a := range activities {
			if a.Type == "DEDUCTION" && (a.Code == "DENDA" || a.Code == "TERLAMBAT") {
				dendaAct = &a
				break
			}
		}
		
		if dendaAct == nil {
			newAct := domain.MasterActivity{
				ID:           uuid.New(),
				ActivityName: "Denda Terlambat",
				Code:         "TERLAMBAT",
				Type:         domain.TypeDeduction,
				DefaultRate:  decimal.NewFromInt(1000),
				IsActive:     true,
				CreatedAt:    time.Now(),
			}
			h.actRepo.Create(&newAct)
			dendaAct = &newAct
		}

		if dendaAct != nil {
			dendaRate := decimal.NewFromInt(1000) // Rp 1000 per menit
			err = h.payService.UpsertDetail(&domain.PayrollDetail{
				ID: uuid.New(),
				PayrollTransactionID: tx.ID,
				ActivityID: &dendaAct.ID,
				Quantity: decimal.NewFromFloat(float64(req.TotalTerlambatMenit)),
				Rate: dendaRate,
				Type: domain.TypeDeduction,
				Description: fmt.Sprintf("Terlambat %d mnt", req.TotalTerlambatMenit),
			})
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Gagal sinkronisasi denda: "+err.Error())
			}
		} else {
			return echo.NewHTTPError(http.StatusBadRequest, "Master aktivitas POTONGAN untuk Denda belum dikonfigurasi. Buat Master Aktivitas baru bertipe Potongan.")
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Data berhasil disinkronisasi ke payroll",
	})
}

// 4. Bulk Sync
func (h *SiakadHandler) SyncAllToPayroll(c echo.Context) error {
	var req struct {
		Bulan int `json:"bulan"`
		Tahun int `json:"tahun"`
	}
	_ = c.Bind(&req)

	now := time.Now()
	bulan := req.Bulan
	if bulan == 0 {
		bulan = int(now.Month())
	}
	tahun := req.Tahun
	if tahun == 0 {
		tahun = now.Year()
	}

	// 1. Fetch Siakad Data
	httpReq, err := http.NewRequest("GET", h.cfg.SiakadAPIURL, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	httpReq.Header.Set("x-api-key", h.cfg.SiakadAPIKey)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil || resp.StatusCode != http.StatusOK {
		return echo.NewHTTPError(http.StatusBadGateway, "Failed to connect to SIAKAD API")
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	employees, err := h.empRepo.List(nil, true)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch employees")
	}

	siakadMap := make(map[string]domain.Employee)
	for _, emp := range employees {
		if emp.SiakadID != nil && *emp.SiakadID != "" {
			siakadMap[*emp.SiakadID] = emp
		}
	}

	activities, _ := h.actRepo.List(true)
	var mengajarAct *domain.MasterActivity
	var dendaAct *domain.MasterActivity
	for _, a := range activities {
		if a.Type == "ADDITION" && (a.Code == "MENGAJAR" || a.Code == "JAM_MENGAJAR") && mengajarAct == nil {
			mengajarAct = &a
		}
		if a.Type == "DEDUCTION" && (a.Code == "DENDA" || a.Code == "TERLAMBAT") && dendaAct == nil {
			dendaAct = &a
		}
	}
	if mengajarAct == nil {
		newAct := domain.MasterActivity{
			ID:           uuid.New(),
			ActivityName: "Jam Mengajar",
			Code:         "JAM_MENGAJAR",
			Type:         domain.TypeAddition,
			DefaultRate:  decimal.NewFromInt(0),
			IsActive:     true,
			CreatedAt:    time.Now(),
		}
		h.actRepo.Create(&newAct)
		mengajarAct = &newAct
	}

	if dendaAct == nil {
		newAct := domain.MasterActivity{
			ID:           uuid.New(),
			ActivityName: "Denda Terlambat",
			Code:         "TERLAMBAT",
			Type:         domain.TypeDeduction,
			DefaultRate:  decimal.NewFromInt(1000),
			IsActive:     true,
			CreatedAt:    time.Now(),
		}
		h.actRepo.Create(&newAct)
		dendaAct = &newAct
	}
	
	sycnedCount := 0

	if dataList, ok := result["data"].([]interface{}); ok {
		for _, item := range dataList {
			pengajar := item.(map[string]interface{})
			idStr := fmt.Sprintf("%v", pengajar["id"])
			
			localEmp, found := siakadMap[idStr]
			if !found {
				continue
			}

			// Ensure Draft Payroll
			err := h.payService.EnsurePayrollForEmployee(localEmp.ID, bulan, tahun)
			if err != nil {
				continue
			}
			
			txs, _ := h.payService.ListTransactions(bulan, tahun)
			var tx *domain.PayrollTransaction
			for _, t := range txs {
				if t.EmployeeID == localEmp.ID && t.Status == domain.StatusDraft {
					tx = &t
					break
				}
			}
			if tx == nil {
				continue
			}

			// extract totals
			totalJam := 0
			if val, ok := pengajar["jumlahAbsenTerverifikasi"]; ok && val != nil {
				totalJam = int(val.(float64))
			}
			totalTerlambat := 0
			if absenDetailRaw, ok := pengajar["absenDetail"]; ok && absenDetailRaw != nil {
				if absenList, isArr := absenDetailRaw.([]interface{}); isArr {
					for _, a := range absenList {
						absen, isMap := a.(map[string]interface{})
						if isMap {
							if tVal, has := absen["terlambatMenit"]; has && tVal != nil {
								totalTerlambat += int(tVal.(float64))
							}
						}
					}
				}
			}

			// Sync Jam Mengajar
			if totalJam > 0 {
				rate := decimal.Zero
				if localEmp.HourlyRate != nil && localEmp.HourlyRate.GreaterThan(decimal.Zero) {
					rate = *localEmp.HourlyRate
				} else if localEmp.Category != nil && localEmp.Category.HourlyRate.GreaterThan(decimal.Zero) {
					rate = localEmp.Category.HourlyRate
				} else {
					rate = mengajarAct.DefaultRate
				}

				h.payService.UpsertDetail(&domain.PayrollDetail{
					ID: uuid.New(),
					PayrollTransactionID: tx.ID,
					ActivityID: &mengajarAct.ID,
					Quantity: decimal.NewFromFloat(float64(totalJam)),
					Rate: rate,
					Type: domain.TypeAddition,
					Description: "Jam Mengajar",
				})
			}

			// Sync Terlambat
			if totalTerlambat > 0 {
				dendaRate := decimal.NewFromInt(1000)
				h.payService.UpsertDetail(&domain.PayrollDetail{
					ID: uuid.New(),
					PayrollTransactionID: tx.ID,
					ActivityID: &dendaAct.ID,
					Quantity: decimal.NewFromFloat(float64(totalTerlambat)),
					Rate: dendaRate,
					Type: domain.TypeDeduction,
					Description: fmt.Sprintf("Terlambat %d mnt", totalTerlambat),
				})
			}

			sycnedCount++
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Berhasil mensinkronisasi data untuk %d pengajar", sycnedCount),
	})
}
