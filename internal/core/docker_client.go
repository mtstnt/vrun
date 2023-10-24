package core

import (
	"github.com/docker/docker/client"
)

var dockerClient *client.Client

// Get the current client singleton.
func GetDockerClient() (*client.Client, error) {
	if dockerClient == nil {
		var err error
		dockerClient, err = client.NewClientWithOpts(
			client.WithAPIVersionNegotiation(),
			client.WithHostFromEnv(),
		)
		if err != nil {
			return nil, err
		}
	}
	return dockerClient, nil
}
