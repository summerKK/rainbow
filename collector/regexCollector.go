package collector

import (
	"rainbow/result"
)

type RegexCollector struct {
}

func (r *RegexCollector) Next() (b bool) {
	panic("implement me")
}

func (r *RegexCollector) Name() (name string) {
	panic("implement me")
}

func (r *RegexCollector) Collection(results chan<- *result.Result) (errorList []error) {
	panic("implement me")
}

func (r *RegexCollector) Config() *Config {
	panic("implement me")
}
