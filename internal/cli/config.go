package cli

import (
	"fmt"
	"os"

	"cloudhub/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config file path",
	Run: func(cmd *cobra.Command, args []string) {
		showConfigPath()
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configPathCmd)
	rootCmd.AddCommand(configCmd)
}

func initConfig() {
	fmt.Println("Initializing configuration...")
	if err := config.CreateConfigDir(); err != nil {
		fmt.Printf("Failed to create config directory: %v\n", err)
		os.Exit(1)
	}

	configPath, _ := config.GetConfigPath()


	if _, err := os.Stat(configPath); os.IsExist(err) {
		fmt.Printf("Config file already exists at: %s\n", configPath)
		fmt.Println("Remove it first if you want to recreate it")
		return
	}

	exampleConfig := `# Cloudhub Configuration File

stack_path: "/opt/stack"
daemon:
  # Check interval in seconds
  interval: 60

notifications:
  # Ntfy notifications
  ntfy:
    enabled: false
    server: ""
    topic: ""
`

	if err := os.WriteFile(configPath, []byte(exampleConfig), 0644); err != nil {
		fmt.Printf("Failed to create config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Config file created at: %s\n", configPath)
	fmt.Println("\nEdit the file to configure your settings:")
	fmt.Printf("   nano %s\n", configPath)
}

func showConfigPath() {
	configPath, err := config.GetConfigPath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Config path: %s\n", configPath)

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("\nConfig file does not exist")
		fmt.Println("Create it with: cloudhub config init")
	} else {
		fmt.Println("\nConfig file exists")
	}
}
