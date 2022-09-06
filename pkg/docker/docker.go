/**
 * This file is part of Privado OSS.
 *
 * Privado is an open source static code analysis tool to discover data flows in the code.
 * Copyright (C) 2022 Privado, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * For more information, contact support@privado.ai
 */

package docker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Privado-Inc/privado-cli/pkg/config"
	"github.com/Privado-Inc/privado-cli/pkg/telemetry"
	"github.com/Privado-Inc/privado-cli/pkg/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"
)

type containerOutputProcessor struct {
	messages []string
	matchFn  func(string)
}

func getDefaultDockerClient() (*client.Client, error) {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return client, nil
}

func getBaseContainerConfig(image string) *container.Config {
	config := &container.Config{
		Image:        image,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		OpenStdin:    true,
		Tty:          true,
	}
	return config
}

func getContainerHostConfig(volumes containerVolumes) *container.HostConfig {
	hostConfig := &container.HostConfig{}

	if volumes.userKeyVolumeEnabled {
		hostConfig.Mounts = append(
			hostConfig.Mounts,
			mount.Mount{
				Type:     "bind",
				Source:   volumes.userKeyVolumeHost,
				Target:   config.AppConfig.Container.UserKeyVolumeDir,
				ReadOnly: true,
			},
		)
	}
	if volumes.dockerKeyVolumeEnabled {
		hostConfig.Mounts = append(
			hostConfig.Mounts,
			mount.Mount{
				Type:     "bind",
				Source:   volumes.dockerKeyVolumeHost,
				Target:   config.AppConfig.Container.DockerKeyVolumeDir,
				ReadOnly: true,
			},
		)
	}
	if volumes.userConfigVolumeEnabled {
		hostConfig.Mounts = append(
			hostConfig.Mounts,
			mount.Mount{
				Type:   "bind",
				Source: volumes.userConfigVolumeHost,
				Target: config.AppConfig.Container.UserConfigVolumeDir,
			},
		)
	}
	if volumes.sourceCodeVolumeEnabled {
		hostConfig.Mounts = append(
			hostConfig.Mounts,
			mount.Mount{
				Type:   "bind",
				Source: volumes.sourceCodeVolumeHost,
				Target: config.AppConfig.Container.SourceCodeVolumeDir,
			},
		)
	}
	if volumes.externalRulesVolumeEnabled {
		hostConfig.Mounts = append(
			hostConfig.Mounts,
			mount.Mount{
				Type:   "bind",
				Source: volumes.externalRulesVolumeHost,
				Target: config.AppConfig.Container.ExternalRulesVolumeDir,
			},
		)
	}
	if volumes.m2PackageCacheVolumeEnabled {
		hostConfig.Mounts = append(
			hostConfig.Mounts,
			mount.Mount{
				Type:   "bind",
				Source: volumes.m2PackageCacheVolumeHost,
				Target: config.AppConfig.Container.M2PackageCacheVolumeDir,
			},
		)
	}
	if volumes.gradlePackageCacheVolumeEnabled {
		hostConfig.Mounts = append(
			hostConfig.Mounts,
			mount.Mount{
				Type:   "bind",
				Source: volumes.gradlePackageCacheVolumeHost,
				Target: config.AppConfig.Container.GradlePackageCacheVolumeDir,
			},
		)
	}

	return hostConfig
}

func GetEnvsFromDockerImage(imageURL string) ([]EnvVar, error) {
	client, err := getDefaultDockerClient()
	if err != nil {
		return nil, err
	}

	imageInfo, _, err := client.ImageInspectWithRaw(context.Background(), imageURL)
	if err != nil {
		return nil, err
	}

	sanitizedEnvs := []EnvVar{}

	for _, env := range imageInfo.Config.Env {
		x := strings.Split(env, "=")
		sanitizedEnvs = append(sanitizedEnvs, EnvVar{Key: x[0], Value: x[1]})
	}

	return sanitizedEnvs, nil
}

func GetPrivadoDockerAccessKey(pullImage bool) (string, error) {
	imageURL := config.AppConfig.Container.ImageURL

	if pullImage {
		client, err := getDefaultDockerClient()
		if err != nil {
			return "", err
		}
		if err := PullLatestImage(imageURL, client); err != nil {
			return "", err
		}
	}

	envs, err := GetEnvsFromDockerImage(imageURL)
	if err != nil {
		return "", err
	}

	for _, env := range envs {
		if env.Key == config.AppConfig.Container.DockerAccessKeyEnv {
			return env.Value, nil
		}
	}

	return "", nil
}

func PullLatestImage(image string, client *client.Client) (err error) {
	if client == nil {
		client, err = getDefaultDockerClient()
		if err != nil {
			return err
		}
	}

	ctx := context.Background()

	fmt.Println("\n> Pulling the latest image:", image)
	reader, err := client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	id, isTerm := term.GetFdInfo(os.Stdout)
	_ = jsonmessage.DisplayJSONMessagesStream(reader, os.Stdout, id, isTerm, nil)

	defer reader.Close()
	io.Copy(os.Stdout, reader)

	return nil
}

func attachContainerOutput(client *client.Client, ctx context.Context, containerId string) (*bufio.Reader, error) {
	waiter, err := client.ContainerAttach(ctx, containerId, types.ContainerAttachOptions{
		Stderr: true,
		Stdout: true,
		Stdin:  true,
		Stream: true,
		Logs:   true,
	})

	if err != nil {
		return nil, err
	}
	// attach stdin by default for now
	go io.Copy(waiter.Conn, os.Stdin)

	return waiter.Reader, err
}

