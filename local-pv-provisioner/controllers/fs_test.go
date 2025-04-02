package controllers

import (
	"errors"
	"time"
)

type notImplementedFS struct{}

var _ fileSystem = &notImplementedFS{}

func (fs *notImplementedFS) Open(name string) (file, error) {
	return nil, errors.New("not implemented")
}
func (fs *notImplementedFS) Stat(name string) (FileInfo, error) {
	return nil, errors.New("not implemented")
}
func (fs *notImplementedFS) OpenFile(name string, flag int, perm FileMode) (file, error) {
	return nil, errors.New("not implemented")
}
func (fs *notImplementedFS) Walk(root string, fn func(path string, info FileInfo, err error) error) error {
	return errors.New("not implemented")
}
func (fs *notImplementedFS) MkdirAll(path string, perm FileMode) error {
	return errors.New("not implemented")
}
func (fs *notImplementedFS) Remove(name string) error {
	return errors.New("not implemented")
}

type constFileInfo struct {
	name    string
	size    int64
	mode    FileMode
	modTime time.Time
	isDir   bool
	sys     any
}

var _ FileInfo = &constFileInfo{}

func (fi *constFileInfo) Name() string {
	return fi.name
}
func (fi *constFileInfo) Size() int64 {
	return fi.size
}
func (fi *constFileInfo) Mode() FileMode {
	return fi.mode
}
func (fi *constFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi *constFileInfo) IsDir() bool {
	return fi.isDir
}
func (fi *constFileInfo) Sys() any {
	return fi.sys
}
