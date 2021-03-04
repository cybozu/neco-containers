package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v33/github"
	"github.com/kelseyhightower/envconfig"
)

type env struct {
	AppID             int64  `split_words:"true"`
	AppInstallationID int64  `split_words:"true"`
	AppPrivateKeyPath string `split_words:"true"`
	Organization      string
}

func main() {
	var e env
	err := envconfig.Process("github", &e)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rt, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, e.AppID, e.AppInstallationID, e.AppPrivateKeyPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	client := github.NewClient(&http.Client{Transport: rt})
	token, _, err := client.Actions.CreateOrganizationRegistrationToken(context.Background(), e.Organization)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(token.GetToken())
}
