package usecase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ibnualwan/bisyaroh/internal/config"
	"github.com/ibnualwan/bisyaroh/internal/domain"
	"github.com/shopspring/decimal"
)

// KBMService menghitung ulang baris "Jam Mengajar" dan "Denda Terlambat"
// pada payroll dari absensi Siakad, SELALU berdasarkan jendela dufah pada
// pemetaan periode bulan tersebut — tidak pernah dari "periode aktif"
// Siakad. Inilah mesin di balik webhook, rekonsiliasi malam, tombol
// sinkron, dan perbaikan data; semua jalur memakai perhitungan yang sama.
type KBMService interface {
	RecomputeMonth(month, year int, trigger string) (*KBMMonthResult, error)
	RecomputeTeacher(siakadID string, month, year int, trigger string) error
}

type KBMChange struct {
	Name        string `json:"name"`
	JamBefore   int    `json:"jam_before"`
	JamAfter    int    `json:"jam_after"`
	TelatBefore int    `json:"telat_before"`
	TelatAfter  int    `json:"telat_after"`
}

type KBMMonthResult struct {
	Month           int         `json:"month"`
	Year            int         `json:"year"`
	DufahName       string      `json:"dufah_name"`
	WindowStart     string      `json:"window_start"`
	WindowEnd       string      `json:"window_end"`
	TeachersUpdated int         `json:"teachers_updated"`
	TeachersSkipped int         `json:"teachers_skipped"`
	SkippedNames    []string    `json:"skipped_names,omitempty"`
	Changes         []KBMChange `json:"changes"`
}

type kbmService struct {
	cfg        *config.Config
	empRepo    domain.EmployeeRepository
	periodRepo domain.PeriodRepository
	payRepo    domain.PayrollRepository
	actRepo    domain.ActivityRepository
	logRepo    domain.KBMSyncLogRepository
	payService PayrollService
}

func NewKBMService(cfg *config.Config, empRepo domain.EmployeeRepository, periodRepo domain.PeriodRepository, payRepo domain.PayrollRepository, actRepo domain.ActivityRepository, logRepo domain.KBMSyncLogRepository, payService PayrollService) KBMService {
	return &kbmService{cfg: cfg, empRepo: empRepo, periodRepo: periodRepo, payRepo: payRepo, actRepo: actRepo, logRepo: logRepo, payService: payService}
}

type siakadWindowData struct {
	ID          string `json:"id"`
	Nama        string `json:"nama"`
	AbsenDetail []struct {
		TerlambatMenit int `json:"terlambatMenit"`
	} `json:"absenDetail"`
}

