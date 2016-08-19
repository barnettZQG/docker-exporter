package exporter

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type PreCPUStats struct {
	Cpu_usage        CPU_usage          `json:cpu_usage`
	System_cpu_usage float64            `json:system_cpu_usage`
	Throttling_data  map[string]float64 `json:throttling_data`
}

type CPU_usage struct {
	Total_usage         float64   `json:total_usage`
	Percpu_usage        []float64 `json:percpu_usage`
	Usage_in_kernelmode float64   `json:usage_in_kernelmode`
	Usage_in_usermode   float64   `json:usage_in_usermode`
}

type MemoryStats struct {
	Usage     float64            `json:usage`
	Max_usage float64            `json:max_usage`
	Stats     map[string]float64 `json:stats`
	Failcnt   float64            `json:failcnt`
	Limit     float64            `json:limit`
}
type StatsExporter struct {
	MemUsage *prometheus.GaugeVec
}

func GetStatsExporter(labels []string) *StatsExporter {
	var sanitizedLabels = make([]string, len(labels))
	for index, labelName := range labels {
		sanitizedLabels[index] = sanitize(labelName)
	}
	//log.Infoln("sanitizedLabels length:", len(sanitizedLabels))
	return &StatsExporter{
		MemUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: Exporters,
			Name:      "container_mem_usage",
			Help:      "the memory usage of the container",
		}, append([]string{"id", "name", "image", "imageId"}, sanitizedLabels...)),
	}
}
func sanitize(label string) string {
	label = strings.Replace(label, ".", "_", -1)
	return strings.Replace(label, "-", "_", -1)
}

type BlkioStats struct {
	Io_service_bytes_recursive []IoData `json:io_service_bytes_recursive`
	Io_serviced_recursive      []IoData `json:io_serviced_recursive`
	Io_queue_recursive         []IoData `json:io_queue_recursive`
	Io_service_time_recursive  []IoData `json:io_service_time_recursive`
	Io_wait_time_recursive     []IoData `json:io_wait_time_recursive`
	Io_merged_recursive        []IoData `json:io_merged_recursive`
	Io_time_recursive          []IoData `json:io_time_recursive`
	Sectors_recursive          []IoData `json:sectors_recursive`
}

type IoData struct {
	Major float64 `json:major`
	Minor float64 `json:minor`
	Op    string  `json:op`
	Value float64 `json:value`
}
