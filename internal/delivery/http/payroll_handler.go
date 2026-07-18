package http

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/ibnualwan/bisyaroh/internal/domain"
	"github.com/ibnualwan/bisyaroh/internal/usecase"
	"github.com/ibnualwan/bisyaroh/pkg/generator"
	"github.com/ibnualwan/bisyaroh/pkg/response"
	"github.com/ibnualwan/bisyaroh/pkg/whatsapp"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
)

type PayrollHandler struct {
	service  usecase.PayrollService
	empRepo  domain.EmployeeRepository
	waClient whatsapp.WhatsAppClient
	slipGen  generator.SlipGenerator
}

func NewPayrollHandler(service usecase.PayrollService, empRepo domain.EmployeeRepository, waClient whatsapp.WhatsAppClient, slipGen generator.SlipGenerator) *PayrollHandler {
	return &PayrollHandler{
		service:  service,
		empRepo:  empRepo,
		waClient: waClient,
		slipGen:  slipGen,
	}
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
		return response.Error(c, http.StatusBadRequest, "invalid request body: "+err.Error())
	}
	detail.PayrollTransactionID = txID

	if err := h.service.AddDetail(&detail); err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to add detail: "+err.Error())
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

func (h *PayrollHandler) PreviewSlipWA(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid transaction id")
	}

	tx, err := h.service.GetTransaction(id)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "transaction not found")
	}

	// Generate issue date (Pare, [Date])
	// Use Indonesian formatting for month
	months := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	now := time.Now()
	issueDate := strconv.Itoa(now.Day()) + " " + months[now.Month()] + " " + strconv.Itoa(now.Year())

	imageBytes, err := h.slipGen.GenerateSlipImage(tx, issueDate)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to generate slip image")
	}

	// Return base64 for preview
	base64Img := base64.StdEncoding.EncodeToString(imageBytes)
	return response.Success(c, http.StatusOK, "success", map[string]string{
		"image_base64": base64Img,
	})
}

func (h *PayrollHandler) SendSlipWA(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid transaction id")
	}

	tx, err := h.service.GetTransaction(id)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "transaction not found")
	}

	if tx.Employee.PhoneWA == "" {
		return response.Error(c, http.StatusBadRequest, "pegawai tidak memiliki nomor whatsapp yang tersimpan")
	}

	months := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	now := time.Now()
	issueDate := strconv.Itoa(now.Day()) + " " + months[now.Month()] + " " + strconv.Itoa(now.Year())

	imageBytes, err := h.slipGen.GenerateSlipImage(tx, issueDate)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to generate slip image")
	}

	caption := "Assalamu'alaikum Warahmatullahi Wabarakatuh\n\n" +
		"Berikut adalah Slip Bisyaroh Anda:\n\n" +
		"Nama: *" + tx.Employee.Name + "*\n" +
		"Jabatan: *" + tx.Employee.Role + "*\n" +
		"Periode: " + months[tx.Month] + " " + strconv.Itoa(tx.Year) + "\n\n" +
		"Terima kasih atas dedikasi dan pengabdian Anda di Markaz Arabiyah. Semoga *Allah senantiasa melimpahkan keberkahan*.\n\n" +
		"Ttd,\n*Karismaning Ulfa Awwalina, S.Pd*\nManager Keuangan Markaz Arabiyah"

	err = h.waClient.SendImage(tx.Employee.PhoneWA, caption, imageBytes)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "gagal mengirim pesan WA: "+err.Error())
	}

	return response.Success(c, http.StatusOK, "slip berhasil dikirim via WhatsApp", nil)
}

func (h *PayrollHandler) GeneratePayroll(c echo.Context) error {
	month, _ := strconv.Atoi(c.QueryParam("month"))
	year, _ := strconv.Atoi(c.QueryParam("year"))
	if month == 0 || year == 0 {
		return response.Error(c, http.StatusBadRequest, "month and year required")
	}

	employees, err := h.empRepo.List(nil, true)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list employees")
	}

	created := 0
	for _, emp := range employees {
		err := h.service.EnsurePayrollForEmployee(emp.ID, month, year)
		if err == nil {
			created++
		}
	}

	return response.Success(c, http.StatusOK, "payroll generated", map[string]int{"count": created})
}

func (h *PayrollHandler) ListDetails(c echo.Context) error {
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid transaction id")
	}

	tx, err := h.service.GetTransaction(txID)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "transaction not found")
	}

	return response.Success(c, http.StatusOK, "success", tx.Details)
}

func (h *PayrollHandler) UpdateStatus(c echo.Context) error {
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid transaction id")
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&body); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body")
	}

	err = h.service.UpdateStatus(txID, domain.PayrollStatus(body.Status))
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to update status: "+err.Error())
	}

	return response.Success(c, http.StatusOK, "status updated", nil)
}

func (h *PayrollHandler) UpdateDetail(c echo.Context) error {
	detailId, err := uuid.Parse(c.Param("detailId"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid detail id")
	}

	var req struct {
		ActivityID  *string         `json:"activity_id"`
		Quantity    decimal.Decimal `json:"quantity"`
		Rate        decimal.Decimal `json:"rate"`
		Description string          `json:"description"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, err.Error())
	}

	detail, err := h.service.GetDetail(detailId)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "detail not found")
	}

	if req.ActivityID != nil && *req.ActivityID != "" {
		actId, err := uuid.Parse(*req.ActivityID)
		if err == nil {
			detail.ActivityID = &actId
		}
	} else if req.ActivityID != nil && *req.ActivityID == "" {
		detail.ActivityID = nil
	}

	detail.Quantity = req.Quantity
	detail.Rate = req.Rate
	detail.Description = req.Description

	err = h.service.UpdateDetail(detail)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, err.Error())
	}

	return response.Success(c, http.StatusOK, "detail updated successfully", nil)
}

func (h *PayrollHandler) BulkSendWA(c echo.Context) error {
	month, _ := strconv.Atoi(c.QueryParam("month"))
	year, _ := strconv.Atoi(c.QueryParam("year"))

	txs, err := h.service.ListTransactions(month, year)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "failed to list transactions")
	}

	months := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	now := time.Now()
	issueDate := strconv.Itoa(now.Day()) + " " + months[now.Month()] + " " + strconv.Itoa(now.Year())

	sent := 0
	for _, t := range txs {
		if t.Employee.PhoneWA == "" {
			continue
		}
		imageBytes, err := h.slipGen.GenerateSlipImage(&t, issueDate)
		if err != nil {
			continue
		}

		caption := "Assalamu'alaikum Warahmatullahi Wabarakatuh\n\n" +
			"Berikut adalah Slip Bisyaroh Anda:\n\n" +
			"Nama: *" + t.Employee.Name + "*\n" +
			"Jabatan: *" + t.Employee.Role + "*\n" +
			"Periode: " + months[t.Month] + " " + strconv.Itoa(t.Year) + "\n\n" +
			"Terima kasih atas dedikasi dan pengabdian Anda di Markaz Arabiyah. Semoga *Allah senantiasa melimpahkan keberkahan*.\n\n" +
			"Ttd,\n*Karismaning Ulfa Awwalina, S.Pd*\nManager Keuangan Markaz Arabiyah"

		if err := h.waClient.SendImage(t.Employee.PhoneWA, caption, imageBytes); err == nil {
			sent++
		}
	}

	return response.Success(c, http.StatusOK, "berhasil mengirim "+strconv.Itoa(sent)+" slip via WhatsApp", nil)
}
