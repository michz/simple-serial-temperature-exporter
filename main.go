package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

// @TODO Variablennamen konfigurierbar via CLI
// @TODO HTTP Port konfigurierbar via CLI

var currentTemperature float64 = 9999.9

func httpRequestHandler(w http.ResponseWriter, r *http.Request) {
	if currentTemperature < 9999 {
		fmt.Fprintf(w, "temperature{} %f", currentTemperature)
	}
}

func httpHealthzRequestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func main() {
	// Start webserver concurrently
	go mainHttp()

	// Set up options.
	options := serial.OpenOptions{
		PortName:        "/dev/ttyUSB0",
		BaudRate:        9600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 2,
	}

	f, err := serial.Open(options)

	if err != nil {
		fmt.Printf("Could not open serial port: %s\n", err)
		os.Exit(1)
	}

	defer f.Close()

	lineBuf := bytes.NewBuffer(make([]uint8, 256))
	rxBuf := make([]byte, 256)
	for {
		n, err := f.Read(rxBuf)
		if err != nil {
			fmt.Printf("Could not read from serial port: %s\n", err)
			os.Exit(2)
		}

		if n > 0 {
			lineBuf.Write(rxBuf[:n])
			//fmt.Printf("countBytes: %d\n", countBytes)
		}

		var lineFound bool = false
		for _, b := range rxBuf[:n] {
			if b == '\n' {
				lineFound = true
				break
			}
		}

		if lineFound {
			line, _ := lineBuf.ReadString('\n')
			line = strings.TrimSpace(line)
			line = strings.Trim(line, "0\x00")
			currentTemperature, _ = strconv.ParseFloat(line, 64)
		}

		// Only sleep if no character was received
		if n == 0 {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func mainHttp() {
	http.HandleFunc("/export", httpRequestHandler)
	http.HandleFunc("/healthz", httpHealthzRequestHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
