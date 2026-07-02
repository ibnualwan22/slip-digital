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

	"github.com/ibnualwan/bisyaroh/internal/domain"
	"golang.org/x/image/font"
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
	fontFace    font.Face
    fontFaceBold font.Face
}

func NewSlipGenerator(bgPath, stempelPath string) (SlipGenerator, error) {
	// Parse TTF font
	f, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    18,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create font face: %w", err)
	}
    
    faceBold, _ := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    18,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	return &slipGenerator{
		bgPath:      bgPath,
		stempelPath: stempelPath,
		fontFace:    face,
        fontFaceBold: faceBold,
	}, nil
}

func addLabel(img draw.Image, x, y int, label string, face font.Face, col color.Color) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	d.DrawString(label)
}

func (s *slipGenerator) GenerateSlipImage(tx *domain.PayrollTransaction, issueDate string) ([]byte, error) {
	// 1. Load Background Template
	bgFile, err := os.Open(s.bgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open background image: %w", err)
	}
	defer bgFile.Close()

	bgImg, _, err := image.Decode(bgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode background image: %w", err)
	}

	bounds := bgImg.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, bgImg, image.Point{}, draw.Src)

	// 2. Draw Text (Employee Data)
    // Approximate coordinates
	textColor := color.Black

	addLabel(rgba, 150, 250, "Nama Pegawai : "+tx.Employee.Name, s.fontFaceBold, textColor)
	addLabel(rgba, 150, 280, "Duf'ah       : -", s.fontFace, textColor)
    addLabel(rgba, 500, 250, "Jabatan      : "+tx.Employee.Role, s.fontFaceBold, textColor)
	addLabel(rgba, 500, 280, "Kategori     : ", s.fontFace, textColor)
    if tx.Employee.Category != nil {
        addLabel(rgba, 600, 280, tx.Employee.Category.Name, s.fontFace, textColor)
    }

	// 3. Draw Additions (Pendapatan)
	yPos := 360
	addLabel(rgba, 100, yPos, "1. Rincian Pendapatan", s.fontFaceBold, textColor)
	yPos += 30

	for _, det := range tx.Details {
		if det.Type == "ADDITION" {
			name := "Bonus/Tunjangan"
			if det.Activity != nil {
				name = det.Activity.ActivityName
			}
			addLabel(rgba, 150, yPos, name, s.fontFace, textColor)
            amountStr := det.TotalAmount.StringFixed(0)
			addLabel(rgba, 650, yPos, "Rp. "+amountStr, s.fontFace, textColor)
			yPos += 25
		}
	}
    yPos += 15
    addLabel(rgba, 150, yPos, "------------- TOTAL PENDAPATAN -------------", s.fontFace, textColor)
    addLabel(rgba, 650, yPos, "Rp. "+tx.GrossIncome.StringFixed(0), s.fontFaceBold, textColor)
    yPos += 45

	// 4. Draw Deductions (Potongan)
	addLabel(rgba, 100, yPos, "2. Rincian Potongan", s.fontFaceBold, textColor)
	yPos += 30

	for _, det := range tx.Details {
		if det.Type == "DEDUCTION" {
			name := "Potongan/Denda"
			if det.Activity != nil {
				name = det.Activity.ActivityName
			}
			addLabel(rgba, 150, yPos, name, s.fontFace, textColor)
            amountStr := det.TotalAmount.StringFixed(0)
			addLabel(rgba, 650, yPos, "Rp. "+amountStr, s.fontFace, textColor)
			yPos += 25
		}
	}
    
    yPos += 15
    addLabel(rgba, 150, yPos, "------------- TOTAL POTONGAN -------------", s.fontFace, textColor)
    addLabel(rgba, 650, yPos, "Rp. "+tx.TotalDeductions.StringFixed(0), s.fontFaceBold, textColor)
    yPos += 60

	// 5. Take Home Pay
	addLabel(rgba, 100, yPos, "3. Take Home Pay (Diterima)", s.fontFaceBold, textColor)
	addLabel(rgba, 650, yPos, "Rp. "+tx.TakeHomePay.StringFixed(0), s.fontFaceBold, textColor)

	// 6. Signatures (Bottom Right)
	sigY := bounds.Max.Y - 200
	addLabel(rgba, 600, sigY, "Pare, "+issueDate, s.fontFace, textColor)
	sigY += 100
	addLabel(rgba, 600, sigY, "Karismaning Ulfa Awalina, S.Pd", s.fontFaceBold, textColor)
	addLabel(rgba, 600, sigY+25, "Manager Keuangan Markaz Arabiyah", s.fontFace, textColor)

	// 7. Render as PNG to byte array
	var buf bytes.Buffer
	err = png.Encode(&buf, rgba)
	if err != nil {
		return nil, fmt.Errorf("failed to encode png: %w", err)
	}

	return buf.Bytes(), nil
}
