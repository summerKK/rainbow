package collector

import (
	"sync"
	"time"

	"rainbow/result"
)

type Manager struct {
	collectors []Collector
	resultChan chan *result.Result
	errorChan  chan error
	doneChan   chan struct{}
}

func NewManager(collectors ...Collector) *Manager {
	m := &Manager{
		resultChan: make(chan *result.Result),
		errorChan:  make(chan error),
		collectors: collectors,
	}

	return m
}

func (m *Manager) Run() {
	go func() {
		var wg sync.WaitGroup
		for _, collector := range m.collectors {
			wg.Add(1)
			go func(c Collector) {
				m.runCollector(c, &wg)
			}(collector)
		}
		wg.Wait()

		close(m.resultChan)
		close(m.errorChan)
	}()
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
