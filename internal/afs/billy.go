// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package afs

import (
	"os"
	"strings"
	"sync"

	"github.com/go-git/go-billy/v5"
	"github.com/spf13/afero"
)

func Billy(fs afero.Fs) billy.Filesystem {
	return &billyFS{fs: fs, root: ""}
}

type billyFS struct {
	fs   afero.Fs
	root string
}

// Capabilities exports the filesystem as readonly
func (billyFS) Capabilities() billy.Capability {
	return billy.ReadCapability | billy.SeekCapability
}

func (b *billyFS) Create(filename string) (billy.File, error) {
	file, err := b.fs.Create(filename)
	if err != nil {
		return nil, err
	}

	return &billyFile{ file: file }, nil
}

func (b *billyFS) Open(filename string) (billy.File, error) {
	file, err := b.fs.Open(b.Join(b.root, filename))
	if err != nil {
		return nil, err
	}

	return &billyFile{ file: file }, nil
}

func (b *billyFS) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	file, err := b.fs.OpenFile(b.Join(b.root, filename), flag, perm)
	if err != nil {
		return nil, err
	}

	return &billyFile{ file: file }, nil
}

func (b *billyFS) Stat(filename string) (os.FileInfo, error) {
	return b.fs.Stat(b.Join(b.root, filename))
}

func (b *billyFS) Rename(oldpath, newpath string) error {
	return b.fs.Rename(b.Join(b.root, oldpath), b.Join(b.root, newpath))
}

func (b *billyFS) Remove(filename string) error {
	return b.fs.Remove(b.Join(b.root, filename))
}

func (b *billyFS) Join(elem ...string) string {
	res := ""

	for i := len(elem); i > 0; i-- {
		if elem[i-1] != "" {
			res = "/" + elem[i-1] + res
		}
	}

	return strings.Join(elem, "/")
}

func (b *billyFS) ReadDir(path string) ([]os.FileInfo, error) {
	f, err := b.fs.Open(b.Join(b.root, path))
	if err != nil {
		return nil, err
	}

	return f.Readdir(0)
}

func (b *billyFS) MkdirAll(filename string, perm os.FileMode) error {
	return b.fs.MkdirAll(b.Join(b.root, filename), perm)
}

func (b *billyFS) Lstat(filename string) (os.FileInfo, error) {
	return b.fs.Stat(b.Join(b.root, filename))
}

func (b *billyFS) Chroot(path string) (billy.Filesystem, error) {
	return &billyFS{
		fs:   b.fs,
		root: b.Join(b.root, path),
	}, nil
}

func (b *billyFS) Root() string {
	return b.root
}

// todo

func (b *billyFS) Readlink(link string) (string, error) {
	return "", billy.ErrNotSupported
}

// unsupported

func (b *billyFS) Symlink(target, link string) error {
	return billy.ErrNotSupported
}

func (b *billyFS) TempFile(dir, prefix string) (billy.File, error) {
	panic("implement me")
}

var _ billy.Filesystem = &billyFS{}



type billyFile struct {
	mu   sync.Mutex
	file afero.File
}

func (b *billyFile) Name() string {
	return b.file.Name()
}

func (b *billyFile) Write(p []byte) (n int, err error) {
	return b.file.Write(p)
}

func (b *billyFile) Read(p []byte) (n int, err error) {
	return b.file.Read(p)
}

func (b *billyFile) ReadAt(p []byte, off int64) (n int, err error) {
	return b.file.ReadAt(p, off)
}

func (b *billyFile) Seek(offset int64, whence int) (int64, error) {
	return b.file.Seek(offset, whence)
}

func (b *billyFile) Close() error {
	return b.file.Close()
}

func (b *billyFile) Lock() error {
	b.mu.Lock()
	return nil
}

func (b *billyFile) Unlock() error {
	b.mu.Unlock()
	return nil
}

func (b *billyFile) Truncate(size int64) error {
	return b.file.Truncate(size)
}

var _ billy.File = &billyFile{}
