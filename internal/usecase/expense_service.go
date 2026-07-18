package usecase

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ibnualwan/bisyaroh/internal/domain"
)

type ExpenseService struct {
	repo         domain.ExpenseRepository
	geminiAPIKey string
}

func NewExpenseService(repo domain.ExpenseRepository, geminiAPIKey string) *ExpenseService {
	return &ExpenseService{repo: repo, geminiAPIKey: geminiAPIKey}
}

// ----- Report CRUD -----

func (s *ExpenseService) CreateReport(month, year int, keterangan string) (*domain.ExpenseReport, error) {
	report := &domain.ExpenseReport{
		ID:         uuid.New(),
		Month:      month,
		Year:       year,
		Keterangan: keterangan,
		Status:     "DRAFT",
	}
	if err := s.repo.CreateReport(report); err != nil {
		return nil, err
	}
	return report, nil
}

func (s *ExpenseService) GetReport(id uuid.UUID) (*domain.ExpenseReport, error) {
	return s.repo.GetReportByID(id)
}

func (s *ExpenseService) ListReports(month, year int) ([]domain.ExpenseReport, error) {
	return s.repo.ListReports(month, year)
}

func (s *ExpenseService) DeleteReport(id uuid.UUID) error {
	return s.repo.DeleteReport(id)
}

// ----- Item CRUD -----

type AddItemRequest struct {
	NamaBarang      string          `json:"nama_barang"`
	HargaSatuan     decimal.Decimal `json:"harga_satuan"`
	Jumlah          int             `json:"jumlah"`
	Kredit          decimal.Decimal `json:"kredit"`
	Debit           decimal.Decimal `json:"debit"`
	BuktiPembayaran string          `json:"bukti_pembayaran"`
	GroupID         string          `json:"group_id"`
}

func (s *ExpenseService) AddItem(reportID uuid.UUID, req AddItemRequest) (*domain.ExpenseItem, error) {
	totalHarga := req.HargaSatuan.Mul(decimal.NewFromInt(int64(req.Jumlah)))

	item := &domain.ExpenseItem{
		ID:              uuid.New(),
		ExpenseReportID: reportID,
		NamaBarang:      req.NamaBarang,
		HargaSatuan:     req.HargaSatuan,
		Jumlah:          req.Jumlah,
		TotalHarga:      totalHarga,
		Kredit:          req.Kredit,
		Debit:           req.Debit,
		BuktiPembayaran: req.BuktiPembayaran,
		GroupID:         req.GroupID,
	}
	if err := s.repo.CreateItem(item); err != nil {
		return nil, err
	}
	if err := s.recalculateTotals(reportID); err != nil {
		return nil, err
	}
	return item, nil
}

type UpdateItemRequest struct {
	NamaBarang      string          `json:"nama_barang"`
	HargaSatuan     decimal.Decimal `json:"harga_satuan"`
	Jumlah          int             `json:"jumlah"`
	Kredit          decimal.Decimal `json:"kredit"`
	Debit           decimal.Decimal `json:"debit"`
	BuktiPembayaran string          `json:"bukti_pembayaran"`
	GroupID         string          `json:"group_id"`
	IsKonfirmasi    bool            `json:"is_konfirmasi"`
	IsTashih        bool            `json:"is_tashih"`
}

func (s *ExpenseService) UpdateItem(itemID uuid.UUID, req UpdateItemRequest) (*domain.ExpenseItem, error) {
	item, err := s.repo.GetItemByID(itemID)
	if err != nil {
		return nil, err
	}

	item.NamaBarang = req.NamaBarang
	item.HargaSatuan = req.HargaSatuan
	item.Jumlah = req.Jumlah
	item.TotalHarga = req.HargaSatuan.Mul(decimal.NewFromInt(int64(req.Jumlah)))
	item.Kredit = req.Kredit
	item.Debit = req.Debit
	item.BuktiPembayaran = req.BuktiPembayaran
	item.GroupID = req.GroupID
	item.IsKonfirmasi = req.IsKonfirmasi
	item.IsTashih = req.IsTashih

	if err := s.repo.UpdateItem(item); err != nil {
		return nil, err
	}
	if err := s.recalculateTotals(item.ExpenseReportID); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ExpenseService) DeleteItem(itemID uuid.UUID) error {
	item, err := s.repo.GetItemByID(itemID)
	if err != nil {
		return err
	}
	reportID := item.ExpenseReportID
	if err := s.repo.DeleteItem(itemID); err != nil {
		return err
	}
	return s.recalculateTotals(reportID)
}

// ----- Tashih All -----

