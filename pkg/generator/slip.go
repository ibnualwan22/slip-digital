package generator

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"os"
	"strings"

	"github.com/ibnualwan/bisyaroh/internal/domain"
	"github.com/shopspring/decimal"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type SlipGenerator interface {
	GenerateSlipImage(tx *domain.PayrollTransaction, issueDate string) ([]byte, error)
}

type slipGenerator struct {
	bgPath      string
	stempelPath string
	fontRegular font.Face
	fontBold    font.Face
	fontSmall   font.Face
	fontSmBold  font.Face
}

func NewSlipGenerator(bgPath, stempelPath string) (SlipGenerator, error) {
	fRegular, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, fmt.Errorf("failed to parse regular font: %w", err)
	}
	fBold, err := opentype.Parse(gobold.TTF)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bold font: %w", err)
	}

	regular, _ := opentype.NewFace(fRegular, &opentype.FaceOptions{Size: 24, DPI: 72, Hinting: font.HintingFull})
	bold, _ := opentype.NewFace(fBold, &opentype.FaceOptions{Size: 24, DPI: 72, Hinting: font.HintingFull})
	small, _ := opentype.NewFace(fRegular, &opentype.FaceOptions{Size: 20, DPI: 72, Hinting: font.HintingFull})
	smBold, _ := opentype.NewFace(fBold, &opentype.FaceOptions{Size: 20, DPI: 72, Hinting: font.HintingFull})

	return &slipGenerator{
		bgPath:      bgPath,
		stempelPath: stempelPath,
		fontRegular: regular,
		fontBold:    bold,
		fontSmall:   small,
		fontSmBold:  smBold,
	}, nil
}

// Helper: draw text at position
func drawText(img draw.Image, x, y int, text string, face font.Face, col color.Color) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	d.DrawString(text)
}

// Helper: draw right-aligned text
func drawTextRight(img draw.Image, rightX, y int, text string, face font.Face, col color.Color) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
	}
	width := d.MeasureString(text)
	d.Dot = fixed.Point26_6{X: fixed.I(rightX) - width, Y: fixed.I(y)}
	d.DrawString(text)
}

// Helper: draw horizontal line
func drawHLine(img draw.Image, x1, x2, y int, col color.Color) {
	for x := x1; x <= x2; x++ {
		img.Set(x, y, col)
	}
}

// Helper: draw vertical line
func drawVLine(img draw.Image, x, y1, y2 int, col color.Color) {
	for y := y1; y <= y2; y++ {
		img.Set(x, y, col)
	}
}

// Helper: draw a filled rectangle
func drawRect(img draw.Image, x1, y1, x2, y2 int, col color.Color) {
	draw.Draw(img, image.Rect(x1, y1, x2, y2), image.NewUniform(col), image.Point{}, draw.Src)
}

// Format number with dot separator: 1395000 -> "Rp1.395.000"
func formatRp(d decimal.Decimal) string {
	str := d.StringFixed(0)
	if str == "0" {
		return ""
	}
	neg := false
	if strings.HasPrefix(str, "-") {
		neg = true
		str = str[1:]
	}
	n := len(str)
	if n <= 3 {
		if neg {
			return "Rp-" + str
		}
		return "Rp" + str
	}
	var result []byte
	for i, c := range str {
		if i > 0 && (n-i)%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, byte(c))
	}
	if neg {
		return "Rp-" + string(result)
	}
	return "Rp" + string(result)
}

