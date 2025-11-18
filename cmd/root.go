package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
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
				log.SetLevel(log.InfoLevel)
			case 2:
				log.SetLevel(log.DebugLevel)
			case 3:
				log.SetLevel(log.TraceLevel)
			default:
				log.SetLevel(log.WarnLevel)
			}
		},
	}
)

func init() {
	rootCmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "verbose output (can be repeated: -v, -vv, -vvv)")
	rootCmd.AddCommand(buildCmd)
}

func Execute() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		DisableLevelTruncation: true,
	})
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
