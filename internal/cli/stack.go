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

var stackUpCmd = &cobra.Command{
	Use: "up <stack>",
	Short: "Start a stack",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stackName := args[0]

		stack, err := compose.GetStack(stackName)
		if err != nil {
			return err
		}

		fmt.Printf("Starting stack '%s'...\n", stack.Name)
		return stack.Up()
	},
}

var stackDownCmd = &cobra.Command{
	Use: "down <stack>",
	Short: "Stop a stack",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stackName := args[0]

		stack, err := compose.GetStack(stackName)
		if err != nil {
			return err
		}

		fmt.Printf("Stopping stack '%s'...\n", stack.Name)
		return stack.Down()
	},
}

var stackRestartCmd = &cobra.Command{
	Use: "restart <stack>",
	Short: "Restart a stack",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stackName := args[0]

		stack, err := compose.GetStack(stackName)
		if err != nil {
			return err
		}

		fmt.Printf("Restarting stack '%s'...\n", stack.Name)
		return stack.Restart()
	},
}

func init() {
	stackCmd.AddCommand(lsStack)
	stackCmd.AddCommand(stackUpCmd)
	stackCmd.AddCommand(stackDownCmd)
	stackCmd.AddCommand(stackRestartCmd)

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


