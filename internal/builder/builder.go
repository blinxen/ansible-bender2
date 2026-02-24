package builder

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/blinxen/ansible-bender2/internal/config"
	log "github.com/sirupsen/logrus"
	"go.podman.io/storage/pkg/reexec"
)

const (
	createWorkingContainerCommandName = "ansible-bender2-create-container"
	commitWorkingContainerCommandName = "ansible-bender2-commit-container"
	deleteWorkingContainerCommandName = "ansible-bender2-delete-container"
)

func init() {
	reexec.Register(createWorkingContainerCommandName, createWorkingContainerHandler)
	reexec.Register(commitWorkingContainerCommandName, commitWorkingContainerHandler)
	reexec.Register(deleteWorkingContainerCommandName, deleteWorkingContainerHandler)
}

func CreateWorkingContainer(config *config.Config) {
	cmd := reexec.Command(createWorkingContainerCommandName)
	cmd.Stderr = os.Stderr

	payload, err := json.Marshal(config)
	if err != nil {
		log.Fatalf("Could not create working container: ", err)
	}
	cmd.Stdin = bytes.NewReader(payload)

	_, err = cmd.Output()
	if err != nil {
		log.Fatalf("Could not create working container: ", err)
	}
}

func CommitWorkingContainer(config *config.Config) string {
	cmd := reexec.Command(commitWorkingContainerCommandName)
	cmd.Stderr = os.Stderr

	payload, err := json.Marshal(config)
	if err != nil {
		log.Fatalf("Could not create image: ", err)
	}
	cmd.Stdin = bytes.NewReader(payload)

	imageId, err := cmd.Output()
	if err != nil {
		log.Fatalf("Could not create image: ", err)
	}
	return string(imageId)
}

func DeleteWorkingContainer(config *config.Config) {
	cmd := reexec.Command(deleteWorkingContainerCommandName)
	cmd.Stderr = os.Stderr

	payload, err := json.Marshal(config)
	if err != nil {
		log.Fatalf("Could not working container: ", err)
	}
	cmd.Stdin = bytes.NewReader(payload)

	_, err = cmd.Output()
	if err != nil {
		log.Fatalf("Could not working container: ", err)
	}
}
