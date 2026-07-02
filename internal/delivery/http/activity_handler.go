package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ibnualwan/bisyaroh/internal/domain"
	"github.com/ibnualwan/bisyaroh/pkg/response"
)

// For simplicity, directly using the repository instead of a service layer for Master Data
type ActivityHandler struct {
	repo domain.ActivityRepository
}

func NewActivityHandler(repo domain.ActivityRepository) *ActivityHandler {
	return &ActivityHandler{repo: repo}
}

func (h *ActivityHandler) List(c echo.Context) error {
	activeOnly := c.QueryParam("active") == "true"
	activities, err := h.repo.List(activeOnly)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to fetch activities")
	}
	return response.Success(c, http.StatusOK, "success", activities)
}

func (h *ActivityHandler) Create(c echo.Context) error {
	var activity domain.MasterActivity
	if err := c.Bind(&activity); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body")
	}

	if err := h.repo.Create(&activity); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to create activity")
	}

	return response.Success(c, http.StatusCreated, "activity created successfully", activity)
}

func (h *ActivityHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid id")
	}

	activity, err := h.repo.GetByID(id)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "activity not found")
	}

	return response.Success(c, http.StatusOK, "success", activity)
}

func (h *ActivityHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid id")
	}

	var activity domain.MasterActivity
	if err := c.Bind(&activity); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body")
	}
	activity.ID = id

	if err := h.repo.Update(&activity); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to update activity")
	}

	return response.Success(c, http.StatusOK, "activity updated successfully", activity)
}
