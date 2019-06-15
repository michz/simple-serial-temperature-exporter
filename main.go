package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

var currentTemperature float64 = 9999.9
var outputMetricName string = "temperature"
var staticLabels string = ""

func httpRequestHandler(w http.ResponseWriter, r *http.Request) {
	if currentTemperature < 9999 {
		fmt.Fprintf(w, "%s{%s} %f", outputMetricName, staticLabels, currentTemperature)
	}
}

func httpHealthzRequestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func main() {
	// Command line parameters
	httpPort := flag.Int("httpPort", 27164, "Port to listen on for HTTP requests")
	serialPort := flag.String("serialPort", "/dev/ttyUSB0", "Serial port to get temperature from")
	metricName := flag.String("metricName", "temperature", "Name for the Prometheus metric")
	labels := flag.String("labels", "", "Static labels that should be appended to the prometheus metric")

	flag.Parse()
	outputMetricName = *metricName
	staticLabels = *labels

	// Start webserver concurrently
	go mainHttp(*httpPort)

	// Set up options.
	options := serial.OpenOptions{
		PortName:        *serialPort,
		BaudRate:        9600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 2,
	}

	f, err := serial.Open(options)

	if err != nil {
		fmt.Printf("Could not open serial port: %s\n", err)
		os.Exit(3)
	}

	defer f.Close()

	lineBuf := bytes.NewBuffer(make([]uint8, 256))
	rxBuf := make([]byte, 256)
	for {
		n, err := f.Read(rxBuf)
		if err != nil {
			fmt.Printf("Could not read from serial port: %s\n", err)
			os.Exit(4)
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

func mainHttp(httpPort int) {
	http.HandleFunc("/export", httpRequestHandler)
	http.HandleFunc("/healthz", httpHealthzRequestHandler)
	log.Fatal(http.ListenAndServe(":"+strconv.FormatInt(int64(httpPort), 10), nil))
}
