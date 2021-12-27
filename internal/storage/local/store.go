// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package local

import (
	"context"
	"encoding/json"

	"github.com/dgraph-io/badger/v3"
)

// Store provides a generic key/value interface backed by badger. Stores are obtained from the database to ensure
// prefix do not collide.
type Store struct {
	db     *badger.DB
	prefix string
}

func (c *Store) Get(ctx context.Context, key string, value interface{}) (err error) {
	return c.db.View(func(txn *badger.Txn) error {
		k := []byte(c.prefix + "/" + key)
		found, err := txn.Get(k)
		if err != nil {
			return err
		}

		return found.Value(func(val []byte) error {
			return json.Unmarshal(val, value)
		})
	})
}

func (c *Store) Put(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.db.Update(func(txn *badger.Txn) error {
		k := []byte(c.prefix + "/" + key)

		return txn.Set(k, data)
	})
}

func (c *Store) Delete(ctx context.Context, key string) error {
	return c.db.Update(func(txn *badger.Txn) error {
		k := []byte(c.prefix + "/" + key)

		return txn.Delete(k)
	})
}
