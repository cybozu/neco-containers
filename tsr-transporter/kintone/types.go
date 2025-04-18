package kintone

import (
	"net/http"
	"time"
)

/*
// Kintoneのレスポンス用
type MetaValue struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
type MetaValue2 struct {
	Value string `json:"value"`
}

type FileAttached struct {
	FileKey string `json:"fileKey"`
	Name    string `json:"name"`
}
type MetaValueFile struct {
	Value []FileAttached `json:"value"`
}

type CreateBy struct {
	Type  string    `json:"type"`
	Value MetaValue `json:"value"`
}

type ValueContent struct {
	ModifiedContents MetaValue `json:"ModifiedContents"`
	Modifiedby       MetaValue `json:"Modifiedby"`
	Date             MetaValue `json:"Date"`
}
type MemoValue struct {
	Id    string       `json:"id"`
	Value ValueContent `json:"value"`
}
type UpdateMemo2 struct {
	Type  string      `json:"type"`
	Value []MemoValue `json:"value"`
}

type UpdateMemo struct {
	Type  string   `json:"type"`
	Value []string `json:"value"`
}
type Updater struct {
	Type  string    `json:"type"`
	Value MetaValue `json:"value"`
}
type FileInfo struct {
	Filekey     string `json:"filekey"`
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Size        string `json:"size"`
}
type File struct {
	Type  string     `json:"type"`
	Value []FileInfo `json:"value"`
}
type ReadRecord struct {
	Id           MetaValue `json:"$id"`
	RecordNumber MetaValue `json:"レコード番号"`
	CreateTime   MetaValue `json:"Createdtime"`
	Createdby    CreateBy  `json:"Createby"`
	UpdateTime   MetaValue `json:"更新日時"`
	//UpdateMemo   UpdateMemo2 `json:"変更内容メモ"` 全部読む必要は無いので、この部分は処理しない
	Updater     Updater   `json:"更新者"`
	Description MetaValue `json:"Description"`
	Title       MetaValue `json:"Title"`
	Revision    MetaValue `json:"$revision"`
	File        File      `json:"File"`
}
type ReadRecord2 struct {
	Id       string `json:"id"`
	Revision string `json:"revision"`
}

type AppRecord struct {
	Record ReadRecord `json:"record"`
}
type AppRecords struct {
	Record []ReadRecord `json:"records"`
}

// キントーンへの書込み用
type WriteRecord struct {
	Title       MetaValue `json:"Title"`
	Description MetaValue `json:"Description"`
}
type WriteRecord2 struct {
	Title       MetaValue     `json:"Title"`
	Description MetaValue     `json:"Description"`
	File        MetaValueFile `json:"File"`
}

type PostRecord struct {
	AppId  string      `json:"app"`
	Recode WriteRecord `json:"record"`
}
type PutRecord struct {
	AppId  string      `json:"app"`
	RecNum int         `json:"id"`
	Recode WriteRecord `json:"record"`
}

type PutRecordWithFile struct {
	AppId  string       `json:"app"`
	RecNum int          `json:"id"`
	Recode WriteRecord2 `json:"record"`
}
*/

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
	Memo Column `json:"memo"`
}
type RecodeForUpdate struct {
	AppId  string         `json:"app"`
	Recode FieldForUpdate `json:"record"`
}

// Kintoneのドメインと接続先アプリなどを設定する
type App struct {
	Domain       string
	AppId        int
	AppToken     string
	IsGuestSpace bool
	SpaceId      int
	Client       *http.Client  // Specialized client.
	Timeout      time.Duration // Timeout for API responses.
	Proxy        http.Transport
}
