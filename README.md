# Cloudhub
Docker homelab monitor tool
================================

## Project Overview
Cloudhub is a CLI tool for monitoring and managing Docker containers in your homelab. It provides features such as listing containers, starting and stopping containers, restarting containers, and managing stacks.

## Core Architecture
The project is structured into the following packages:

* `cmd`: contains the main entry point of the application
* `internal/cli`: contains the CLI commands and their implementations
* `internal/compose`: contains the stack management functionality
* `internal/config`: contains the configuration management functionality
* `internal/daemon`: contains the daemon functionality
* `internal/docker`: contains the Docker client functionality
* `internal/notify`: contains the notification functionality

## How the Code Works
The main flow of the application is as follows:

1. The user runs the `cloudhub` command with a specific subcommand (e.g. `ls`, `start`, `stop`, etc.)
2. The `cmd/main.go` file executes the corresponding CLI command
3. The CLI command interacts with the Docker client to perform the desired action
4. The Docker client communicates with the Docker daemon to perform the action
5. The result of the action is returned to the user

## Implemented Features
The following features are implemented:

* Listing containers: `cloudhub ls`
* Starting a container: `cloudhub start <container>`
* Stopping a container: `cloudhub stop <container>`
* Restarting a container: `cloudhub restart <container>`
* Managing stacks: `cloudhub stack <command>`
* Daemon management: `cloudhub daemon <command>`
* Notification support: `cloudhub daemon run` with notification flags

## Installation
To install Cloudhub, you can use the following command:
```bash
curl -sSL https://raw.githubusercontent.com/sansoune/cloudhub/main/scripts/quick-install.sh | bash
```
Alternatively, you can build from source by running the following commands:
```bash
git clone https://github.com/sansoune/cloudhub.git
cd cloudhub
go build -o cloudhub cmd/main.go
sudo mv cloudhub /usr/local/bin/cloudhub
```
## Configuration
The configuration file is located at `~/.config/cloudhub/config.yaml`. You can edit this file to configure the stack path, daemon interval, and notification settings.

The following flags are available:

* `--interval`: sets the daemon interval
* `--ntfy-server`: sets the ntfy server URL
* `--ntfy-topic`: sets the ntfy topic name

## Development Guide
To contribute to the project, you can follow these steps:

1. Fork the repository
2. Clone the repository to your local machine
3. Make changes to the code
4. Run `go build` to build the application
5. Run `go test` to run the tests
6. Submit a pull request with your changes

## License
Cloudhub is licensed under the MIT license.