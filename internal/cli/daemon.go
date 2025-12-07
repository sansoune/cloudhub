package cli

import (
	"fmt"
	"time"
	"os"
	"os/signal"
	"syscall"
	"github.com/spf13/cobra"
	"cloudhub/internal/daemon"
)

var daemonInterval int

var daemonCmd = &cobra.Command{
	Use: "daemon",
	Short: "Daemon Management",
}

var daemonRunCmd = &cobra.Command{
	Use: "run",
	Short: "Run the daemon",
	Run: func(cmd *cobra.Command, args []string) {
		runDaemon()
	},
}

func init() {
	daemonRunCmd.Flags().IntVar(&daemonInterval, "interval", 10, "Tick interval in seconds")
	
	daemonCmd.AddCommand(daemonRunCmd)
	rootCmd.AddCommand(daemonCmd)
}

func runDaemon() {
	fmt.Println("starting daemon...")
	
	interval := time.Duration(daemonInterval) * time.Second
	monitor := daemon.NewMonitor(interval)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		errChan <- monitor.Start()
	}()

	select {
		case <-sigChan:
			fmt.Println("\nInterrupt received, stopping...")
			monitor.Stop()
case err := <-errChan:
if err != nil {
	fmt.Printf("error: %v\n", err)
	os.Exit(1)
	}
	}
}
