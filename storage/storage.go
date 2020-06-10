package storage

import "errors"

type Storage interface {
	Exist(k string) (ok bool)
	Get(k string) (v string)
	Delete(k string) (ok bool)
	AddOrUpdate(k string, v interface{}) (err error)
	GetAll() (collection map[string]string)
	Close()
	GetRandomOne() (v string)
	Len() uint64
}

func NewStorage(path string, bucket string) (Storage, error) {
	if path == "" || bucket == "" {
		return nil, errors.New("nil path/bucket")
	}

	return NewBlotDbStorage(path, bucket)
}
