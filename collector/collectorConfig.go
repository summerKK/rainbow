package collector

import (
	"encoding/xml"
	"os"
)

type Configs struct {
	Configs []Config `xml:"config"`
}

type Config struct {
	Name             string `xml:"name,attr"`
	UrlFormat        string `xml:"urlFormat"`
	Type             Type   `xml:"collectType"`
	Charset          string `xml:"charset"`
	ValueNameRuleMap struct {
		Items []struct {
			Name string `xml:"name,attr"`
			Rule string `xml:"rule,attr"`
			Attr string `xml:"attribute,attr"`
		} `xml:"item"`
	} `xml:"valueNameRuleMap"`
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
