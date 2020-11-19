package main

// Install cups package
// sudo apt-get install libcups2-dev

import (
	"bytes"
	"flag"
	"image/jpeg"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/jung-kurt/gofpdf"
	log "github.com/sirupsen/logrus"

	stdlog "log"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"

	"github.com/boombuler/barcode"

	bc "github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/qr"
)

var isPoweredOn = false
var scanMutex = sync.Mutex{}
var bleAddr string

var dryRun *bool
var printQRCode *bool
var printBarcode *bool
var minRSSI *int

func beginScan(d gatt.Device) {
	scanMutex.Lock()
	for isPoweredOn {
		d.Scan(nil, true) //Scan for five seconds and then restart
		time.Sleep(5 * time.Hour)
		d.StopScanning()
	}
	scanMutex.Unlock()
}

func onStateChanged(d gatt.Device, s gatt.State) {
	log.Debugf("State: %s", s)
	switch s {
	case gatt.StatePoweredOn:
		log.Debug("scanning...")
		isPoweredOn = true
		go beginScan(d)
		return
	case gatt.StatePoweredOff:
		log.Debug("REINIT ON POWER OFF")
		isPoweredOn = false
		d.Init(onStateChanged)
	default:
		log.Debugf("WARN: unhandled state: %s", string(s))
	}
}

func insertNth(s string, n int) string {
	var buffer bytes.Buffer
	var n_1 = n - 1
	var l_1 = len(s) - 1
	for i, rune := range s {
		buffer.WriteRune(rune)
		if i%n == n_1 && i != l_1 {
			buffer.WriteRune('\n')
		}
	}
	return buffer.String()
}

func createBarcode(addr string, rssi int, printBarcode bool, printQRCode bool) string {

	const fileName = "/tmp/oink.pdf"

	text := strings.ToLower(strings.Replace(addr, ":", "", -1))
	log.Infof("Found new beacon: %s rssi:%d", addr, rssi)

	filenameQR := makeQR(text)

	// Compute length based on chosen codes
	var length float64 = 0
	if printQRCode {
		length += 15
	}

	if printBarcode {
		length += 43
	}

	pdf := gofpdf.NewCustom(
		&gofpdf.InitType{
			UnitStr:        "mm",
			Size:           gofpdf.SizeType{Wd: 12, Ht: 48},
			FontDirStr:     "",
			OrientationStr: "L",
		})
	pdf.SetFont("Helvetica", "", 11)

	space := "             "
	formattedText := insertNth(text, 2)
	// CellFormat(width, height, text, border, position after, align, fill, link, linkStr)
	pdf.CellFormat(0, -2, space+formattedText, "0", 0, "LB", false, 0, "")
	// pdf.CellFormat(0, -0, space+formattedText[6:], "0", 0, "LB", false, 0, "")
	// pdf.MultiCell(10, -6, "             "+insertNth(text, 2), "0", "LB", false)

	// QRCode
	if printQRCode {
		log.Infof("Filename QR: %s", filenameQR)
		pdf.ImageOptions(
			filenameQR,
			1, 1,
			10, 10,
			false,
			gofpdf.ImageOptions{ImageType: "JPG", ReadDpi: true},
			0,
			"",
		)
	}

	err := pdf.OutputFileAndClose(fileName)
	if err != nil {
		log.Fatal(err)
	}

	return fileName
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	log.Infof("Peripheral ID:%s, NAME:(%s) rssi (%d)", p.ID(), p.Name(), rssi)
	if p.ID() == bleAddr && bleAddr != "" {
		log.Info("Skipping previously printed beacon....")
		return
	}
	// only print the really close ones
	if rssi > *minRSSI && strings.Contains(p.Name(), "Google") {
		bleAddr = p.ID()

		fileName := createBarcode(p.ID(), rssi, *printBarcode, *printQRCode)

		if !*dryRun {
			cmd := exec.Command("lp", fileName)
			var out bytes.Buffer
			cmd.Stdout = &out
			err := cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
			log.Infof("**************** Done printing barcode %s", out.String())
			if err != nil {
				panic(err)
			}
		} else {
			log.Info("Dry run - skipping printing.")
		}
	}
}

func makeCode39(text string) string {
	filename := "/tmp/barcode.jpg"
	// Create the barcode
	// Code39
	//bCode, err := bc.Encode(text, false, false)

	// Code128
	bCode, err := bc.Encode(text) //, true, false)

	if err != nil {
		panic(err)
	}

	// Scale the barcode to 200x200 pixels
	bCode2, err := barcode.Scale(bCode, 1000, 100)
	if err != nil {
		panic(err)
	}

	// Crop it
	// bCode2, err := cutter.Crop(bCode2, cutter.Config{
	// 	Width:  350,
	// 	Height: 5,
	// 	Mode:   cutter.Centered, // optional, default value
	// })
	// if err != nil {
	// 	panic(err)
	// }

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
	bCode2, err := barcode.Scale(bCode, 800, 800)

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

	dryRun = flag.Bool("n", false, "Dry run - don't print")
	printQRCode = flag.Bool("q", false, "Print QR Code")
	printBarcode = flag.Bool("b", false, "Print barode")
	minRSSI = flag.Int("r", -20, "Minimun RSSI required for printing.")
	displayHelp := flag.Bool("h", false, "Display this help.")
	flag.Parse()

	if *displayHelp {
		flag.PrintDefaults()
		return
	}

	// Supress extraneous log output from GATT
	stdlog.SetOutput(ioutil.Discard)

	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	log.SetFormatter(&log.TextFormatter{})

	d, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}

	// Register handlers.
	d.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	d.Init(onStateChanged)
	log.Info("Done setting up Bluetooth handlers.")

	// waiting
	select {}
}
