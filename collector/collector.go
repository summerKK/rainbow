package collector

import "rainbow/result"

type Collector interface {
	Next() (b bool)
	Name() (name string)
	Collection(chan<- *result.Result) (errorList []error)
}

type Type uint8

const (
	COLLECTOR_TYPE_SELECTOR = iota
	COLLECTOR_TYPE_REGEX
)
