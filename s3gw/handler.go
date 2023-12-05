package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type credentialsProvider struct{}

func (credentialsProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return awsCredentials, nil
}

var client *s3.Client

const bucketPathPrefix = len("/bucket/")
const partSize = 128 << 20 // 128MiB

// getHeaderOrNil gets header value or return nil if the key does not exist
func getHeaderOrNil(header http.Header, key string) *string {
	value := header.Get(key)
	if value == "" {
		return nil
	} else {
		return &value
	}
}

// setHeaderOrNil sets header value or delete key if the value is nil
func setHeaderOrNil(header http.Header, key string, value *string) {
	if value == nil {
		header.Del(key)
	} else {
		header.Set(key, *value)
	}
}

type Object struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last-modified"`
	ETag         string    `json:"etag"`
}

type listResult struct {
	Objects []Object `json:"objects"`
}

func newListResult() *listResult {
	result := listResult{
		Objects: []Object{},
	}
	return &result
}

func listHandlerFunc(res http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	input := &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	}
	output, err := client.ListObjectsV2(req.Context(), input)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest) // XXX we cannot determine appropriate status code
		res.Write([]byte(err.Error()))
		return
	}

	result := newListResult()
	for _, c := range output.Contents {
		result.Objects = append(result.Objects, Object{
			Key:          *c.Key,
			Size:         c.Size,
			LastModified: c.LastModified.UTC(),
			ETag:         *c.ETag,
		})
	}
	marshalled, err := json.Marshal(result)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}
	res.Header().Set("Content-Length", fmt.Sprint(len(marshalled)))
	res.Write(marshalled)
}

func objectHandlerFunc(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		objectGetHandlerFunc(res, req)
	case "PUT":
		objectPutHandlerFunc(res, req)
	case "DELETE":
		objectDeleteHandlerFunc(res, req)
	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func objectGetHandlerFunc(res http.ResponseWriter, req *http.Request) {
	objectKey := req.URL.Path[bucketPathPrefix:]
	input := &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	}
	output, err := client.GetObject(req.Context(), input)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest) // XXX we cannot determine appropriate status code
		res.Write([]byte(err.Error()))
		return
	}

	setHeaderOrNil(res.Header(), "Content-Type", output.ContentType)
	setHeaderOrNil(res.Header(), "ETag", output.ETag)
	res.Header().Set("Content-Length", fmt.Sprint(output.ContentLength))
	if output.LastModified != nil {
		res.Header().Set("Last-Modified", output.LastModified.UTC().Format(http.TimeFormat))
	}
	io.Copy(res, output.Body)
}

func objectPutHandlerFunc(res http.ResponseWriter, req *http.Request) {
	objectKey := req.URL.Path[bucketPathPrefix:]
	input := &s3.PutObjectInput{
		Bucket:        &bucketName,
		Key:           &objectKey,
		Body:          req.Body,
		ContentLength: req.ContentLength,
		ContentType:   getHeaderOrNil(req.Header, "Content-Type"),
	}
	uploader := s3manager.NewUploader(client, func(u *s3manager.Uploader) {
		u.Concurrency = 1
		u.PartSize = partSize
		u.LeavePartsOnError = false
	})
	_, err := uploader.Upload(req.Context(), input)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest) // XXX we cannot determine appropriate status code
		res.Write([]byte(err.Error()))
		return
	}
	res.WriteHeader(http.StatusOK)
}

func objectDeleteHandlerFunc(res http.ResponseWriter, req *http.Request) {
	objectKey := req.URL.Path[bucketPathPrefix:]
	input := &s3.DeleteObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	}
	_, err := client.DeleteObject(req.Context(), input)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest) // XXX we cannot determine appropriate status code
		res.Write([]byte(err.Error()))
		return
	}
	res.WriteHeader(http.StatusOK) // ?
}

func healthHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
}
