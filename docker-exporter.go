package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/barnettZQG/docker-exporter/exporter"
	"github.com/barnettZQG/docker-exporter/manager"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

var (
	showVersion = flag.Bool(
		"version", false,
		"Print version information.",
	)
	listenAddress = flag.String(
		"web.listen-address", ":8888",
		"Address to listen on for web interface and telemetry.",
	)
	metricPath = flag.String(
		"web.telemetry-path", "/metrics",
		"Path under which to expose metrics.",
	)
)

// Metric name parts.
const (
	// Namespace for all metrics.
	Namespace = "docker"
	// Subsystem(s).
	Exporters = "exporter"
)

// landingPage contains the HTML served at '/'.
// TODO: Make this nicer and more informative.
var landingPage = []byte(`<html>
<head><title>Cloud exporter</title></head>
<body>
<h1>Docker exporter</h1>
<p><a href='` + *metricPath + `'>Metrics</a></p>
</body>
</html>
`)

// Exporter collects Docker metrics. It implements prometheus.Collector.
type Exporter struct {
	dsn          string
	duration     prometheus.Gauge
	error        prometheus.Gauge
	totalScrapes prometheus.Counter
	scrapeErrors *prometheus.CounterVec
	daemonUp     prometheus.Gauge
	containerNum prometheus.Gauge
}

// NewExporter returns a new Docker exporter for the provided DSN.
func NewExporter() *Exporter {
	return &Exporter{
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: Exporters,
			Name:      "last_scrape_duration_seconds",
			Help:      "Duration of the last scrape of metrics from Cloud.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: Exporters,
			Name:      "scrapes_total",
			Help:      "Total number of times Cloud was scraped for metrics.",
		}),
		scrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: Namespace,
			Subsystem: Exporters,
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occured scraping a Cloud.",
		}, []string{"collector"}),
		error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: Exporters,
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics from Cloud resulted in an error (1 for error, 0 for success).",
		}),
		daemonUp: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: Namespace,
			Name:      "up",
			Help:      "Whether the docker daemon is up.",
		}),
		containerNum: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: Exporters,
			Name:      "container_num",
			Help:      "the number of container in this instance",
		}),
	}
}

// Describe implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {

	metricCh := make(chan prometheus.Metric)
	doneCh := make(chan struct{})

	go func() {
		for m := range metricCh {
			ch <- m.Desc()
		}
		close(doneCh)
	}()

	e.Collect(metricCh)
	close(metricCh)
	<-doneCh
}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	e.scrape(ch)
	e.duration.Collect(ch)
	ch <- e.totalScrapes
	ch <- e.error
	e.scrapeErrors.Collect(ch)
	ch <- e.daemonUp
	ch <- e.containerNum
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) {

	log.Infoln("begin a new scrape")
	//define the number of scrape
	e.totalScrapes.Inc()
	var err error
	//define scrape time
	defer func(begun time.Time) {
		e.duration.Set(time.Since(begun).Seconds())
		if err == nil {
			e.error.Set(0)
		} else {
			e.error.Set(1)
		}
	}(time.Now())
	//get the container number
	dockermanager := manager.GetDockerManager()
	//check cloud server stats
	_, err = dockermanager.GetClient()
	if err != nil {
		e.daemonUp.Set(0)
	} else {
		e.daemonUp.Set(1)
	}
	e.containerNum.Set(float64(dockermanager.GetNumberOfContainer()))
	//获取容器状态
	exporter.GetContainerStatus().Collect(ch)
	//获取容器资源状态
	exporter.GetContainerStat().Collect(ch)
}

func init() {
	prometheus.MustRegister(version.NewCollector("cloud_exporter"))
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("cloud_exporter"))
		os.Exit(0)
	}
	log.Infoln("Starting cloud_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	exporter := NewExporter()
	prometheus.MustRegister(exporter)

	http.Handle(*metricPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(landingPage)
	})
	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
