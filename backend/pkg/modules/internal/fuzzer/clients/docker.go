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
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/config"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/logging"
)

const (
	ContainerAutoremove  = true
	ContainerNetworkMode = "host"
	ContainerNameDefault = "fuzzer"
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
	logging.Logf("[Fuzzer][DockerClient] TriggerFuzzingJob(%v, %v, %v, %v): -->", apiID, endpoint, securityItem, timeBudget)

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("unable to create new docker client: %w", err)
	}

	containerName := ContainerNameDefault // TODO must be unique. For demo only

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
	logging.Logf("[Fuzzer][DockerClient] inputEnv=%v", inputEnv)

	// Pull the docker image if needed
	if strings.Contains(c.imageName, "/") {
		reader, err := cli.ImagePull(ctx, c.imageName, types.ImagePullOptions{})
		if err != nil {
			return fmt.Errorf("unable to pull docker Fuzzer image %v: %w", c.imageName, err)
		}
		_, err = io.Copy(os.Stdout, reader)
		if err != nil {
			// Not blocking I guess
			logging.Errorf("[Fuzzer][DockerClient] can't get output of pull action (%v)", err)
		}
		logging.Logf("[Fuzzer][DockerClient] Image is pulled (%v)", c.imageName)
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
		return fmt.Errorf("unable to create docker container for %v: %w", c.imageName, err)
	}
	logging.Logf("[Fuzzer][DockerClient] Container Creation Ok")

	// Start it
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("unable to start docker container for %v: %w", c.imageName, err)
	}
	logging.Logf("[Fuzzer][DockerClient] Container Start Ok")

	if c.showDockerLog {
		out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, Follow: true})
		if err != nil {
			return fmt.Errorf("unable to retrieve docker container logs for %v: %w", c.imageName, err)
		}
		_, err = io.Copy(os.Stdout, out)
		if err != nil {
			// Not blocking I guess
			logging.Errorf("[Fuzzer][DockerClient] can't get output of container logs (%v)", err)
		}
	}

	logging.Logf("[Fuzzer][DockerClient] TriggerFuzzingJob(): <--")
	return nil
}

func (c *DockerClient) StopFuzzingJob(apiID int64) error {
	logging.Logf("[Fuzzer][DockerClient] StopFuzzingJob(%v): -->", apiID)
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("unable to create new docker client: %w", err)
	}
	containerName := ContainerNameDefault // TODO must be unique. For demo only

	if err := cli.ContainerStop(ctx, containerName, nil); err != nil {
		logging.Errorf("unable to stop container %s: %s", containerName, err)
	}

	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}
	if err := cli.ContainerRemove(ctx, containerName, removeOptions); err != nil {
		// Just log this as warning, as it's not an error, specially if the container has "autoremove" flag
		logging.Warningf("unable to remove container: %s", err)
	}
	logging.Logf("[Fuzzer][DockerClient] StopFuzzingJob(%v): <--", apiID)
	return nil
}

//nolint: ireturn,nolintlint
func NewDockerClient(config *config.Config) (Client, error) {
	client := &DockerClient{
		imageName:         config.GetImageName(),
		showDockerLog:     config.GetShowDockerLogFlag(),
		platformType:      config.GetPlatformType(),
		platformHost:      config.GetPlatformHostFromFuzzer(),
		subFuzzer:         config.GetSubFuzzerList(),
		tokenInjectorPath: config.GetRestlerTokenInjectorPath(),
	}
	return client, nil
}
