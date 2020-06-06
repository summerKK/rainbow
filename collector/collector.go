package collector

import "rainbow/result"

type Collector interface {
	Next() (b bool)
	Name() (name string)
	Collection(chan<- *result.Result) (errorList []error)
}

type Type uint8
