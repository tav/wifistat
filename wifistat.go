// Public Domain (-) 2012 The Wifistat Authors.
// See the Wifistat UNLICENSE file for details.

package main

import (
	"amp/log"
	"amp/optparse"
	"amp/runtime"
	"bufio"
	"bytes"
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

const (
	csvUrlPrefix = "https://docs.google.com/spreadsheet/pub?key="
	csvUrlSuffix = "&single=true&gid=0&output=csv"
)

var (
	csvDir                string
	devicesUrlKey         string
	enableMemberAnalytics bool
	membersUrlKey         string
	openingUrlKey         string
	parseLock             sync.RWMutex
	parsedAlready         bool
	wifiLogDir            string
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
			if i >= 0 {
				break
			}
		}
		break
	}

	log.Info("Finished parsing (%s)", time.Since(startTime))

}

func getCsvFile(name, urlpath string, timestamp time.Time, force bool) (*csv.Reader, error) {
	filename := filepath.Join(csvDir, name+timestamp.Format(".2006-01-02")+".csv")
	if !force {
		file, err := os.Open(filename)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		} else {
			return csv.NewReader(file), nil
		}
	}
	log.Info("Downloading %s.csv from Google Spreadsheets", name)
	resp, err := http.Get(csvUrlPrefix + urlpath + csvUrlSuffix)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(filename, body, 0644)
	if err != nil {
		return nil, err
	}
	return csv.NewReader(bytes.NewBuffer(body)), nil
}

func parseCsv(force bool) {
	now := time.Now()
	_, err := getCsvFile("devices", devicesUrlKey, now, force)
	if err != nil {
		runtime.Error("Couldn't load devices.csv: %s", err)
		return
	}
	_, err = getCsvFile("members", membersUrlKey, now, force)
	if err != nil {
		runtime.Error("Couldn't load members.csv: %s", err)
		return
	}
	_, err = getCsvFile("opening", openingUrlKey, now, force)
	if err != nil {
		runtime.Error("Couldn't load opening.csv: %s", err)
		return
	}
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

	membership := opts.BoolConfig("member-analytics", false,
		"enable membership-based analytics")

	devices := opts.StringConfig("devices-url", "",
		"the url key for the devices.csv Google Spreadsheet")

	members := opts.StringConfig("members-url", "",
		"the url key for the members.csv Google Spreadsheet")

	opening := opts.StringConfig("opening-url", "",
		"the url key for the opening.csv Google Spreadsheet")

	// Parse the command line options.
	os.Args[0] = "wifistat"
	_, root, _ := runtime.DefaultOpts("wifistat", opts, os.Args)

	// Compute option variables.
	wifiLogDir = runtime.JoinPath(root, *wifi)
	csvDir = runtime.JoinPath(root, *csv)
	err := os.MkdirAll(csvDir, 0755)
	if err != nil {
		runtime.StandardError(err)
	}

	// Handle member analytics options.
	if *membership {
		enableMemberAnalytics = true
		devicesUrlKey = *devices
		if devicesUrlKey == "" {
			runtime.Error("You need to specify the `devices-url` command-line option.")
		}
		membersUrlKey = *members
		if membersUrlKey == "" {
			runtime.Error("You need to specify the `members-url` command-line option.")
		}
		openingUrlKey = *opening
		if openingUrlKey == "" {
			runtime.Error("You need to specify the `opening-url` command-line option.")
		}
	}

	// Parse the logs.
	parseCsv(false)
	parseWifi()

	// Register the various handlers.
	http.HandleFunc("/", handleRequest)
	http.HandleFunc("/reload", handleReload)

	// Start the web server.
	log.Info("Running wifistat on %s", *addr)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		runtime.StandardError(err)
	}

	runtime.Exit(0)

}
