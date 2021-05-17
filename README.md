## Install cups package
```
sudo apt-get install cups libcups2-dev
```

## Install Go

Google for instructions.

## Installing Brother PT-P700 Printer

* Hook up the label printer to the linux host over USB
* Switch off the USB drive (left button)

## Allowing user access to USB devices
```
sudo bash -c "echo 'SUBSYSTEM==\"usb\", ATTR{idVendor}==\"04f9\", ATTR{idProduct}==\"2061\", GROUP=\"plugdev\", MODE=\"0666\"'>/etc/udev/rules.d/00-usb-permissions.rules"
sudo usermod -a -G plugdev pi
sudo service udev restart
sudo udevadm control --reload-rules
```

## Compile

```
go get -u github.com/fogleman/gg
go build
```

## Run

```
sudo ./beacon-barcode
```
Use  ```-h```  to display command line flags.

## Autostart
Edit the autostart config ``sudo /etc/rc.local``. Add this line before ```exit 0```:

```nohup /home/pi/src/beacon-barcode &```

