package runner

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
)

// Find an image by name (filters by 'reference')
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

// Builds an image from a Dockerfile, provided as a string.
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
		Remove:     true,
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

// Add a file, specified as relative path and content, to an existing *tar.Writer
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
func writeMapToTar(tw *tar.Writer, sourceFileMap FileMap) error {
	for filePath, fileContent := range sourceFileMap {
		if err := addFileToTar(tw, filePath, fileContent); err != nil {
			return err
		}
	}
	return nil
}

func disposeContainer(c *client.Client, ctx context.Context, containerID string) error {
	return c.ContainerRemove(
		ctx,
		containerID,
		types.ContainerRemoveOptions{},
	)
}
