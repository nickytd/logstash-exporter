package collector

import (
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Namespace const string
	Namespace = "logstash"
)

var _log log.Logger

func SetLogger(logger log.Logger) {
	_log = logger
}

// Collector interface implement Collect function
type Collector interface {
	Collect(ch chan<- prometheus.Metric) (err error)
}
