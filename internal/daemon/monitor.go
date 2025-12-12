package daemon

import (
	"cloudhub/internal/docker"
	"cloudhub/internal/notify"
	"fmt"
	"time"
)

type Monitor struct {
	interval time.Duration
	stopChan chan struct{}
	client *docker.Client
	previousState map[string]string
	notifiers []notify.Notifier
}

func NewMonitor(interval time.Duration, notifiers []notify.Notifier) (*Monitor, error) {
	client, err := docker.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}


	return &Monitor{
		interval: interval,
		stopChan: make(chan struct{}),
		client: client,
		previousState: make(map[string]string),
		notifiers: notifiers,
	}, nil
}

func (m *Monitor) Start() error {
	defer m.client.Close()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	fmt.Println("Daemon started")

	if err := m.checkContainers(); err != nil {
		fmt.Printf("Initial check failed: %v\n", err)
	}

	for {
		select {
			case <- ticker.C:
				if err := m.checkContainers(); err != nil {
					fmt.Printf("Check failed: %v\n", err)
				}
			case <-m.stopChan:
				fmt.Println("Daemon stopped")
				return nil
		}
	}
}

func (m *Monitor) Stop() {
	close(m.stopChan)
}


func (m *Monitor) checkContainers() error {
	containers, err := m.client.ListContainers(true)
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	current_state := make(map[string]string)
	for _, c := range containers {
		current_state[c.Name] = c.State
	}

	if len(m.previousState) > 0 {
		m.detectChanges(current_state)
	}

	m.previousState = current_state

	return nil
}

func (m *Monitor) detectChanges(currentState map[string]string) {
	for name, currentStatus := range currentState {
		previousStatus, existed := m.previousState[name]
		if !existed {
			msg := fmt.Sprintf("NEW: Container '%s' appeared (%s)\n", name, currentStatus)
			fmt.Println(msg)
			m.SendNotification(msg)
		} else if previousStatus != currentStatus {
			// State changed
			if previousStatus == "running" && currentStatus == "exited" {
				msg := fmt.Sprintf("ALERT: Container '%s' stopped (running → exited)\n", name)
				fmt.Println(msg)
				m.SendNotification(msg)
			} else if previousStatus == "exited" && currentStatus == "running" {
				msg := fmt.Sprintf("RECOVERED: Container '%s' started (exited → running)\n", name)
				fmt.Println(msg)
				m.SendNotification(msg)
			} else {
				msg := fmt.Sprintf("CHANGE: Container '%s' changed state: %s → %s\n", name, previousStatus, currentStatus)
				fmt.Println(msg)
				m.SendNotification(msg)
			}
		}
	}

	for name, previousStatus := range m.previousState {
		if _, exists := currentState[name]; !exists {
			msg := fmt.Sprintf("DISAPPEARED: Container '%s' is gone (was: %s)\n", name, previousStatus)
			fmt.Println(msg)
			m.SendNotification(msg)
		}
	}
}

		
func (m *Monitor) SendNotification(message string) {
	for _, notifier := range m.notifiers {
		if err := notifier.Send(message); err != nil {
			fmt.Printf("Failed to send notification: %v\n", err)
		}
	}
}
