package collector

import (
	"fmt"
	"testing"
)

func TestNewCollectorConfig(t *testing.T) {
	fileName := "collectorConfig_test.xml"
	config := NewCollectorConfig(fileName)
	fmt.Printf("%+v\n", config)
}
