package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
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
	server.HandleFunc(redfishPath, redfishMock)
	go func() {
		http.ListenAndServeTLS(b.host, "testdata/ssl/localhost.crt", "testdata/ssl/localhost.key", server)
	}()
}

// Redfish REST Service
func redfishMock(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json;odata.metadata=minimal;charset=utf-8")
	// basic authentication
	if user, pass, ok := r.BasicAuth(); !ok || user != basicAuthUser || pass != basicAuthPassword {
		w.Header().Add("WWW-Authenticate", `Basic realm="my private area"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	mutex.Lock()
	key := string(r.Host)
	fn := responseFiles[key][accessCounter[key]]
	responseFile := path.Join(responseDir[key], fn)
	accessCounter[key] = accessCounter[key] + 1 // race
	mutex.Unlock()

	file, err := os.Open(responseFile)
	if err != nil {
		// create not found response
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer file.Close()
	//time.Sleep(5 * time.Second)
	stringJSON, _ := io.ReadAll(file)
	fmt.Fprint(w, string(stringJSON))
}
