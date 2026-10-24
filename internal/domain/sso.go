package domain

import "time"

// SSOConsumedToken mencatat jti token SSO Siakad yang sudah dipakai,
// agar satu token tidak bisa diputar ulang (replay).
type SSOConsumedToken struct {
	JTI        string    `gorm:"primaryKey;type:varchar(64)" json:"jti"`
	ConsumedAt time.Time `json:"consumed_at"`
}

// SSORepository menyimpan/mengonsumsi jti token SSO.
type SSORepository interface {
	// ConsumeJTI mengembalikan true bila jti ini baru pertama kali dipakai.
	ConsumeJTI(jti string) (bool, error)
}
