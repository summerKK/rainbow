package storage

import (
	"fmt"
	"os"
	"testing"
)

var blotDbStorage *BoltDbStorage
var path = "test.db"

type value struct {
	Name string `json:"name"`
}

var testData = []struct {
	k    string
	v    *value
	want string
}{
	{k: "0", v: &value{"hello,world"}, want: `{"name":"hello,world"}`},
	{k: "1", v: &value{"hello,world"}, want: `{"name":"hello,world"}`},
	{k: "2", v: &value{"hello,world"}, want: `{"name":"hello,world"}`},
	{k: "3", v: nil, want: ""},
}

func init() {
	var err error
	if _, err = os.Stat(path); !os.IsNotExist(err) {
		err = os.Remove(path)
		if err != nil {
			panic(err)
		}
	}

	blotDbStorage, err = NewBlotDbStorage(path, "testBucket")
	if err != nil {
		panic(err)
	}
}

func TestBoltDbStorage_Get(t *testing.T) {
	for _, datum := range testData {
		if datum.v == nil {
			continue
		}
		err := blotDbStorage.AddOrUpdate(datum.k, datum.v)
		if err != nil {
			fmt.Println(err)
		}
	}

	for _, datum := range testData {
		if v := blotDbStorage.Get(datum.k); v != datum.want {
			t.Errorf("BlotDbStorage.Get() = %v,want %v\n", v, datum.want)
		}
	}
}

func TestBoltDbStorage_GetAll(t *testing.T) {
	var keys []string
	for _, datum := range testData {
		if datum.v == nil {
			continue
		}
		keys = append(keys, datum.k)
		err := blotDbStorage.AddOrUpdate(datum.k, datum.v)
		if err != nil {
			fmt.Println(err)
		}
	}

	collection := blotDbStorage.GetAll()
	for _, key := range keys {
		if _, ok := collection[key]; !ok {
			t.Errorf("BlotDbStorage.GetAll() %v key not exists\n", key)
		}

	}

}

func TestBoltDbStorage_GetRandomOne(t *testing.T) {
	for _, datum := range testData {
		if datum.v == nil {
			continue
		}
		err := blotDbStorage.AddOrUpdate(datum.k, datum.v)
		if err != nil {
			fmt.Println(err)
		}
	}

	v := blotDbStorage.GetRandomOne()
	if v == "" {
		t.Errorf("BlotDbStorage.GetRandomOne() failed\n")
	}
}

func TestBoltDbStorage_Exist(t *testing.T) {
	var key string
	for _, datum := range testData {
		if datum.v == nil {
			continue
		}
		key = datum.k
		err := blotDbStorage.AddOrUpdate(datum.k, datum.v)
		if err != nil {
			fmt.Println(err)
		}
	}
	ok := blotDbStorage.Exist(key)
	if !ok {
		t.Errorf("BlotDbStorage.Exist(key) failed\n")
	}
}
