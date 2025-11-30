package compose

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const defaultStacksDir = "/opt/cloudhub"

type Stack struct {
	Name string
	Path string
}

func FindStacks() ([]Stack, error) {
	if _, err := os.Stat(defaultStacksDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("stacks directory does not exist: %s", defaultStacksDir)
	}

	entries, err := os.ReadDir(defaultStacksDir)

	if err != nil {
		return nil, fmt.Errorf("failed to read stacks directory: %w", err)
	}

	var stacks []Stack

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		stackDir := filepath.Join(defaultStacksDir, entry.Name())
		composeFile := filepath.Join(stackDir, "docker-compose.yml")

		if _, err := os.Stat(composeFile); err == nil {
			stacks = append(stacks, Stack{
				Name: entry.Name(),
				Path: stackDir,
			})
		}

	}

	return stacks, nil
}

func GetStack(name string) (*Stack, error) {
	stacks, err := FindStacks()
	if err != nil {
		return nil, err
	}

	for _, stack := range stacks {
		if stack.Name == name {
			return &stack, nil
		}
	}

	return nil, fmt.Errorf("stack '%s' not found", name)
}


// running compose command with args under the dedicated dir
func (s *Stack) runComposeCommand(args ...string) error {

	cmdArgs := append([]string{"compose"}, args...)

	cmd := exec.Command("docker", cmdArgs...)
	cmd.Dir = s.Path

	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("docker compose error: %v\nOutput:\n%s", err, string(output))
	}

	return nil
}

func (s *Stack) Up() error {
	return s.runComposeCommand("up", "-d")
}

func (s *Stack) Down() error {
	return s.runComposeCommand("down")
}

func (s *Stack) Restart() error {
	if err := s.runComposeCommand("down"); err != nil {
		return err
	}
	return s.runComposeCommand("up", "-d")
}
