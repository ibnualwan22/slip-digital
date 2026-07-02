package http

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ibnualwan/bisyaroh/internal/domain"
	"github.com/ibnualwan/bisyaroh/internal/usecase"
	"github.com/ibnualwan/bisyaroh/pkg/response"
)

type PayrollHandler struct {
	service usecase.PayrollService
}

func NewPayrollHandler(service usecase.PayrollService) *PayrollHandler {
	return &PayrollHandler{service: service}
}

func (h *PayrollHandler) CreateTransaction(c echo.Context) error {
	var tx domain.PayrollTransaction
	if err := c.Bind(&tx); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body")
	}

	if err := h.service.CreateTransaction(&tx); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to create transaction")
	}

	return response.Success(c, http.StatusCreated, "transaction created", tx)
}

func (h *PayrollHandler) ListTransactions(c echo.Context) error {
	month, _ := strconv.Atoi(c.QueryParam("month"))
	year, _ := strconv.Atoi(c.QueryParam("year"))

	txs, err := h.service.ListTransactions(month, year)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list transactions")
	}

	return response.Success(c, http.StatusOK, "success", txs)
}

func (h *PayrollHandler) GetTransaction(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid id")
	}

	tx, err := h.service.GetTransaction(id)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "transaction not found")
	}

	return response.Success(c, http.StatusOK, "success", tx)
}

func (h *PayrollHandler) DeleteTransaction(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid id")
	}

	if err := h.service.DeleteTransaction(id); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to delete transaction")
	}

	return response.Success(c, http.StatusOK, "transaction deleted successfully", nil)
}

func (h *PayrollHandler) AddDetail(c echo.Context) error {
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid transaction id")
	}

	var detail domain.PayrollDetail
	if err := c.Bind(&detail); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body")
	}
	detail.PayrollTransactionID = txID

	if err := h.service.AddDetail(&detail); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to add detail")
	}

	return response.Success(c, http.StatusCreated, "detail added", detail)
}

func (h *PayrollHandler) RemoveDetail(c echo.Context) error {
	detailId, err := uuid.Parse(c.Param("detailId"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid detail id")
	}

	if err := h.service.RemoveDetail(detailId); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to remove detail")
	}

	return response.Success(c, http.StatusOK, "detail removed successfully", nil)
}

func (h *PayrollHandler) CalculateTHP(c echo.Context) error {
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid transaction id")
	}

	if err := h.service.CalculateTransactionTHP(txID); err != nil {
		return response.Error(c, http.StatusInternalServerError, err.Error())
	}

	return response.Success(c, http.StatusOK, "calculation successful", nil)
}
