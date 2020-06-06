package collector

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cihub/seelog"
	"rainbow/result"
	"rainbow/util"
)

type subCollector struct {
	// 配置文件
	config *Config
	//  当前url
	currentUrl   string
	currentIndex int
	urls         []string
	selectorMap  map[string][]string
}

func newSubCollector(config *Config, t Type) (*subCollector, error) {
	if config == nil {
		return nil, errors.New("nil config")
	}

	// 验证参数
	if !config.Verify() || config.Type != t {
		return nil, fmt.Errorf("参数配置错误:%+v", config)
	}

	selectorMap := make(map[string][]string)

	for _, item := range config.ValueNameRuleMap.Items {
		if item.Name == "" || item.Rule == "" {
			_ = seelog.Warnf("config:%s name和rule不能为空,请检查参数", item.Name)
			continue
		}
		if item.Attr != "" {
			selectorMap[item.Name] = []string{item.Rule, item.Attr}
		} else {
			selectorMap[item.Name] = []string{item.Rule}
		}
	}

	if t == COLLECTOR_TYPE_SELECTOR {
		// 通常 ValueNameRuleMap.Items的第一个是用来定位元素的大致位置,标识符:$position
		if _, ok := selectorMap[SYMOBL_POSITION]; !ok {
			return nil, errors.New("valueNameRuleMap 缺少标识符:$position,请检查xml配置")
		}
	}

	parameters := strings.Split(config.Parameter, ",")
	urls, err := util.MakeUrls(config.UrlFormat, parameters)
	if err != nil {
		return nil, err
	}

	return &subCollector{
		config:      config,
		urls:        urls,
		selectorMap: selectorMap,
	}, nil

}

func (s *subCollector) Next() (b bool) {
	if s.currentIndex >= len(s.urls) {
		return false
	}
	s.currentUrl = s.urls[s.currentIndex]
	s.currentIndex++

	seelog.Debugf("["+s.Name()+"]"+"current url:%s", s.currentUrl)
	return true
}

func (s *subCollector) Name() (name string) {
	name = s.config.Name
	return
}

func (s *subCollector) Collection(resultChan chan<- *result.Result) (errorList []error) {
	panic("implement me")
}

func (s *subCollector) Config() *Config {
	return s.config
}
