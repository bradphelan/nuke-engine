package engine

import (
	"encoding/json"
	"time"

	bolt "go.etcd.io/bbolt"
)

var bucketName = []byte("builds")

type CacheEntry struct {
	OutputURIs     []string  `json:"output_uris"`
	OutputFPs      []string  `json:"output_fps"`
	DiscoveredURIs []string  `json:"discovered_uris"`
	DiscoveredFPs  []string  `json:"discovered_fps"`
	Timestamp      time.Time `json:"timestamp"`
}

type Cache struct {
	db *bolt.DB
}

func NewCache(path string) (*Cache, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return &Cache{db: db}, nil
}

func (c *Cache) Close() error { return c.db.Close() }

func (c *Cache) Store(key string, entry *CacheEntry) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		data, err := json.Marshal(entry)
		if err != nil {
			return err
		}
		return b.Put([]byte(key), data)
	})
}

func (c *Cache) Lookup(key string) (*CacheEntry, error) {
	var entry *CacheEntry
	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		data := b.Get([]byte(key))
		if data == nil {
			return nil
		}
		entry = &CacheEntry{}
		return json.Unmarshal(data, entry)
	})
	return entry, err
}
