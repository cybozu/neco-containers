package main

import (
	"crypto/tls"
	//"errors"
	"fmt"
	"io"
	"log/slog"
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
		slog.Error(fmt.Sprintf("%s", err))
		return nil, err
	}
	req.SetBasicAuth(os.Getenv("BMC_USER"), os.Getenv("BMC_PASS"))
	resp, err := client.Do(req)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err := fmt.Errorf("HTTP status code = %d", resp.StatusCode)
		slog.Error(fmt.Sprintf("%s", err))
		return nil, err
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return nil, err
	}
	return buf, nil
}
