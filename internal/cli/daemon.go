package cli

import (
	"cloudhub/internal/config"
	"cloudhub/internal/daemon"
	"cloudhub/internal/notify"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	daemonInterval int
	ntfyServer string
	ntfyTopic string
)

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
	daemonRunCmd.Flags().IntVar(&daemonInterval, "interval", 0, "Tick interval in seconds")
	daemonRunCmd.Flags().StringVar(&ntfyServer, "ntfy-server", "", "Ntfy server URL")
	daemonRunCmd.Flags().StringVar(&ntfyTopic, "ntfy-topic", "", "Ntfy topic name")
	
	daemonCmd.AddCommand(daemonRunCmd)
	rootCmd.AddCommand(daemonCmd)
}

func runDaemon() {
	fmt.Println("starting daemon...")

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Config file not found, using defaults: %v\n", err)
		cfg = config.DefaultConfig()

	} else {
		fmt.Println("Config loaded from file")
	}

	intervalDuration := cfg.Daemon.Interval
	if daemonInterval > 0 {
		intervalDuration = daemonInterval
		fmt.Printf("Overriding interval from flag: %d seconds\n", intervalDuration)
	}

	fmt.Printf("Check interval: %d seconds\n", intervalDuration)

	ntfyEnabled := cfg.Notifications.Ntfy.Enabled
	ntfySrvr := cfg.Notifications.Ntfy.Server
	ntfyTop := cfg.Notifications.Ntfy.Topic

	if ntfyServer != "" {
		ntfySrvr = ntfyServer
		ntfyEnabled = true
		fmt.Printf("Overriding ntfy server from flag: %s\n", ntfySrvr)
	}
	if ntfyTopic != "" {
		ntfyTop = ntfyTopic
		ntfyEnabled = true
		fmt.Printf("Overriding ntfy topic from flag: %s\n", ntfyTop)
	}

	var notifiers []notify.Notifier

	if ntfyEnabled && ntfyTop != "" {
		fmt.Printf("ntfy notification enabled (topic: %s)\n", ntfyTop)
		ntfyNotifier :=  notify.NewNtfyNotifier(ntfySrvr, ntfyTop)
		notifiers = append(notifiers, ntfyNotifier)
	}
	
	interval := time.Duration(intervalDuration) * time.Second
	monitor, err := daemon.NewMonitor(interval, notifiers)
	if err != nil {
		fmt.Printf("Failed to create monitor: %v\n", err)
		os.Exit(1)
	}

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
