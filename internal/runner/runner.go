package runner

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"text/template"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Run accepts a pre-filled Task object containing info needed to run the code and setup the environment.
// Will typically reside in worker.
// TODO: Still not done. Cache running container for a session or user. Will be specified later on.
// TODO: Associate a task with an ID to prevent container name collision.
func Run(c *client.Client, ctx context.Context, task Task) (Result, error) {
	// imageSummary, err := findImageByTagName(c, ctx, task.Config.Dockerfile)
	// if err != nil {
	// 	return Result{}, err
	// }
	var (
		imageSummary *types.ImageSummary
		err          error
		imageID      string
	)

	if imageSummary == nil {
		imageID, err = buildImageFromString(c, ctx, task.Config.Dockerfile)
		if err != nil {
			return Result{}, err
		}
	} else {
		imageID = imageSummary.ID
	}

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
				Memory:  int64(task.Config.MemoryLimit),
				Devices: nil,
			},
			Privileged: false,
		},
		nil, nil,
		"runner",
	)
	if err != nil {
		return Result{}, err
	}

	// Writing to tar.
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	content, err := os.ReadFile("./builtin/timer.sh")
	if err != nil {
		return Result{}, err
	}
	tmpl, err := template.New("").Parse(string(content))
	if err != nil {
		return Result{}, err
	}
	var resultBuf bytes.Buffer
	if err := tmpl.Execute(&resultBuf, task); err != nil {
		return Result{}, err
	}
	task.FileSystem["timer.sh"] = resultBuf.Bytes()
	if err := writeMapToTar(tw, task.FileSystem); err != nil {
		return Result{}, err
	}
	tw.Close()
	// end writing to tar.
	if err := c.CopyToContainer(ctx, ccr.ID, "/code", &buf, types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	}); err != nil {
		if err := disposeContainer(c, ctx, ccr.ID); err != nil {
			log.Println("Dispose err: " + err.Error())
		}
		return Result{}, err
	}

	// Writing testcases to tar.
	testCases := FileMap{}
	for i, testCase := range task.TestCases {
		testCases[fmt.Sprintf("tests/%d.txt", i)] = []byte(testCase.Input)
	}
	var buf1 bytes.Buffer
	tw1 := tar.NewWriter(&buf1)
	if err := writeMapToTar(tw1, testCases); err != nil {
		return Result{}, err
	}
	tw1.Close()

	if err := c.CopyToContainer(ctx, ccr.ID, "/", &buf1, types.CopyToContainerOptions{
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
