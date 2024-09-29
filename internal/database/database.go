package database

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/yplog/ticktockbox/internal/model"
)

type Database struct {
	db *badger.DB
}

func NewDatabase(path string) (*Database, error) {
	opts := badger.DefaultOptions(path)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) SetItem(item model.Item, expireTime time.Time) error {
	return d.db.Update(func(txn *badger.Txn) error {
		itemJSON, err := json.Marshal(item)
		if err != nil {
			return err
		}
		return txn.Set(model.MakeKey(expireTime), itemJSON)
	})
}

func (d *Database) GetItem(expireTime time.Time) (model.Item, error) {
	var item model.Item
	err := d.db.View(func(txn *badger.Txn) error {
		i, err := txn.Get(model.MakeKey(expireTime))
		if err != nil {
			return err
		}
		return i.Value(func(val []byte) error {
			return json.Unmarshal(val, &item)
		})
	})
	if err == badger.ErrKeyNotFound {
		return item, errors.New("item not found")
	}
	return item, err
}

func (d *Database) DeleteItem(key []byte) error {
	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}
func (d *Database) GetExpiredItemsWithKeys(now time.Time) ([]model.Item, [][]byte, error) {
	var expiredItems []model.Item
	var expiredKeys [][]byte

	err := d.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()

		seekKey := model.MakeKey(now)

		for it.Seek(seekKey); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			expireTimeNano := binary.BigEndian.Uint64(k)
			expireTime := time.Unix(0, int64(expireTimeNano))

			if expireTime.After(now) {
				continue
			}

			err := item.Value(func(v []byte) error {
				var itemObj model.Item
				if err := json.Unmarshal(v, &itemObj); err != nil {
					return err
				}
				expiredItems = append(expiredItems, itemObj)
				expiredKeys = append(expiredKeys, item.KeyCopy(nil))
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return expiredItems, expiredKeys, err
}

func (d *Database) GetExpiredItems(now time.Time) ([]model.Item, error) {
	var expiredItems []model.Item

	err := d.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Reverse = true // Ters sıralama için
		it := txn.NewIterator(opts)
		defer it.Close()

		seekKey := model.MakeKey(now)

		for it.Seek(seekKey); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			expireTimeNano := binary.BigEndian.Uint64(k)
			expireTime := time.Unix(0, int64(expireTimeNano))

			if expireTime.After(now) {
				continue // Henüz süresi dolmamış öğeleri atla
			}

			err := item.Value(func(v []byte) error {
				var itemObj model.Item
				if err := json.Unmarshal(v, &itemObj); err != nil {
					return err
				}
				expiredItems = append(expiredItems, itemObj)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return expiredItems, err
}

func (d *Database) RunGC() error {
	return d.db.RunValueLogGC(0.5)
}
