package daemon

import (
	"fmt"
	"time"
)

type Monitor struct {
	interval time.Duration
	stopChan chan struct{}
}

func NewMonitor(interval time.Duration) *Monitor {
	return &Monitor{
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

func (m *Monitor) Start() error {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	fmt.Println("Daemon started")

	for {
		select {
			case <- ticker.C:
				fmt.Printf("[%s] Tick...\n", time.Now().Format("15:04:05"))
			case <-m.stopChan:
				fmt.Println("Daemon stopped")
				return nil
		}
	}
}

func (m *Monitor) Stop() {
	close(m.stopChan)
}
