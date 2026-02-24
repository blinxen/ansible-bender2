package cmd

import (
	"os"

	"github.com/blinxen/ansible-bender2/internal/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	verbose int = 0
	rootCmd     = &cobra.Command{
		Use:   "ansible-bender2",
		Short: "Build OCI images using ansible",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			switch verbose {
			case 1:
				logging.InitLogger(logrus.InfoLevel)
			case 2:
				logging.InitLogger(logrus.DebugLevel)
			case 3:
				logging.InitLogger(logrus.TraceLevel)
			default:
				logging.InitLogger(logrus.WarnLevel)
			}
		},
	}
)

func init() {
	rootCmd.PersistentFlags().CountVarP(
		&verbose,
		"verbose",
		"v",
		"verbose output (can be repeated: -v, -vv, -vvv)",
	)
	rootCmd.AddCommand(buildCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
