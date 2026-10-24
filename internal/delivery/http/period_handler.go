package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ibnualwan/bisyaroh/internal/domain"
	"github.com/ibnualwan/bisyaroh/pkg/response"
	"github.com/labstack/echo/v4"
)

type PeriodHandler struct {
	repo domain.PeriodRepository
}

func NewPeriodHandler(repo domain.PeriodRepository) *PeriodHandler {
	return &PeriodHandler{repo: repo}
}

type periodRequest struct {
	DufahName string `json:"dufah_name"`
	Usbu      int    `json:"usbu"`
	Month     int    `json:"month"`
	Year      int    `json:"year"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type periodView struct {
	ID        string `json:"id"`
	DufahName string `json:"dufah_name"`
	Usbu      int    `json:"usbu"`
	Month     int    `json:"month"`
	Year      int    `json:"year"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func toPeriodView(p *domain.PeriodMapping) periodView {
	return periodView{
		ID:        p.ID.String(),
		DufahName: p.DufahName,
		Usbu:      p.Usbu,
		Month:     p.Month,
		Year:      p.Year,
		StartDate: p.StartDate.Format("2006-01-02"),
		EndDate:   p.EndDate.Format("2006-01-02"),
	}
}

func (h *PeriodHandler) parseRequest(c echo.Context) (*domain.PeriodMapping, error) {
	var req periodRequest
	if err := c.Bind(&req); err != nil {
		return nil, fmt.Errorf("format permintaan tidak valid")
	}
	if req.DufahName == "" {
		return nil, fmt.Errorf("nama dufah wajib diisi")
	}
	if req.Usbu < 0 || req.Usbu > 3 {
		return nil, fmt.Errorf("usbu harus 0-3")
	}
	if req.Month < 1 || req.Month > 12 {
		return nil, fmt.Errorf("bulan harus 1-12")
	}
	if req.Year < 2020 || req.Year > 2100 {
		return nil, fmt.Errorf("tahun tidak valid")
	}
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("tanggal mulai harus format YYYY-MM-DD")
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("tanggal selesai harus format YYYY-MM-DD")
	}
	if end.Before(start) {
		return nil, fmt.Errorf("tanggal selesai tidak boleh sebelum tanggal mulai")
	}
	return &domain.PeriodMapping{
		DufahName: req.DufahName,
		Usbu:      req.Usbu,
		Month:     req.Month,
		Year:      req.Year,
		StartDate: start,
		EndDate:   end,
	}, nil
}

// findOverlap mencari periode lain yang rentang tanggalnya bertabrakan.
func (h *PeriodHandler) findOverlap(candidate *domain.PeriodMapping, excludeID *uuid.UUID) (*domain.PeriodMapping, error) {
	all, err := h.repo.List()
	if err != nil {
		return nil, err
	}
	for i := range all {
		p := &all[i]
		if excludeID != nil && p.ID == *excludeID {
			continue
		}
		if !candidate.StartDate.After(p.EndDate) && !p.StartDate.After(candidate.EndDate) {
			return p, nil
		}
	}
	return nil, nil
}

func (h *PeriodHandler) List(c echo.Context) error {
	periods, err := h.repo.List()
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list periods")
	}
	views := make([]periodView, 0, len(periods))
	for i := range periods {
		views = append(views, toPeriodView(&periods[i]))
	}
	return response.Success(c, http.StatusOK, "success", views)
}

func (h *PeriodHandler) Create(c echo.Context) error {
	period, err := h.parseRequest(c)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error())
	}
	overlap, err := h.findOverlap(period, nil)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to validate period")
	}
	if overlap != nil {
		return response.Error(c, http.StatusBadRequest, fmt.Sprintf(
			"rentang tanggal tumpang tindih dengan periode %s (%s s.d. %s)",
			overlap.DufahName, overlap.StartDate.Format("2006-01-02"), overlap.EndDate.Format("2006-01-02")))
	}
	if err := h.repo.Create(period); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to create period")
	}
	return response.Success(c, http.StatusCreated, "period created successfully", toPeriodView(period))
}

func (h *PeriodHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid id")
	}
	existing, err := h.repo.GetByID(id)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "period not found")
	}
	period, err := h.parseRequest(c)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error())
	}
	overlap, err := h.findOverlap(period, &id)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to validate period")
	}
	if overlap != nil {
		return response.Error(c, http.StatusBadRequest, fmt.Sprintf(
			"rentang tanggal tumpang tindih dengan periode %s (%s s.d. %s)",
			overlap.DufahName, overlap.StartDate.Format("2006-01-02"), overlap.EndDate.Format("2006-01-02")))
	}
	existing.DufahName = period.DufahName
	existing.Usbu = period.Usbu
	existing.Month = period.Month
	existing.Year = period.Year
	existing.StartDate = period.StartDate
	existing.EndDate = period.EndDate
	if err := h.repo.Update(existing); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to update period")
	}
	return response.Success(c, http.StatusOK, "period updated successfully", toPeriodView(existing))
}

func (h *PeriodHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid id")
	}
	if _, err := h.repo.GetByID(id); err != nil {
		return response.Error(c, http.StatusNotFound, "period not found")
	}
	if err := h.repo.Delete(id); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to delete period")
	}
	return response.Success(c, http.StatusOK, "period deleted successfully", nil)
}
