// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clients

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerClient "github.com/docker/docker/client"
	logging "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/config"
)

const (
	ContainerAutoremove  = true
	ContainerNetworkMode = "host"
)

type DockerClient struct {
	imageName         string
	showDockerLog     bool
	platformType      string
	platformHost      string
	subFuzzer         string
	tokenInjectorPath string
}

func (c *DockerClient) TriggerFuzzingJob(apiID int64, endpoint string, securityItem string, timeBudget string) error {
	logging.Infof("[Fuzzer][DockerClient] TriggerFuzzingJob(%v, %v, %v, %v): --> <--", apiID, endpoint, securityItem, timeBudget)

	ctx := context.Background()
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("unable to create new docker client: %v", err)
	}

	containerName := c.GetContainerNameForAPI(apiID)

	// Define environment for container
	inputEnv := []string{
		fmt.Sprintf("URI=%s", endpoint),
		fmt.Sprintf("PLATFORM_TYPE=%s", c.platformType),
		fmt.Sprintf("PLATFORM_HOST=%s", c.platformHost),
		fmt.Sprintf("API_ID=%v", apiID),
		fmt.Sprintf("RESTLER_TIME_BUDGET=%s", timeBudget),
		fmt.Sprintf("RESTLER_TOKEN_INJECTOR_PATH=%s", c.tokenInjectorPath),
		fmt.Sprintf("FUZZER=%s", c.subFuzzer),
	}
	if len(securityItem) > 0 {
		inputEnv = append(inputEnv, fmt.Sprintf("SERVICE_AUTH=%s", securityItem))
	}
	logging.Debugf("[Fuzzer][DockerClient] inputEnv=%v", inputEnv)

	// Pull the docker image if needed
	if strings.Contains(c.imageName, "/") {
		logging.Infof("[Fuzzer][DockerClient] pull image (%v)", c.imageName)
		reader, err := cli.ImagePull(ctx, c.imageName, types.ImagePullOptions{})
		if err != nil {
			return fmt.Errorf("unable to pull docker Fuzzer image %v: %w", c.imageName, err)
		}
		_, err = io.Copy(os.Stdout, reader)
		if err != nil {
			// Not blocking I guess
			logging.Errorf("[Fuzzer][DockerClient] can't get output of pull action (%v)", err)
		}
	}

	// Create the container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: c.imageName,
		Env:   inputEnv,
	}, &container.HostConfig{
		AutoRemove:  ContainerAutoremove,
		NetworkMode: ContainerNetworkMode,
	}, nil, nil, containerName)
	if err != nil {
		return fmt.Errorf("unable to create docker container (%v) from image (%v): err=(%v)", containerName, c.imageName, err)
	}

	// Start it
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("unable to start docker container (%v) from image (%v): err=(%v)", containerName, c.imageName, err)
	}

	if c.showDockerLog {
		logging.Infof("[Fuzzer][DockerClient] Activate Show docker logs")
		out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, Follow: true})
		if err != nil {
			return fmt.Errorf("unable to retrieve docker container logs for %v, err=(%v)", c.imageName, err)
		}
		_, err = io.Copy(os.Stdout, out)
		if err != nil {
			// Not blocking I guess
			logging.Errorf("[Fuzzer][DockerClient] can't get output of container logs, err=(%v)", err)
		}
	}

	return nil
}

func (c *DockerClient) StopFuzzingJob(apiID int64, complete bool) error {
	logging.Infof("[Fuzzer][DockerClient] StopFuzzingJob(%v): --> <--", apiID)
	ctx := context.Background()
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("unable to create new docker client: %w", err)
	}

	containerName := c.GetContainerNameForAPI(apiID)

	if err := cli.ContainerStop(ctx, containerName, nil); err != nil {
		logging.Errorf("[Fuzzer][DockerClient] can't stop container (%v), err=(%v)", containerName, err)
	}

	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}
	if err := cli.ContainerRemove(ctx, containerName, removeOptions); err != nil {
		// Just log this as warning, as it's not an error, specially if the container has "autoremove" flag it can be "in progress"
		logging.Infof("[Fuzzer][DockerClient] can't remove container (%v), err=(%v)", containerName, err)
	}

	return nil
}

func (c *DockerClient) GetContainerNameForAPI(apiID int64) string {
	return containerNamePrefix + strconv.FormatInt(apiID, 10) //nolint:gomnd
}

// nolint: ireturn,nolintlint
func NewDockerClient(config *config.Config) (Client, error) {
	client := &DockerClient{
		imageName:         config.GetImageName(),
		showDockerLog:     config.GetShowDockerLogFlag(),
		platformType:      config.GetPlatformType(),
		platformHost:      config.GetPlatformHost(),
		subFuzzer:         config.GetSubFuzzerList(),
		tokenInjectorPath: config.GetRestlerTokenInjectorPath(),
	}
	return client, nil
}
