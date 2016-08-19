package exporter

import (
	"strings"

	"github.com/barnettZQG/docker-exporter/manager"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type ContainerStatus struct {
	status *prometheus.GaugeVec
}

// Metric name parts.
const (
	// Namespace for all metrics.
	Namespace = "docker"
	// Subsystem(s).
	Exporters = "exporter"
)

func GetContainerStatus() *ContainerStatus {
	return &ContainerStatus{
		status: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace,
			Subsystem: Exporters,
			Name:      "container_up",
			Help:      "the status of container",
		}, []string{"id", "name", "image", "imageid"}),
	}
}
func (e *ContainerStatus) Collect(ch chan<- prometheus.Metric) {
	containers, err := manager.GetDockerManager().GetAllContainer()
	if err != nil {
		log.Errorln("Get containers error,", err.Error())
	}
	for _, container := range containers {
		var status float64

		if strings.Contains(container.Status, "Up") {
			status = 1
		} else if strings.Contains(container.Status, "Exited") {
			status = 0
		} else if strings.Contains(container.Status, "Error") {
			status = -1
		}
		//log.Infoln("container status:" + container.Status + " container name:" + container.Names[0])
		e.status.WithLabelValues(container.ID, container.Names[0], container.Image, container.ImageID).Set(status)
	}
	e.status.Collect(ch)
}
