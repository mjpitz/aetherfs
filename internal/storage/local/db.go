// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package local

import (
	"context"
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
	"go.uber.org/zap"

	"github.com/mjpitz/myago/dirset"
	"github.com/mjpitz/myago/vfs"
	"github.com/mjpitz/myago/zaputil"
)

// SetupDB will setup the local database and attach it to the provided context.
func SetupDB(ctx context.Context, dirs dirset.DirectorySet) (context.Context, error) {
	path := filepath.Join(dirs.LocalStateDir, "localdb")
	passphrase := os.Getenv("STORAGE_LOCAL_PASSPHRASE")

	err := vfs.Extract(ctx).MkdirAll(path, 0755)
	if err != nil {
		return nil, err
	}

	db, err := Open(ctx, path, passphrase)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, contextKey, db), nil
}

// Open opens the database at dir using the provided passphrase. We generally want a passphrase to avoid having
// credentials written to disk in plaintext.
func Open(ctx context.Context, path string, passphrase string) (*DB, error) {
	hash := sha256.Sum256([]byte(passphrase))
	opts := badger.DefaultOptions(path).
		WithEncryptionKey(hash[:]).
		WithIndexCacheSize(100 << 20).
		WithLogger(zaputil.Badger(zap.NewNop()))

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &DB{
		db: db,
	}, nil
}

// DB encapsulates the set of values
type DB struct {
	db *badger.DB
}

func (d *DB) Credentials() *Store {
	return &Store{
		db:     d.db,
		prefix: "credentials",
	}
}

func (d *DB) Tokens() *Store {
	return &Store{
		db:     d.db,
		prefix: "tokens",
	}
}

func (d *DB) Close() error {
	return d.db.Close()
}

var _ io.Closer = &DB{}
