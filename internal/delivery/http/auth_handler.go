package http

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ibnualwan/bisyaroh/internal/config"
	"github.com/ibnualwan/bisyaroh/internal/domain"
	"github.com/ibnualwan/bisyaroh/pkg/response"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler menangani dua jalur login:
//  1. Admin — kredensial tunggal dari environment (ADMIN_USERNAME +
//     ADMIN_PASSWORD_HASH bcrypt).
//  2. Pengajar — kredensial diverifikasi ke Siakad lewat endpoint internal
//     /api/external/auth/verify (dijaga API key), lalu identitas Siakad
//     dicocokkan ke data karyawan lewat employees.siakad_id.
//     E-Maliyah TIDAK menyimpan password pengajar.
//
// Nanti jalur (2) dapat diganti SSO Bawwabah di titik yang sama
// (verifySiakad) tanpa mengubah token maupun halaman.
type AuthHandler struct {
	adminUsername string
	adminHash     string
	jwtSecret     string
	verifyURL     string
	siakadAPIKey  string
	empRepo       domain.EmployeeRepository

	mu       sync.Mutex
	attempts map[string][]time.Time
}

func NewAuthHandler(cfg *config.Config, empRepo domain.EmployeeRepository) *AuthHandler {
	verifyURL := ""
	if cfg.SiakadAPIURL != "" {
		verifyURL = strings.TrimSuffix(cfg.SiakadAPIURL, "/pengajar") + "/auth/verify"
	}
	return &AuthHandler{
		adminUsername: cfg.AdminUsername,
		adminHash:     cfg.AdminPasswordHash,
		jwtSecret:     cfg.JWTSecret,
		verifyURL:     verifyURL,
		siakadAPIKey:  cfg.SiakadAPIKey,
		empRepo:       empRepo,
		attempts:      make(map[string][]time.Time),
	}
}

const (
	loginMaxAttempts = 5
	loginWindow      = 5 * time.Minute
	tokenLifetime    = 12 * time.Hour
)

func (h *AuthHandler) rateLimited(ip string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	now := time.Now()
	recent := h.attempts[ip][:0]
	for _, t := range h.attempts[ip] {
		if now.Sub(t) < loginWindow {
			recent = append(recent, t)
		}
	}
	h.attempts[ip] = recent
	return len(recent) >= loginMaxAttempts
}

func (h *AuthHandler) recordFailure(ip string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.attempts[ip] = append(h.attempts[ip], time.Now())
}

func (h *AuthHandler) clearFailures(ip string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.attempts, ip)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) issueToken(claims jwt.MapClaims) (string, error) {
	now := time.Now()
	claims["iat"] = now.Unix()
	claims["exp"] = now.Add(tokenLifetime).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

type siakadIdentity struct {
	ID       string `json:"id"`
	Nama     string `json:"nama"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// verifySiakad memeriksa kredensial ke endpoint internal Siakad.
func (h *AuthHandler) verifySiakad(username, password string) (*siakadIdentity, error) {
	if h.verifyURL == "" {
		return nil, fmt.Errorf("integrasi Siakad belum dikonfigurasi")
	}
	payload, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req, err := http.NewRequest(http.MethodPost, h.verifyURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", h.siakadAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tidak dapat menghubungi Siakad: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kredensial ditolak Siakad (status %d)", resp.StatusCode)
	}

	var result struct {
		Success bool           `json:"success"`
		Data    siakadIdentity `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || !result.Success {
		return nil, fmt.Errorf("respons Siakad tidak valid")
	}
	return &result.Data, nil
}

func (h *AuthHandler) findEmployeeBySiakadID(siakadID string) (*domain.Employee, error) {
	employees, err := h.empRepo.List(nil, false)
	if err != nil {
		return nil, err
	}
	for i := range employees {
		if employees[i].SiakadID != nil && *employees[i].SiakadID == siakadID {
			return &employees[i], nil
		}
	}
	return nil, nil
}

func (h *AuthHandler) Login(c echo.Context) error {
	ip := c.RealIP()
	if h.rateLimited(ip) {
		return response.Error(c, http.StatusTooManyRequests, "Terlalu banyak percobaan login. Coba lagi dalam beberapa menit.")
	}

	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Format permintaan tidak valid")
	}

	// Jalur 1: admin (kredensial environment)
	if h.adminUsername != "" && subtle.ConstantTimeCompare([]byte(req.Username), []byte(h.adminUsername)) == 1 {
		if h.adminHash != "" && bcrypt.CompareHashAndPassword([]byte(h.adminHash), []byte(req.Password)) == nil {
			h.clearFailures(ip)
			token, err := h.issueToken(jwt.MapClaims{
				"sub":  req.Username,
				"role": "admin",
				"name": "Administrator",
			})
			if err != nil {
				return response.Error(c, http.StatusInternalServerError, "Gagal membuat token")
			}
			return response.Success(c, http.StatusOK, "Login berhasil", map[string]interface{}{
				"token": token, "username": req.Username, "role": "admin", "name": "Administrator",
			})
		}
		h.recordFailure(ip)
		return response.Error(c, http.StatusUnauthorized, "Username atau password salah")
	}

	// Jalur 2: pengajar (verifikasi ke Siakad)
	identity, err := h.verifySiakad(req.Username, req.Password)
	if err != nil {
		h.recordFailure(ip)
		return response.Error(c, http.StatusUnauthorized, "Username atau password salah")
	}

	emp, err := h.findEmployeeBySiakadID(identity.ID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Gagal memeriksa data karyawan")
	}
	if emp == nil {
		// Kredensial Siakad sah, tetapi belum tertaut ke data karyawan.
		return response.Error(c, http.StatusForbidden, "Akun Siakad Anda belum terhubung ke data karyawan E-Maliyah. Silakan hubungi admin.")
	}

	// Karyawan bertanda admin (mis. bendahara) mendapat token admin,
	// bukan pengajar — ia mengelola payroll, bukan hanya melihat miliknya.
	if emp.IsAdmin {
		h.clearFailures(ip)
		token, err := h.issueToken(jwt.MapClaims{
			"sub":         req.Username,
			"role":        "admin",
			"name":        identity.Nama,
			"employee_id": emp.ID.String(),
			"siakad_id":   identity.ID,
		})
		if err != nil {
			return response.Error(c, http.StatusInternalServerError, "Gagal membuat token")
		}
		return response.Success(c, http.StatusOK, "Login berhasil", map[string]interface{}{
			"token": token, "username": req.Username, "role": "admin", "name": identity.Nama,
		})
	}

	h.clearFailures(ip)
	token, err := h.issueToken(jwt.MapClaims{
		"sub":         req.Username,
		"role":        "pengajar",
		"name":        identity.Nama,
		"employee_id": emp.ID.String(),
		"siakad_id":   identity.ID,
	})
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "Gagal membuat token")
	}
	return response.Success(c, http.StatusOK, "Login berhasil", map[string]interface{}{
		"token": token, "username": req.Username, "role": "pengajar", "name": identity.Nama,
	})
}
