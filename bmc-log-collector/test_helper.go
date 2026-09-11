package main

import (
	"bufio"
	"encoding/json"
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

var (
	redfishPath   string = "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
	redfishLcPath string = "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Lclog/Entries"
)

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

	// Lifecycle log mock: each file in lcFiles is a whole LC log snapshot
	// (newest first) used for one scraping cycle. The handler serves only the
	// newest lcPageSize entries of it, as the real iDRAC returns only the
	// latest page of the LC log.
	lcFiles    []string
	lcPageSize int
	lcCounter  int
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
	server.HandleFunc(redfishLcPath, b.redfishLclog)
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

// DELL Lifecycle Log Service at Redfish REST.
// Each request serves the latest page of the next snapshot file.
func (b *bmcMock) redfishLclog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json;odata.metadata=minimal;charset=utf-8")
	// Basic authentication
	if user, pass, ok := r.BasicAuth(); !ok || user != basicAuthUser || pass != basicAuthPassword {
		w.Header().Add("WWW-Authenticate", `Basic realm="my private area"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	idx := b.lcCounter
	if idx > len(b.lcFiles)-1 {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	b.lcCounter = idx + 1

	fd, err := os.Open(path.Join(b.resDir, b.lcFiles[idx]))
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer fd.Close()

	var snapshot struct {
		Members []json.RawMessage `json:"Members"`
	}
	if err := json.NewDecoder(fd).Decode(&snapshot); err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	pageSize := b.lcPageSize
	if pageSize == 0 {
		pageSize = 3
	}
	response := map[string]any{
		"Members@odata.count": len(snapshot.Members),
		"Members":             snapshot.Members[:min(pageSize, len(snapshot.Members))],
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Println("failed to encode the LC log response", err)
	}
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
	for range 20 {
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
	fd, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		return err
	}
	defer fd.Close()
	_, err = fmt.Fprintln(fd, byteJson)
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

// failingLogWriter always fails; for testing the write-failure handling
type failingLogWriter struct{}

func (f failingLogWriter) write(stringJson string, serial string) error {
	return errors.New("write failed for test")
}
