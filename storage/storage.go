package storage

type Storage interface {
	Exist(k string) (ok bool)
	Get(k string) (v string)
	Delete(k string) (ok bool)
	AddOrUpdate(k string, v interface{}) (err error)
	GetAll() (collection map[string]string)
	Close()
	GetRandomOne() (v string)
}
