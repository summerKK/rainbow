package collector

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

var configs *Configs
var selectorCollector Collector
var manager *Manager

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

	manager = NewManager(selectorCollector)
}

func TestSelectorCollector(t *testing.T) {
	manager.Run()
	var wg = &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for result := range manager.ResultChan() {
			fmt.Printf("%+v\n", result)
		}
		wg.Done()
	}()

	go func() {
		for err := range manager.ErrorChan() {
			fmt.Printf("%+v\n", err)
		}
		wg.Done()
	}()

	wg.Wait()
	fmt.Println("采集完成")
}
