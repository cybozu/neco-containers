package controllers

import (
	"errors"
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
