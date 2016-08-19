package manager

import (
	"time"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/prometheus/common/log"
	"golang.org/x/net/context"
)

type DockerManager struct {
	Client     *client.Client
	containers []types.Container
}

var dockerManager *DockerManager

func GetDockerManager() *DockerManager {
	if dockerManager == nil {
		dockerManager = &DockerManager{}
	}
	return dockerManager
}

func (dm *DockerManager) GetAllContainer() ([]types.Container, error) {
	if dm.containers == nil {
		cli, err := dm.GetClient()
		if err != nil {
			log.Errorln("Get client error,", err.Error())
		}
		//true -> all container
		options := types.ContainerListOptions{All: true}
		bengun := time.Now()
		dm.containers, err = cli.ContainerList(context.Background(), options)
		over := time.Since(bengun).Seconds()
		log.Infoln("get contaiers take time ", over)
		if err != nil {
			log.Errorln("Get containers error,", err.Error())
			return nil, err
		}
	}

	return dm.containers, nil
}

func (dm *DockerManager) GetNumberOfContainer() int {
	containers, err := dm.GetAllContainer()
	if err != nil {
		log.Errorln("Get containers error,", err.Error())
		return 0
	}
	return len(containers)
}
func (dm *DockerManager) GetClient() (*client.Client, error) {

	if dm.Client == nil {
		defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
		cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, defaultHeaders)
		if err != nil {
			log.Errorln("Get a docker client error,", err.Error())
			return nil, err
		} else {
			dm.Client = cli
		}
	}
	return dm.Client, nil
}
