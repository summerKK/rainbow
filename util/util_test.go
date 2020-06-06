package util

import (
	"fmt"
	"testing"
)

func TestMakeUrls(t *testing.T) {
	urlFormat := "http://baidu.com?s=%s"
	params := []string{"1", "2", "3"}
	urls, err := MakeUrls(urlFormat, params)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(urls)
}
