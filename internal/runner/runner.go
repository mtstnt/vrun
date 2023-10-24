package runner

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/mtstnt/vrun/internal/core"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type SourceFileMap map[string][]byte

type Config struct {
	ImageName      string
	CompileCommand string
	RunCommand     string
}

type Task struct {
	Config     Config
	FileSystem SourceFileMap
	TestCases  []string
}

type Result struct {
	CorrectTestcases int
}

func Run(ctx context.Context, task Task) (Result, error) {
	c, err := core.GetDockerClient()
	if err != nil {
		return Result{}, err
	}

	imageSummary, err := findImageByTagName(c, ctx, task.Config.ImageName)
	if err != nil {
		return Result{}, err
	}

	var imageID string
	if imageSummary == nil {
		imageID, err = buildImageFromString(c, ctx, `
FROM python:3.12.0-slim
CMD ["sleep", "infinity"]
`) // TODO: Diisi dockerfile beneran.
		if err != nil {
			return Result{}, err
		}
	} else {
		imageID = imageSummary.ID
	}

	// 10MB memory limit
	memoryLimit := 10_000_000

	ccr, err := c.ContainerCreate(ctx,
		&container.Config{
			Image:           imageID,
			NetworkDisabled: true,
			WorkingDir:      "/code",
			Cmd: []string{
				"sh", "./timer.sh",
			},
		},
		&container.HostConfig{
			Resources: container.Resources{
				Memory:  int64(memoryLimit),
				Devices: nil,
			},
			Privileged: false,
		},
		&network.NetworkingConfig{},
		&v1.Platform{},
		"runner",
	)
	if err != nil {
		return Result{}, err
	}

	fstar, err := writeMapToTar(task.FileSystem)
	if err != nil {
		return Result{}, err
	}

	if err := c.CopyToContainer(ctx, ccr.ID, "/code", fstar, types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	}); err != nil {
		if err := disposeContainer(c, ctx, ccr.ID); err != nil {
			log.Println("Dispose err: " + err.Error())
		}
		return Result{}, err
	}

	if err := c.ContainerStart(
		ctx,
		ccr.ID,
		types.ContainerStartOptions{},
	); err != nil {
		if err := disposeContainer(c, ctx, ccr.ID); err != nil {
			log.Println("Dispose err: " + err.Error())
		}
		return Result{}, err
	}

	wr, errCh := c.ContainerWait(
		ctx,
		ccr.ID,
		container.WaitConditionNotRunning,
	)

	select {
	case waitResp := <-wr:
		if waitResp.Error != nil {
			if err := disposeContainer(c, ctx, ccr.ID); err != nil {
				log.Println("Dispose err: " + err.Error())
			}
			return Result{}, err
		}
	case err := <-errCh:
		if err := disposeContainer(c, ctx, ccr.ID); err != nil {
			log.Println("Dispose err: " + err.Error())
		}
		return Result{}, err
	}

	containerLogs, err := c.ContainerLogs(
		ctx,
		ccr.ID,
		types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Details:    true,
		},
	)
	if err != nil {
		return Result{}, err
	}

	var (
		bufStdout = bytes.NewBuffer(nil)
		bufStderr = bytes.NewBuffer(nil)
	)

	if _, err := stdcopy.StdCopy(bufStdout, bufStderr, containerLogs); err != nil {
		return Result{}, err
	}

	fmt.Println("STDOUT: " + bufStdout.String())
	fmt.Println("STDERR: " + bufStderr.String())

	if err := disposeContainer(c, ctx, ccr.ID); err != nil {
		log.Println("Dispose err: " + err.Error())
	}

	return Result{}, nil
}
