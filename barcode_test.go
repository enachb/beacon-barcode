package main

import (
	"testing"
)

//TestQRCode Test print qrcode
func TestQRCode(*testing.T) {

	createBarcode("aabbccddeeff", -12, false, true)
}
