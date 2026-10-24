package postgres

import (
	"gorm.io/gorm"

	"github.com/ibnualwan/bisyaroh/internal/domain"
)

type ssoRepository struct {
	db *gorm.DB
}

func NewSSORepository(db *gorm.DB) domain.SSORepository {
	return &ssoRepository{db: db}
}

func (r *ssoRepository) ConsumeJTI(jti string) (bool, error) {
	res := r.db.Exec("INSERT INTO sso_consumed_tokens (jti, consumed_at) VALUES (?, NOW()) ON CONFLICT (jti) DO NOTHING", jti)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}
