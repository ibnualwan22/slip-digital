package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/ibnualwan/bisyaroh/internal/domain"
	"github.com/ibnualwan/bisyaroh/pkg/response"
	"github.com/labstack/echo/v4"
)

type CategoryHandler struct {
	repo domain.CategoryRepository
}

func NewCategoryHandler(repo domain.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{repo: repo}
}

func (h *CategoryHandler) Create(c echo.Context) error {
	var cat domain.EmployeeCategory
	if err := c.Bind(&cat); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body")
	}

	if err := h.repo.Create(&cat); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to create category")
	}

	return response.Success(c, http.StatusCreated, "category created successfully", cat)
}

func (h *CategoryHandler) List(c echo.Context) error {
	categories, err := h.repo.List()
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list categories")
	}

	return response.Success(c, http.StatusOK, "success", categories)
}

func (h *CategoryHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid id")
	}

	cat, err := h.repo.GetByID(id)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "category not found")
	}

	return response.Success(c, http.StatusOK, "success", cat)
}

func (h *CategoryHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid id")
	}

	var cat domain.EmployeeCategory
	if err := c.Bind(&cat); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body")
	}
	cat.ID = id

	if err := h.repo.Update(&cat); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to update category")
	}

	return response.Success(c, http.StatusOK, "category updated successfully", cat)
}

func (h *CategoryHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid id")
	}

	if err := h.repo.Delete(id); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to delete category")
	}

	return response.Success(c, http.StatusOK, "category deleted successfully", nil)
}