func (s *ExpenseService) TashihAll(reportID uuid.UUID) (*domain.ExpenseReport, error) {
	items, err := s.repo.GetItemsByReportID(reportID)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if !item.IsTashih {
			return nil, fmt.Errorf("masih ada transaksi yang belum di-tashih")
		}
	}
	report, err := s.repo.GetReportByID(reportID)
	if err != nil {
		return nil, err
	}
	report.Status = "CONFIRMED"
	if err := s.repo.UpdateReport(report); err != nil {
		return nil, err
	}
	return report, nil
}

// ----- Internal helper -----

func (s *ExpenseService) recalculateTotals(reportID uuid.UUID) error {
	items, err := s.repo.GetItemsByReportID(reportID)
	if err != nil {
		return err
	}
	report, err := s.repo.GetReportByID(reportID)
	if err != nil {
		return err
	}

	totalHarga := decimal.Zero
	totalPengeluaran := decimal.Zero
	totalKembalian := decimal.Zero

	for _, item := range items {
		totalHarga = totalHarga.Add(item.TotalHarga)
		totalPengeluaran = totalPengeluaran.Add(item.Kredit)
		totalKembalian = totalKembalian.Add(item.Debit)
	}

	report.TotalHargaBarang = totalHarga
	report.TotalPengeluaran = totalPengeluaran
	report.TotalKembalian = totalKembalian
	return s.repo.UpdateReport(report)
}

// ----- AI Scan Receipt (Gemini) -----

type ScannedItem struct {
	NamaBarang  string  `json:"nama_barang"`
	HargaSatuan float64 `json:"harga_satuan"`
	Jumlah      int     `json:"jumlah"`
	TotalHarga  float64 `json:"total_harga"`
	Kredit      float64 `json:"kredit"`
	Debit       float64 `json:"debit"`
}

type ScanReceiptResult struct {
	Items  []ScannedItem `json:"items"`
	RawText string       `json:"raw_text"`
}

func (s *ExpenseService) ScanReceipt(imageBase64 string, mimeType string) (*ScanReceiptResult, error) {
	if s.geminiAPIKey == "" {
		return nil, fmt.Errorf("Gemini API key tidak dikonfigurasi")
	}

	prompt := `Baca nota/kwitansi pembelian ini dan ekstrak SEMUA item (multi item / lebih dari 1 item).
Penting: Format HANYA JSON murni yang sesuai struktur ini, jangan tambahkan penjelasan apapun:
{
  "items": [
    {
      "nama_barang": "Nama barang (contoh: Baterai ABC / Kertas Folio)",
      "harga_satuan": 21000,
      "jumlah": 3,
      "total_harga": 63000,
      "kredit": 70000,
      "debit": 7000
    }
  ],
  "raw_text": "text utuh dari nota"
}
Catatan:
1. Ekstrak semua barang yang ada secara berurutan.
2. harga_satuan, total_harga, kredit, debit adalah Number/Angka (tanpa Rp, tanpa koma/titik pemisah ribuan).
3. Jika total kredit/debit tidak spesifik per barang, taruh nominalnya di item pertama saja, yang lain isi 0.`

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"inline_data": map[string]string{
							"mime_type": mimeType,
							"data":      imageBase64,
						},
					},
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.1,
			"response_mime_type": "application/json",
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", s.geminiAPIKey)
	resp, err := http.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungi Gemini API: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
		log.Printf("Gemini unmarshal error: %v, raw response: %s", err, string(respBytes))
		return nil, fmt.Errorf("gagal parse respons Gemini: %v", err)
	}
	if geminiResp.Error != nil {
		log.Printf("Gemini API error from Google: %s", geminiResp.Error.Message)
		return nil, fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		log.Printf("Gemini empty candidates. Raw response: %s", string(respBytes))
		return nil, fmt.Errorf("Gemini tidak mendeteksi teks (respon kosong)")
	}

	text := strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text)
	// Bersihkan markdown code block jika ada
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	log.Printf("Gemini JSON result: %s", text)

	var result ScanReceiptResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		log.Printf("Failed to unmarshal Gemini result into ScannedItem: %v", err)
		return nil, fmt.Errorf("gagal membaca format data AI: %v", err)
	}
	return &result, nil
}

// ScanReceiptFromBase64 helper: terima raw base64 string
func (s *ExpenseService) ScanReceiptFromBase64(b64data string, mimeType string) (*ScanReceiptResult, error) {
	// Hapus prefix data URL jika ada
	if idx := strings.Index(b64data, ","); idx != -1 {
		b64data = b64data[idx+1:]
	}
	// Validasi base64
	if _, err := base64.StdEncoding.DecodeString(b64data); err != nil {
		// Coba URL-safe base64
		if _, err2 := base64.URLEncoding.DecodeString(b64data); err2 != nil {
			return nil, fmt.Errorf("data bukan base64 valid")
		}
	}
	return s.ScanReceipt(b64data, mimeType)
}
