package http

import (
	"crypto/subtle"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ibnualwan/bisyaroh/internal/config"
	"github.com/ibnualwan/bisyaroh/internal/domain"
	"github.com/ibnualwan/bisyaroh/internal/usecase"
	"github.com/ibnualwan/bisyaroh/pkg/response"
	"github.com/labstack/echo/v4"
)

// KBMHandler melayani: hitung ulang manual (admin), info asal data KBM
// (admin), dan webhook dari Siakad (publik, dijaga secret bersama).
// Webhook dan rekonsiliasi malam memakai mesin yang sama persis dengan
// hitung ulang manual.
type KBMHandler struct {
	svc        usecase.KBMService
	periodRepo domain.PeriodRepository
	logRepo    domain.KBMSyncLogRepository
	secret     string

	mu      sync.Mutex
	running map[string]bool
	pending map[string]bool
	lastRun map[string]time.Time
}

func NewKBMHandler(svc usecase.KBMService, periodRepo domain.PeriodRepository, logRepo domain.KBMSyncLogRepository, cfg *config.Config) *KBMHandler {
	return &KBMHandler{
		svc: svc, periodRepo: periodRepo, logRepo: logRepo, secret: cfg.WebhookSecret,
		running: make(map[string]bool), pending: make(map[string]bool), lastRun: make(map[string]time.Time),
	}
}

type kbmMonthRequest struct {
	Bulan int `json:"bulan"`
	Tahun int `json:"tahun"`
}

// Recompute (admin): hitung ulang satu bulan dari jendela dufahnya.
func (h *KBMHandler) Recompute(c echo.Context) error {
	var req kbmMonthRequest
	if err := c.Bind(&req); err != nil || req.Bulan < 1 || req.Bulan > 12 || req.Tahun < 2020 {
		return response.Error(c, http.StatusBadRequest, "bulan/tahun tidak valid")
	}
	result, err := h.svc.RecomputeMonth(req.Bulan, req.Tahun, "manual")
	if err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error())
	}
	return response.Success(c, http.StatusOK, "hitung ulang KBM selesai", result)
}

// Info (admin): asal data KBM satu bulan untuk label di halaman payroll.
func (h *KBMHandler) Info(c echo.Context) error {
	month, _ := strconv.Atoi(c.QueryParam("bulan"))
	year, _ := strconv.Atoi(c.QueryParam("tahun"))
	if month < 1 || month > 12 || year < 2020 {
		return response.Error(c, http.StatusBadRequest, "bulan/tahun tidak valid")
	}
	var periodView interface{}
	if p, err := h.periodRepo.GetByMonthYear(month, year); err == nil {
		periodView = map[string]interface{}{
			"dufah_name": p.DufahName,
			"start_date": p.StartDate.Format("2006-01-02"),
			"end_date":   p.EndDate.Format("2006-01-02"),
		}
	}
	var lastSync interface{}
	if log, err := h.logRepo.Latest(month, year); err == nil {
		lastSync = map[string]interface{}{
			"at":               log.CreatedAt.Format("2006-01-02 15:04:05"),
			"trigger":          log.Trigger,
			"teachers_updated": log.TeachersUpdated,
			"teachers_skipped": log.TeachersSkipped,
		}
	}
	return response.Success(c, http.StatusOK, "success", map[string]interface{}{
		"period":    periodView,
		"last_sync": lastSync,
	})
}

type siakadWebhookRequest struct {
	Tanggal    string `json:"tanggal"`
	PengajarID string `json:"pengajar_id"`
}

