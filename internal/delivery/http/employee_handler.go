package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ibnualwan/bisyaroh/internal/domain"
	"github.com/ibnualwan/bisyaroh/internal/usecase"
	"github.com/ibnualwan/bisyaroh/pkg/response"
)

type EmployeeHandler struct {
	service usecase.EmployeeService
}

func NewEmployeeHandler(service usecase.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{service: service}
}

func (h *EmployeeHandler) Create(c echo.Context) error {
	var emp domain.Employee
	if err := c.Bind(&emp); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body")
	}

	if err := h.service.CreateEmployee(&emp); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to create employee")
	}

	return response.Success(c, http.StatusCreated, "employee created successfully", emp)
}

func (h *EmployeeHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid id")
	}

	emp, err := h.service.GetEmployee(id)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "employee not found")
	}

	return response.Success(c, http.StatusOK, "success", emp)
}

func (h *EmployeeHandler) List(c echo.Context) error {
	categoryStr := c.QueryParam("category")
	activeOnly := c.QueryParam("active") == "true"

	var categoryID *uuid.UUID
	if categoryStr != "" {
		if id, err := uuid.Parse(categoryStr); err == nil {
			categoryID = &id
		}
	}

	employees, err := h.service.ListEmployees(categoryID, activeOnly)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to fetch employees")
	}

	return response.Success(c, http.StatusOK, "success", employees)
}

func (h *EmployeeHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid id")
	}

	var emp domain.Employee
	if err := c.Bind(&emp); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body")
	}
	emp.ID = id

	if err := h.service.UpdateEmployee(&emp); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to update employee")
	}

	return response.Success(c, http.StatusOK, "employee updated successfully", emp)
}

func (h *EmployeeHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid id")
	}

	if err := h.service.DeleteEmployee(id); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to delete employee")
	}

	return response.Success(c, http.StatusOK, "employee deleted successfully", nil)
}
