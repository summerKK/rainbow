package util

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
)

func MakeUrls(urlFormat string, params []string) (urls []string, err error) {
	if urlFormat == "" {
		err = errors.New("urlFormat 不能为空")
		return
	}

	if strings.Contains(urlFormat, "%s") && len(params) == 0 {
		err = fmt.Errorf("params参数不能为空.urlFormat:%s params:%+v", urlFormat, params)
		return
	}

	if !strings.Contains(urlFormat, "%s") {
		urls = append(urls, urlFormat)
		return
	}

	for _, param := range params {
		if param == "" {
			continue
		}
		urls = append(urls, fmt.Sprintf(urlFormat, param))
	}

	return
}

func RandomUA() string {
	userAgent := [...]string{
		"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, 360SE)",
		"Mozilla/4.0 (compatible, MSIE 8.0, Windows NT 6.0, Trident/4.0)",
		"Mozilla/5.0 (compatible, MSIE 9.0, Windows NT 6.1, Trident/5.0)",
		"Opera/9.80 (Windows NT 6.1, U, en) Presto/2.8.131 Version/11.11",
		"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, TencentTraveler 4.0)",
		"Mozilla/5.0 (Windows, U, Windows NT 6.1, en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
		"Mozilla/5.0 (Macintosh, Intel Mac OS X 10_7_0) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
		"Mozilla/5.0 (Macintosh, U, Intel Mac OS X 10_6_8, en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
		"Mozilla/5.0 (Linux, U, Android 3.0, en-us, Xoom Build/HRI39) AppleWebKit/534.13 (KHTML, like Gecko) Version/4.0 Safari/534.13",
		"Mozilla/5.0 (iPad, U, CPU OS 4_3_3 like Mac OS X, en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
		"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, Trident/4.0, SE 2.X MetaSr 1.0, SE 2.X MetaSr 1.0, .NET CLR 2.0.50727, SE 2.X MetaSr 1.0)",
		"Mozilla/5.0 (iPhone, U, CPU iPhone OS 4_3_3 like Mac OS X, en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
		"MQQBrowser/26 Mozilla/5.0 (Linux, U, Android 2.3.7, zh-cn, MB200 Build/GRJ22, CyanogenMod-7) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
	}

	return userAgent[rand.New(rand.NewSource(time.Now().Unix())).Intn(len(userAgent))]
}

func VerifyProxyIp(ip string, port int) bool {
	if ip == "" || !IsIp(ip) || port <= 0 {
		return false
	}

	proxyStr := "http://" + ip + ":" + strconv.Itoa(port)
	response, _, errs := gorequest.New().
		Proxy(proxyStr).
		Get("http://httpbin.org/get").
		Timeout(time.Second * 6).
		End()
	if len(errs) > 0 {
		return false
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return false
	}

	return true
}

// IsIp will match the given parameter is ip address or not.
func IsIp(ip string) bool {
	return IsInputMatchRegex(ip,
		"^((?:(?:25[0-5]|2[0-4]\\d|((1\\d{2})|([1-9]?\\d)))\\.){3}(?:25[0-5]|2[0-4]\\d|((1\\d{2})|([1-9]?\\d))))")
}

// IsInputMatchRegex will verify the input string is match the regex or not.
// This function will recover the panic if regex can't be parsed.
func IsInputMatchRegex(input, regex string) bool {
	result := false
	reg := regexp.MustCompile(regex)
	result = reg.MatchString(input)

	defer func() {
		r := recover()
		if r != nil {
			result = false
			fmt.Println(r)
		}
	}()

	return result
}
