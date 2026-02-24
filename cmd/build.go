package cmd

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/blinxen/ansible-bender2/internal/ansible"
	"github.com/blinxen/ansible-bender2/internal/builder"
	"github.com/blinxen/ansible-bender2/internal/config"
)

// flags
var (
	noCache              bool
	createImageOnFailure bool
	noSquash             bool
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
	buildCmd.Flags().BoolVar(&createImageOnFailure, "create-image-on-failure", false, "if the playbook run fails then create a image with the current state")
	buildCmd.Flags().BoolVar(&noSquash, "no-squash", false, "do not squash image")
}

func build(cmd *cobra.Command, args []string) {
	// TODO: Add checks for dependencies like ansible-playbook, callback plugin
	config := config.ParseConfig(args[0], noCache, noSquash, createImageOnFailure)
	ansible.PreprocessPlaybook(&config)
	builder.CreateWorkingContainer(&config)
	err := ansible.RunPlaybook(&config)
	if err != nil {
		log.Error("Ansible run was not successful")
		config.TargetImage.Name = fmt.Sprintf("%s-failed-%d", config.TargetImage.Name, time.Now().Unix())
	}
	if err == nil || config.CreateFailImage {
		imageId := builder.CommitWorkingContainer(&config)
		log.Infof("Created OCI image (%s) with id: %s\n", config.TargetImage.Name, imageId)
	}
	builder.DeleteWorkingContainer(&config)
}
