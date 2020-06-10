package main

import (
	"rainbow/scheduler"
)

func main() {
	scheduler.Run("logConfig.xml", "collectorConfig.xml", "proxy.db", "IP")
}
