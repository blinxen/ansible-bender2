package builder

import (
	"context"
	"os"
	"strings"

	"github.com/containers/buildah"
	log "github.com/sirupsen/logrus"
	imageStore "go.podman.io/image/v5/storage"
	"go.podman.io/storage"
	"go.podman.io/storage/pkg/unshare"

	"github.com/blinxen/ansible-bender2/internal/config"
)

type BuildahBuilder struct {
	builder    buildah.Builder
	buildStore storage.Store
	ctx        context.Context
}

func init() {
	if buildah.InitReexec() {
		os.Exit(0)
	}
}

func newBuildah(ctx context.Context, config *config.Config) Builder {
	unshare.MaybeReexecUsingUserNamespace(false)

	buildStoreOptions, err := storage.DefaultStoreOptions()
	if err != nil {
		log.Fatalf("Could not retrieve default storage options: %s", err)
	}

	buildStore, err := storage.GetStore(buildStoreOptions)
	if err != nil {
		log.Fatalf("Could not retrieve storage: %s", err)
	}

	builderOpts := buildah.BuilderOptions{
		FromImage: config.BaseImage,
		Container: config.WorkingContainer.Name,
	}

	builder, err := buildah.OpenBuilder(buildStore, config.WorkingContainer.Name)
	if err == nil && config.WorkingContainer.NoCache {
		builder.Delete()
		builder = nil
	}

	if builder == nil {
		builder, err = buildah.NewBuilder(ctx, buildStore, builderOpts)
		if err != nil {
			log.Fatalf("Could not create buildah builder: %s", err)
		}
	}

	for k, v := range config.TargetImage.Environment {
		builder.SetEnv(k, v)
	}

	if len(config.WorkingContainer.User) > 0 {
		builder.SetUser(config.WorkingContainer.User)
	}

	for _, v := range config.WorkingContainer.Volumes {
		volume := strings.Split(v, ":")
		builder.Add(volume[1], false, buildah.AddAndCopyOptions{}, volume[0])
	}

	return &BuildahBuilder{builder: *builder, buildStore: buildStore, ctx: ctx}

}

func (b *BuildahBuilder) Commit(config *config.Config) string {
	imageRef, err := imageStore.Transport.ParseStoreReference(b.buildStore, config.TargetImage.Name)
	if err != nil {
		log.Fatalf("Could not image reference from image store: %s", err)
	}

	if len(config.TargetImage.User) > 0 {
		b.builder.SetUser(config.TargetImage.User)
	}
	if len(config.TargetImage.Workdir) > 0 {
		b.builder.SetWorkDir(config.TargetImage.Workdir)
	}

	if len(config.TargetImage.Entrypoint) > 0 {
		b.builder.SetEntrypoint(config.TargetImage.Entrypoint)
	}
	for k, v := range config.TargetImage.Labels {
		b.builder.SetLabel(k, v)
	}
	for k, v := range config.TargetImage.Annotations {
		b.builder.SetAnnotation(k, v)
	}
	for _, p := range config.TargetImage.Ports {
		b.builder.SetPort(p)
	}
	for _, v := range config.TargetImage.Volumes {
		b.builder.AddVolume(v)
	}
	if len(config.TargetImage.Cmd) > 0 {
		b.builder.SetCmd(config.TargetImage.Cmd)
	}
	if len(config.TargetImage.Entrypoint) > 0 {
		b.builder.SetEntrypoint(config.TargetImage.Entrypoint)
	}

	imageId, _, _, err := b.builder.Commit(b.ctx, imageRef, buildah.CommitOptions{})
	if err != nil {
		log.Fatalf("Could not commit image: %s", err)
	}

	return imageId
}

func (b *BuildahBuilder) Delete() {
	b.buildStore.Shutdown(false)
	err := b.builder.Delete()
	if err != nil {
		log.Fatalf("Could not delete image: %s", err)
	}
}

func (b *BuildahBuilder) ContainerName() string {
	return b.builder.Container
}
