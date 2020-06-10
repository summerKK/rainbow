package verify

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/cihub/seelog"
	"rainbow/result"
	"rainbow/storage"
	"rainbow/util"
)

const chunkCount = 10

type Verify struct {
	saveChan   chan *result.Result
	deleteChan chan *result.Result
	s          storage.Storage
	resultChan <-chan *result.Result
}

func NewVerify(resultChan <-chan *result.Result, s storage.Storage) (v *Verify, e error) {
	if resultChan == nil || s == nil {
		return nil, errors.New("nil resultChan/storage")
	}

	v = &Verify{
		saveChan:   make(chan *result.Result, 50),
		deleteChan: make(chan *result.Result, 50),
		s:          s,
		resultChan: resultChan,
	}
	go v.validation()

	return v, nil
}

func (verify *Verify) ValidationAndDelete() {
	collection := verify.s.GetAll()
	var res result.Result
	for _, v := range collection {
		err := json.Unmarshal([]byte(v), &res)
		if err != nil {
			continue
		}
		verify.deleteChan <- &res
	}
}

func (verify *Verify) ValidationAndSave() {
	for r := range verify.resultChan {
		verify.saveChan <- r
	}
}

func (verify *Verify) validation() {
	go func() {
		for res := range verify.saveChan {
			go func(res *result.Result) {
				if util.VerifyProxyIp(res.Ip, res.Port) {
					_ = verify.s.AddOrUpdate(res.Ip, res)
					seelog.Debugf("insert %s to DB", res.Ip)
				}
			}(res)
		}
	}()

	go func() {
		for res := range verify.deleteChan {
			go func(res *result.Result) {
				if !util.VerifyProxyIp(res.Ip, res.Port) {
					verify.s.Delete(res.Ip)
				}
			}(res)
		}
	}()
}

// 先获取所有数据,然后对数据进行分块(chunkCount).然后通过 chunkCount个goroutine并发验证IP
// 需要注意最后一次的分块可能小于标准分块数量.需要额外处理(len(collection) % chunkCount)
func ValidationAndDelete(s storage.Storage) error {
	if s == nil {
		return errors.New("nil storage")
	}

	var wg sync.WaitGroup
	collection := s.GetAll()
	chunk := len(collection) / chunkCount
	lastChunk := len(collection) % chunkCount
	count := 0
	rangeLen := chunkCount
	if lastChunk > 0 {
		rangeLen += 1
	}
	for i := 0; i < rangeLen; i++ {
		var temp []*result.Result
		for k, v := range collection {
			var res result.Result
			count++
			err := json.Unmarshal([]byte(v), &res)
			if err != nil {
				s.Delete(k)
				continue
			}
			if count >= chunk || ((i == chunkCount-1) && count == lastChunk) {
				count = 0
				go func(temp []*result.Result, k string, wg *sync.WaitGroup) {
					defer wg.Done()
					for _, res := range temp {
						if !util.VerifyProxyIp(res.Ip, res.Port) {
							s.Delete(k)
						}
					}
				}(temp, k, &wg)
				wg.Add(1)
			} else {
				temp = append(temp, &res)
				continue
			}
		}
	}

	wg.Wait()

	return nil
}

func ValidationAndSave(resultChan <-chan *result.Result, s storage.Storage) error {
	if resultChan == nil || s == nil {
		return errors.New("nil resultChan/storage")
	}

	for r := range resultChan {
		if util.VerifyProxyIp(r.Ip, r.Port) {
			_ = s.AddOrUpdate(r.Ip, r)
			seelog.Debugf("insert %s to DB", r.Ip)
		}
	}

	return nil
}
