package cli

import (
	"fmt"

	"cloudhub/internal/docker"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show Docker version information",
	Long:  `Display version information about the Docker daemon.`,
	Run: func(cmd *cobra.Command, args []string) {
		showVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func showVersion() {

	client, err := docker.NewClient()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer client.Close()

	version, err := client.GetVersion()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Println("🐳 Docker Version Information")
	fmt.Println("─────────────────────────────────────")
	fmt.Printf("Version:        %s\n", version.Version)
	fmt.Printf("API Version:    %s\n", version.APIVersion)
	fmt.Printf("OS/Arch:        %s/%s\n", version.Os, version.Arch)
	fmt.Printf("Git Commit:     %s\n", version.GitCommit)
	fmt.Println("─────────────────────────────────────")

}
