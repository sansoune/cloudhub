package cli

import (
	"fmt"

	"cloudhub/internal/compose"
	"github.com/spf13/cobra"
)

var stackCmd = &cobra.Command{
	Use: "stack [command]",
	Short: "manage stacks",
}

var lsStack = &cobra.Command{
	Use: "ls",
	Short: "list available stacks under the specified dir",
	Run: func(cmd *cobra.Command, args []string) {
		listStacks()
	},
}

func init() {
	stackCmd.AddCommand(lsStack)

	rootCmd.AddCommand(stackCmd)
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


