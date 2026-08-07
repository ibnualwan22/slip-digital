package usecase

import (
	"github.com/ibnualwan/bisyaroh/internal/domain"
)

type dashboardService struct {
	repo domain.DashboardRepository
}

func NewDashboardService(repo domain.DashboardRepository) domain.DashboardService {
	return &dashboardService{repo}
}

func (s *dashboardService) GetDashboardStats(year int) (*domain.DashboardStats, error) {
	return s.repo.GetDashboardStats(year)
}
