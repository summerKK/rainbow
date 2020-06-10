package main

import (
	_ "net/http/pprof"

	"rainbow/scheduler"
)

func main() {
	scheduler.Run("logConfig.xml", "collectorConfig.xml", "proxy.db", "IP")
}
