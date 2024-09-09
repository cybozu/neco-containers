package main

import (
	"encoding/json"
	"os"
)

// BMCPassword represents password for a BMC user.
type BMCPassword struct {
	Raw  string `json:"raw"`
	Hash string `json:"hash"`
	Salt string `json:"salt"`
}

// Credentials represents credentials of a BMC user.
type Credentials struct {
	Password BMCPassword `json:"password"`
}

// UserConfig represents a set of BMC user credentials in JSON format.
type UserConfig struct {
	Root    Credentials `json:"root"`
	Repair  Credentials `json:"repair"`
	Power   Credentials `json:"power"`
	Support Credentials `json:"support"`
}

// LoadConfig loads UserConfig.
func LoadConfig(userFile string) (*UserConfig, error) {
	g, err := os.Open(userFile)
	if err != nil {
		return nil, err
	}
	defer g.Close()

	bmcUsers := new(UserConfig)
	err = json.NewDecoder(g).Decode(bmcUsers)
	if err != nil {
		return nil, err
	}

	return bmcUsers, nil
}
