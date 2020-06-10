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
	// 控制并发
	saveChan chan struct{}
	// 控制并发
	deleteChan chan struct{}
	s          storage.Storage
	resultChan <-chan *result.Result
}

func NewVerify(resultChan <-chan *result.Result, s storage.Storage) (v *Verify, e error) {
	if resultChan == nil || s == nil {
		return nil, errors.New("nil resultChan/storage")
	}

	v = &Verify{
		saveChan:   make(chan struct{}, 20),
		deleteChan: make(chan struct{}, 20),
		s:          s,
		resultChan: resultChan,
	}

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
		verify.deleteChan <- struct{}{}
		go func() {
			defer func() {
				<-verify.deleteChan
			}()

			if !util.VerifyProxyIp(res.Ip, res.Port) {
				verify.s.Delete(res.Ip)
				seelog.Debugf("delete %s from DB", res.Ip)
			}
		}()
	}
}

func (verify *Verify) ValidationAndSave() {
	for res := range verify.resultChan {
		verify.saveChan <- struct{}{}
		go func(res *result.Result) {
			defer func() {
				<-verify.saveChan
			}()

			if util.VerifyProxyIp(res.Ip, res.Port) {
				_ = verify.s.AddOrUpdate(res.Ip, res)
				seelog.Debugf("insert %s to DB", res.Ip)
			} else {
				seelog.Debugf("ignore %s", res.Ip)
			}
		}(res)
	}
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
