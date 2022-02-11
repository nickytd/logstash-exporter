package main

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"logstash_exporter/collector"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"
)

var (
	scrapeDurations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: collector.Namespace,
			Subsystem: "exporter",
			Name:      "scrape_duration_seconds",
			Help:      "logstash_exporter: Duration of a scrape job.",
		},
		[]string{"collector", "result"},
	)
	_log log.Logger
)

// LogstashCollector collector type
type LogstashCollector struct {
	collectors map[string]collector.Collector
}

// NewLogstashCollector register a logstash collector
func NewLogstashCollector(logstashEndpoint string) (*LogstashCollector, error) {
	nodeStatsCollector, err := collector.NewNodeStatsCollector(logstashEndpoint)
	if err != nil {
		_ = level.Error(_log).Log("msg", "Cannot register a new collector", "err", err)
	}

	nodeInfoCollector, err := collector.NewNodeInfoCollector(logstashEndpoint)
	if err != nil {
		_ = level.Error(_log).Log("msg", "Cannot register a new collector", "err", err)
	}

	return &LogstashCollector{
		collectors: map[string]collector.Collector{
			"node": nodeStatsCollector,
			"info": nodeInfoCollector,
		},
	}, nil
}

func listen(exporterBindAddress string) {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/metrics", http.StatusMovedPermanently)
	})

	_ = level.Info(_log).Log("msg", "Starting server on", "bindAddress", exporterBindAddress)
	if err := http.ListenAndServe(exporterBindAddress, nil); err != nil {
		_ = level.Error(_log).Log("msg", "Cannot start Logstash exporter", "err", err)
	}
}

// Describe logstash metrics
func (coll LogstashCollector) Describe(ch chan<- *prometheus.Desc) {
	scrapeDurations.Describe(ch)
}

// Collect logstash metrics
func (coll LogstashCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(coll.collectors))
	for name, c := range coll.collectors {
		go func(name string, c collector.Collector) {
			execute(name, c, ch)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
	scrapeDurations.Collect(ch)
}

func execute(name string, c collector.Collector, ch chan<- prometheus.Metric) {
	begin := time.Now()
	err := c.Collect(ch)
	duration := time.Since(begin)
	var result string

	if err != nil {
		_ = level.Error(_log).Log("msg", "collector failed ", "name", name, "after", duration.Seconds(), "err", err)
		result = "error"
	} else {
		_ = level.Debug(_log).Log("msg", "collector succeeded", "name", name, "after", duration.Seconds())
		result = "success"
	}
	scrapeDurations.WithLabelValues(name, result).Observe(duration.Seconds())
}

func init() {
	prometheus.MustRegister(version.NewCollector("logstash_exporter"))
}

func main() {
	var (
		logstashEndpoint    = kingpin.Flag("logstash.endpoint", "The protocol, host and port on which logstash metrics API listens").Default("http://localhost:9600").String()
		exporterBindAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9198").String()
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("logstash_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	_log = promlog.New(promlogConfig)
	collector.SetLogger(_log)

	logstashCollector, err := NewLogstashCollector(*logstashEndpoint)
	if err != nil {
		_ = level.Error(_log).Log("msg", "Cannot register a new Logstash Collector", "err", err)
	}

	prometheus.MustRegister(logstashCollector)

	_ = level.Info(_log).Log("msg", "Starting Logstash exporter", "ver", version.Info())
	_ = level.Info(_log).Log("build context", version.BuildContext())

	listen(*exporterBindAddress)
}
