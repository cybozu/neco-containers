package kintone

import (
	"encoding/json"
	"os"
)

type TestConfig struct {
	Domain  string `json:"domain"`
	AppId   int    `json:"app_id"`
	SpaceId int    `json:"space_id"`
	Guest   bool   `json:"is_guest"`
	Proxy   string `json:"proxy"`
	Token   string `json:"token"`
}

func setKintoneAppParam(configFilename string) (*TestConfig, error) {
	fd, err := os.Open(configFilename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	conf := new(TestConfig)
	err = json.NewDecoder(fd).Decode(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
