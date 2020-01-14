package main

import (
	"fmt"
	"image/jpeg"
	"os"
	"strings"

	"github.com/boombuler/barcode"
	bc "github.com/boombuler/barcode/code39"
	"github.com/boombuler/barcode/qr"
	"github.com/jung-kurt/gofpdf"
)

func makeCode39(text string) string {
	filename := "barcode.jpg"
	// Create the barcode
	bCode, err := bc.Encode(text, true, true)
	if err != nil {
		panic(err)
	}

	// Scale the barcode to 200x200 pixels
	bCode2, err := barcode.Scale(bCode, 142, 5)

	if err != nil {
		panic(err)
	}

	// create the output file
	file, _ := os.Create(filename)
	defer file.Close()

	// encode the barcode as png
	jpeg.Encode(file, bCode2,
		&jpeg.Options{
			Quality: 100,
		})
	return filename
}

func makeQR(text string) string {
	filename := "qrcode.jpg"
	// Create the barcode
	bCode, err := qr.Encode(text, qr.L, qr.Auto)
	if err != nil {
		panic(err)
	}

	// Scale the barcode to 200x200 pixels
	bCode2, err := barcode.Scale(bCode, 140, 140)

	if err != nil {
		panic(err)
	}

	// create the output file
	file, _ := os.Create(filename)
	defer file.Close()

	// encode the barcode as png
	jpeg.Encode(file, bCode2,
		&jpeg.Options{
			Quality: 100,
		})
	return filename
}

func main() {

	text := strings.Replace("AA:FF:DD:99", ":", "", -1)
	fmt.Printf("Text %s\n", text)

	filenameBC := makeCode39(text)
	filenameQR := makeQR(text)

	pdf := gofpdf.NewCustom(
		&gofpdf.InitType{
			UnitStr:        "mm",
			Size:           gofpdf.SizeType{Wd: 12, Ht: 80},
			FontDirStr:     "",
			OrientationStr: "L",
		})
	//	pdf.AddPage()
	pdf.SetFont("Arial", "B", 8)

	// CellFormat(width, height, text, border, position after, align, fill, link, linkStr)
	pdf.CellFormat(0, 0, text, "0", 0, "LB", false, 0, "")

	// ImageOptions(src, x, y, width, height, flow, options, link, linkStr)
	pdf.ImageOptions(
		filenameBC,
		1, 1,
		35, 5,
		false,
		gofpdf.ImageOptions{ImageType: "JPG", ReadDpi: true},
		0,
		"",
	)

	pdf.ImageOptions(
		filenameQR,
		70, 1,
		10, 10,
		false,
		gofpdf.ImageOptions{ImageType: "JPG", ReadDpi: true},
		0,
		"",
	)

	err := pdf.OutputFileAndClose("oink.pdf")
	if err != nil {
		panic(err)
	}
}
