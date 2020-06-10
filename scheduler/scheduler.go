package scheduler

import (
	"os"
	"sync"
	"time"

	"github.com/cihub/seelog"
	"rainbow/collector"
	"rainbow/server"
	"rainbow/storage"
	"rainbow/verify"
)

func Run(logConfigFile, collectorConfigFile, path, bucket string) {
	var wg sync.WaitGroup
	manager, err := collector.NewManager(collectorConfigFile)

	if err != nil {
		panic(err)
	}

	s, err := storage.NewStorage(path, bucket)
	if err != nil {
		panic(err)
	}

	// 开始爬取
	go func() {
		wg.Add(1)
		defer wg.Done()
		manager.Run()
	}()

	err = SetLogger(logConfigFile)
	if err != nil {
		panic(err)
	}
	defer seelog.Flush()

	// 把爬取的数据存下来
	go func() {
		wg.Add(1)
		defer wg.Done()
		err := verify.ValidationAndSave(manager.ResultChan(), s)
		if err != nil {
			panic(err)
		}
	}()

	// 监控ip是否有效
	go func() {
		ticker := time.NewTicker(time.Minute * 10)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := verify.ValidationAndDelete(s)
				if err != nil {
					panic(err)
				}
			}
		}
	}()

	// 开启http服务
	go func() {
		err := server.NewServer(s)
		if err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}

func SetLogger(fileName string) error {
	if _, err := os.Stat(fileName); err == nil {
		logger, err := seelog.LoggerFromConfigAsFile(fileName)
		if err != nil {
			return err
		}

		_ = seelog.ReplaceLogger(logger)
	} else {
		configString := `<seelog>
                        <outputs formatid="main">
                            <filter levels="info,error,critical">
                                <rollingfile type="date" filename="log/AppLog.log" namemode="prefix" datepattern="02.01.2006"/>
                            </filter>
                            <console/>
                        </outputs>
                        <formats>
                            <format id="main" format="%Date %Time [%LEVEL] %Msg%n"/>
                        </formats>
                        </seelog>`
		logger, err := seelog.LoggerFromConfigAsString(configString)
		if err != nil {
			return err
		}

		_ = seelog.ReplaceLogger(logger)
	}

	seelog.Info("log initialize finish.")

	return nil
}
