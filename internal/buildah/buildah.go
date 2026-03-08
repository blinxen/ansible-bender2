package buildah

import (
	"context"
	"fmt"
	"strings"

	"github.com/blinxen/ansible-bender2/internal/config"
	"github.com/containers/buildah"
	imageStore "go.podman.io/image/v5/storage"
	"go.podman.io/storage"
)

func newWorkingContainer(ctx context.Context, config config.Config) error {
	logger := GetLogger()
	buildStore, err := getBuildStore()
	if err != nil {
		return err
	}

	builderOpts := buildah.BuilderOptions{
		FromImage: config.BaseImage,
		Container: config.WorkingContainer.Name,
		Logger:    logger,
	}

	builder, err := buildah.OpenBuilder(buildStore, config.WorkingContainer.Name)
	if err == nil && config.WorkingContainer.NoCache {
		err = builder.Delete()
		if err != nil {
			return err
		}
		builder = nil
	}

	if builder == nil {
		// TODO: Add log entry to tell user if image is being pulled
		builder, err = buildah.NewBuilder(ctx, buildStore, builderOpts)
		if err != nil {
			return err
		}
	}
	builder.Logger = logger

	for k, v := range config.TargetImage.Environment {
		builder.SetEnv(k, v)
	}

	if len(config.WorkingContainer.User) > 0 {
		builder.SetUser(config.WorkingContainer.User)
	}

	for _, v := range config.WorkingContainer.Volumes {
		volume := strings.Split(v, ":")
		err = builder.Add(volume[1], false, buildah.AddAndCopyOptions{}, volume[0])
		if err != nil {
			logger.Error("could not add volume", "src", volume[0], "dest", volume[1])
		}
	}

	_, err = buildStore.Shutdown(false)
	return err
}

func commitWorkingContainer(ctx *context.Context, config config.Config) (string, error) {
	buildStore, err := getBuildStore()
	if err != nil {
		return "", err
	}

	imageRef, err := imageStore.Transport.ParseStoreReference(buildStore, config.TargetImage.Name)
	if err != nil {
		return "", err
	}

	builder, err := buildah.OpenBuilder(buildStore, config.WorkingContainer.Name)
	builder.Logger = GetLogger()
	if err == nil && config.WorkingContainer.NoCache {
		err = builder.Delete()
		if err != nil {
			return "", err
		}
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

	imageId, _, _, err := builder.Commit(
		*ctx,
		imageRef,
		buildah.CommitOptions{Squash: config.Squash},
	)
	if err != nil {
		return "", err
	}

	_, err = buildStore.Shutdown(false)
	if err != nil {
		return "", err
	}

	return imageId, nil
}

func deleteWorkingContainer(config config.Config) error {
	buildStore, err := getBuildStore()
	if err != nil {
		return err
	}

	builder, err := buildah.OpenBuilder(buildStore, config.WorkingContainer.Name)
	builder.Logger = GetLogger()
	if err != nil {
		err = builder.Delete()
		if err != nil {
			return err
		}
	}

	_, err = buildStore.Shutdown(false)
	return err
}

func getBuildStore() (storage.Store, error) {
	buildStoreOptions, err := storage.DefaultStoreOptions()
	if err != nil {
		return nil, fmt.Errorf("could not get default storage options: %s", err)
	}

	buildStore, err := storage.GetStore(buildStoreOptions)
	if err != nil {
		return nil, fmt.Errorf("could not open storage: %s", err)
	}

	return buildStore, nil
}
