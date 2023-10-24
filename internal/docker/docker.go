package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// Wrapper for Docker SDK functionalities used in the runner.
// This can be mocked for testing.
type Docker interface {
	FindImageByTagName(context.Context, string) (*types.ImageSummary, error)
}

var _ Docker = (*docker)(nil)

// TODO: In the future, complete this struct instead of populating helpers.
type docker struct {
	client *client.Client
}

func New(client *client.Client) Docker {
	return &docker{client}
}

func (d *docker) FindImageByTagName(ctx context.Context, tagName string) (*types.ImageSummary, error) {
	resultsList, err := d.client.ImageList(
		ctx,
		types.ImageListOptions{
			All: true,
			Filters: filters.NewArgs(
				filters.KeyValuePair{
					Key:   "reference",
					Value: tagName,
				},
			),
		},
	)
	if err != nil {
		return nil, err
	}

	// Return empty result
	if len(resultsList) == 0 {
		return nil, nil
	}
	// Return an existing image.
	return &resultsList[0], nil
}
