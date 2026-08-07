package http

import (
	"strconv"
	"time"

	"github.com/ibnualwan/bisyaroh/internal/domain"
	"github.com/ibnualwan/bisyaroh/pkg/response"
	"github.com/labstack/echo/v4"
)

type DashboardHandler struct {
	service domain.DashboardService
}

func NewDashboardHandler(service domain.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) GetStats(c echo.Context) error {
	yearStr := c.QueryParam("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year == 0 {
		year = time.Now().Year() // default to current year
	}

	stats, err := h.service.GetDashboardStats(year)
	if err != nil {
		return response.Error(c, 500, "Failed to retrieve dashboard stats")
	}

	return response.Success(c, 200, "Dashboard stats retrieved successfully", stats)
}
