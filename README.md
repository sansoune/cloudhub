# Cloudhub
Docker homelab monitor tool

## Project Overview
Cloudhub is a CLI tool for monitoring and managing Docker containers in your homelab. It provides features such as listing containers, starting and stopping containers, restarting containers, and managing stacks.

## Core Architecture
The project is divided into several packages:
- `internal/cli`: contains the command-line interface logic
- `internal/compose`: handles Docker Compose operations
- `internal/config`: manages configuration files and settings
- `internal/daemon`: implements the daemon functionality for monitoring containers
- `internal/docker`: provides a client for interacting with the Docker API
- `internal/notify`: defines notifiers for sending notifications

## How the Code Works
The main flow of the application starts with the `cmd/main.go` file, which executes the `cli.Execute()` function. This function sets up the command-line interface and handles user input. The `internal/cli` package defines various commands, such as `ls`, `start`, `stop`, `restart`, and `daemon`, each with its own set of subcommands.

## Implemented Features
- List containers: `cloudhub ls`
- Start a container: `cloudhub start <container>`
- Stop a container: `cloudhub stop <container>`
- Restart a container: `cloudhub restart <container>`
- Manage stacks: `cloudhub stack ls`, `cloudhub stack up <stack>`, `cloudhub stack down <stack>`, `cloudhub stack restart <stack>`
- Daemon management: `cloudhub daemon run`, `cloudhub daemon install`, `cloudhub daemon uninstall`, `cloudhub daemon start`, `cloudhub daemon stop`, `cloudhub daemon restart`, `cloudhub daemon status`, `cloudhub daemon logs`
- Configuration management: `cloudhub config init`, `cloudhub config path`
- Notifications: supports Ntfy notifications

## Installation
To install Cloudhub, you can use the provided `install.sh` script or build the project manually using Go modules:
```bash
go build -o cloudhub cmd/main.go
sudo mv cloudhub /usr/local/bin/cloudhub
cloudhub config init
cloudhub daemon install
```

## Configuration
Cloudhub uses a configuration file located at `~/.config/cloudhub/config.yaml`. You can initialize the configuration file using `cloudhub config init`. The configuration file supports the following settings:
- `stack_path`: the directory where Docker Compose files are stored
- `daemon.interval`: the interval at which the daemon checks for container changes
- `notifications.ntfy.enabled`: enables or disables Ntfy notifications
- `notifications.ntfy.server`: the Ntfy server URL
- `notifications.ntfy.topic`: the Ntfy topic name

## API Endpoints
Cloudhub does not provide a REST API. Instead, it uses a command-line interface to interact with the user.

## Development Guide
To contribute to Cloudhub, follow these steps:
1. Clone the repository: `git clone https://github.com/sansoune/cloudhub.git`
2. Install dependencies: `go mod tidy`
3. Build the project: `go build -o cloudhub cmd/main.go`
4. Run the tests: `go test ./...`
5. Submit a pull request with your changes

## License
Cloudhub is licensed under the MIT License.