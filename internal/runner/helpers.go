package runner

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
)

// Internal helpers
func findImageByTagName(c *client.Client, ctx context.Context, imageName string) (*types.ImageSummary, error) {
	resultsList, err := c.ImageList(
		ctx,
		types.ImageListOptions{
			All: true,
			Filters: filters.NewArgs(
				filters.KeyValuePair{
					Key:   "reference",
					Value: imageName,
				},
			),
		},
	)
	if err != nil {
		return nil, err
	}

	if len(resultsList) == 0 {
		return nil, nil
	}

	return &resultsList[0], nil
}

func buildImageFromString(c *client.Client, ctx context.Context, dockerFile string) (string, error) {
	var (
		uuidGen   = uuid.New()
		randomStr = uuidGen.String()
		imageName = "runner1" + randomStr
	)

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := addFileToTar(tw, "Dockerfile", []byte(dockerFile)); err != nil {
		return "", err
	}
	tw.Close()

	reader := bytes.NewReader(buf.Bytes())
	buildResp, err := c.ImageBuild(ctx, reader, types.ImageBuildOptions{
		Tags:       []string{imageName + ":latest"},
		Dockerfile: "Dockerfile",
		Context:    reader,
	})
	if err != nil {
		return "", err
	}

	image, err := c.ImageList(ctx, types.ImageListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key:   "reference",
				Value: imageName + "*",
			},
		),
	})
	if err != nil {
		return "", err
	}

	if len(image) == 0 {
		v, err := io.ReadAll(buildResp.Body)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("image not built. this is image build body: %s", string(v))
	}

	return image[0].ID, nil
}

func addFileToTar(tw *tar.Writer, filePath string, fileContent []byte) error {
	if err := tw.WriteHeader(&tar.Header{
		Name: filePath,
		Mode: 0777,
		Size: int64(len(fileContent)),
	}); err != nil {
		return err
	}
	if _, err := tw.Write(fileContent); err != nil {
		return err
	}
	return nil
}

// Converts a map of [filepath]: [file contents] into a tar. Returning the io.Reader of it.
func writeMapToTar(sourceFileMap map[string][]byte) (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	content, err := os.ReadFile("./builtin/timer.sh")
	if err != nil {
		return nil, err
	}

	if err := addFileToTar(tw, "timer.sh", content); err != nil {
		return nil, err
	}

	for filePath, fileContent := range sourceFileMap {
		if err := addFileToTar(tw, filePath, fileContent); err != nil {
			return nil, err
		}
	}
	tw.Close()
	return &buf, nil
}

func disposeContainer(c *client.Client, ctx context.Context, containerID string) error {
	return c.ContainerRemove(
		ctx,
		containerID,
		types.ContainerRemoveOptions{},
	)
}
