package cli

import (
	"fmt"

	"cloudhub/internal/docker"

	"github.com/spf13/cobra"

)

var startCmd = &cobra.Command{
	Use:   "start <container>",
	Short: "Start a stopped container",
	Long:  `Start a container by name or ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runStart(args[0])
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart(nameOrID string) {
	client, err := docker.NewClient()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Printf("Starting container '%s'...\n", nameOrID)

	err = client.StartConatiner(nameOrID)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Container '%s' started successfully\n", nameOrID)
}
