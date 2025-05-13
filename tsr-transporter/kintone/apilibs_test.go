package kintone

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Kintone Application Interface Library", func() {
	Context("Basic API test", Ordered, func() {
		ctx := context.Background()
		var KintoneApp *App
		var returnVals Records
		var statusCode int
		var err error
		var rec []byte

		It("Create new Kintone endpoint", func() {
			ka, _ := ReadAppConfig("../config/kintone-test-config.json")
			KintoneApp, err = NewKintoneEp(
				ka.Domain,
				ka.AppId,
				ka.SpaceId,
				ka.Guest,
				ka.Proxy,
				ka.Token,
				ka.WkDir)
			Expect(err).NotTo(HaveOccurred())
		})

		var registerdRecNum int
		It("Put record as TSR request", func() {
			var rec RecodeForUpdate
			rec.AppId = strconv.Itoa(KintoneApp.AppId)
			rec.Recode.Memo.Value = "ABCDEFGHIJ123"
			statusCode, retData, err := KintoneApp.UpdateRecord(ctx, rec, http.MethodPost)
			Expect(err).NotTo(HaveOccurred())
			Expect(statusCode).NotTo(Equal("200"))

			var retVals RecordForRead
			err = json.Unmarshal(retData, &retVals)
			Expect(err).NotTo(HaveOccurred())

			registerdRecNum, err = strconv.Atoi(retVals.Id)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Check arrived TSR request", func() {
			httpStatus, recs, err := KintoneApp.CheckTsrRequest(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(httpStatus).NotTo(Equal("200"))
			Expect(len(recs.Record)).To(Equal(1))
		})

		It("Get Kintone Record", func() {
			var returnVal Record
			statusCode, rec, err := KintoneApp.GetRecord(ctx, registerdRecNum)
			Expect(err).NotTo(HaveOccurred())
			Expect(statusCode).To(Equal(200))

			err = json.Unmarshal(rec, &returnVal)
			Expect(err).NotTo(HaveOccurred())
			Expect(returnVal.Record.Memo.Value).To(Equal("ABCDEFGHIJ123"))
		})

		It("Check query", func() {
			query := `Created_datetime = TODAY() and log_archive not like "*.zip"`
			statusCode, rec, err = KintoneApp.GetRecords(ctx, query)
			Expect(err).NotTo(HaveOccurred())
			Expect(statusCode).To(Equal(200))
		})

		It("Check record", func() {
			err = json.Unmarshal(rec, &returnVals)
			Expect(err).NotTo(HaveOccurred())
		})

		var recWithFile RecordWithFile
		It("Upload TSR", func() {
			for _, record := range returnVals.Record {
				recWithFile.AppId = strconv.Itoa(KintoneApp.AppId)
				recWithFile.RecNum, _ = strconv.Atoi(record.RecordNumber.Value)
				recWithFile.Recode.File.Value = make([]AttachedFile, 1)
				recWithFile.Recode.File.Value[0].FileKey = ""
				recWithFile.Recode.File.Value[0].Name = "../testdata/test-tsr.zip"
				recWithFile.Recode.TsrDate.Value = time.Now().Format(time.RFC3339)
				statusCode, err = KintoneApp.UploadFile(ctx, recWithFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(statusCode).To(Equal(200))
			}
		})
	})
})
