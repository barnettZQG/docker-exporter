package exporter

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/barnettZQG/docker-exporter/manager"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"golang.org/x/net/context"
)

type ContainerStat struct {
	Read         string                        `json:read`
	Precpu_stats PreCPUStats                   `json:precpu_stats`
	Cpu_stats    PreCPUStats                   `json:cpu_stats`
	Memory_stats MemoryStats                   `json:memory_stats`
	Blkio_stats  BlkioStats                    `json:blkio_stats`
	Networks     map[string]map[string]float64 `json:networks`
	container    *types.Container
}

func GetContainerStat() *ContainerStat {
	return &ContainerStat{}
}

func (e *ContainerStat) Collect(ch chan<- prometheus.Metric) {

	cli, err := manager.GetDockerManager().GetClient()
	if err != nil {
		log.Errorln("container stat get client error " + err.Error())
		return
	}
	containers, err := manager.GetDockerManager().GetAllContainer()
	if err != nil {
		log.Errorln("container stat get all container error " + err.Error())
	} else {
		var cha chan *ContainerStat = make(chan *ContainerStat, len(containers))
		index := 0
		begun := time.Now()
		for _, contaier := range containers {
			if strings.Contains(contaier.Status, "Up") {
				index++
				go e.GetStat(cli, contaier, cha)
			}
			// go func(cli *client.Client, container types.Container, cc chan *ContainerStat) {
			// 	e.GetStat(cli, container, cc)
			// }(cli, contaier, cha)
		}
		for {
			select {
			case stat := <-cha:
				if stat != nil {
					e.Exporter(ch, stat)
				}
				if index--; index == 0 {
					// e.Do(ch, statExporter)
					overTime := time.Since(begun).Seconds()
					log.Infoln("get all container stats take time ", overTime)
					close(cha)
					return
				}
			case <-time.After(time.Second * 10):
				log.Errorln("get container stat time out 20 second")
				return
			}
		}

	}
}

/**
 *获取每个容器的资源状态信息
 */
func (e *ContainerStat) GetStat(client *client.Client, container types.Container, ch chan<- *ContainerStat) {
	log.Info("begin get stat ", container.Names[0])
	bengun := time.Now()
	log.Info("begin get stat ", container.Names[0], bengun)
	reader, err := client.ContainerStats(context.Background(), container.ID, false)
	over := time.Since(bengun).Seconds()
	log.Infoln("get contaier", container.Names[0], "stats take time ", over)
	if err != nil {
		log.Errorln("container stat get container stats error " + err.Error())
		ch <- nil
		return
	}
	defer reader.Close()
	var containerStat *ContainerStat = &ContainerStat{}
	containerStat.container = &container
	err = json.NewDecoder(reader).Decode(containerStat)
	if err != nil {
		log.Errorln("container stats decode error" + err.Error())
		ch <- nil
	} else {
		//log.Infoln("container stats decode complete, container read time: " + containerStat.Read)
		ch <- containerStat
	}

}

/*
 *写回每个容器信息到指标通道
 */
func (e *ContainerStat) Exporter(ch chan<- prometheus.Metric, stat *ContainerStat) {
	//log.Infoln("receive a container stat and begin to a statExporter ", stat.container.Names[0])
	delete(stat.container.Labels, " ")
	lakey := []string{}
	lavalve := []string{}
	lavalve = append(lavalve, stat.container.ID,
		stat.container.Names[0],
		stat.container.Image,
		stat.container.ImageID,
	)
	// for k, v := range stat.container.Labels {
	// 	if k != "name" && k != "id" && k != "image" && k != "imageId" {
	// 		lakey = append(lakey, k)
	// 		lavalve = append(lavalve, v)
	// 	}
	// }
	//log.Infoln("lavalve length:", len(lavalve), "lakey length:", len(lakey))
	//根据容器不同的label获取exporter
	statExporter := GetStatsExporter(lakey)

	statExporter.MemUsage.WithLabelValues(lavalve...).Set(stat.Memory_stats.Usage)

	e.Do(ch, statExporter)
}

var j int

func (e *ContainerStat) Do(ch chan<- prometheus.Metric, stat *StatsExporter) {
	//log.Infoln("Begin write container stats metric to chan ")
	//写回内存使用信息
	stat.MemUsage.Collect(ch)
}
