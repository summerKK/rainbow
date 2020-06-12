package collector

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cihub/seelog"
	"rainbow/result"
)

type Manager struct {
	collectors []Collector
	resultChan chan *result.Result
	errorChan  chan error
	doneChan   chan struct{}
}

func NewManager(collectorConfigFile string) (*Manager, error) {

	if _, err := os.Stat(collectorConfigFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s not exists", collectorConfigFile)
	}

	configs := NewCollectorConfig(collectorConfigFile)
	if len(configs.Configs) == 0 {
		return nil, fmt.Errorf("%s parse failed", collectorConfigFile)
	}

	var collectors []Collector
	for _, config := range configs.Configs {
		collector, err := config.Collector()
		if err != nil {
			_ = seelog.Error(err)
			continue
		}
		collectors = append(collectors, collector)
	}

	m := &Manager{
		resultChan: make(chan *result.Result),
		errorChan:  make(chan error),
		collectors: collectors,
	}

	return m, nil
}

func (m *Manager) Run() {
	defer func() {
		close(m.resultChan)
		close(m.errorChan)
	}()

	for {
		var wg sync.WaitGroup
		for _, collector := range m.collectors {
			wg.Add(1)
			go func(c Collector) {
				m.runCollector(c, &wg)
			}(collector)
		}

		wg.Wait()
		// 所有网站都爬取完了,间歇30分钟
		seelog.Info("所有网站都爬取完了,间歇15分钟")
		time.Sleep(time.Minute * 15)
	}
}

func (m *Manager) runCollector(collector Collector, wg *sync.WaitGroup) {
	defer wg.Done()
	for collector.Next() {
		errorList := collector.Collection(m.resultChan)
		if len(errorList) > 0 {
			for _, err := range errorList {
				m.errorChan <- err
			}
		}
		// 间歇一下,避免爬取太快被封IP
		time.Sleep(time.Second * time.Duration(collector.Config().Interval))
	}
}

func (m *Manager) ResultChan() <-chan *result.Result {
	return m.resultChan
}

func (m *Manager) ErrorChan() <-chan error {
	return m.errorChan
}

func (m *Manager) DoneChan() <-chan struct{} {
	return m.doneChan
}
