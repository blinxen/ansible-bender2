package main

import (
	"os"

	"go.podman.io/storage/pkg/reexec"

	"github.com/blinxen/ansible-bender2/cmd"
	_ "github.com/blinxen/ansible-bender2/internal/builder"
)

func main() {
	if reexec.Init() {
		os.Exit(0)
	}

	cmd.Execute()
}
