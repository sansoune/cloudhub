package compose

import (
	"fmt"
	"os"
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

