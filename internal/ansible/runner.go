package ansible

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/blinxen/ansible-bender2/internal/config"
	"github.com/blinxen/ansible-bender2/internal/logging"
	"go.podman.io/storage/pkg/reexec"
	"go.podman.io/storage/pkg/unshare"
)

const (
	runAnsibleCommandName = "ansible-bender2-run-playbook"
)

func init() {
	reexec.Register(
		runAnsibleCommandName,
		func() {
			unshare.MaybeReexecUsingUserNamespace(true)
		},
	)
	reexec.Register(
		runAnsibleCommandName+"-in-a-user-namespace",
		func() {
			var config config.Config
			if err := json.NewDecoder(os.Stdin).Decode(&config); err != nil {
				GetLogger().WithError(err).Fatal("unexpected error when unmarshalling config")
			}
			logging.InitLogger(config.LogLevel)
			err := executePlaybook(&config)
			if err != nil {
				os.Exit(1)
			}
		},
	)
}

func RunPlaybook(config *config.Config) error {
	cmd := reexec.Command(runAnsibleCommandName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	payload, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("unexpected error when marshalling config: %s", err)
	}
	cmd.Stdin = bytes.NewReader(payload)

	return cmd.Run()
}
