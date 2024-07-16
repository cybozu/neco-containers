package main

import (
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)

// Get iDRAC server list from CSV file
func machineListReader(filename string) (Machines, error) {
	var mlist Machines
	file, err := os.Open(filename)
	if err != nil {
		slog.Error("failed open file")
		return mlist, err
	}
	defer file.Close()
	csvReader := csv.NewReader(file)
	for {
		item, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("failed read file")
			return mlist, err
		}
		mlist.machine = append(mlist.machine, Machine{Serial: item[0], BmcIP: item[1], NodeIP: item[2]})
	}
	return mlist, nil
}

// Get from Redfish on iDRAC webserver
func bmcClient(url string) ([]byte, error) {
	fmt.Println("================ server ", url)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
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

// Print iDRAC log without duplicated
func printLogs(byteJSON []byte, server Machine, ptrDir string) {

	// ポインタDIRは環境変数

	var members Redfish
	if err := json.Unmarshal(byteJSON, &members); err != nil {
		slog.Error("failed to convert struct from JSON")
		return
	}

	lastPtr, err := readLastPointer(server.Serial, ptrDir)
	if err != nil {
		slog.Error("failed to get last log pointer")
		return
	}

	layout := "2006-01-02T15:04:05Z07:00"
	var createUnixtime int64
	var lastId int

	for i := len(members.Sel) - 1; i >= 0; i-- {
		t, _ := time.Parse(layout, members.Sel[i].Create)
		createUnixtime = t.Unix()
		lastId, _ = strconv.Atoi(members.Sel[i].Id)
		members.Sel[i].Serial = server.Serial
		members.Sel[i].BmcIP = server.BmcIP
		members.Sel[i].NodeIP = server.NodeIP

		// IDの大小で比較して出力 クリアでId=1に戻った時はシリアル時刻の大小で比較
		if lastPtr.LastReadId < lastId {
			v, _ := json.Marshal(members.Sel[i])
			fmt.Println(string(v))
		} else if lastPtr.LastReadId > lastId {
			if lastPtr.LastReadTime < createUnixtime {
				v, _ := json.Marshal(members.Sel[i])
				fmt.Println(string(v))
			}
		}
	}

	err = updateLastPointer(LastPointer{
		Serial:       server.Serial,
		LastReadTime: createUnixtime,
		LastReadId:   lastId,
	}, ptrDir)
	if err != nil {
		//slog.Error("failed to update log pointer")
		return
	}
}

// 排他制御を入れること！！
func readLastPointer(serial string, ptrDir string) (LastPointer, error) {
	var lptr LastPointer
	f, err := os.Open(path.Join(ptrDir, serial))
	if errors.Is(err, os.ErrNotExist) {
		f, err = os.Create(path.Join(ptrDir, serial))
		if err != nil {
			//slog.Error("failed to create pointer file")
			return lptr, err
		}
		lptr := LastPointer{
			Serial:       serial,
			LastReadTime: 0,
			LastReadId:   0,
		}
		f.Close()
		return lptr, err
	} else if err != nil {
		//slog.Error("failed to open pointer file")
		return lptr, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		//slog.Error("failed to get the status of the file")
		return lptr, err
	}
	if st.Size() == 0 {
		return lptr, nil
	}
	byteJSON, err := io.ReadAll(f)
	if err != nil {
		//slog.Error("failed to read pointer file")
		return lptr, err
	}
	if json.Unmarshal(byteJSON, &lptr) != nil {
		//slog.Error("failed to convert the struct from JSON")
		return lptr, err
	}
	return lptr, err
}

func updateLastPointer(lptr LastPointer, ptrDir string) error {
	file, err := os.Create(path.Join(ptrDir, lptr.Serial))
	if err != nil {
		//slog.Error("failed to open pointer file")
		return err
	}
	defer file.Close()
	byteJSON, err := json.Marshal(lptr)
	if err != nil {
		//slog.Error("failed to convert JSON")
		return err
	}
	n, err := file.WriteString(string(byteJSON))
	if err != nil { //|| n == 0 {
		//slog.Error("failed to save the log pointer")
		fmt.Println("wrote bytes=", n)
		return err
	}
	return nil
}
