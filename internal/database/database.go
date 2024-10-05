package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/yplog/ticktockbox/internal/model"
	"log"
)

type Database struct {
	db *badger.DB
}

var (
	idKey        = []byte("last_id")
	expirePrefix = []byte("expire_")
	dataPrefix   = []byte("data_")
)

func NewDatabase(path string) (*Database, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) GetNextID() (model.ID, error) {
	var nextID model.ID

	err := d.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(idKey)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				nextID = model.ID(1)
				return txn.Set(idKey, nextID.ToBytes())
			}
			return err
		}

		err = item.Value(func(val []byte) error {
			currentID := model.IDFromBytes(val)
			nextID = model.ID(currentID.ToUint64() + 1)
			return nil
		})
		if err != nil {
			return err
		}

		return txn.Set(idKey, nextID.ToBytes())
	})

	if err != nil {
		return 0, err
	}

	return nextID, nil
}

func (d *Database) SetExpireData(key *model.ID, value *model.ExpireData) error {
	return d.db.Update(func(txn *badger.Txn) error {
		value, err := json.Marshal(value)

		if err != nil {
			return fmt.Errorf("failed to marshal item: %v", err)
		}

		kwp := makeKeyWithPrefix(expirePrefix, key)

		log.Printf("Setting ExpireData with key: %s, value: %s", string(kwp), string(value))

		return txn.Set(kwp, value)
	})
}

func (d *Database) SetData(key *model.ID, value *model.Data) error {
	return d.db.Update(func(txn *badger.Txn) error {
		value, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal item: %v", err)
		}

		kwp := makeKeyWithPrefix(dataPrefix, key)

		log.Printf("Setting Data with key: %s, value: %s", string(kwp), string(value))

		return txn.Set(kwp, value)
	})
}

func (d *Database) GetExpireData(key *model.ID) (*model.ExpireData, error) {
	var value *model.ExpireData

	err := d.db.View(func(txn *badger.Txn) error {
		kwp := makeKeyWithPrefix(expirePrefix, key)

		i, err := txn.Get(kwp)
		if err != nil {
			return err
		}
		return i.Value(func(val []byte) error {
			var err error
			value = model.ExpireDataFromBytes(val)
			return err
		})
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, errors.New("item not found")
	}

	return value, err
}

func (d *Database) GetData(key *model.ID) (*model.Data, error) {
	var data *model.Data

	err := d.db.View(func(txn *badger.Txn) error {
		kwp := makeKeyWithPrefix(dataPrefix, key)

		i, err := txn.Get(kwp)
		if err != nil {
			return err
		}
		return i.Value(func(val []byte) error {
			var err error
			data = model.DataFromBytes(val)
			return err
		})
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, errors.New("item not found")
	}

	return data, err
}

/*func (d *Database) GetItem(expireTime time.Time) (*model.Item, error) {
	var item *model.Item
	err := d.db.View(func(txn *badger.Txn) error {
		i, err := txn.Get(model.MakeKey(expireTime))
		if err != nil {
			return err
		}
		return i.Value(func(val []byte) error {
			var err error
			item, err = model.ItemFromValue(val)
			return err
		})
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, errors.New("item not found")
	}
	return item, err
}*/

func (d *Database) DeleteItem(key []byte) error {
	return d.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(key)
		if err != nil {
			return fmt.Errorf("failed to delete item: %v", err)
		}
		log.Printf("Item with key %x successfully deleted from database", key)
		return nil
	})
}

/*func (d *Database) GetExpiredItems(now time.Time) ([]*model.Item, error) {
	var expiredItems []*model.Item

	err := d.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		opts.Prefix = []byte(model.PrefixExpire)
		it := txn.NewIterator(opts)

		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			k := item.Key()
			if len(k) < model.PrefixLength+8 {
				continue
			}

			expireTimeNano := binary.BigEndian.Uint64(k[model.PrefixLength : model.PrefixLength+8])
			expireTime := time.Unix(0, int64(expireTimeNano)).UTC()

			log.Printf("Checking item with key: %x, expire time: %s", k, expireTime.Format(time.RFC3339))

			if expireTime.After(now) {
				log.Printf("Found non-expired item, stopping iteration")
				break
			}
			err := item.Value(func(v []byte) error {
				log.Printf("Raw item value: %s", string(v))
				itemObj, err := model.ItemFromValue(v)
				if err != nil {
					return fmt.Errorf("failed to unmarshal item: %v, raw value: %s", err, string(v))
				}
				itemObj.ExpireTime = expireTime
				expiredItems = append(expiredItems, itemObj)
				log.Printf("Added expired item to list: %+v", itemObj)
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return expiredItems, err
}*/

func (d *Database) RunGC() error {
	return d.db.RunValueLogGC(0.5)
}

func makeKeyWithPrefix(prefix []byte, key *model.ID) []byte {
	keyBytes := key.ToBytes()

	return append(prefix, keyBytes...)
}
