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

// SSOHandler menangani handoff login sekali pakai (payload_b64 +
// HMAC-SHA256 hex): satu instance per penerbit, masing-masing dengan
// secret sendiri — instance Siakad (secret EMALIYAH_SSO_SECRET) dan
// instance Bawwabah (secret BAWWABAH_EMALIYAH_SSO_SECRET).
// Handler memverifikasi, menandai jti sudah dipakai, lalu menerbitkan
// sesi E-Maliyah sendiri (JWT 12 jam, sama seperti login biasa).
// Instance Bawwabah (allowSuperadmin) juga menghormati klaim superadmin:
// super admin Bawwabah mendapat sesi admin E-Maliyah tanpa baris pegawai.
// Instance Siakad TIDAK PERNAH menghormati klaim itu.
type SSOHandler struct {
	empRepo         domain.EmployeeRepository
	ssoRepo         domain.SSORepository
	jwtSecret       string
	ssoSecret       string
	allowSuperadmin bool
	issuerLabel     string
}

func NewSSOHandler(empRepo domain.EmployeeRepository, ssoRepo domain.SSORepository, jwtSecret, ssoSecret string, allowSuperadmin bool, issuerLabel string) *SSOHandler {
	return &SSOHandler{empRepo: empRepo, ssoRepo: ssoRepo, jwtSecret: jwtSecret, ssoSecret: ssoSecret, allowSuperadmin: allowSuperadmin, issuerLabel: issuerLabel}
}

type ssoTokenPayload struct {
	SiakadID   string `json:"siakad_id"`
	Superadmin bool   `json:"superadmin"`
	Exp        int64  `json:"exp"`
	JTI        string `json:"jti"`
}

func ssoErrorPage(issuer, msg string) string {
	return `<!DOCTYPE html><html lang="id"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>E-Maliyah</title></head><body style="font-family:system-ui,sans-serif;background:#F3F4F6;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0"><div style="background:#fff;border:1px solid #E5E7EB;border-radius:14px;padding:24px;max-width:420px"><h2 style="color:#B91C1C;margin:0 0 8px">Gagal masuk lewat ` + issuer + `</h2><p style="color:#374151;font-size:14px;line-height:1.6;margin:0">` + msg + ` Silakan kembali ke ` + issuer + ` dan coba lagi, atau login langsung di E-Maliyah.</p></div></body></html>`
}

// successPage menerbitkan JWT E-Maliyah (12 jam) dan mengembalikan halaman
// sukses yang menyimpan sesi ke localStorage lalu mengarahkan ke tujuan.
func (h *SSOHandler) successPage(c echo.Context, role, name, loadingText, redirect string, claims jwt.MapClaims) error {
	now := time.Now()
	claims["iat"] = now.Unix()
	claims["exp"] = now.Add(12 * time.Hour).Unix()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(h.jwtSecret))
	if err != nil {
		return c.HTML(http.StatusUnauthorized, ssoErrorPage(h.issuerLabel, "Terjadi kesalahan pada server."))
	}
	tokenJS, _ := json.Marshal(token)
	nameJS, _ := json.Marshal(name)
	roleJS, _ := json.Marshal(role)
	html := fmt.Sprintf(`<!DOCTYPE html><html lang="id"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>E-Maliyah</title></head><body style="font-family:system-ui,sans-serif;background:#F3F4F6;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0"><p style="color:#6B7280">%s</p><script>try{localStorage.setItem('emaliyah_token',%s);localStorage.setItem('emaliyah_role',%s);localStorage.setItem('emaliyah_name',%s);}catch(e){}window.location.href='%s';</script></body></html>`, loadingText, string(tokenJS), string(roleJS), string(nameJS), redirect)
	return c.HTML(http.StatusOK, html)
}

func (h *SSOHandler) Callback(c echo.Context) error {
	fail := func(msg string) error {
		return c.HTML(http.StatusUnauthorized, ssoErrorPage(h.issuerLabel, msg))
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
	// Super admin Bawwabah -> admin E-Maliyah (tanpa baris pegawai).
	// Cabang ini HANYA hidup pada instance Bawwabah (allowSuperadmin).
	if h.allowSuperadmin && p.Superadmin {
		return h.successPage(c, "admin", "Administrator", "Mengalihkan ke Dashboard…", "/dashboard", jwt.MapClaims{
			"sub":  p.SiakadID,
			"role": "admin",
			"name": "Administrator",
		})
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
	// Karyawan bertanda admin mendapat sesi admin, bukan pengajar —
	// berlaku untuk SSO Siakad maupun SSO Bawwabah non-superadmin.
	if emp.IsAdmin {
		return h.successPage(c, "admin", emp.Name, "Mengalihkan ke Dashboard…", "/dashboard", jwt.MapClaims{
			"sub":         p.SiakadID,
			"role":        "admin",
			"name":        emp.Name,
			"employee_id": emp.ID.String(),
			"siakad_id":   p.SiakadID,
		})
	}
	return h.successPage(c, "pengajar", emp.Name, "Mengalihkan ke Rincian Bisyaroh…", "/rincian-bisyaroh", jwt.MapClaims{
		"sub":         p.SiakadID,
		"role":        "pengajar",
		"name":        emp.Name,
		"employee_id": emp.ID.String(),
		"siakad_id":   p.SiakadID,
	})
}
