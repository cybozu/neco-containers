package controllers

import (
	"io"
	"os"
	"path/filepath"
)

// cf. https://go.dev/talks/2012/10things.slide#8
var fs fileSystem = osFS{}

type fileSystem interface {
	Open(name string) (file, error)
	Stat(name string) (FileInfo, error)
	OpenFile(name string, flag int, perm FileMode) (file, error)
	Walk(root string, fn func(path string, info FileInfo, err error) error) error
	MkdirAll(path string, perm FileMode) error
}

type file interface {
	io.Closer
	io.Writer
	io.Seeker
	Stat() (FileInfo, error)
	Sync() error
}

// osFS implements fileSystem using the local disk.
type osFS struct{}

func (osFS) Open(name string) (file, error)     { return os.Open(name) }
func (osFS) Stat(name string) (FileInfo, error) { return os.Stat(name) }
func (osFS) OpenFile(name string, flag int, perm FileMode) (file, error) {
	return os.OpenFile(name, flag, perm)
}
func (osFS) Walk(root string, f func(path string, info FileInfo, err error) error) error {
	return filepath.Walk(root, f)
}
func (osFS) MkdirAll(path string, perm FileMode) error {
	return os.MkdirAll(path, perm)
}

const O_WRONLY = os.O_WRONLY
const O_CREATE = os.O_CREATE

type FileInfo = os.FileInfo
type FileMode = os.FileMode
