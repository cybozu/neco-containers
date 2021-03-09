package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var interval = 5 * time.Minute

func init() {
	rootCmd.Flags().String("api-key", "", "opsgenie API key")
	rootCmd.Flags().DurationVar(&interval, "interval", interval, "interval between heartbeats")
	viper.BindEnv("api-key", "OPSGENIE_APIKEY")
	viper.BindPFlag("api-key", rootCmd.Flags().Lookup("api-key"))
}

var rootCmd = &cobra.Command{
	Use:   "heartbeat TARGET",
	Short: "continuously send ping to Opsgenie heartbeat API",
	Long: `continuously send ping to Opsgenie heartbeat API.

API key can be given through OPSGENIE_APIKEY environment variable
or --api-key command-line flag.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return subMain(args[0])
	},
}

func subMain(target string) error {
	apiKey := viper.GetString("api-key")
	if apiKey == "" {
		return errors.New("no api key")
	}
	beatURL := fmt.Sprintf("https://api.opsgenie.com/v2/heartbeats/%s/ping", target)
	authVal := fmt.Sprintf("GenieKey %s", apiKey)

	tick := time.NewTicker(interval)
	defer tick.Stop()

	hc := &http.Client{}

	for {
		<-tick.C

		req, err := http.NewRequest(http.MethodGet, beatURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", authVal)

		resp, err := hc.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "request failed: %v\n", err)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read request body: %v\n", err)
			continue
		}

		if resp.StatusCode < 200 && resp.StatusCode >= 300 {
			fmt.Fprintf(os.Stderr, "got error response %d: %s\n", resp.StatusCode, body)
		}

		fmt.Fprintf(os.Stderr, "success\n")
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
