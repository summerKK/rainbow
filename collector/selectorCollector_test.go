package collector

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/cihub/seelog"
	"rainbow/result"
)

var configs *Configs
var selectorCollector Collector

func init() {
	rand.Seed(time.Now().Unix())
	configs = NewCollectorConfig("collectorConfig_test.xml")
	var config *Config
	var n int
	l := len(configs.Configs) - 1
	for n = rand.Intn(l); configs.Configs[n].Type != COLLECTOR_TYPE_SELECTOR; n = rand.Intn(l) {

	}
	config = configs.Configs[n]

	selectorCollector = NewSelectorCollector(config)
	if selectorCollector == nil {
		panic("初始化失败")
	}
}

func TestSelectorCollector(t *testing.T) {
	resultChan := make(chan *result.Result)
	doneChan := make(chan struct{})
	go func() {
		for selectorCollector.Next() {
			errorList := selectorCollector.Collection(resultChan)
			if len(errorList) > 0 {
				_ = seelog.Error(errorList)
			}
		}
		doneChan <- struct{}{}
	}()

	go func() {
		for r := range resultChan {
			fmt.Printf("%+v\n", r)
		}
	}()

	<-doneChan
	close(resultChan)
	time.Sleep(time.Second)
}
