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
	*subCollector
}

func NewSelectorCollector(config *Config) Collector {
	collector, err := newSubCollector(config, COLLECTOR_TYPE_SELECTOR)
	if err != nil {
		panic(err)
	}
	return &SelectorCollector{
		collector,
	}
}

func (s *SelectorCollector) Collection(resultChan chan<- *result.Result) (errorList []error) {
	response, _, errors := gorequest.New().Get(s.currentUrl).Set("User-Agent", util.RandomUA()).End()
	if response.Body != nil {
		defer response.Body.Close()
	}

	if len(errors) > 0 {
		errorList = append(errorList, errors...)
		return
	}

	if response.StatusCode != http.StatusOK {
		errorList = append(errorList, fmt.Errorf("[SelectorCollector] GET %s failed,status code:%d", s.currentUrl, response.StatusCode))
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