func (s *kbmService) fetchWindow(period *domain.PeriodMapping) (map[string][2]int, error) {
	url := fmt.Sprintf("%s?from=%s&to=%s", s.cfg.SiakadAPIURL,
		period.StartDate.Format("2006-01-02"), period.EndDate.Format("2006-01-02"))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", s.cfg.SiakadAPIKey)
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tidak dapat menghubungi Siakad: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Siakad membalas status %d", resp.StatusCode)
	}
	var result struct {
		Data []siakadWindowData `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("respons Siakad tidak valid")
	}
	out := make(map[string][2]int)
	for _, p := range result.Data {
		telat := 0
		for _, a := range p.AbsenDetail {
			telat += a.TerlambatMenit
		}
		out[p.ID] = [2]int{len(p.AbsenDetail), telat}
	}
	return out, nil
}

// kbmActivities menyelesaikan aktivitas kanonis KBM. Sejak penyatuan,
// hanya ada SATU aktivitas jam: JAM_MENGAJAR. Baris manual dan baris
// Siakad dibedakan oleh kolom Source pada detail (bukan oleh aktivitas),
// sehingga mesin hanya boleh menghapus baris Source="siakad" (baris lama
// tanpa penanda diperlakukan sebagai siakad demi kompatibilitas).
func (s *kbmService) kbmActivities() (jamAct *domain.MasterActivity, jamIDs map[uuid.UUID]bool, dendaAct *domain.MasterActivity, dendaIDs map[uuid.UUID]bool, err error) {
	activities, err := s.actRepo.List(true)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	jamIDs = make(map[uuid.UUID]bool)
	dendaIDs = make(map[uuid.UUID]bool)
	for i := range activities {
		a := &activities[i]
		if a.Type == domain.TypeAddition && a.Code == "JAM_MENGAJAR" {
			jamIDs[a.ID] = true
			jamAct = a
		}
		if a.Type == domain.TypeDeduction && (a.Code == "TERLAMBAT" || a.Code == "DENDA") {
			dendaIDs[a.ID] = true
			if a.Code == "TERLAMBAT" {
				dendaAct = a
			}
		}
	}
	if jamAct == nil {
		jamAct = &domain.MasterActivity{ID: uuid.New(), ActivityName: "Jam Mengajar", Code: "JAM_MENGAJAR", Type: domain.TypeAddition, DefaultRate: decimal.Zero, IsActive: true, CreatedAt: time.Now()}
		if err := s.actRepo.Create(jamAct); err != nil {
			return nil, nil, nil, nil, err
		}
		jamIDs[jamAct.ID] = true
	}
	if dendaAct == nil {
		dendaAct = &domain.MasterActivity{ID: uuid.New(), ActivityName: "Denda Terlambat", Code: "TERLAMBAT", Type: domain.TypeDeduction, DefaultRate: decimal.NewFromInt(1000), IsActive: true, CreatedAt: time.Now()}
		if err := s.actRepo.Create(dendaAct); err != nil {
			return nil, nil, nil, nil, err
		}
		dendaIDs[dendaAct.ID] = true
	}
	return jamAct, jamIDs, dendaAct, dendaIDs, nil
}

func jamRateFor(emp *domain.Employee, jamAct *domain.MasterActivity) decimal.Decimal {
	if emp.HourlyRate != nil && emp.HourlyRate.GreaterThan(decimal.Zero) {
		return *emp.HourlyRate
	}
	if emp.Category != nil && emp.Category.HourlyRate.GreaterThan(decimal.Zero) {
		return emp.Category.HourlyRate
	}
	return jamAct.DefaultRate
}

// applyToTx menimpa baris KBM satu transaksi dengan nilai baru.
// Baris lama selalu dihapus dulu — termasuk saat nilai baru 0 — agar
// tidak ada sisa basi (perilaku lama membiarkan denda basi tertinggal).
func (s *kbmService) applyToTx(tx *domain.PayrollTransaction, emp *domain.Employee, jamVal, telatVal int, jamAct *domain.MasterActivity, jamIDs map[uuid.UUID]bool, dendaAct *domain.MasterActivity, dendaIDs map[uuid.UUID]bool) (*KBMChange, error) {
	details, err := s.payRepo.GetDetailsByTransactionID(tx.ID)
	if err != nil {
		return nil, err
	}
	change := &KBMChange{Name: emp.Name, JamAfter: jamVal, TelatAfter: telatVal}
	for _, d := range details {
		if d.ActivityID == nil {
			continue
		}
		if jamIDs[*d.ActivityID] && (d.Source == "siakad" || d.Source == "") {
			change.JamBefore += int(d.Quantity.IntPart())
			if err := s.payService.RemoveDetail(d.ID); err != nil {
				return nil, err
			}
		}
		if dendaIDs[*d.ActivityID] && (d.Source == "siakad" || d.Source == "") {
			change.TelatBefore += int(d.Quantity.IntPart())
			if err := s.payService.RemoveDetail(d.ID); err != nil {
				return nil, err
			}
		}
	}
	if jamVal > 0 {
		err := s.payService.AddDetail(&domain.PayrollDetail{
			ID: uuid.New(), PayrollTransactionID: tx.ID, ActivityID: &jamAct.ID,
			Quantity: decimal.NewFromInt(int64(jamVal)), Rate: jamRateFor(emp, jamAct),
			Type: domain.TypeAddition, Description: "Jam Mengajar", Source: "siakad",
		})
		if err != nil {
			return nil, err
		}
	}
	if telatVal > 0 {
		err := s.payService.AddDetail(&domain.PayrollDetail{
			ID: uuid.New(), PayrollTransactionID: tx.ID, ActivityID: &dendaAct.ID,
			Quantity: decimal.NewFromInt(int64(telatVal)), Rate: decimal.NewFromInt(1000),
			Type: domain.TypeDeduction, Description: fmt.Sprintf("Terlambat %d mnt", telatVal), Source: "siakad",
		})
		if err != nil {
			return nil, err
		}
	}
	if err := s.payService.CalculateTransactionTHP(tx.ID); err != nil {
		return nil, err
	}
	if change.JamBefore == change.JamAfter && change.TelatBefore == change.TelatAfter {
		return nil, nil
	}
	return change, nil
}

func (s *kbmService) RecomputeMonth(month, year int, trigger string) (*KBMMonthResult, error) {
	period, err := s.periodRepo.GetByMonthYear(month, year)
	if err != nil {
		return nil, fmt.Errorf("bulan %d/%d belum terhubung ke dufah pada pemetaan periode", month, year)
	}
	window, err := s.fetchWindow(period)
	if err != nil {
		return nil, err
	}
	jamAct, jamIDs, dendaAct, dendaIDs, err := s.kbmActivities()
	if err != nil {
		return nil, err
	}
	employees, err := s.empRepo.List(nil, false)
	if err != nil {
		return nil, err
	}
	empBySiakad := make(map[string]*domain.Employee)
	for i := range employees {
		if employees[i].SiakadID != nil && *employees[i].SiakadID != "" {
			empBySiakad[*employees[i].SiakadID] = &employees[i]
		}
	}
	txs, err := s.payService.ListTransactions(month, year)
	if err != nil {
		return nil, err
	}
	txByEmp := make(map[uuid.UUID]*domain.PayrollTransaction)
	for i := range txs {
		txByEmp[txs[i].EmployeeID] = &txs[i]
	}

	result := &KBMMonthResult{
		Month: month, Year: year, DufahName: period.DufahName,
		WindowStart: period.StartDate.Format("2006-01-02"), WindowEnd: period.EndDate.Format("2006-01-02"),
		Changes: []KBMChange{},
	}

	// Sasaran: semua pengajar tertaut Siakad yang muncul di jendela,
	// ditambah semua pemilik transaksi bulan ini yang tertaut.
	targets := make(map[string]*domain.Employee)
	for sid, emp := range empBySiakad {
		if _, ok := window[sid]; ok {
			targets[sid] = emp
		}
	}
	for empID, tx := range txByEmp {
		_ = tx
		for sid, emp := range empBySiakad {
			if emp.ID == empID {
				targets[sid] = emp
			}
		}
	}

	for sid, emp := range targets {
		vals := window[sid] // [jam, telat]; nol bila tidak ada di jendela
		tx := txByEmp[emp.ID]
		if tx == nil {
			if err := s.payService.EnsurePayrollForEmployee(emp.ID, month, year); err != nil {
				result.TeachersSkipped++
				result.SkippedNames = append(result.SkippedNames, emp.Name+" (gagal membuat payroll)")
				continue
			}
			fresh, err := s.payService.ListTransactions(month, year)
			if err != nil {
				return nil, err
			}
			for i := range fresh {
				if fresh[i].EmployeeID == emp.ID {
					tx = &fresh[i]
					break
				}
			}
			if tx == nil {
				result.TeachersSkipped++
				result.SkippedNames = append(result.SkippedNames, emp.Name+" (payroll tidak ditemukan)")
				continue
			}
		}
		if tx.Status != domain.StatusDraft {
			result.TeachersSkipped++
			result.SkippedNames = append(result.SkippedNames, fmt.Sprintf("%s (%s)", emp.Name, tx.Status))
			continue
		}
		change, err := s.applyToTx(tx, emp, vals[0], vals[1], jamAct, jamIDs, dendaAct, dendaIDs)
		if err != nil {
			return nil, fmt.Errorf("gagal memperbarui %s: %w", emp.Name, err)
		}
		if change != nil {
			result.Changes = append(result.Changes, *change)
		}
		result.TeachersUpdated++
	}

	changesJSON, _ := json.Marshal(result.Changes)
	_ = s.logRepo.Create(&domain.KBMSyncLog{
		Month: month, Year: year, DufahName: period.DufahName, Trigger: trigger,
		TeachersUpdated: result.TeachersUpdated, TeachersSkipped: result.TeachersSkipped,
		ChangesJSON: string(changesJSON),
	})
	return result, nil
}

func (s *kbmService) RecomputeTeacher(siakadID string, month, year int, trigger string) error {
	_, err := s.RecomputeMonth(month, year, trigger)
	return err
}
