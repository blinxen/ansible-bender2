package builder

import (
	"context"

	"github.com/blinxen/ansible-bender2/internal/config"
)

type Builder interface {
	// Commit image to local container registry and return the image ID
	Commit(*config.Config) string
	// Container name of the started container
	ContainerName() string
	// Deletes this builder and does some cleanup jobs
	Delete()
}

func NewBuilder(ctx context.Context, config *config.Config) Builder {
	// TODO: Consider adding more builders
	return newBuildah(ctx, config)
}
