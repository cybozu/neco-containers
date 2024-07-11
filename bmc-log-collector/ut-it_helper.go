package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"
)

// Start BMC Simulator

func start_iDRAC_Simulator_ut() {
	go func() {
		fmt.Println("*** Web server running for UT")
		server := http.Server{
			Addr:    ":8080",
			Handler: nil,
		}
		uri := "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
		access_counter = make(map[string]int)
		access_counter[uri] = 0
		http.HandleFunc(uri, redfish_svc)
		server.ListenAndServeTLS("testdata/ssl/localhost.crt", "testdata/ssl/localhost.key")
	}()
}

func start_iDRAC_Simulator_it() {
	go func() {
		fmt.Println("*** Web server running for IT")
		server := http.Server{
			Addr:    ":9080",
			Handler: nil,
		}
		uri := "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries2"
		access_counter = make(map[string]int)
		access_counter[uri] = 0
		http.HandleFunc(uri, redfish_svc)
		server.ListenAndServeTLS("testdata/ssl/localhost.crt", "testdata/ssl/localhost.key")
	}()
}

// カウンターが、二つのサーバーで共有されている。これでバグが起きる
// JSON data for response
// var counter int = 0
var resp []string = []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json", "683FPQ3-4.json"}
var testRspDir string = "testdata/redfish_response"
var access_counter map[string]int

// id & password for basic authentication
const (
	basicAuthUser     = "user"
	basicAuthPassword = "pass"
)

// Redfish REST Service
func redfish_svc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json;odata.metadata=minimal;charset=utf-8")
	// basic authentication
	if user, pass, ok := r.BasicAuth(); !ok || user != basicAuthUser || pass != basicAuthPassword {
		w.Header().Add("WWW-Authenticate", `Basic realm="my private area"`)
		w.WriteHeader(http.StatusUnauthorized)
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}
	// get response file & increment a access counter
	i := access_counter[string(r.URL.Path)]
	fn := path.Join(testRspDir, resp[i])
	fmt.Println("================", r.URL.Path, "=============")

	// ここも問題
	access_counter[string(r.URL.Path)] = access_counter[string(r.URL.Path)] + 1

	file, err := os.Open(fn)
	if err != nil {
		// create not found response
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "")
		return
	}
	defer file.Close()
	fmt.Println(access_counter[string(r.URL.Path)], fn, file.Name())
	time.Sleep(5 * time.Second)
	stringJSON, _ := io.ReadAll(file)
	fmt.Fprint(w, string(stringJSON))
}
