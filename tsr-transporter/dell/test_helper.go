package dell

import (
	"encoding/json"
	"os"
)

type BmcConfig struct {
	IpV4 string `json:"idrac_ipv4"`
	User string `json:"user"`
	Pass string `json:"pass"`
}

func setBmcParam(configFilename string) (*BmcConfig, error) {
	fd, err := os.Open(configFilename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	conf := new(BmcConfig)
	err = json.NewDecoder(fd).Decode(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