func processAttachedContainerOutput(reader *bufio.Reader, attachStdOut bool, outputProcessors []containerOutputProcessor) {
	// noticed we are missing output due to
	// this kind of usage
	// rather print
	// if attachStdOut {
	// 	go io.Copy(os.Stdout, reader)
	// 	go io.Copy(os.Stderr, reader)
	// }

	if len(outputProcessors) <= 0 {
		return
	}

	go func() {
		for {
			outputLine, _ := reader.ReadString('\n')
			if attachStdOut {
				fmt.Print(outputLine)
			}

			// process each line in parallel goroutines so output
			// does not get blocked and we do not skip anything in
			// either approach of getting stdout (copy, print)
			// also, this is efficient
			go func(outputLine string) {
				for _, outputProcessor := range outputProcessors {
					for _, message := range outputProcessor.messages {
						if strings.Contains(outputLine, message) {
							processedLine := strings.TrimSpace(strings.TrimSuffix(outputLine, "\n"))
							outputProcessor.matchFn(processedLine)
						}
					}

				}
			}(outputLine)
		}
	}()
}

func WaitForContainer(client *client.Client, ctx context.Context, containerId string) error {
	statusCh, errCh := client.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	return nil
}

func RemoveContainerForcefully(client *client.Client, ctx context.Context, containerId string) error {
	return client.ContainerRemove(
		ctx,
		containerId,
		types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		},
	)
}

func StopContainer(client *client.Client, ctx context.Context, containerId string) error {
	return client.ContainerStop(ctx, containerId, nil)
}

func RunImage(opts ...RunImageOption) error {
	runOptions := newRunImageHandler(opts)
	ctx := context.Background()

	client, err := getDefaultDockerClient()
	if err != nil {
		return err
	}

	image := config.AppConfig.Container.ImageURL
	// Pull image
	if runOptions.pullLatestImage {
		if err := PullLatestImage(image, client); err != nil {
			return err
		}
	}

	// Generate container configurations
	containerConfig := getBaseContainerConfig(image)
	containerConfig.Entrypoint = runOptions.entrypoint
	containerConfig.Cmd = runOptions.args
	containerConfig.Env = runOptions.environmentVars
	hostConfig := getContainerHostConfig(runOptions.volumes)

	telemetry.DefaultInstance.RecordAtomicMetric("dockerCmd", strings.Join(containerConfig.Cmd, " "))

	// Create container
	creationResponse, err := client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return err
	}
	if len(creationResponse.Warnings) > 0 {
		fmt.Println("\n> Encountered warnings:")
		for i, warn := range creationResponse.Warnings {
			fmt.Println(i+1, warn)
			telemetry.DefaultInstance.RecordArrayMetric("warning", warn)
		}
	}

	// always remove the container in the end
	defer RemoveContainerForcefully(client, ctx, creationResponse.ID)

	// Attach input/output streams with container
	containerOutputProcessors := []containerOutputProcessor{}
	if runOptions.spawnWebBrowserOnURLMessage {
		containerOutputProcessors = append(containerOutputProcessors, containerOutputProcessor{
			messages: runOptions.spawnWebBrowserOnURLTriggerMessages,
			matchFn: func(message string) {
				telemetry.DefaultInstance.RecordAtomicMetric("didReceiveCloudLinkMessage", true)
				url := utils.ExtractURLFromString(message)
				if url != "" {
					telemetry.DefaultInstance.RecordAtomicMetric("didParseCloudLink", true)
					if err := utils.OpenURLInBrowser(url); err != nil {
						telemetry.DefaultInstance.RecordArrayMetric("error", err)
					}
					telemetry.DefaultInstance.RecordAtomicMetric("didAutoSpawnBrowser", err == nil)
				}
			},
		})
	}

	if runOptions.exitOnError {
		containerOutputProcessors = append(containerOutputProcessors, containerOutputProcessor{
			messages: runOptions.exitOnErrorTriggerMessages,
			matchFn: func(message string) {
				fmt.Println("\n> Some error occurred")
				if message != "" {
					// reset any color from internal process
					fmt.Println("Find more details below:\n", message, "\033[0m")
					telemetry.DefaultInstance.RecordArrayMetric("warning", message)
				}
				fmt.Println("\n> If this is an unexpected output, please try again or open an issue here: ", config.AppConfig.PrivadoRepository)
				fmt.Println("> Terminating..")
				RemoveContainerForcefully(client, ctx, creationResponse.ID)
			},
		})
	}

	if runOptions.attachOutput || len(containerOutputProcessors) > 0 {
		reader, err := attachContainerOutput(client, ctx, creationResponse.ID)
		if err != nil {
			return err
		}

		processAttachedContainerOutput(reader, runOptions.attachOutput, containerOutputProcessors)
	}

	// Start container
	fmt.Println("\n> Starting container with the latest image")
	fmt.Println("> Container ID:", creationResponse.ID)
	if err := client.ContainerStart(ctx, creationResponse.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// Setup interrupt fns if enabled
	if runOptions.setupInterrupt {
		// Listen for interrupt, clear signal after execution
		// Remove container when received
		// All cleanup here: The process ends after this
		// and defer functions are not executed
		sgn := utils.RunOnCtrlC(func() {
			fmt.Println("\n> Received interrupt signal")
			fmt.Println("> Terminating..")
			RemoveContainerForcefully(client, ctx, creationResponse.ID)
		})
		defer utils.ClearSignals(sgn)
	}

	// Image output after this point
	fmt.Println("\n> Waiting for process to complete:")

	// wait for container to stop (automatically or by interrupt)
	if err := WaitForContainer(client, ctx, creationResponse.ID); err != nil {
		return err
	}

	return nil
}
