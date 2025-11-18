package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/blinxen/ansible-bender2/internal/ansible"
	"github.com/blinxen/ansible-bender2/internal/builder"
	"github.com/blinxen/ansible-bender2/internal/config"
)

const OriginalCwdEnv = "ANSIBLE_BENDER2_ORIGINAL_CWD"

// flags
var (
	noCache     bool
	noFailImage bool
	squash      bool
)

var buildCmd = &cobra.Command{
	Use:          "build <playbook>",
	Short:        "Build the OCI image",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	Run:          build,
}

func init() {
	buildCmd.Flags().BoolVar(&noCache, "no-cache", false, "do not use caching mecahnism")
	// TODO: Reverse this?
	buildCmd.Flags().BoolVar(&noFailImage, "no-fail-image", false, "do not create the failure image when an error occurs")
	buildCmd.Flags().BoolVar(&squash, "squash", false, "squash image to exactly one layer")
}

// This method preserves the initial working directory
// If unshare was used then we lose it since it creates a new user namespace
// and re-executes the binary at / in the namespace
func preserveCwd() {
	cwd := os.Getenv(OriginalCwdEnv)
	if cwd != "" {
		err := os.Chdir(cwd)
		if err != nil {
			log.Fatalf("Could not restore original working directory")
		}
		return
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Could not get working directory: %s", err)
	}
	os.Setenv(OriginalCwdEnv, cwd)
}

func build(cmd *cobra.Command, args []string) {
	preserveCwd()
	// TODO: Add checks for dependencies like ansible-playbook, callback plugin
	ctx := context.Background()
	config := config.ParseConfig(args[0], noCache, squash, noFailImage)
	ansible.PreprocessPlaybook(&config)
	builder := builder.NewBuilder(ctx, &config)
	err := ansible.RunPlaybook(&config)
	if err != nil {
		log.Error("Ansible run was not successful")
		config.TargetImage.Name = fmt.Sprintf("%s-failed-%d", config.TargetImage.Name, time.Now().Unix())
	}
	if err == nil || !config.NoFailImage {
		imageId := builder.Commit(&config)
		log.Infof("Created OCI image (%s) with id: %s\n", config.TargetImage.Name, imageId)
	}
	builder.Delete()
}
