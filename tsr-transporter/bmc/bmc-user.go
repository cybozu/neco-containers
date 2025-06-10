package bmc

import (
	"encoding/json"
	"fmt"
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

// Read users of BMC
func ReadUsers(userFile string) (*UserConfig, error) {
	fmt.Println("DEBUG userfile", userFile)
	fd, err := os.Open(userFile)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	bmcUsers := new(UserConfig)
	err = json.NewDecoder(fd).Decode(bmcUsers)
	if err != nil {
		return nil, err
	}
	return bmcUsers, nil
}
