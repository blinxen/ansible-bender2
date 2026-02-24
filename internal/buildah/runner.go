package buildah

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/blinxen/ansible-bender2/internal/config"
	"github.com/blinxen/ansible-bender2/internal/logging"
	"go.podman.io/storage/pkg/reexec"
	"go.podman.io/storage/pkg/unshare"
)

type buildahAction int

const (
	actionCreateContainer buildahAction = iota
	actionCommitContainer
	actionDeleteContainer
)

const (
	createWorkingContainerCommandName = "ansible-bender2-create-container"
	commitWorkingContainerCommandName = "ansible-bender2-commit-container"
	deleteWorkingContainerCommandName = "ansible-bender2-delete-container"
)

func init() {
	reexec.Register(createWorkingContainerCommandName, enterNamespace)
	reexec.Register(
		createWorkingContainerCommandName+"-in-a-user-namespace",
		doAction(actionCreateContainer),
	)
	reexec.Register(commitWorkingContainerCommandName, enterNamespace)
	reexec.Register(
		commitWorkingContainerCommandName+"-in-a-user-namespace",
		doAction(actionCommitContainer),
	)
	reexec.Register(deleteWorkingContainerCommandName, enterNamespace)
	reexec.Register(
		deleteWorkingContainerCommandName+"-in-a-user-namespace",
		doAction(actionDeleteContainer),
	)
}

func CreateWorkingContainer(config *config.Config) error {
	cmd := reexec.Command(createWorkingContainerCommandName)
	configJson, err := configToJson(config)
	if err != nil {
		return err
	}
	cmd.Stdin = bytes.NewReader(configJson)
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func CommitWorkingContainer(config *config.Config) (string, error) {
	cmd := reexec.Command(commitWorkingContainerCommandName)
	configJson, err := configToJson(config)
	if err != nil {
		return "", err
	}
	cmd.Stdin = bytes.NewReader(configJson)
	cmd.Stderr = os.Stderr

	imageId, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(imageId), nil
}

func DeleteWorkingContainer(config *config.Config) error {
	cmd := reexec.Command(deleteWorkingContainerCommandName)
	configJson, err := configToJson(config)
	if err != nil {
		return err
	}
	cmd.Stdin = bytes.NewReader(configJson)
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func enterNamespace() {
	unshare.MaybeReexecUsingUserNamespace(true)
}

func doAction(action buildahAction) func() {
	return func() {
		logger := GetLogger()
		ctx := context.Background()
		var config config.Config
		if err := json.NewDecoder(os.Stdin).Decode(&config); err != nil {
			logger.WithError(err).Fatal("unexpected error when unmarshalling config")
		}
		logging.InitLogger(config.LogLevel)
		switch action {
		case actionCreateContainer:
			{
				err := newWorkingContainer(ctx, config)
				if err != nil {
					logger.WithError(err).Fatal("could not create working container")
				}
			}
		case actionCommitContainer:
			{
				id, err := commitWorkingContainer(&ctx, config)
				if err != nil {
					logger.WithError(err).Fatal("could not commit target image")
				}
				fmt.Printf("%s", id)
			}
		case actionDeleteContainer:
			{
				err := deleteWorkingContainer(config)
				if err != nil {
					logger.WithError(err).Fatal("could not delete working container")
				}
			}
		}
	}
}

func configToJson(config *config.Config) ([]byte, error) {
	payload, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("unexpected error when marshalling config: %s", err)
	}

	return payload, nil
}
