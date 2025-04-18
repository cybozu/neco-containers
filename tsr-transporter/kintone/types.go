package kintone

import (
	"net/http"
	"time"
)

type Column struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
type ColumnCreateBy struct {
	Type  string `json:"type"`
	Value Column `json:"value"`
}
type ColumnUpdater struct {
	Type  string `json:"type"`
	Value Column `json:"value"`
}
type FileInfo struct {
	FileKey     string `json:"filekey"`
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Size        string `json:"size"`
}
type ColumnFile struct {
	Type  string     `json:"type"`
	Value []FileInfo `json:"value"`
}
type Fields struct {
	Id           Column         `json:"$id"`
	RecordNumber Column         `json:"Record_number"`
	CreateTime   Column         `json:"Created_datetime"`
	CreatedBy    ColumnCreateBy `json:"Created_by"`
	UpdateTime   Column         `json:"Updated_datetime"`
	Updater      ColumnUpdater  `json:"Updated_by"`
	Hostname     Column         `json:"hostname"`
	Revision     Column         `json:"$revision"`
	Memo         Column         `json:"memo"`
	File         ColumnFile     `json:"log_archive"`
}
type Record struct {
	Record Fields `json:"record"`
}
type Records struct {
	Record []Fields `json:"records"`
}
type RecordForRead struct {
	Id       string `json:"id"`
	Revision string `json:"revision"`
}

// For Attache File
type AttachedFile struct {
	FileKey string `json:"fileKey"`
	Name    string `json:"name"`
}
type ColumnFileAttached struct {
	Type  string         `json:"type"`
	Value []AttachedFile `json:"value"`
}
type FieldWithFile struct {
	TsrDate Column             `json:"datetime"`
	File    ColumnFileAttached `json:"log_archive"`
}
type RecordWithFile struct {
	AppId  string        `json:"app"`
	RecNum int           `json:"id"`
	Recode FieldWithFile `json:"record"`
}

// For Update Test
type FieldForUpdate struct {
	Memo    Column `json:"memo"`
	TsrDate Column `json:"datetime"`
}
type RecodeForUpdate struct {
	AppId  string         `json:"app"`
	RecNum int            `json:"id"`
	Recode FieldForUpdate `json:"record"`
}
type App struct {
	Domain       string
	AppId        int
	AppToken     string
	IsGuestSpace bool
	SpaceId      int
	Client       *http.Client  // Specialized client.
	Timeout      time.Duration // Timeout for API responses.
	Proxy        http.Transport
	WkDir        string
}
