package builder

import (
	"context"
	"strings"

	"github.com/containers/buildah"
	log "github.com/sirupsen/logrus"
	imageStore "go.podman.io/image/v5/storage"
	"go.podman.io/storage"
	"go.podman.io/storage/pkg/unshare"

	"github.com/blinxen/ansible-bender2/internal/config"
)

func newWorkingContainer(ctx context.Context, config config.Config) {
	unshare.MaybeReexecUsingUserNamespace(false)

	buildStore := getBuildStore()
	defer buildStore.Shutdown(false)

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
}

func commitWorkingContainer(ctx *context.Context, config config.Config) string {
	unshare.MaybeReexecUsingUserNamespace(false)

	buildStore := getBuildStore()
	defer buildStore.Shutdown(false)

	imageRef, err := imageStore.Transport.ParseStoreReference(buildStore, config.TargetImage.Name)
	if err != nil {
		log.Fatalf("Could not image reference from image store: %s", err)
	}

	builder, err := buildah.OpenBuilder(buildStore, config.WorkingContainer.Name)
	if err == nil && config.WorkingContainer.NoCache {
		builder.Delete()
		builder = nil
	}

	if len(config.TargetImage.User) > 0 {
		builder.SetUser(config.TargetImage.User)
	}
	if len(config.TargetImage.Workdir) > 0 {
		builder.SetWorkDir(config.TargetImage.Workdir)
	}

	if len(config.TargetImage.Entrypoint) > 0 {
		builder.SetEntrypoint(config.TargetImage.Entrypoint)
	}
	for k, v := range config.TargetImage.Labels {
		builder.SetLabel(k, v)
	}
	for k, v := range config.TargetImage.Annotations {
		builder.SetAnnotation(k, v)
	}
	for _, p := range config.TargetImage.Ports {
		builder.SetPort(p)
	}
	for _, v := range config.TargetImage.Volumes {
		builder.AddVolume(v)
	}
	if len(config.TargetImage.Cmd) > 0 {
		builder.SetCmd(config.TargetImage.Cmd)
	}
	if len(config.TargetImage.Entrypoint) > 0 {
		builder.SetEntrypoint(config.TargetImage.Entrypoint)
	}

	imageId, _, _, err := builder.Commit(*ctx, imageRef, buildah.CommitOptions{Squash: config.Squash})
	if err != nil {
		log.Fatalf("Could not commit image: %s", err)
	}

	return imageId
}

func deleteWorkingContainer(config config.Config) {
	unshare.MaybeReexecUsingUserNamespace(false)

	buildStore := getBuildStore()
	defer buildStore.Shutdown(false)

	builder, err := buildah.OpenBuilder(buildStore, config.WorkingContainer.Name)
	if err == nil {
		builder.Delete()
	}
}

func getBuildStore() storage.Store {
	buildStoreOptions, err := storage.DefaultStoreOptions()
	if err != nil {
		log.Fatalf("Could not retrieve default storage options: %s", err)
	}

	buildStore, err := storage.GetStore(buildStoreOptions)
	if err != nil {
		log.Fatalf("Could not retrieve storage: %s", err)
	}

	return buildStore
}
