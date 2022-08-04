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
	if volumes.packageCacheVolumeEnabled {
		hostConfig.Mounts = append(
			hostConfig.Mounts,
			mount.Mount{
				Type:   "bind",
				Source: volumes.packageCacheVolumeHost,
				Target: config.AppConfig.Container.M2PackageCacheVolumeDir,
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

func attachContainerOutput(client *client.Client, ctx context.Context, containerId string, attachStdOut bool) (*bufio.Reader, error) {
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

	reader := waiter.Reader

	if attachStdOut {
		go io.Copy(os.Stdout, reader)
		go io.Copy(os.Stderr, reader)
	}

	// attach io by default for now
	go io.Copy(waiter.Conn, os.Stdin)

	return reader, err
}

// [TODO] Rewrite this function to call matchFn respective to messages
// somehow maintain a map of message(s) and matchFn
// and only call the respective one
// handle the rest outside
// as it is non of fn's concern
type containerOutputProcessor struct {
	messages []string
	matchFn  func(string)
}

func processAttachedContainerOutput(reader *bufio.Reader, outputProcessors []containerOutputProcessor) {
	if len(outputProcessors) <= 0 {
		return
	}

	go func() {
		for {
			outputLine, _ := reader.ReadString('\n')
			for _, outputProcessor := range outputProcessors {
				for _, message := range outputProcessor.messages {
					if strings.Contains(outputLine, message) {
						processedMessage := strings.TrimSpace(strings.TrimSuffix(message, "\n"))
						outputProcessor.matchFn(processedMessage)
					}
				}

			}
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
	containerConfig.Cmd = runOptions.args
	containerConfig.Env = runOptions.environmentVars
	hostConfig := getContainerHostConfig(runOptions.volumes)

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
				url := utils.ExtractURLFromString(message)
				if url != "" {
					utils.OpenURLInBrowser(url)
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
		// processContainerOutput(attachStdIO, runOnMatch)
		reader, err := attachContainerOutput(client, ctx, creationResponse.ID, runOptions.attachOutput)
		if err != nil {
			return err
		}

		processAttachedContainerOutput(reader, containerOutputProcessors)
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
