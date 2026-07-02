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
)

type PayrollHandler struct {
	service  usecase.PayrollService
	waClient whatsapp.WhatsAppClient
	slipGen  generator.SlipGenerator
}

func NewPayrollHandler(service usecase.PayrollService, waClient whatsapp.WhatsAppClient, slipGen generator.SlipGenerator) *PayrollHandler {
	return &PayrollHandler{
		service:  service,
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
		"Terima kasih atas dedikasi dan pengabdian Anda di Markaz Arabiyah. Semoga *Qolil Bilbarokah Khoir Min Katsir Bila Barokah*.\n\n" +
		"Ttd,\n*Karismaning Ulfa Awwalina, S.Pd*\nManager Keuangan Markaz Arabiyah"

	err = h.waClient.SendImage(tx.Employee.PhoneWA, caption, imageBytes)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "gagal mengirim pesan WA: "+err.Error())
	}

	return response.Success(c, http.StatusOK, "slip berhasil dikirim via WhatsApp", nil)
}
