// Public Domain (-) 2012 The Wifistat Authors.
// See the Wifistat UNLICENSE file for details.

package main

import (
	"amp/log"
	"amp/optparse"
	"amp/runtime"
	"bufio"
	"encoding/csv"
	// "fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	csvDir        string
	parseLock     sync.RWMutex
	parsedAlready bool
	wifiLogDir    string
)

func parseWifi() {

	if parsedAlready {
		log.Info("Re-parsing logs: %s", wifiLogDir)
	} else {
		log.Info("Parsing logs: %s", wifiLogDir)
		parsedAlready = true
	}

	files, err := ioutil.ReadDir(wifiLogDir)
	if err != nil {
		log.StandardError(err)
		return
	}

	if len(files) == 0 {
		log.Error("No files found to parse in %s", wifiLogDir)
		return
	}

	var (
		c, l     int
		current  []byte
		file     *os.File
		filename string
		isPrefix bool
		key      string
		line     []byte
		reader   *bufio.Reader
		received int64
		sent     int64
		session  string
		split    []string
		status   string
		// pending  map[string]string
	)

	startTime := time.Now()

	var i int

	for _, info := range files {
		filename = filepath.Join(wifiLogDir, info.Name())
		log.Info("Parsing: %s", filename)
		file, err = os.Open(filename)
		if err != nil {
			log.StandardError(err)
			return
		}
		reader = bufio.NewReader(file)
		for {
			line, isPrefix, err = reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.StandardError(err)
				return
			}
			current = append(current, line...)
			if isPrefix {
				continue
			}
			split = strings.Split(string(current), ",")
			l = len(split)
			c = 6
			session = ""
			status = ""
			received = 0
			sent = 0
			for c < l {
				key = split[c]
				c += 2
				switch key {
				case "40":
					status = split[c+1]
				case "42":
					received, _ = strconv.ParseInt(split[c+1], 10, 64)
				case "43":
					sent, _ = strconv.ParseInt(split[c+1], 10, 64)
				case "44":
					session = split[c+1]
				}
				// log.Info("key: %s", key)
			}
			_ = session
			_ = status
			_ = received
			_ = sent
			current = current[0:0]
			i += 1
			// if i >= 0 {
			// 	break
			// }
		}
		// break
	}

	log.Info("Finished parsing (%s)", time.Since(startTime))

}

func getCsvFile(name, urlpath string, force bool, timestamp time.Time) (*csv.Reader, error) {
	filename := filepath.Join(csvDir, name+timestamp.Format(".2006-01-02")+".csv")
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return csv.NewReader(file), nil
}

func parseCsv(force bool) {

}

func handleRequest(w http.ResponseWriter, req *http.Request) {

	parseLock.RLock()
	defer parseLock.RUnlock()

	io.WriteString(w, "hello, world!\n")

}

func handleReload(w http.ResponseWriter, r *http.Request) {

	parseLock.Lock()
	defer parseLock.Unlock()

	if r.FormValue("csv") == "1" {
		parseCsv(true)
	}

	parseWifi()
	http.Redirect(w, r, "/", http.StatusFound)

}

func main() {

	// Define the options for the command line and config file options parser.
	opts := optparse.Parser(
		"Usage: wifistat <config.yaml> [options]\n",
		"wifistat 0.0.1")

	addr := opts.StringConfig("addr", ":9040",
		"the host:port address for the web server [:9040]")

	csv := opts.StringConfig("csv-dir", "csv",
		"the path to the csv files directory [csv]")

	wifi := opts.StringConfig("wifi-logs-dir", "iaslogs",
		"the path to the Wi-Fi logs directory [iaslogs]")

	// Parse the command line options.
	os.Args[0] = "wifistat"
	_, root, _ := runtime.DefaultOpts("wifistat", opts, os.Args)

	// Parse the logs.
	csvDir = runtime.JoinPath(root, *csv)
	wifiLogDir = runtime.JoinPath(root, *wifi)
	parseCsv(false)
	parseWifi()

	// Register the various handlers.
	http.HandleFunc("/", handleRequest)
	http.HandleFunc("/reload", handleReload)

	// Start the web server.
	log.Info("Running wifistat on %s", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		runtime.StandardError(err)
	}

	runtime.Exit(0)

}
