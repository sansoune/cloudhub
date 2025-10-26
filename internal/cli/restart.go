package cli

import (
	"fmt"

	"cloudhub/internal/docker"

	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart <container>",
	Short: "Restart a container",
	Long:  `Restart a container by name or ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runRestart(args[0])
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}

func runRestart(nameOrID string) {
	client, err := docker.NewClient()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Printf("Restarting container '%s'...\n", nameOrID)

	err = client.RestartContainer(nameOrID)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Container '%s' restarted successfully\n", nameOrID)
}
