package notify

import (
	"fmt"
	"net/http"
	"strings"
)

type Notifier interface {
	Send(message string) error
}

type NtfyNotifier struct {
	server string
	topic string
}

func NewNtfyNotifier(server, topic string) *NtfyNotifier {
	return &NtfyNotifier{
		server: server,
		topic: topic,
	}
}

func (n *NtfyNotifier) Send(message string) error {
	url := fmt.Sprintf("%s/%s", n.server, n.topic)
	
	resp, err := http.Post(url, "text/plain", strings.NewReader(message))
	if err != nil {
		return fmt.Errorf("failed to send ntfy notification: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ntfy returned status %d", resp.StatusCode)
	}

	return nil
}
