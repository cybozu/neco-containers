package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"regexp"
	"sync"
	"time"
)

var redfishPath string = "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"

// ID & password for basic authentication
const (
	basicAuthUser     = "support"
	basicAuthPassword = "raw password for support user"
)

type bmcMock struct {
	host          string
	resDir        string
	files         []string
	accessCounter map[string]int
	responseFiles map[string][]string
	responseDir   map[string]string
	isInitmap     bool
	mutex         sync.Mutex
}

// Mock server of iDRAC
func (b *bmcMock) startMock() {
	b.mutex.Lock()
	b.accessCounter[b.host] = 0
	b.responseFiles[b.host] = b.files
	b.responseDir[b.host] = b.resDir
	b.mutex.Unlock()

	server := http.NewServeMux()
	server.HandleFunc(redfishPath, b.redfishSel)
	go func() {
		slog.Error("error at ListenAndServeTLS", "err", http.ListenAndServeTLS(b.host, "testdata/ssl/localhost.crt", "testdata/ssl/localhost.key", server))
	}()
}

// DELL System Event Log Service at Redfish REST
func (b *bmcMock) redfishSel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json;odata.metadata=minimal;charset=utf-8")
	// Basic authentication
	if user, pass, ok := r.BasicAuth(); !ok || user != basicAuthUser || pass != basicAuthPassword {
		w.Header().Add("WWW-Authenticate", `Basic realm="my private area"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Exclusive lock against other mock server which parallel running
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Check a response file is available
	key := string(r.Host)
	if b.accessCounter[key] > (len(b.responseFiles[key]) - 1) {
		time.Sleep(3 * time.Second)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		fmt.Println("error accessCounter[key]", b.accessCounter[key], key, r)
		return
	}

	fn := b.responseFiles[key][b.accessCounter[key]]
	responseFile := path.Join(b.responseDir[key], fn)
	b.accessCounter[key] = b.accessCounter[key] + 1
	fmt.Println("accessCounter[key]", b.accessCounter[key], key, r)

	// Create HTTP response from the response file
	fd, err := os.Open(responseFile)
	if err != nil {
		// Create not found response
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer fd.Close()
	// BMC working time
	time.Sleep(1 * time.Second)

	// Reply
	stringJSON, _ := io.ReadAll(fd)
	fmt.Fprint(w, string(stringJSON))
}

// Method for Test
func OpenTestResultLog(fn string) (*os.File, error) {
	var file *os.File
	var err error
	for {
		file, err = os.Open(fn)
		if errors.Is(err, os.ErrNotExist) {
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}
	return file, err
}

// Method for Test
func ReadingTestResultLogNext(b *bufio.Reader) (string, error) {
	var stringJSON string
	var err error
	for i := 0; i < 20; i++ {
		stringJSON, err = b.ReadString('\n')
		if err == io.EOF {
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	return stringJSON, err
}

type logTest struct {
	outputDir string
}

func (l logTest) write(byteJson string, serial string) error {
	fn := path.Join(l.outputDir, serial)
	fd, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer fd.Close()
	_, err = fd.WriteString(fmt.Sprintln(string(byteJson)))
	if err != nil {
		return err
	}
	fmt.Println(byteJson)
	return nil
}

func searchMetricsComment(lines []string, keyword string) bool {
	pattern := "^" + keyword
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	for _, line := range lines {
		matches := re.FindAllString(line, -1)
		if len(matches) > 0 {
			return true
		}
	}
	return false
}

func findMetrics(lines []string, keyword string) (string, error) {
	re, err := regexp.Compile(keyword)
	if err != nil {
		return "", err
	}
	for _, line := range lines {
		matches := re.FindAllString(line, -1)
		if len(matches) > 0 {
			return line + "\n", nil
		}
	}
	return "", fmt.Errorf("not Found %v", keyword)
}
