package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/ibnualwan/bisyaroh/internal/domain"
)

// SSOHandler menangani handoff login dari Siakad: Siakad menandatangani
// token sekali pakai (payload_b64 + HMAC-SHA256 hex) berisi siakad_id;
// handler memverifikasi, menandai jti sudah dipakai, lalu menerbitkan
// sesi E-Maliyah sendiri (JWT 12 jam, sama seperti login biasa).
type SSOHandler struct {
	empRepo   domain.EmployeeRepository
	ssoRepo   domain.SSORepository
	jwtSecret string
	ssoSecret string
}

func NewSSOHandler(empRepo domain.EmployeeRepository, ssoRepo domain.SSORepository, jwtSecret, ssoSecret string) *SSOHandler {
	return &SSOHandler{empRepo: empRepo, ssoRepo: ssoRepo, jwtSecret: jwtSecret, ssoSecret: ssoSecret}
}

type ssoTokenPayload struct {
	SiakadID string `json:"siakad_id"`
	Exp      int64  `json:"exp"`
	JTI      string `json:"jti"`
}

func ssoErrorPage(msg string) string {
	return `<!DOCTYPE html><html lang="id"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>E-Maliyah</title></head><body style="font-family:system-ui,sans-serif;background:#F3F4F6;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0"><div style="background:#fff;border:1px solid #E5E7EB;border-radius:14px;padding:24px;max-width:420px"><h2 style="color:#B91C1C;margin:0 0 8px">Gagal masuk lewat Siakad</h2><p style="color:#374151;font-size:14px;line-height:1.6;margin:0">` + msg + ` Silakan kembali ke Siakad dan coba lagi, atau login langsung di E-Maliyah.</p></div></body></html>`
}

func (h *SSOHandler) Callback(c echo.Context) error {
	fail := func(msg string) error {
		return c.HTML(http.StatusUnauthorized, ssoErrorPage(msg))
	}
	if h.ssoSecret == "" {
		return fail("Integrasi SSO belum dikonfigurasi di server.")
	}
	raw := c.QueryParam("token")
	parts := strings.Split(raw, ".")
	if len(parts) != 2 {
		return fail("Token SSO tidak valid.")
	}
	payloadB64, sigHex := parts[0], parts[1]
	mac := hmac.New(sha256.New, []byte(h.ssoSecret))
	mac.Write([]byte(payloadB64))
	expected := hex.EncodeToString(mac.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(sigHex)) != 1 {
		return fail("Token SSO tidak valid.")
	}
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return fail("Token SSO tidak valid.")
	}
	var p ssoTokenPayload
	if err := json.Unmarshal(payloadJSON, &p); err != nil {
		return fail("Token SSO tidak valid.")
	}
	if p.SiakadID == "" || p.JTI == "" || time.Now().Unix() > p.Exp {
		return fail("Token SSO sudah kedaluwarsa.")
	}
	fresh, err := h.ssoRepo.ConsumeJTI(p.JTI)
	if err != nil {
		return fail("Terjadi kesalahan pada server.")
	}
	if !fresh {
		return fail("Token SSO sudah dipakai.")
	}
	employees, err := h.empRepo.List(nil, false)
	if err != nil {
		return fail("Terjadi kesalahan pada server.")
	}
	var emp *domain.Employee
	for i := range employees {
		if employees[i].SiakadID != nil && *employees[i].SiakadID == p.SiakadID {
			emp = &employees[i]
			break
		}
	}
	if emp == nil {
		return fail("Akun Siakad Anda belum terhubung ke data karyawan E-Maliyah.")
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":         p.SiakadID,
		"role":        "pengajar",
		"name":        emp.Name,
		"employee_id": emp.ID.String(),
		"siakad_id":   p.SiakadID,
		"iat":         now.Unix(),
		"exp":         now.Add(12 * time.Hour).Unix(),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(h.jwtSecret))
	if err != nil {
		return fail("Terjadi kesalahan pada server.")
	}
	tokenJS, _ := json.Marshal(token)
	nameJS, _ := json.Marshal(emp.Name)
	html := fmt.Sprintf(`<!DOCTYPE html><html lang="id"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>E-Maliyah</title></head><body style="font-family:system-ui,sans-serif;background:#F3F4F6;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0"><p style="color:#6B7280">Mengalihkan ke Rincian Bisyaroh…</p><script>try{localStorage.setItem('emaliyah_token',%s);localStorage.setItem('emaliyah_role','pengajar');localStorage.setItem('emaliyah_name',%s);}catch(e){}window.location.href='/rincian-bisyaroh';</script></body></html>`, string(tokenJS), string(nameJS))
	return c.HTML(http.StatusOK, html)
}
