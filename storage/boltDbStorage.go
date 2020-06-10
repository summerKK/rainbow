package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	bolt "go.etcd.io/bbolt"
)

type BoltDbStorage struct {
	db         *bolt.DB
	bucketName []byte
	content    sync.Map
	contentKey []string
	count      uint64
}

func NewBlotDbStorage(path string, bucketName string) (boltDb *BoltDbStorage, err error) {
	if path == "" || bucketName == "" {
		err = fmt.Errorf("nil path or nil bucketName.(path:%s,bucketName:%s)", path, bucketName)
		return
	}

	db, err := bolt.Open(path, 0666, &bolt.Options{Timeout: time.Second * 1})
	if err != nil {
		return
	}

	bucketBytes := []byte(bucketName)
	// 创建bucket
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(bucketBytes)
		return err
	})

	if err != nil {
		return
	}

	boltDb = &BoltDbStorage{
		db:         db,
		bucketName: bucketBytes,
		contentKey: make([]string, 64),
	}

	// 同步数据到内存中
	err = boltDb.sync()

	return
}

func (b *BoltDbStorage) Exist(k string) (ok bool) {
	return b.Get(k) != ""
}

func (b *BoltDbStorage) Get(k string) (v string) {
	if value, ok := b.content.Load(k); ok {
		if s, ok := value.(string); ok {
			v = s
		}
	}

	return
}

func (b *BoltDbStorage) Delete(k string) (ok bool) {
	err := b.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(b.bucketName).Delete([]byte(k))
		return err
	})

	if err == nil {
		ok = true
		if _, o := b.content.Load(k); o {
			b.content.Delete(k)
			atomic.AddUint64(&b.count, ^uint64(0))
		}
	}

	return
}

func (b *BoltDbStorage) AddOrUpdate(k string, v interface{}) (err error) {
	if v == nil {
		err = errors.New("nil v")
		return
	}
	jsonStr, err := json.Marshal(v)
	if err != nil {
		return
	}

	err = b.db.Update(func(tx *bolt.Tx) error {
		err = tx.Bucket(b.bucketName).Put([]byte(k), jsonStr)
		return err
	})

	if err == nil {
		if _, loaded := b.content.LoadOrStore(k, string(jsonStr)); !loaded {
			atomic.AddUint64(&b.count, 1)
		}
	}

	return
}

func (b *BoltDbStorage) GetAll() (collection map[string]string) {
	collection = make(map[string]string, b.count)
	b.content.Range(func(key, value interface{}) bool {
		if k, ok := key.(string); ok {
			if v, ok := value.(string); ok {
				collection[k] = v
			}
		}
		return true
	})

	return collection
}

func (b *BoltDbStorage) Close() {
	_ = b.db.Close()
}

// TODO 随机获取待优化
func (b *BoltDbStorage) GetRandomOne() (v string) {
	var defaultKey string
	var randomKey string
	rand.Seed(time.Now().Unix())
	var n int
	if b.count > 0 {
		n = rand.Intn(int(b.count))
	} else {
		n = 0
	}
	b.content.Range(func(key, value interface{}) bool {
		if defaultKey == "" {
			defaultKey = key.(string)
		}
		if n == 0 {
			randomKey = key.(string)
		}
		n--
		return true
	})

	if randomKey == "" {
		randomKey = defaultKey
	}

	return b.Get(randomKey)
}

func (b *BoltDbStorage) sync() error {
	err := b.db.View(func(tx *bolt.Tx) error {
		err := tx.Bucket(b.bucketName).ForEach(func(k, v []byte) error {
			key, value := make([]byte, len(k)), make([]byte, len(v))
			copy(key, k)
			copy(value, v)
			b.content.Store(string(key), string(value))
			atomic.AddUint64(&b.count, 1)
			return nil
		})

		return err
	})

	return err
}
