package cmd

import (
	"fmt"
	"time"

	"github.com/blinxen/ansible-bender2/internal/ansible"
	"github.com/blinxen/ansible-bender2/internal/buildah"
	"github.com/blinxen/ansible-bender2/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
	SilenceUsage: true,
	Run:          build,
}

func init() {
	buildCmd.Flags().BoolVar(&noCache, "no-cache", false, "do not use caching mecahnism")
	buildCmd.Flags().BoolVar(
		&createImageOnFailure,
		"create-image-on-failure",
		false,
		"if the playbook run fails then create a image with the current state",
	)
	buildCmd.Flags().BoolVar(&noSquash, "no-squash", false, "do not squash image")
}

func build(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		logrus.Fatalf("expected single playbook argument, got %d arguments instead", len(args))
	}
	// TODO: Add checks for dependencies like ansible-playbook, callback plugin
	benderConfig, err := config.ParseConfig(args[0], noCache, noSquash, createImageOnFailure)
	if err != nil {
		config.GetLogger().WithError(err).Fatal("playbook parsing failed")
	}
	err = ansible.PreprocessPlaybook(benderConfig)
	if err != nil {
		ansible.GetLogger().WithError(err).Fatal("preprocessing playbook failed")
	}
	err = buildah.CreateWorkingContainer(benderConfig)
	if err != nil {
		buildah.GetLogger().Fatal("working container creation failed")
	}
	err = ansible.RunPlaybook(benderConfig)
	if err != nil {
		ansible.GetLogger().Error("playbook execution failed")
		if benderConfig.CreateFailImage {
			benderConfig.Squash = false
			benderConfig.TargetImage.Name = fmt.Sprintf(
				"%s-failed-%d",
				benderConfig.TargetImage.Name,
				time.Now().Unix(),
			)
			commitImage(benderConfig)
			logrus.Fatal("failure image was created")
		}
	}
	commitImage(benderConfig)
	err = buildah.DeleteWorkingContainer(benderConfig)
	if err != nil {
		buildah.GetLogger().Fatal("working container deletion failed")
	}
}

func commitImage(config *config.Config) {
	imageId, err := buildah.CommitWorkingContainer(config)
	if err != nil {
		buildah.GetLogger().WithError(err).Error("working container committing failed")
		return
	}
	buildah.GetLogger().Infof(
		"created OCI image (%s) with id: %s\n",
		config.TargetImage.Name,
		imageId,
	)
	fmt.Println(imageId)
}
