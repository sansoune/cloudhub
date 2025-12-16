package cli

import (
	"cloudhub/internal/config"
	"cloudhub/internal/daemon"
	"cloudhub/internal/notify"
	"fmt"
	"os"
	"os/signal"
	"os/exec"
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
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		runDaemon()
	},
}

var daemonInstallCmd = &cobra.Command{
	Use: "install",
	Short: "Install daemon as systemd service",
	Long:  `Install Cloudhub as a systemd service that runs on boot.`,
	Run: func(cmd *cobra.Command, args []string) {
		installService()
	},
}

func init() {
	daemonRunCmd.Flags().IntVar(&daemonInterval, "interval", 0, "Tick interval in seconds")
	daemonRunCmd.Flags().StringVar(&ntfyServer, "ntfy-server", "", "Ntfy server URL")
	daemonRunCmd.Flags().StringVar(&ntfyTopic, "ntfy-topic", "", "Ntfy topic name")
	
	daemonCmd.AddCommand(daemonRunCmd)
	daemonCmd.AddCommand(daemonInstallCmd)
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

func installService() {
	fmt.Println("Installing Cloudhub as systemd service...")

	if _, err := os.Stat("/usr/local/bin/cloudhub"); os.IsNotExist(err) {
		fmt.Println("Cloudhub binary not found at /usr/local/bin/cloudhub")
		fmt.Println("Copy Cloudhub binary to /usr/local/bin/ then install the daemon")
		return
	}

	if _, err := os.Stat("/etc/systemd/system/cloudhub.service"); err == nil {
		fmt.Println("Service already installed")
		fmt.Println("Uninstall first: cloudhub daemon uninstall")
		return
	}

	serviceContent := `[Unit]
Description=Cloudhub Docker Monitoring Daemon
After=docker.service network.target
Requires=docker.service

[Service]
Type=simple
ExecStart=/usr/local/bin/cloudhub daemon run
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`

	tmpFile := "/tmp/cloudhub.service"
	if err := os.WriteFile(tmpFile, []byte(serviceContent), 0644); err != nil {
		fmt.Printf("Failed to create service file: %v\n", err)
		return
	}

	cmd := exec.Command("sudo", "cp", tmpFile, "/etc/systemd/system/cloudhub.service")
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Failed to install service: %v\n%s\n", err, output)
		return
	}

	cmd = exec.Command("sudo", "systemctl", "daemon-reload")
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Failed to reload systemd: %v\n%s\n", err, output)
		return
	}

	cmd = exec.Command("sudo", "systemctl", "enable", "cloudhub")
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Failed to enable service: %v\n%s\n", err, output)
	}

	fmt.Println("Service installed successfully!")
	fmt.Println("\nQuick Start:")
	fmt.Println("   cloudhub daemon start    - Start the daemon")
	fmt.Println("   cloudhub daemon status   - Check status")
	fmt.Println("   cloudhub daemon logs     - View logs")
	fmt.Println("   cloudhub daemon stop     - Stop the daemon")
	fmt.Println("\nThe daemon will now start automatically on boot!")
}
