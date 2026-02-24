package main

import (
	"os"

	"github.com/blinxen/ansible-bender2/cmd"
	_ "github.com/blinxen/ansible-bender2/internal/ansible"
	_ "github.com/blinxen/ansible-bender2/internal/buildah"
	"go.podman.io/storage/pkg/reexec"
)

func main() {
	if reexec.Init() {
		os.Exit(0)
	}

	cmd.Execute()
}
