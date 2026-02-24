package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/blinxen/ansible-bender2/internal/config"
)

// Writes the new working container name into stdout
func createWorkingContainerHandler() {
	ctx := context.Background()

	newWorkingContainer(ctx, readConfig())
}

// Writes the new image ID into stdout
func commitWorkingContainerHandler() {
	ctx := context.Background()
	id := commitWorkingContainer(&ctx, readConfig())
	fmt.Printf("%s", id)
}

func deleteWorkingContainerHandler() {
	deleteWorkingContainer(readConfig())
}

func readConfig() config.Config {
	var config config.Config
	if err := json.NewDecoder(os.Stdin).Decode(&config); err != nil {
		fmt.Fprintf(os.Stderr, "buildah-commit: decode request: %v\n", err)
		os.Exit(1)
	}

	return config
}
