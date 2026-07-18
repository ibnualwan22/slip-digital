package http

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"

	"github.com/ibnualwan/bisyaroh/internal/usecase"
	"github.com/ibnualwan/bisyaroh/pkg/response"
)

type ExpenseHandler struct {
	service             *usecase.ExpenseService
	cloudinaryCloudName string
	cloudinaryAPIKey    string
	cloudinaryAPISecret string
}

func NewExpenseHandler(service *usecase.ExpenseService, cloudName, apiKey, apiSecret string) *ExpenseHandler {
	return &ExpenseHandler{
		service:             service,
		cloudinaryCloudName: cloudName,
		cloudinaryAPIKey:    apiKey,
		cloudinaryAPISecret: apiSecret,
	}
}

// ListReports GET /expenses
func (h *ExpenseHandler) ListReports(c echo.Context) error {
	month, _ := strconv.Atoi(c.QueryParam("month"))
	year, _ := strconv.Atoi(c.QueryParam("year"))

	reports, err := h.service.ListReports(month, year)
	if err != nil {
		return response.Error(c, 500, "Gagal mengambil data laporan: "+err.Error())
	}
	return response.Success(c, 200, "OK", reports)
}

// CreateReport POST /expenses
func (h *ExpenseHandler) CreateReport(c echo.Context) error {
	var req struct {
		Month      int    `json:"month"`
		Year       int    `json:"year"`
		Keterangan string `json:"keterangan"`
	}
	if err := c.Bind(&req); err != nil {
		return response.Error(c, 400, "Request tidak valid: "+err.Error())
	}
	if req.Month < 1 || req.Month > 12 {
		return response.Error(c, 400, "Bulan tidak valid (harus 1-12)")
	}
	if req.Year < 2020 {
		return response.Error(c, 400, "Tahun tidak valid (minimal 2020)")
	}

	report, err := h.service.CreateReport(req.Month, req.Year, req.Keterangan)
	if err != nil {
		return response.Error(c, 500, "Gagal membuat laporan: "+err.Error())
	}
	return response.Success(c, 201, "Laporan berhasil dibuat", report)
}

// GetReport GET /expenses/:id
func (h *ExpenseHandler) GetReport(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, 400, "ID tidak valid")
	}
	report, err := h.service.GetReport(id)
	if err != nil {
		return response.Error(c, 404, "Laporan tidak ditemukan")
	}
	return response.Success(c, 200, "OK", report)
}

// DeleteReport DELETE /expenses/:id
func (h *ExpenseHandler) DeleteReport(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, 400, "ID tidak valid")
	}
	if err := h.service.DeleteReport(id); err != nil {
		return response.Error(c, 500, "Gagal menghapus laporan: "+err.Error())
	}
	return response.Success(c, 200, "Laporan berhasil dihapus", nil)
}

// AddItem POST /expenses/:id/items
func (h *ExpenseHandler) AddItem(c echo.Context) error {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, 400, "ID tidak valid")
	}

	var req struct {
		NamaBarang      string  `json:"nama_barang"`
		HargaSatuan     float64 `json:"harga_satuan"`
		Jumlah          int     `json:"jumlah"`
		Kredit          float64 `json:"kredit"`
		Debit           float64 `json:"debit"`
		BuktiPembayaran string  `json:"bukti_pembayaran"`
		GroupID         string  `json:"group_id"`
	}
	if err := c.Bind(&req); err != nil {
		return response.Error(c, 400, "Request tidak valid: "+err.Error())
	}
	if req.Jumlah < 1 {
		req.Jumlah = 1
	}

	item, err := h.service.AddItem(reportID, usecase.AddItemRequest{
		NamaBarang:      req.NamaBarang,
		HargaSatuan:     decimal.NewFromFloat(req.HargaSatuan),
		Jumlah:          req.Jumlah,
		Kredit:          decimal.NewFromFloat(req.Kredit),
		Debit:           decimal.NewFromFloat(req.Debit),
		BuktiPembayaran: req.BuktiPembayaran,
		GroupID:         req.GroupID,
	})
	if err != nil {
		return response.Error(c, 500, "Gagal menambah item: "+err.Error())
	}
	return response.Success(c, 201, "Item berhasil ditambahkan", item)
}

// UpdateItem PUT /expenses/items/:itemId
func (h *ExpenseHandler) UpdateItem(c echo.Context) error {
	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		return response.Error(c, 400, "ID item tidak valid")
	}

	var req struct {
		NamaBarang      string  `json:"nama_barang"`
		HargaSatuan     float64 `json:"harga_satuan"`
		Jumlah          int     `json:"jumlah"`
		Kredit          float64 `json:"kredit"`
		Debit           float64 `json:"debit"`
		BuktiPembayaran string  `json:"bukti_pembayaran"`
		GroupID         string  `json:"group_id"`
		IsKonfirmasi    bool    `json:"is_konfirmasi"`
		IsTashih        bool    `json:"is_tashih"`
	}
	if err := c.Bind(&req); err != nil {
		return response.Error(c, 400, "Request tidak valid: "+err.Error())
	}
	if req.Jumlah < 1 {
		req.Jumlah = 1
	}

	item, err := h.service.UpdateItem(itemID, usecase.UpdateItemRequest{
		NamaBarang:      req.NamaBarang,
		HargaSatuan:     decimal.NewFromFloat(req.HargaSatuan),
		Jumlah:          req.Jumlah,
		Kredit:          decimal.NewFromFloat(req.Kredit),
		Debit:           decimal.NewFromFloat(req.Debit),
		BuktiPembayaran: req.BuktiPembayaran,
		GroupID:         req.GroupID,
		IsKonfirmasi:    req.IsKonfirmasi,
		IsTashih:        req.IsTashih,
	})
	if err != nil {
		return response.Error(c, 500, "Gagal mengupdate item: "+err.Error())
	}
	return response.Success(c, 200, "Item berhasil diupdate", item)
}

