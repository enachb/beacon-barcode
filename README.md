## Install cups package
```
sudo apt-get install cups libcups2-dev
```

## Install Go

Google for instructions.

## Installing Brother PT-P700 Printer

* Hook up the label printer to the linux host over USB
* Install printer driver Brother PT-2420PC Foomatic/ptouch

    ```sudo lpadmin -p brother -v usb://Brother/PT-P700 -P ./Brother-PT-PC.ppd```
* Make it the default printer 

    ```lpoptions -d brother```

* Enable printer

    ```sudo cupsenable brother```

* Switch off the USB drive (left button)

## Compile

```
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

## Troubleshooting
* Printer not ready error message

    ```sudo cupsaccept brother```

