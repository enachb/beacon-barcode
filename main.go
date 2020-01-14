package main

import (
	"image/jpeg"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jung-kurt/gofpdf"
	log "github.com/sirupsen/logrus"

	stdlog "log"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"

	"github.com/boombuler/barcode"
	bc "github.com/boombuler/barcode/code39"
	"github.com/boombuler/barcode/qr"
)

var isPoweredOn = false
var scanMutex = sync.Mutex{}

func beginScan(d gatt.Device) {
	scanMutex.Lock()
	for isPoweredOn {
		d.Scan(nil, true) //Scan for five seconds and then restart
		time.Sleep(5 * time.Second)
		d.StopScanning()
	}
	scanMutex.Unlock()
}

func onStateChanged(d gatt.Device, s gatt.State) {
	log.Debugf("State:", s)
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
		log.Debugf("WARN: unhandled state: ", string(s))
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	log.Debugf("Peripheral ID:%s, NAME:(%s) power (%d)", p.ID(), p.Name(), a.TxPowerLevel)
	if rssi < -30 {

		text := strings.Replace(p.ID(), ":", "", -1)
		log.Infof("Found new beacon: %s rssi:%d", p.ID(), rssi)

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
		log.Info("Done printing barcode")
		if err != nil {
			panic(err)
		}
	}
}

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

	// Supress extraneous log output from GATT
	stdlog.SetOutput(ioutil.Discard)

	//dataChan = make(chan string, 1)

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

	//	text := strings.Replace("AA:FF:DD:99", ":", "", -1)
	//fmt.Printf("Text %s\n", text)

	// waiting
	select {}
}
