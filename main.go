package main

// Install cups package
// sudo apt-get install libcups2-dev

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	stdlog "log"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"

	"github.com/boombuler/barcode"

	"github.com/boombuler/barcode/qr"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
)

var isPoweredOn = false
var scanMutex = sync.Mutex{}
var bleAddr string

var dryRun *bool
var minRSSI *int
var face font.Face

var labels chan string

var scanned map[string]int64

func insertNth(s string, n int) string {
	var buffer bytes.Buffer
	var n_1 = n - 1
	var l_1 = len(s) - 1
	for i, rune := range s {
		buffer.WriteRune(rune)
		if i%n == n_1 && i != l_1 {
			buffer.WriteRune(' ')
		}
	}
	return buffer.String()
}

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

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {

	if !strings.Contains(p.Name(), "Google") {
		return
	}

	log.Infof("Peripheral ID:%s, NAME:(%s) rssi (%d)", p.ID(), p.Name(), rssi)

	// only print the really close ones
	if rssi > *minRSSI {
		text := strings.ToLower(strings.Replace(p.ID(), ":", "", -1))
		_, exists := scanned[text]
		if exists {
			log.Infof("Skipping previously printed beacon: %s", text)
			return
		}
		log.Infof("Found new beacon: %s rssi:%d", p.ID(), rssi)

		img := gg.NewContext(320, 76)
		img.SetRGB(1, 1, 1)
		img.Clear()
		qrImage := makeQR(text)
		img.DrawImage(qrImage, 0, 0)
		img.SetFontFace(face)
		img.SetRGB(0, 0, 0)
		img.DrawString(insertNth(text, 2), 80, 48)

		imgPath := fmt.Sprintf("/tmp/%s.png", text)
		img.SavePNG(imgPath)

		scanned[text] = time.Now().Unix()

		if !*dryRun {
			labels <- imgPath
		} else {
			log.Info("Dry run - skipping printing.")
		}

	}
}

func makeQR(text string) *image.Gray {
	// filename := "qrcode.png"
	// Create the barcode
	bCode, err := qr.Encode(text, qr.L, qr.Auto)
	if err != nil {
		panic(err)
	}

	// Scale the barcode to 200x200 pixels
	bCode2, err := barcode.Scale(bCode, 76, 76)
	if err != nil {
		panic(err)
	}

	bounds := bCode2.Bounds()
	gray := image.NewGray(bounds)
	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			var rgba = bCode2.At(x, y)
			gray.Set(x, y, rgba)
		}
	}

	return gray
}

func main() {

	dryRun = flag.Bool("n", false, "Dry run - don't print")
	minRSSI = flag.Int("r", -20, "Minimun RSSI required for printing.")
	flag.Parse()

	labels = make(chan string, 100)
	scanned = make(map[string]int64)

	// Supress extraneous log output from GATT
	stdlog.SetOutput(ioutil.Discard)

	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	log.SetFormatter(&log.TextFormatter{})

	font, err := truetype.Parse(gobold.TTF)
	if err != nil {
		log.Fatal(err)
	}

	face = truetype.NewFace(font,
		&truetype.Options{
			Size: 11,
			DPI:  180,
		},
	)

	d, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}

	// Register handlers.
	d.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	d.Init(onStateChanged)
	log.Info("Done setting up Bluetooth handlers.")

	// Printing
	for l := range labels {
		//Print
		if !*dryRun {
			cmd := exec.Command("./ptouch-print-x86", "--image", l)
			var out bytes.Buffer
			cmd.Stdout = &out
			err := cmd.Run()
			// if err != nil {
			// 	log.Fatal(err)
			// }
			err = os.Remove(l)
			if err != nil {
				log.Fatal(err)
			}
			log.Infof("Done printing barcode %s", out.String())
			if err != nil {
				panic(err)
			}
		} else {
			log.Info("Dry run - skipping printing.")
		}

	}
}