// DeleteItem DELETE /expenses/items/:itemId
func (h *ExpenseHandler) DeleteItem(c echo.Context) error {
	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		return response.Error(c, 400, "ID item tidak valid")
	}
	if err := h.service.DeleteItem(itemID); err != nil {
		return response.Error(c, 500, "Gagal menghapus item: "+err.Error())
	}
	return response.Success(c, 200, "Item berhasil dihapus", nil)
}

// TashihAll POST /expenses/:id/tashih-all
func (h *ExpenseHandler) TashihAll(c echo.Context) error {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.Error(c, 400, "ID tidak valid")
	}
	report, err := h.service.TashihAll(reportID)
	if err != nil {
		return response.Error(c, 400, "Gagal tashih semua: "+err.Error())
	}
	return response.Success(c, 200, "Semua transaksi berhasil di-tashih", report)
}


// UploadBukti POST /expenses/upload
func (h *ExpenseHandler) UploadBukti(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return response.Error(c, 400, "File tidak ditemukan: "+err.Error())
	}

	src, err := file.Open()
	if err != nil {
		return response.Error(c, 500, "Gagal membuka file: "+err.Error())
	}
	defer src.Close()

	cloudName := h.cloudinaryCloudName
	apiKey := h.cloudinaryAPIKey
	apiSecret := h.cloudinaryAPISecret
	if cloudName == "" {
		cloudName = os.Getenv("CLOUDE_NAME")
	}
	if apiKey == "" {
		apiKey = os.Getenv("API_KEY_CLOUDINARY")
	}
	if apiSecret == "" {
		apiSecret = os.Getenv("API_SECRET_CLOUDINARY")
	}

	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cloudName)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Build signature: sha1("folder=expense-reports&timestamp={ts}{apiSecret}")
	sigStr := fmt.Sprintf("folder=expense-reports&timestamp=%s%s", timestamp, apiSecret)
	sha := sha1.New()
	sha.Write([]byte(sigStr))
	signature := hex.EncodeToString(sha.Sum(nil))

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for _, kv := range [][2]string{
		{"api_key", apiKey},
		{"timestamp", timestamp},
		{"signature", signature},
		{"folder", "expense-reports"},
	} {
		fw, _ := writer.CreateFormField(kv[0])
		fw.Write([]byte(kv[1]))
	}

	part, err := writer.CreateFormFile("file", file.Filename)
	if err != nil {
		return response.Error(c, 500, "Gagal membuat form upload: "+err.Error())
	}
	if _, err := io.Copy(part, src); err != nil {
		return response.Error(c, 500, "Gagal menulis file: "+err.Error())
	}
	writer.Close()

	httpReq, err := http.NewRequest("POST", uploadURL, &buf)
	if err != nil {
		return response.Error(c, 500, "Gagal membuat request upload: "+err.Error())
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return response.Error(c, 500, "Gagal upload ke Cloudinary: "+err.Error())
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var cloudResp map[string]interface{}
	if err := json.Unmarshal(respBody, &cloudResp); err != nil {
		return response.Error(c, 500, "Gagal parse respons Cloudinary: "+string(respBody))
	}
	if errVal, ok := cloudResp["error"]; ok {
		return response.Error(c, 500, fmt.Sprintf("Cloudinary error: %v", errVal))
	}

	secureURL, _ := cloudResp["secure_url"].(string)
	return response.Success(c, 200, "Upload berhasil", map[string]string{"url": secureURL})
}

// ScanReceipt POST /expenses/scan-receipt
func (h *ExpenseHandler) ScanReceipt(c echo.Context) error {
	var req struct {
		ImageBase64 string `json:"image_base64"`
		MimeType    string `json:"mime_type"`
	}
	if err := c.Bind(&req); err != nil {
		return response.Error(c, 400, "Request tidak valid: "+err.Error())
	}
	if req.ImageBase64 == "" {
		return response.Error(c, 400, "image_base64 wajib diisi")
	}
	if req.MimeType == "" {
		req.MimeType = "image/jpeg"
	}

	result, err := h.service.ScanReceiptFromBase64(req.ImageBase64, req.MimeType)
	if err != nil {
		return response.Error(c, 500, "Gagal scan nota: "+err.Error())
	}
	return response.Success(c, 200, "Scan berhasil", result)
}

// ScanReceiptFromURL POST /expenses/scan-receipt-url
func (h *ExpenseHandler) ScanReceiptFromURL(c echo.Context) error {
	var req struct {
		ImageURL string `json:"image_url"`
		MimeType string `json:"mime_type"`
	}
	if err := c.Bind(&req); err != nil {
		return response.Error(c, 400, "Request tidak valid: "+err.Error())
	}
	if req.ImageURL == "" {
		return response.Error(c, 400, "image_url wajib diisi")
	}
	if req.MimeType == "" {
		req.MimeType = "image/jpeg"
	}

	resp, err := http.Get(req.ImageURL)
	if err != nil {
		return response.Error(c, 500, "Gagal mengunduh gambar: "+err.Error())
	}
	defer resp.Body.Close()

	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return response.Error(c, 500, "Gagal membaca gambar: "+err.Error())
	}

	b64 := base64.StdEncoding.EncodeToString(imgBytes)
	result, err := h.service.ScanReceipt(b64, req.MimeType)
	if err != nil {
		return response.Error(c, 500, "Gagal scan nota: "+err.Error())
	}
	return response.Success(c, 200, "Scan berhasil", result)
}
