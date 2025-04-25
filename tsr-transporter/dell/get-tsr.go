package dell

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	jobMaxWaitingCount int = 40 // 15 * 40 -> 600 sec
)

type Bmc struct {
	BmcIpv4  string
	UserInfo *url.Userinfo
	Client   *http.Client  // Specialized client.
	Timeout  time.Duration // Timeout for API responses.
}

func NewBmcEp(bmcIpv4 string, username string, password string) (Bmc, error) {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}
	hc := &http.Client{Transport: tr}
	return Bmc{
		BmcIpv4:  bmcIpv4,
		UserInfo: url.UserPassword(username, password),
		Client:   hc,
		Timeout:  10 * time.Second,
	}, nil
}

func (bmc Bmc) httpRequest(req *http.Request) (int, []byte, *url.URL, error) {
	type result struct {
		resp *http.Response
		err  error
	}
	done := make(chan result, 1)

	go func() {
		resp, err := bmc.Client.Do(req)
		done <- result{resp, err}
	}()
	select {
	case r := <-done:
		// Error handling for app.Client.Do(req)
		if r.err != nil {
			return r.resp.StatusCode, nil, nil, r.err
		}

		byteJSON, err := io.ReadAll(r.resp.Body)
		defer r.resp.Body.Close()
		if err != nil {
			return r.resp.StatusCode, nil, nil, err
		}
		jobURL, errx := r.resp.Location()
		if errx != nil {
			if errx.Error() == "http: no Location header in response" {
				return r.resp.StatusCode, byteJSON, nil, nil
			} else {
				return r.resp.StatusCode, byteJSON, nil, err
			}
		}
		return r.resp.StatusCode, byteJSON, jobURL, nil
	case <-time.After(bmc.Timeout):
		// If the cancellation is valid, it will be a cancellation request, otherwise it will be discarded.
		type requestCanceler interface {
			CancelRequest(*http.Request)
		}
		if canceller, ok := bmc.Client.Transport.(requestCanceler); ok {
			canceller.CancelRequest(req)
		} else {
			go func() {
				r := <-done
				if r.err == nil {
					r.resp.Body.Close()
				}
			}()
		}
		return http.StatusRequestTimeout, nil, nil, fmt.Errorf("timeout occurred while connecting to iDRAC")
	}
}

func (bmc *Bmc) StartCollection(ctx context.Context) (*url.URL, error) {
	// https://developer.dell.com/apis/2978/versions/6.xx/openapi.yaml/paths/~1redfish~1v1~1Managers~1%7BManagerId%7D~1Oem~1Dell~1DellLCService~1Actions~1DellLCService.SupportAssistCollection/post
	u := &url.URL{
		Scheme: "https",
		Host:   bmc.BmcIpv4,
		// This is legacy endpoint. iDRAC9 firmware 4.x.y does not support /redfish/v1/Managers/{ManagerId}/Oem/Dell/DellLCService/Actions/DellLCService.SupportAssistCollection.
		Path: "/redfish/v1/Dell/Managers/iDRAC.Embedded.1/DellLCService/Actions/DellLCService.SupportAssistCollection",
		User: bmc.UserInfo,
	}
	payload := map[string]interface{}{
		"ShareType":           "Local",
		"DataSelectorArrayIn": []string{"HWData"},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	// Without this header, the response will be 406 Not Acceptable.
	// https://blog.csdn.net/fwtyyds/article/details/128896633
	req.Header.Set("Accept", "*/*")
	httpStatus, _, jobUrl, err := bmc.httpRequest(req)
	if err != nil {
		return nil, err
	}
	if httpStatus != http.StatusAccepted {
		return nil, fmt.Errorf("failed to accept TSR Job")
	}
	return jobUrl, nil
}

func (bmc *Bmc) checkJobCondition(ctx context.Context, jobURL *url.URL) (bool, error) {
	// https://developer.dell.com/apis/2978/versions/6.xx/openapi.yaml/paths/~1redfish~1v1~1Managers~1%7BManagerId%7D~1Oem~1Dell~1Jobs~1%7BDellJobId%7D/get
	u := &url.URL{
		Scheme: "https",
		Host:   bmc.BmcIpv4,
		Path:   jobURL.Path,
		User:   bmc.UserInfo,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	// Without this header, the response will be 406 Not Acceptable.
	// https://blog.csdn.net/fwtyyds/article/details/128896633
	req.Header.Set("Accept", "*/*")
	httpStatus, byteJSON, _, err := bmc.httpRequest(req)
	if err != nil {
		return false, err
	}
	if httpStatus != http.StatusOK {
		return false, fmt.Errorf("status code err %v", httpStatus)
	}
	var job struct {
		JobState string `json:"JobState"`
		Message  string `json:"Message"`
	}
	if err := json.Unmarshal(byteJSON, &job); err != nil {
		return false, fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	if job.JobState == "Failed" {
		return false, fmt.Errorf("job failed: %s", job.Message)
	}
	if job.JobState != "Completed" {
		return false, nil
	}
	return true, nil
}

func (bmc *Bmc) WaitCollection(ctx context.Context, jobURL *url.URL) error {
	var wait time.Duration = 15
	var c int
	for i := 0; i < jobMaxWaitingCount; i++ {
		c = i
		complete, err := bmc.checkJobCondition(ctx, jobURL)
		if err != nil {
			return err
		}
		if complete {
			return nil
		}
		time.Sleep(wait * time.Second)
	}
	return fmt.Errorf("timeout %v", int(wait.Seconds())*c)
}

func (bmc *Bmc) DownloadSupportAssist(ctx context.Context, w io.Writer) error {
	// https://github.com/dell/iDRAC-Redfish-Scripting/blob/643e08194c48433b894f7ca31a77d49289fcefa3/Redfish%20Python/SupportAssistCollectionLocalREDFISH.py#L265C117-L265C151
	// https://${HOST}/redfish/v1/Dell/sacollect.zip
	u := &url.URL{
		Scheme: "https",
		Host:   bmc.BmcIpv4,
		Path:   "/redfish/v1/Dell/sacollect.zip",
		User:   bmc.UserInfo,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	// Without this header, the response will be 406 Not Acceptable.
	// https://blog.csdn.net/fwtyyds/article/details/128896633
	req.Header.Set("Accept", "*/*")
	httpStatus, byteTSR, _, err := bmc.httpRequest(req)
	if httpStatus != http.StatusOK {
		fmt.Println("status code", httpStatus)
		return fmt.Errorf("HTTP Status code %v", httpStatus)
	}
	if err != nil {
		return err
	}
	_, err = w.Write(byteTSR)
	if err != nil {
		return err
	}
	return nil
}
