package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cloudhub",
	Short: "Docker homelab monitor tool",
	Long:  `Cloudhub is a CLI tool for monitoring and managing Docker containers in your homelab.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
