package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// タイムアウトが必要？
// Get from Redfish on iDRAC webserver
func bmcClient(url string) ([]byte, error) {
	fmt.Println("client start to ", url)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// タイムアウトの処理エラーのセットが欲しい
	client := &http.Client{Timeout: time.Duration(10) * time.Second, Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		//slog.Error("failed to setup HTTP client")
		return nil, err
	}
	req.SetBasicAuth(os.Getenv("BMC_USER"), os.Getenv("BMC_PASS"))
	resp, err := client.Do(req)
	if err != nil {
		//slog.Error("failed to iDRAC accessing")
		return nil, err
	}
	defer resp.Body.Close()

	//fmt.Println("HTTP status code ", resp.StatusCode)
	if resp.StatusCode == 401 {
		//slog.Error("unauthorized for iDRAC accessing")
		err := errors.New("unauthorized")
		return nil, err
	} else if resp.StatusCode != 200 {
		//slog.Error("failed to access web-page in iDRAC accessing")
		err := errors.New("not found contents")
		return nil, err
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		//slog.Error("read error web-pages")
		err := errors.New("can not read contents")
		return nil, err
	}
	return buf, nil
}
