package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/neco-containers/tsr-transporter/bmc"
	"github.com/neco-containers/tsr-transporter/dell"
	"github.com/neco-containers/tsr-transporter/kintone"
	"github.com/neco-containers/tsr-transporter/sabakan"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	basicAuthUser     = "support"
	basicAuthPassword = "raw password for support user"
)

var _ = Describe("TSR Transporter", Ordered, func() {
	ka, _ := readKintoneAppParam("local/kintone-test-config.json")
	App, err := kintone.NewKintoneEp(
		ka.Domain,
		ka.AppId,
		ka.SpaceId,
		ka.Guest,
		ka.Proxy,
		ka.Token)
	if err != nil {
		slog.Error("Error setting up the endpoint of Kintone app", "err", err)
		os.Exit(1)
	}

	BeforeAll(func(ctx SpecContext) {
		GinkgoWriter.Println("Start stub servers")
		saba := sabakanMock{
			host:   "127.0.0.1:7180",
			path:   "/api/v1/machines",
			resDir: "testdata/sabakan",
		}
		saba.startMock()
		By("Wait for mock server become up: " + saba.getEndpoint())
		Eventually(func(ctx SpecContext) error {
			req, _ := http.NewRequest("GET", saba.getEndpoint(), nil)
			client := &http.Client{Timeout: time.Duration(3) * time.Second}
			_, err := client.Do(req)
			return err
		}).WithContext(ctx).Should(Succeed())
	}, NodeTimeout(10*time.Second))

	//var user *bmc.UserConfig
	var cnf config

	Context("BMC USER access test", func() {
		It("bmc user test", func() {
			user, err := bmc.LoadBMCUserConfig("testdata/etc/bmc-user.json")
			Expect(err).ToNot(HaveOccurred())
			Expect(user.Support.Password.Raw).To(Equal(basicAuthPassword))
			GinkgoWriter.Println("user=", user.Support.Password.Raw)
		})
		cnf = config{
			bmcUsername: "support",
			//bmcPassword:     user.Support.Password.Raw,
			sabakanEndpoint: *flgSabakanEndpoint,
			kintone:         App,
			intervalTime:    time.Duration(*flgIntervalTime) * time.Second,
		}
		fmt.Println("cnf=", cnf)
	})

	Context("Kintone access test", func() {
		ctx, cancelCause := context.WithCancelCause(context.Background())
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			sig := <-c
			cancelCause(fmt.Errorf("%v", sig))
		}()
		var serial string
		var recId int
		It("Kintone request check", func() {
			httpStatus, records, err := cnf.kintone.CheckReq(ctx)
			fmt.Println("len=", len(records.Record))
			Expect(err).ToNot(HaveOccurred())
			//Expect(len(records.Record)).To(Equal(1))
			Expect(httpStatus).To(Equal(200))

			for _, record := range records.Record {
				//fmt.Println("serial", record.Hostname.Value)
				//Expect(record.Hostname.Value).To(Equal("3Z5P854"))
				serial = record.Hostname.Value
				recId, err = strconv.Atoi(record.Id.Value)
				Expect(err).ToNot(HaveOccurred())
				//fmt.Println("serial", serial)

				var iDRAC_ipv4 string
				//It("Convert IPv4 from Serial", func(ctx SpecContext) {
				ipv4, err := sabakan.GetBmcIpv4("http://127.0.0.1:7180/api/v1/machines", serial)
				Expect(err).ToNot(HaveOccurred())
				//Expect(ipv4).To(Equal("10.72.17.9"))
				iDRAC_ipv4 = ipv4
				//Expect(iDRAC_ipv4).To(Equal("10.72.17.9"))
				//}, SpecTimeout(3*time.Second))

				var bmc dell.Bmc
				var job *url.URL
				var bf *BmcConfig
				//It("setup iDRAC endpoint", func(ctx SpecContext) {
				bf, _ = setBmcParam("local/idrac-test-config.json")
				bmc, err = dell.NewBmcEp(iDRAC_ipv4, bf.User, bf.Pass)
				Expect(err).NotTo(HaveOccurred())
				//})

				//It("Request iDRAC to create TSR", func() {
				job, err = bmc.StartCollection(ctx)
				Expect(err).NotTo(HaveOccurred())
				//})

				//It("Waiting JOB to collect TSR in iDRAC", func() {
				err = bmc.WaitCollection(ctx, job)
				Expect(err).NotTo(HaveOccurred())
				//})

				//It("Download TSR from iDRAC", func() {
				downloadDir, _ := os.Getwd()
				filename := filepath.Join(downloadDir, "test-tsr.zip")
				f, err := os.Create(filename)
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()
				err = bmc.DownloadSupportAssist(ctx, f)
				Expect(err).NotTo(HaveOccurred())
				//})
				var recWithFile kintone.RecordWithFile

				//It("Upload TSR", func() {
				recWithFile.AppId = strconv.Itoa(App.AppId)
				recWithFile.RecNum = recId
				recWithFile.Recode.File.Value = make([]kintone.AttachedFile, 1)
				recWithFile.Recode.File.Value[0].FileKey = ""
				recWithFile.Recode.File.Value[0].Name = "test-tsr.zip"
				recWithFile.Recode.TsrDate.Value = time.Now().Format(time.RFC3339)
				httpStatus, err := cnf.kintone.UploadFile(ctx, recWithFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(httpStatus).To(Equal(200))
				//})
			}
		})
	})
})
