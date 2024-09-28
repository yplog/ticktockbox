package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
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

func (d *Database) SetItem(item model.Item) error {
	return d.db.Update(func(txn *badger.Txn) error {
		item.ExpireDate = item.ExpireDate.UTC()
		itemJSON, err := json.Marshal(item)
		if err != nil {
			return err
		}
		return txn.Set([]byte(item.ID), itemJSON)
	})
}

func (d *Database) GetItem(id string) (model.Item, error) {
	var item model.Item
	err := d.db.View(func(txn *badger.Txn) error {
		i, err := txn.Get([]byte(id))
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

func (d *Database) DeleteItem(id string) error {
	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(id))
	})
}

func (d *Database) GetExpiredItems(now time.Time) ([]model.Item, error) {
	var expiredItems []model.Item

	log.Printf("Starting GetExpiredItems function with now=%s", now.Format(time.RFC3339))

	err := d.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			err := item.Value(func(v []byte) error {
				var storedItem model.Item
				if err := json.Unmarshal(v, &storedItem); err != nil {
					log.Printf("Error unmarshaling item: %v", err)
					return err
				}

				expireDate := storedItem.ExpireDate

				if now.After(expireDate) || now.Equal(expireDate) {
					expiredItems = append(expiredItems, storedItem)
					log.Printf("Expired item found: ID=%s, ExpireDate=%s, Now=%s",
						storedItem.ID, expireDate.Format(time.RFC3339), now.Format(time.RFC3339))
				} else {
					log.Printf("Item not expired: ID=%s, ExpireDate=%s, Now=%s",
						storedItem.ID, expireDate.Format(time.RFC3339), now.Format(time.RFC3339))
				}
				return nil
			})
			if err != nil {
				log.Printf("Error processing item: %v", err)
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("Error in GetExpiredItems: %v", err)
		return nil, fmt.Errorf("error getting expired items: %w", err)
	}

	log.Printf("Total expired items found: %d", len(expiredItems))

	return expiredItems, nil
}

func (d *Database) RunGC() error {
	return d.db.RunValueLogGC(0.5)
}
