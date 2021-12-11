// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package afs

import (
	"context"
	"io/fs"
	"os"

	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

func Webdav(fs afero.Fs) *webdav.Handler {
	return &webdav.Handler{
		FileSystem: &aferoWebdev{fs: fs},
		LockSystem: webdav.NewMemLS(),
	}
}

type aferoWebdev struct {
	fs afero.Fs
}

func (a *aferoWebdev) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return a.fs.Mkdir(name, perm)
}

func (a *aferoWebdev) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	return a.fs.OpenFile(name, flag, perm)
}

func (a *aferoWebdev) RemoveAll(ctx context.Context, name string) error {
	return a.fs.RemoveAll(name)
}

func (a *aferoWebdev) Rename(ctx context.Context, oldName, newName string) error {
	return a.fs.Rename(oldName, newName)
}

func (a *aferoWebdev) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return a.fs.Stat(name)
}

var _ webdav.FileSystem = &aferoWebdev{}

type aferoWebdavFile struct {
	file afero.File
}

func (a *aferoWebdavFile) Close() error {
	return a.file.Close()
}

func (a *aferoWebdavFile) Read(p []byte) (n int, err error) {
	return a.file.Read(p)
}

func (a *aferoWebdavFile) Seek(offset int64, whence int) (int64, error) {
	return a.file.Seek(offset, whence)
}

func (a *aferoWebdavFile) Readdir(count int) ([]fs.FileInfo, error) {
	return a.file.Readdir(count)
}

func (a *aferoWebdavFile) Stat() (fs.FileInfo, error) {
	return a.file.Stat()
}

func (a *aferoWebdavFile) Write(p []byte) (n int, err error) {
	return a.file.Write(p)
}

var _ webdav.File = &aferoWebdavFile{}
