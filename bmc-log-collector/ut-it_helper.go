package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

//var mu0 sync.Mutex

// access counter foreach web-server
var access_counter map[string]int
var response_files map[string][]string
var response_dir map[string]string
var is_initmap bool = false

// id & password for basic authentication
const (
	basicAuthUser     = "user"
	basicAuthPassword = "pass"
)

func init_map() {
	access_counter = make(map[string]int)
	response_files = make(map[string][]string)
	response_dir = make(map[string]string)
	is_initmap = true
}

// BMC Simulator for Unit Test
func start_iDRAC_Simulator_ut(mu *sync.Mutex) {
	path := "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
	host := "127.0.0.1:8080"
	if !is_initmap {
		init_map()
	}

	mu.Lock()
	access_counter[host] = 0
	response_files[host] = []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"}
	response_dir[host] = "testdata/redfish_response_ut1"
	mu.Unlock()

	server := http.NewServeMux()
	server.HandleFunc(path, redfish_svc)
	go func() {
		http.ListenAndServeTLS(host, "testdata/ssl/localhost.crt", "testdata/ssl/localhost.key", server)
	}()
}

// BMC Simulator for Integration Test #1
func startIdracMock_it(mu *sync.Mutex) {
	path := "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
	host := "127.0.0.1:9080"
	if !is_initmap {
		init_map()
	}

	mu.Lock()
	access_counter[host] = 0
	response_files[host] = []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"}
	response_dir[host] = "testdata/redfish_response_it1"
	mu.Unlock()

	server := http.NewServeMux()
	server.HandleFunc(path, redfish_svc)
	go func() {
		http.ListenAndServeTLS(host, "testdata/ssl/localhost.crt", "testdata/ssl/localhost.key", server)
	}()
}

// BMC Simulator for Integration Test #2
func start_iDRAC_Simulator_it2_idrac1(mu *sync.Mutex) {
	path := "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
	host := "127.0.0.1:7180"
	if !is_initmap {
		init_map()
	}

	mu.Lock()
	access_counter[host] = 0
	response_files[host] = []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"}
	response_dir[host] = "testdata/redfish_response_it2"
	mu.Unlock()

	server := http.NewServeMux()
	server.HandleFunc(path, redfish_svc)
	go func() {
		http.ListenAndServeTLS(host, "testdata/ssl/localhost.crt", "testdata/ssl/localhost.key", server)
	}()
}

func start_iDRAC_Simulator_it2_idrac2(mu *sync.Mutex) {
	path := "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
	host := "127.0.0.1:7280"
	if !is_initmap {
		init_map()
	}
	mu.Lock()
	access_counter[host] = 0
	response_files[host] = []string{"HN3CLP3-1.json", "HN3CLP3-2.json", "HN3CLP3-3.json"}
	response_dir[host] = "testdata/redfish_response_it2"
	mu.Unlock()

	server := http.NewServeMux()
	server.HandleFunc(path, redfish_svc)
	go func() {
		http.ListenAndServeTLS(host, "testdata/ssl/localhost.crt", "testdata/ssl/localhost.key", server)
	}()
}

// Redfish REST Service
func redfish_svc(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json;odata.metadata=minimal;charset=utf-8")
	// basic authentication
	if user, pass, ok := r.BasicAuth(); !ok || user != basicAuthUser || pass != basicAuthPassword {
		w.Header().Add("WWW-Authenticate", `Basic realm="my private area"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	//mu0.Lock()
	key := string(r.Host)
	fn := response_files[key][access_counter[key]]
	response_file := path.Join(response_dir[key], fn)
	access_counter[key] = access_counter[key] + 1
	//mu0.Unlock()

	file, err := os.Open(response_file)
	if err != nil {
		// create not found response
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer file.Close()
	time.Sleep(5 * time.Second)
	stringJSON, _ := io.ReadAll(file)
	fmt.Fprint(w, string(stringJSON))
}
