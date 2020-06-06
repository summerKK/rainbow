package collector

import (
	"encoding/xml"
	"os"
	"strings"
)

const SYMOBL_POSITION = "$position"

type Configs struct {
	Configs []*Config `xml:"config"`
}

type Config struct {
	Name             string `xml:"name,attr"`
	UrlFormat        string `xml:"urlFormat"`
	Parameter        string `xml:"urlParameters"`
	Type             Type   `xml:"collectType"`
	Charset          string `xml:"charset"`
	ValueNameRuleMap struct {
		Items []struct {
			Name string `xml:"name,attr"`
			Rule string `xml:"rule,attr"`
			Attr string `xml:"attribute,attr"`
		} `xml:"item"`
	} `xml:"valueNameRuleMap"`
	Interval int `xml:"interval"`
}

func NewCollectorConfig(filename string) *Configs {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var configXml Configs
	decoder := xml.NewDecoder(file)
	decoder.Strict = false
	err = decoder.Decode(&configXml)
	if err != nil {
		panic(err)
	}

	return &configXml
}

func (s *Config) Verify() bool {
	if s.UrlFormat == "" {
		return false
	}

	// 时间间隔默认2s
	if s.Interval == 0 {
		s.Interval = 2
	}

	if s.Charset == "" {
		s.Charset = "utf-8"
	} else {
		s.Charset = strings.ToLower(s.Charset)
	}

	return true
}
