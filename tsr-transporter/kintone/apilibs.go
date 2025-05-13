package kintone

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Config struct {
	Domain  string `json:"domain"`
	AppId   int    `json:"app_id"`
	SpaceId int    `json:"space_id"`
	Guest   bool   `json:"is_guest"`
	Proxy   string `json:"proxy"`
	Token   string `json:"token"`
	WkDir   string `json:"working_dir"`
}

func ReadAppConfig(configFilename string) (*Config, error) {
	fd, err := os.Open(configFilename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	conf := new(Config)
	err = json.NewDecoder(fd).Decode(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func (app *App) createEndpoint(upload bool, recNum int) string {
	var urlString string
	if app.IsGuestSpace {
		urlString = fmt.Sprintf("%s/k/guest/%d/v1", app.Domain, app.SpaceId)
	} else {
		urlString = fmt.Sprintf("%s/k/v1", app.Domain)
	}
	if upload {
		urlString = fmt.Sprintf("%s/file.json?app=%d", urlString, app.AppId)
	} else {
		if recNum > 0 {
			urlString = fmt.Sprintf("%s/record.json?app=%d&id=%d", urlString, app.AppId, recNum)
		} else {
			urlString = fmt.Sprintf("%s/record.json?app=%d", urlString, app.AppId)
		}
	}
	return urlString
}

func (app *App) createEndpointforRecords(query string) string {
	var urlString string
	if app.IsGuestSpace {
		urlString = fmt.Sprintf("%s/k/guest/%d/v1", app.Domain, app.SpaceId)
	} else {
		urlString = fmt.Sprintf("%s/k/v1", app.Domain)
	}
	queryString := url.QueryEscape(query)
	return fmt.Sprintf("%s/records.json?app=%d&query=%s", urlString, app.AppId, queryString)
}

func (app *App) httpRequest(req *http.Request) (
	int, // HTTP response status code
	[]byte, // HTTP Body
	error,
) {
	type result struct {
		resp *http.Response
		err  error
	}
	done := make(chan result, 1)
	go func() {
		resp, err := app.Client.Do(req)
		done <- result{resp, err}
	}()
	select {
	case r := <-done:
		if r.err != nil {
			return r.resp.StatusCode, nil, r.err
		}
		byteJSON, err := io.ReadAll(r.resp.Body)
		defer r.resp.Body.Close()
		if err != nil {
			return r.resp.StatusCode, nil, err
		}
		return r.resp.StatusCode, byteJSON, err
	case <-time.After(app.Timeout):
		type requestCanceler interface {
			CancelRequest(*http.Request)
		}
		canceller, ok := app.Client.Transport.(requestCanceler)
		if ok {
			canceller.CancelRequest(req)
		} else {
			go func() {
				r := <-done
				if r.err == nil {
					r.resp.Body.Close()
				}
			}()
		}
		return http.StatusRequestTimeout, nil, fmt.Errorf("timeout of HTTP response")
	}
}

func NewKintoneEp(
	Domain string, // Kintone Domain (URL)
	AppId int, // Kintone Appication ID
	SpaceId int, // Space ID
	Guest bool, // Is guest space: false or true
	Proxy string, // Proxy URL
	Token string, // Access token of Kintone Application
	WkDir string, // Working Directory in process
) (*App, error) {
	var proxyReq func(*http.Request) (*url.URL, error)
	if len(Proxy) > 0 {
		proxyUrl, err := url.Parse(Proxy)
		if err != nil {
			return nil, err
		}
		proxyReq = http.ProxyURL(proxyUrl)
	} else {
		proxyReq = nil
	}
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
		Proxy:              proxyReq,
	}
	return &App{
		Domain:       Domain,
		AppId:        AppId,
		AppToken:     Token,
		IsGuestSpace: Guest,
		SpaceId:      SpaceId,
		Client:       &http.Client{Transport: tr},
		Timeout:      10 * time.Second,
		WkDir:        WkDir,
	}, nil
}

func (app *App) GetRecord(ctx context.Context, recNum int) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		app.createEndpoint(false, recNum),
		nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Add("X-Cybozu-API-Token", app.AppToken)
	return app.httpRequest(req)
}

func (app *App) GetRecords(ctx context.Context, query string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		app.createEndpointforRecords(query),
		nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Add("X-Cybozu-API-Token", app.AppToken)
	return app.httpRequest(req)
}

func (app *App) CheckTsrRequest(ctx context.Context) (int, Records, error) {
	var recs Records
	query := `Created_datetime = TODAY() and datetime = ""`
	httpStatus, rec, err := app.GetRecords(ctx, query)
	if err != nil {
		return httpStatus, recs, err
	}
	err = json.Unmarshal(rec, &recs)
	if err != nil {
		return httpStatus, recs, err
	}
	return httpStatus, recs, nil
}

func (app *App) UpdateRecord(ctx context.Context, data interface{}, method string) (int, []byte, error) {
	byteJson, err := json.Marshal(data)
	if err != nil {
		return 0, byteJson, err
	}
	stringJson := string(byteJson)
	req, err := http.NewRequestWithContext(ctx,
		method,
		app.createEndpoint(false, 0),
		bytes.NewBufferString(stringJson))
	if err != nil {
		return 0, byteJson, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("X-Cybozu-API-Token", app.AppToken)
	return app.httpRequest(req)
}

func (app *App) UploadFile(ctx context.Context, data RecordWithFile) (int, error) {
	fd, err := os.Open(data.Recode.File.Value[0].Name)
	if err != nil {
		return 0, err
	}
	defer fd.Close()
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	fw, err := mw.CreateFormFile("file", data.Recode.File.Value[0].Name)
	if err != nil {
		return 0, err
	}
	_, err = io.Copy(fw, fd)
	if err != nil {
		return 0, err
	}
	contentType := mw.FormDataContentType()
	err = mw.Close()
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		app.createEndpoint(true, 0),
		body)
	if err != nil {
		return 0, err
	}
	req.Header.Add("X-Cybozu-API-Token", app.AppToken)
	req.Header.Set("Content-Type", contentType)
	httpStatus, byteJSON, err := app.httpRequest(req)
	if err != nil {
		return httpStatus, err
	}
	var fa AttachedFile
	err = json.Unmarshal(byteJSON, &fa)
	if err != nil {
		return 0, err
	}
	data.Recode.File.Value[0].FileKey = fa.FileKey
	httpStatus, _, err = app.UpdateRecord(ctx, data, http.MethodPut)
	if err != nil {
		return httpStatus, err
	}
	return httpStatus, nil
}
