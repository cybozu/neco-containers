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
	"sync"
	"time"
)

var redfishPath string = "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"

// access counter foreach web-server
var accessCounter map[string]int
var responseFiles map[string][]string
var responseDir map[string]string
var isInitmap bool = false
var mutex sync.Mutex

// id & password for basic authentication
const (
	basicAuthUser     = "user"
	basicAuthPassword = "pass"
)

func init_map() {
	accessCounter = make(map[string]int)
	responseFiles = make(map[string][]string)
	responseDir = make(map[string]string)
	isInitmap = true
}

type bmcMock struct {
	host   string
	resDir string
	files  []string
}

// Mock server of iDRAC
func (b *bmcMock) startMock() {
	if !isInitmap {
		init_map()
	}
	mutex.Lock()
	accessCounter[b.host] = 0
	responseFiles[b.host] = b.files
	responseDir[b.host] = b.resDir
	mutex.Unlock()

	server := http.NewServeMux()
	server.HandleFunc(redfishPath, redfishSel)
	go func() {
		http.ListenAndServeTLS(b.host, "testdata/ssl/localhost.crt", "testdata/ssl/localhost.key", server)
	}()
}

// DELL System Event Log Service at Redfish REST
func redfishSel(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json;odata.metadata=minimal;charset=utf-8")
	// basic authentication
	if user, pass, ok := r.BasicAuth(); !ok || user != basicAuthUser || pass != basicAuthPassword {
		w.Header().Add("WWW-Authenticate", `Basic realm="my private area"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()
	key := string(r.Host)
	fmt.Println("------------------------------------- ", key, " =============== ", accessCounter[key], "---", len(responseFiles[key]))
	if accessCounter[key] > (len(responseFiles[key]) - 1) {
		time.Sleep(3 * time.Second)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// Exclusive lock against other mock server which parallel running
	//mutex.Lock()

	fn := responseFiles[key][accessCounter[key]]
	responseFile := path.Join(responseDir[key], fn)
	accessCounter[key] = accessCounter[key] + 1
	//mutex.Unlock()

	// Create HTTP response from the response file
	file, err := os.Open(responseFile)
	if err != nil {
		// create not found response
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer file.Close()
	// BMC working time
	time.Sleep(1 * time.Second)

	// Reply
	stringJSON, _ := io.ReadAll(file)
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
	for {
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
	file, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		slog.Error("OpenFile()", "err", err, "filename", fn) // テストでは不要
		return err
	}
	defer file.Close()
	fmt.Println("=============================================  ", serial) // 消す
	file.WriteString(fmt.Sprintln(string(byteJson)))
	fmt.Println(byteJson)
	return nil
}