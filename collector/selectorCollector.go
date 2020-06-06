package collector

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
	"github.com/cihub/seelog"
	"github.com/parnurzeal/gorequest"
	"rainbow/result"
	"rainbow/util"
)

type SelectorCollector struct {
	// 配置文件
	config *Config
	//  当前url
	currentUrl   string
	currentIndex int
	urls         []string
	selectorMap  map[string][]string
}

func NewSelectorCollector(config *Config) *SelectorCollector {
	if config == nil {
		_ = seelog.Error("[selectorCollector] nil config")
		return nil
	}

	// 验证参数
	if !config.Verify() || config.Type != COLLECTOR_TYPE_SELECTOR {
		_ = seelog.Errorf("参数配置错误:%+v", config)
		return nil
	}

	selectorMap := make(map[string][]string)

	for _, item := range config.ValueNameRuleMap.Items {
		if item.Name == "" || item.Rule == "" {
			_ = seelog.Errorf("[selectorCollector] config:%s name和rule不能为空,请检查参数", item.Name)
			continue
		}
		if item.Attr != "" {
			selectorMap[item.Name] = []string{item.Rule, item.Attr}
		} else {
			selectorMap[item.Name] = []string{item.Rule}
		}
	}

	// 通常 ValueNameRuleMap.Items的第一个是用来定位元素的大致位置,标识符:$position
	if _, ok := selectorMap[SYMOBL_POSITION]; !ok {
		_ = seelog.Error("[selectorCollector] valueNameRuleMap 缺少标识符:$position,请检查xml配置")
		return nil
	}

	parameters := strings.Split(config.Parameter, ",")
	urls, err := util.MakeUrls(config.UrlFormat, parameters)
	if err != nil {
		_ = seelog.Error(err)
		return nil
	}

	return &SelectorCollector{
		config:      config,
		urls:        urls,
		selectorMap: selectorMap,
	}
}

func (s *SelectorCollector) Next() (b bool) {
	if s.currentIndex >= len(s.urls) {
		return false
	}
	s.currentUrl = s.urls[s.currentIndex]
	s.currentIndex++

	seelog.Debugf("[selectorCollector] current url:%s", s.currentUrl)
	return true
}

func (s *SelectorCollector) Name() (name string) {
	name = s.config.Name
	return
}

func (s *SelectorCollector) Collection(resultChan chan<- *result.Result) (errorList []error) {
	response, _, errors := gorequest.New().Get(s.currentUrl).Set("User-Agent", util.RandomUA()).End()
	if response.Body != nil {
		defer response.Body.Close()
	}

	if len(errors) > 0 {
		_ = seelog.Error(errors)
		errorList = append(errorList, errors...)
		return
	}

	if response.StatusCode != http.StatusOK {
		errorList = append(errorList, fmt.Errorf("[selectorCollector] GET %s failed,status code:%d", s.currentUrl, response.StatusCode))
		return
	}

	var decoder mahonia.Decoder
	if s.config.Charset != "utf-8" {
		decoder = mahonia.NewDecoder(s.config.Charset)
	}

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		errorList = append(errorList, err)
		return
	}

	document.Find(s.selectorMap[SYMOBL_POSITION][0]).Each(func(i int, selection *goquery.Selection) {
		var (
			ip       string
			port     int
			speed    float64
			location string
		)

		var collection = make(map[string]string)
		for name, rule := range s.selectorMap {
			if name == SYMOBL_POSITION {
				continue
			}

			val := ""
			if len(rule) > 1 {
				val, _ = selection.Find(rule[0]).Attr(rule[1])
			} else {
				val = selection.Find(rule[0]).Text()
			}

			// 转换成utf-8
			if decoder != nil {
				val = decoder.ConvertString(val)
			}

			collection[name] = val
		}

		if _, ok := collection["ip"]; ok {
			ip = collection["ip"]
		}

		if _, ok := collection["port"]; ok {
			port, _ = strconv.Atoi(collection["port"])
		}

		if _, ok := collection["location"]; ok {
			location = collection["location"]
		}

		if _, ok := collection["speed"]; ok {
			reg := regexp.MustCompile(`^[1-9]\d*\.*\d*|0\.\d*[1-9]\d*`)
			if strings.Contains(collection["speed"], "秒") {
				speed, _ = strconv.ParseFloat(reg.FindString(collection["speed"]), 64)
			}
		}

		// 过滤响应速度大于3秒的ip
		if ip != "" && port > 0 && speed >= 0 {
			r := &result.Result{
				Ip:       ip,
				Port:     port,
				Location: location,
				Speed:    speed,
			}

			resultChan <- r
		}

	})

	seelog.Debugf("采集 url:%s 完成", s.currentUrl)
	return
}
