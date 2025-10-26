package cli

import (
	"fmt"

	"cloudhub/internal/docker"

	"github.com/spf13/cobra"

)

var stopCmd = &cobra.Command{
	Use:   "stop <container>",
	Short: "Stop a running container",
	Long:  `Stop a container by name or ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runStop(args[0])
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(nameOrID string) {
	client, err := docker.NewClient()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Printf("Stopping container '%s'...\n", nameOrID)

	err = client.StopContainer(nameOrID)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Container '%s' stopped successfully\n", nameOrID)
}
