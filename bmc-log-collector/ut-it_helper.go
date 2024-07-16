package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

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
}

// BMC Simulator for Unit Test
func start_iDRAC_Simulator_ut() {
	go func() {
		server := http.Server{
			Addr:    ":8080",
			Handler: nil,
		}
		uri := "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/EntriesUT1"
		key := "127.0.0.1:8080" + uri
		if !is_initmap {
			init_map()
		}
		access_counter[key] = 0
		response_files[key] = []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"}
		response_dir[key] = "testdata/redfish_response_ut1"

		http.HandleFunc(uri, redfish_svc)
		server.ListenAndServeTLS("testdata/ssl/localhost.crt", "testdata/ssl/localhost.key")
	}()
}

// BMC Simulator for Integration Test #1
func start_iDRAC_Simulator_it() {
	go func() {
		server := http.Server{
			Addr:    ":9080",
			Handler: nil,
		}
		uri := "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/EntriesIT1"
		key := "127.0.0.1:9080" + uri

		if !is_initmap {
			init_map()
		}
		access_counter[key] = 0
		response_files[key] = []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"}
		response_dir[key] = "testdata/redfish_response_it1"

		http.HandleFunc(uri, redfish_svc)
		server.ListenAndServeTLS("testdata/ssl/localhost.crt", "testdata/ssl/localhost.key")
	}()
}

// BMC Simulator for Integration Test #2
func start_iDRAC_Simulator_it2_idrac1() {
	go func() {
		server := http.Server{
			Addr:    ":7080",
			Handler: nil,
		}
		uri := "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries"
		key := "127.0.0.1:7080" + uri

		if !is_initmap {
			init_map()
		}
		access_counter[key] = 0
		response_files[key] = []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"}
		response_dir[key] = "testdata/redfish_response_it2"

		http.HandleFunc(uri, redfish_svc)
		server.ListenAndServeTLS("testdata/ssl/localhost.crt", "testdata/ssl/localhost.key")
	}()
}

func start_iDRAC_Simulator_it2_idrac2() {
	go func() {
		server := http.Server{
			Addr:    ":7082",
			Handler: nil,
		}
		uri := "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries2"
		key := "127.0.0.1:7082" + uri

		if !is_initmap {
			init_map()
		}
		access_counter[key] = 0
		response_files[key] = []string{"683FPQ3-1.json", "683FPQ3-2.json", "683FPQ3-3.json"}
		response_dir[key] = "testdata/redfish_response_it2"

		http.HandleFunc(uri, redfish_svc)
		server.ListenAndServeTLS("testdata/ssl/localhost.crt", "testdata/ssl/localhost.key")
	}()
}

// Redfish REST Service
func redfish_svc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json;odata.metadata=minimal;charset=utf-8")
	// basic authentication
	if user, pass, ok := r.BasicAuth(); !ok || user != basicAuthUser || pass != basicAuthPassword {
		w.Header().Add("WWW-Authenticate", `Basic realm="my private area"`)
		w.WriteHeader(http.StatusUnauthorized)
		//http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	// バグあり
	// キーをポート番号＋パスに変更
	// 排他制御をかける

	// get response file & increment a access counter
	key := strings.TrimSpace(string(r.Host)) + string(r.URL.Path)
	//key := string(r.URL.Path)
	fmt.Println("KEY KEY KEY =", key)
	fn := response_files[key][access_counter[key]]
	response_file := path.Join(response_dir[key], fn)
	access_counter[key] = access_counter[key] + 1

	file, err := os.Open(response_file)
	if err != nil {
		// create not found response
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		//fmt.Fprint(w, "")
		return
	}
	defer file.Close()
	//fmt.Println(access_counter[string(r.URL.Path)], response_file, file.Name())
	time.Sleep(5 * time.Second)
	stringJSON, _ := io.ReadAll(file)
	fmt.Fprint(w, string(stringJSON))
}
