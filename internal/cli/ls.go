package cli

import (
	"fmt"

	"cloudhub/internal/docker"
	"cloudhub/internal/compose"
	"github.com/spf13/cobra"
)

// Main ls command
var lsCmd = &cobra.Command{
	Use:   "ls [type]",
	Short: "List containers",
	Long:  `List all Docker containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		listContainers("") // List all containers
	},
}

// ls running subcommand
var lsRunningCmd = &cobra.Command{
	Use:   "running",
	Short: "List running containers",
	Run: func(cmd *cobra.Command, args []string) {
		listContainers("running")
	},
}

// ls stopped subcommand
var lsStoppedCmd = &cobra.Command{
	Use:   "stopped",
	Short: "List stopped containers",
	Run: func(cmd *cobra.Command, args []string) {
		listContainers("exited")
	},
}

// ls paused subcommand
var lsPausedCmd = &cobra.Command{
	Use:   "paused",
	Short: "List paused containers",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) > 0 && args[0] == "stacks" {
			listStacks()
			return
		}
		listContainers("paused")
	},
}

func init() {
	lsCmd.AddCommand(lsRunningCmd)
	lsCmd.AddCommand(lsStoppedCmd)
	lsCmd.AddCommand(lsPausedCmd)

	rootCmd.AddCommand(lsCmd)
}

func listContainers(state string) {
	client, err := docker.NewClient()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer client.Close()

	listAll := state != ""
	if state == "" {
		listAll = true // "ls" without args shows all
	}

	containers, err := client.ListContainers(listAll)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	if state != "" {
		filtered := []docker.Container{}
		for _, c := range containers {
			if c.State == state {
				filtered = append(filtered, c)
			}
		}
		containers = filtered
	}

	if len(containers) == 0 {
		fmt.Println("No containers found")
		return
	}

	// Print header
	fmt.Printf("\n%-20s %-12s %-20s %s\n", "NAME", "STATE", "STATUS", "ID")
	fmt.Println("────────────────────────────────────────────────────────────────────")

	// Print containers
	for _, c := range containers {
		fmt.Printf("%-20s %-12s %-20s %s\n", c.Name, c.State, c.Status, c.ID)
	}

	fmt.Printf("\nTotal: %d containers\n", len(containers))
}

func listStacks() {
	stacks, err := compose.FindStacks()
	if err != nil {
		fmt.Println("❌ Error: %v\n", err)
		return
	}

	if len(stacks) == 0 {
		fmt.Println("No stacks found")
		return
	}

	fmt.Printf("\n%-20s %s\n", "STACK NAME", "PATH")
	fmt.Println("────────────────────────────────────────────────────────────────────")

	for _, stack := range stacks {
		fmt.Printf("%-20s %s\n", stack.Name, stack.Path)
	}

	fmt.Printf("\nTotal: %d stacks\n", len(stacks))
}