func (s *slipGenerator) GenerateSlipImage(tx *domain.PayrollTransaction, issueDate string) ([]byte, error) {
	// ========== 1. LOAD BACKGROUND ==========
	bgFile, err := os.Open(s.bgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open background: %w", err)
	}
	defer bgFile.Close()

	bgImg, _, err := image.Decode(bgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode background: %w", err)
	}

	bounds := bgImg.Bounds()
	canvas := image.NewRGBA(bounds)
	draw.Draw(canvas, bounds, bgImg, image.Point{}, draw.Src)

	// Colors
	black := color.Black
	lineColor := color.RGBA{80, 80, 80, 255}
	totalBg := color.RGBA{255, 235, 150, 255} // Light yellow for total header

	// Layout constants (image is 1242x1757)
	leftMargin := 90
	rightMargin := 1152
	tableLeft := leftMargin
	tableRight := rightMargin
	rowH := 40 // Modest taller rows for 1.5x fonts

	// ========== 2. GROUP DETAILS ==========
	details := tx.Details
	
	type slipItem struct {
		Name     string
		Intensif string
		Qty      string
		Total    string
	}

	var pengajaranItems []slipItem
	var liqoatItems []slipItem
	var dendaItems []slipItem

	pengajaranCodes := map[string]bool{
		"GAJI_POKOK": true, "TUNJANGAN": true, "JAM_AJAR": true, "WALI_KELAS": true,
	}

	for _, det := range details {
		code := ""
		name := det.Description
		if det.Activity != nil {
			code = det.Activity.Code
			if name == "" {
				name = det.Activity.ActivityName
			}
		}
		if name == "" {
			name = "Item"
		}

		if det.Type == "DEDUCTION" {
			qtyStr := ""
			if !det.Quantity.IsZero() {
				qtyStr = det.Quantity.StringFixed(0)
			}
			dendaItems = append(dendaItems, slipItem{
				Name:     name,
				Intensif: formatRp(det.Rate),
				Qty:      qtyStr,
				Total:    formatRp(det.TotalAmount),
			})
		} else {
			nameLower := strings.ToLower(name)
			isPengajaran := pengajaranCodes[code] || strings.Contains(nameLower, "gaji") || strings.Contains(nameLower, "tunjangan") || strings.Contains(nameLower, "mengajar") || strings.Contains(nameLower, "wali kelas")

			if isPengajaran {
				qtyStr := ""
				if !det.Quantity.IsZero() {
					qtyStr = det.Quantity.StringFixed(0)
				}
				pengajaranItems = append(pengajaranItems, slipItem{
					Name:     name,
					Intensif: formatRp(det.Rate),
					Qty:      qtyStr,
					Total:    formatRp(det.TotalAmount),
				})
			} else {
				qtyStr := ""
				if !det.Quantity.IsZero() {
					qtyStr = det.Quantity.StringFixed(0)
				}
				liqoatItems = append(liqoatItems, slipItem{
					Name:     name,
					Intensif: formatRp(det.Rate),
					Qty:      qtyStr,
					Total:    formatRp(det.TotalAmount),
				})
			}
		}
	}

	// ========== 3. EMPLOYEE INFO ==========
	// Push everything down significantly as requested
	y := 330

	// Row 1: Nama Pegawai | Jabatan
	drawText(canvas, leftMargin, y, "Nama Pegawai", s.fontSmall, black)
	drawText(canvas, leftMargin+135, y, ":  "+tx.Employee.Name, s.fontSmBold, black)
	drawText(canvas, 660, y, "Jabatan", s.fontSmall, black)
	drawText(canvas, 760, y, ":  "+tx.Employee.Role, s.fontSmBold, black)

	// Row 2: Duf'ah | Tanggal
	y += 22
	drawText(canvas, leftMargin, y, "Duf'ah", s.fontSmall, black)
	drawText(canvas, leftMargin+135, y, ":  -", s.fontSmall, black)
	drawText(canvas, 660, y, "Tanggal", s.fontSmall, black)
	drawText(canvas, 760, y, ":  "+issueDate, s.fontSmall, black)

	drawHLine(canvas, leftMargin, tableRight, y+10, lineColor)

	y += 45 // Increased gap

	y += 25

	// ---------- SECTION 1: PENGAJARAN ----------
	drawText(canvas, leftMargin, y, "1.  Pengajaran", s.fontBold, black)
	y += 15

	// Table header
	col1 := tableLeft
	col2 := tableLeft + 390
	col3 := col2 + 225
	col4 := col3 + 165
	col5 := tableRight

	headerY := y
	// No header Bg
	drawHLine(canvas, col1, col5, headerY, lineColor)
	drawHLine(canvas, col1, col5, headerY+rowH, lineColor)
	drawVLine(canvas, col1, headerY, headerY+rowH, lineColor)
	drawVLine(canvas, col2, headerY, headerY+rowH, lineColor)
	drawVLine(canvas, col3, headerY, headerY+rowH, lineColor)
	drawVLine(canvas, col4, headerY, headerY+rowH, lineColor)
	drawVLine(canvas, col5, headerY, headerY+rowH, lineColor)

	drawText(canvas, col1+10, headerY+20, "Nama Pengajaran", s.fontSmBold, black)
	drawText(canvas, col2+10, headerY+20, "Intensif", s.fontSmBold, black)
	drawText(canvas, col3+10, headerY+20, "Jumlah", s.fontSmBold, black)
	drawText(canvas, col4+10, headerY+20, "Total", s.fontSmBold, black)

	y = headerY + rowH
	pengajaranTotal := decimal.Zero
	for _, item := range pengajaranItems {
		drawHLine(canvas, col1, col5, y+rowH, lineColor)
		drawVLine(canvas, col1, y, y+rowH, lineColor)
		drawVLine(canvas, col2, y, y+rowH, lineColor)
		drawVLine(canvas, col3, y, y+rowH, lineColor)
		drawVLine(canvas, col4, y, y+rowH, lineColor)
		drawVLine(canvas, col5, y, y+rowH, lineColor)

		drawText(canvas, col1+10, y+20, item.Name, s.fontSmall, black)
		drawTextRight(canvas, col3-10, y+20, item.Intensif, s.fontSmall, black)
		drawText(canvas, col3+10, y+20, item.Qty, s.fontSmall, black)
		drawTextRight(canvas, col5-10, y+20, item.Total, s.fontSmall, black)
		y += rowH
	}

	// Pengajaran total row
	drawHLine(canvas, col1, col5, y, lineColor)
	drawHLine(canvas, col1, col5, y+rowH, lineColor)
	drawVLine(canvas, col1, y, y+rowH, lineColor)
	drawVLine(canvas, col4, y, y+rowH, lineColor)
	drawVLine(canvas, col5, y, y+rowH, lineColor)
	drawText(canvas, col4-110, y+20, "TOTAL", s.fontSmBold, black)
	for _, det := range tx.Details {
		code := ""
		name := det.Description
		if det.Activity != nil {
			code = det.Activity.Code
			if name == "" {
				name = det.Activity.ActivityName
			}
		}
		nameLower := strings.ToLower(name)
		isPengajaran := pengajaranCodes[code] || strings.Contains(nameLower, "gaji") || strings.Contains(nameLower, "tunjangan") || strings.Contains(nameLower, "mengajar") || strings.Contains(nameLower, "wali kelas")

		if det.Type == "ADDITION" && isPengajaran {
			pengajaranTotal = pengajaranTotal.Add(det.TotalAmount)
		}
	}
	drawTextRight(canvas, col5-10, y+20, formatRp(pengajaranTotal), s.fontSmBold, black)
	y += rowH + 30 // Larger gap

	// ---------- SECTION 2: LIQO'AT BERINTESIF ----------
	drawText(canvas, leftMargin, y, "2.  Liqo'at Berintesif", s.fontBold, black)
	y += 15

	lCol1 := tableLeft
	lCol2 := tableLeft + 420
	lCol3 := lCol2 + 250
	lCol4 := lCol3 + 180
	lCol5 := tableRight

	lHeaderY := y
	drawHLine(canvas, lCol1, lCol5, lHeaderY, lineColor)
	drawHLine(canvas, lCol1, lCol5, lHeaderY+rowH, lineColor)
	drawVLine(canvas, lCol1, lHeaderY, lHeaderY+rowH, lineColor)
	drawVLine(canvas, lCol2, lHeaderY, lHeaderY+rowH, lineColor)
	drawVLine(canvas, lCol3, lHeaderY, lHeaderY+rowH, lineColor)
	drawVLine(canvas, lCol4, lHeaderY, lHeaderY+rowH, lineColor)
	drawVLine(canvas, lCol5, lHeaderY, lHeaderY+rowH, lineColor)

	drawText(canvas, lCol1+10, lHeaderY+20, "Nama Liqo'at", s.fontSmBold, black)
	drawText(canvas, lCol2+10, lHeaderY+20, "Intensif", s.fontSmBold, black)
	drawText(canvas, lCol3+10, lHeaderY+20, "Jumlah", s.fontSmBold, black)
	drawText(canvas, lCol4+10, lHeaderY+20, "Total", s.fontSmBold, black)

	y = lHeaderY + rowH
	for _, item := range liqoatItems {
		drawHLine(canvas, lCol1, lCol5, y+rowH, lineColor)
		drawVLine(canvas, lCol1, y, y+rowH, lineColor)
		drawVLine(canvas, lCol2, y, y+rowH, lineColor)
		drawVLine(canvas, lCol3, y, y+rowH, lineColor)
		drawVLine(canvas, lCol4, y, y+rowH, lineColor)
		drawVLine(canvas, lCol5, y, y+rowH, lineColor)

		drawText(canvas, lCol1+10, y+20, item.Name, s.fontSmall, black)
		drawTextRight(canvas, lCol3-10, y+20, item.Intensif, s.fontSmall, black)
		drawText(canvas, lCol3+10, y+20, item.Qty, s.fontSmall, black)
		drawTextRight(canvas, lCol5-10, y+20, item.Total, s.fontSmall, black)
		y += rowH
	}

	if len(liqoatItems) == 0 {
		drawHLine(canvas, lCol1, lCol5, y+rowH, lineColor)
		drawVLine(canvas, lCol1, y, y+rowH, lineColor)
		drawVLine(canvas, lCol5, y, y+rowH, lineColor)
		drawText(canvas, lCol1+10, y+20, "-", s.fontSmall, black)
		y += rowH
	}

	y += 30 // Larger gap

	// ---------- SECTION 3: DENDA ----------
	drawText(canvas, leftMargin, y, "3.  Denda", s.fontBold, black)
	y += 15

	dCol1 := tableLeft
	dCol2 := tableLeft + 420
	dCol3 := dCol2 + 250
	dCol4 := dCol3 + 180
	dCol5 := tableRight

	dHeaderY := y
	drawHLine(canvas, dCol1, dCol5, dHeaderY, lineColor)
	drawHLine(canvas, dCol1, dCol5, dHeaderY+rowH, lineColor)
	drawVLine(canvas, dCol1, dHeaderY, dHeaderY+rowH, lineColor)
	drawVLine(canvas, dCol2, dHeaderY, dHeaderY+rowH, lineColor)
	drawVLine(canvas, dCol3, dHeaderY, dHeaderY+rowH, lineColor)
	drawVLine(canvas, dCol4, dHeaderY, dHeaderY+rowH, lineColor)
	drawVLine(canvas, dCol5, dHeaderY, dHeaderY+rowH, lineColor)

	drawText(canvas, dCol1+10, dHeaderY+20, "Jenis Denda", s.fontSmBold, black)
	drawText(canvas, dCol2+10, dHeaderY+20, "Pengurangan", s.fontSmBold, black)
	drawText(canvas, dCol3+10, dHeaderY+20, "Jumlah", s.fontSmBold, black)
	drawText(canvas, dCol4+10, dHeaderY+20, "Total", s.fontSmBold, black)

	y = dHeaderY + rowH
	for _, item := range dendaItems {
		drawHLine(canvas, dCol1, dCol5, y+rowH, lineColor)
		drawVLine(canvas, dCol1, y, y+rowH, lineColor)
		drawVLine(canvas, dCol2, y, y+rowH, lineColor)
		drawVLine(canvas, dCol3, y, y+rowH, lineColor)
		drawVLine(canvas, dCol4, y, y+rowH, lineColor)
		drawVLine(canvas, dCol5, y, y+rowH, lineColor)

		drawText(canvas, dCol1+10, y+20, item.Name, s.fontSmall, black)
		drawTextRight(canvas, dCol3-10, y+20, item.Intensif, s.fontSmall, black)
		drawText(canvas, dCol3+10, y+20, item.Qty, s.fontSmall, black)
		drawTextRight(canvas, dCol5-10, y+20, item.Total, s.fontSmall, black)
		y += rowH
	}

	if len(dendaItems) == 0 {
		drawHLine(canvas, dCol1, dCol5, y+rowH, lineColor)
		drawVLine(canvas, dCol1, y, y+rowH, lineColor)
		drawVLine(canvas, dCol5, y, y+rowH, lineColor)
		drawText(canvas, dCol1+10, y+20, "-", s.fontSmall, black)
		y += rowH
	}

	y += 40

	// ========== 5. TOTAL DITERIMA ==========
	drawText(canvas, leftMargin, y, "4.  Total Diterima", s.fontBold, black)
	y += 8

	// Gold box for total
	totalBoxLeft := leftMargin
	totalBoxRight := leftMargin + 600
	totalBoxH := 70 // Row 1 height
	
	// Top Header box
	drawRect(canvas, totalBoxLeft, y, totalBoxRight, y+totalBoxH, totalBg)
	drawHLine(canvas, totalBoxLeft, totalBoxRight, y, lineColor)
	drawVLine(canvas, totalBoxLeft, y, y+totalBoxH, lineColor)
	drawVLine(canvas, totalBoxRight, y, y+totalBoxH, lineColor)
	
	headerWidth := (&font.Drawer{Face: s.fontBold}).MeasureString("TOTAL PENDAPATAN") / 64
	textX := totalBoxLeft + (600 - int(headerWidth))/2
	drawText(canvas, textX, y+48, "TOTAL PENDAPATAN", s.fontBold, black)

	// Bottom Value box
	y += totalBoxH
	drawHLine(canvas, totalBoxLeft, totalBoxRight, y, lineColor)
	drawVLine(canvas, totalBoxLeft, y, y+totalBoxH, lineColor)
	drawVLine(canvas, totalBoxRight, y, y+totalBoxH, lineColor)
	drawHLine(canvas, totalBoxLeft, totalBoxRight, y+totalBoxH, lineColor)
	
	valStr := formatRp(tx.TakeHomePay)
	valWidth := (&font.Drawer{Face: s.fontBold}).MeasureString(valStr) / 64
	valX := totalBoxLeft + (600 - int(valWidth))/2
	drawText(canvas, valX, y+48, valStr, s.fontBold, black)
	
	y += 180 // Spacer for signature

	// ========== 6. STEMPEL + SIGNATURE ==========
	sigX := leftMargin + 60
	sigY := y - 30 
	
	drawText(canvas, sigX, sigY, "Pare, "+issueDate, s.fontSmall, black)

	stempelFile, err := os.Open(s.stempelPath)
	if err == nil {
		defer stempelFile.Close()
		stempelImg, _, err := image.Decode(stempelFile)
		if err == nil {
			targetW, targetH := 165, 168
			dstRect := image.Rect(0, 0, targetW, targetH)
			scaledStempel := image.NewRGBA(dstRect)
			xdraw.CatmullRom.Scale(scaledStempel, dstRect, stempelImg, stempelImg.Bounds(), draw.Over, nil)
			
			stX := sigX - 10 
			stY := sigY + 5 
			draw.Draw(canvas, image.Rect(stX, stY, stX+targetW, stY+targetH), scaledStempel, image.Point{}, draw.Over)
		}
	}

	sigY += 185 // accommodate the larger stamp size
	drawText(canvas, sigX, sigY, "Karismaning Ulfa Awwalina, S.Pd", s.fontBold, black)
	drawText(canvas, sigX, sigY+22, "Manager Keuangan Markaz Arabiyah", s.fontSmall, black)

	// ========== 7. ENCODE ==========
	var buf bytes.Buffer
	err = png.Encode(&buf, canvas)
	if err != nil {
		return nil, fmt.Errorf("failed to encode png: %w", err)
	}

	return buf.Bytes(), nil
}
