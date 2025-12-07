package daemon

import (
	"cloudhub/internal/docker"
	"fmt"
	"time"
)

type Monitor struct {
	interval time.Duration
	stopChan chan struct{}
	client *docker.Client
}

func NewMonitor(interval time.Duration) (*Monitor, error) {
	client, err := docker.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}


	return &Monitor{
		interval: interval,
		stopChan: make(chan struct{}),
		client: client,
	}, nil
}

func (m *Monitor) Start() error {
	defer m.client.Close()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	fmt.Println("Daemon started")

	for {
		select {
			case <- ticker.C:
				constiners, err := m.client.ListContainers(true)
				if err != nil {
					return fmt.Errorf("failed to list containers: %w", err)
			}
				fmt.Printf("[%s] checked %d \n", time.Now().Format("15:04:05"), len(constiners))
			case <-m.stopChan:
				fmt.Println("Daemon stopped")
				return nil
		}
	}
}

func (m *Monitor) Stop() {
	close(m.stopChan)
}
