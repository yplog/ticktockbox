package database

import (
	"github.com/dgraph-io/badger/v4"
)

type Database struct {
	db *badger.DB
}

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

/*func (d *Database) SetItem(item *model.Item) error {
	return d.db.Update(func(txn *badger.Txn) error {
		key := model.MakeKey(item.ExpireTime)
		value, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed to marshal item: %v", err)
		}
		log.Printf("Setting item with key: %x, value: %s", key, string(value))
		return txn.Set(key, value)
	})
}

func (d *Database) GetItem(expireTime time.Time) (*model.Item, error) {
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
	if err == badger.ErrKeyNotFound {
		return nil, errors.New("item not found")
	}
	return item, err
}

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

func (d *Database) GetExpiredItems(now time.Time) ([]*model.Item, error) {
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
}

func (d *Database) DeleteAndVerify(key []byte) error {
	return d.db.Update(func(txn *badger.Txn) error {
		if err := txn.Delete(key); err != nil {
			return fmt.Errorf("failed to delete item: %v", err)
		}

		_, err := txn.Get(key)
		if err == nil {
			return fmt.Errorf("item still exists after deletion")
		} else if err != badger.ErrKeyNotFound {
			return fmt.Errorf("unexpected error during verification: %v", err)
		}

		log.Printf("Item with key %x successfully deleted and verified", key)
		return nil
	})
}*/

func (d *Database) RunGC() error {
	return d.db.RunValueLogGC(0.5)
}