// Webhook (publik + secret): Siakad memberi tahu ada absensi berubah.
// Bulan yang dihitung ulang ditentukan dari pemetaan periode yang
// memuat tanggal perubahan — bukan dari pengirim.
func (h *KBMHandler) Webhook(c echo.Context) error {
	if h.secret == "" {
		return response.Error(c, http.StatusServiceUnavailable, "webhook belum dikonfigurasi")
	}
	got := c.Request().Header.Get("x-webhook-secret")
	if subtle.ConstantTimeCompare([]byte(got), []byte(h.secret)) != 1 {
		return response.Error(c, http.StatusUnauthorized, "secret tidak valid")
	}

	var req siakadWebhookRequest
	_ = c.Bind(&req)

	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc == nil {
		loc = time.Local
	}
	target := time.Now().In(loc)
	if req.Tanggal != "" {
		if parsed, err := time.Parse("2006-01-02", req.Tanggal); err == nil {
			target = parsed
		}
	}

	periods, err := h.periodRepo.List()
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "gagal membaca pemetaan periode")
	}
	for i := range periods {
		p := &periods[i]
		if !target.Before(p.StartDate) && !target.After(p.EndDate) {
			h.trigger(p.Month, p.Year, "webhook")
			return response.Success(c, http.StatusOK, "diterima; hitung ulang dijadwalkan", map[string]interface{}{
				"bulan": p.Month, "tahun": p.Year, "dufah": p.DufahName,
			})
		}
	}
	return response.Success(c, http.StatusOK, "diterima; tanggal di luar semua periode terpetakan", nil)
}

// trigger menjalankan hitung ulang dengan penggabungan: bila bulan yang
// sama sedang berjalan atau baru saja berjalan (<20 detik), cukup satu
// putaran lanjutan — webhook absensi bisa datang beruntun.
func (h *KBMHandler) trigger(month, year int, triggerName string) {
	key := strconv.Itoa(year) + "-" + strconv.Itoa(month)
	h.mu.Lock()
	if h.running[key] {
		h.pending[key] = true
		h.mu.Unlock()
		return
	}
	if time.Since(h.lastRun[key]) < 20*time.Second {
		if !h.pending[key] {
			h.pending[key] = true
			go func() {
				time.Sleep(20 * time.Second)
				h.mu.Lock()
				should := h.pending[key] && !h.running[key]
				if should {
					h.pending[key] = false
					h.running[key] = true
				}
				h.mu.Unlock()
				if should {
					h.runLoop(key, month, year, triggerName)
				}
			}()
		}
		h.mu.Unlock()
		return
	}
	h.running[key] = true
	h.mu.Unlock()
	go h.runLoop(key, month, year, triggerName)
}

func (h *KBMHandler) runLoop(key string, month, year int, triggerName string) {
	for {
		_, _ = h.svc.RecomputeMonth(month, year, triggerName)
		h.mu.Lock()
		h.lastRun[key] = time.Now()
		if h.pending[key] {
			h.pending[key] = false
			h.mu.Unlock()
			continue
		}
		h.running[key] = false
		h.mu.Unlock()
		return
	}
}

// StartReconciliation menjalankan rekonsiliasi otomatis tiap malam
// (~02:30 WIB): bulan yang jendelanya sedang berjalan atau baru berakhir
// (<= 2 hari) dihitung ulang penuh — jaminan bila ada webhook terlewat.
func (h *KBMHandler) StartReconciliation() {
	go func() {
		loc, _ := time.LoadLocation("Asia/Jakarta")
		if loc == nil {
			loc = time.Local
		}
		lastDate := ""
		for {
			time.Sleep(10 * time.Minute)
			now := time.Now().In(loc)
			today := now.Format("2006-01-02")
			if now.Hour() != 2 || now.Minute() < 30 || lastDate == today {
				continue
			}
			lastDate = today
			periods, err := h.periodRepo.List()
			if err != nil {
				continue
			}
			for i := range periods {
				p := &periods[i]
				endsRecently := !p.EndDate.Before(now.AddDate(0, 0, -2)) && p.EndDate.Before(now)
				running := !now.Before(p.StartDate) && !now.After(p.EndDate)
				if running || endsRecently {
					_, _ = h.svc.RecomputeMonth(p.Month, p.Year, "rekonsiliasi")
				}
			}
		}
	}()
}
