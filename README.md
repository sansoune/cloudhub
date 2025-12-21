# Cloudhub
Cloudhub is a CLI tool for monitoring and managing Docker containers in your homelab.

## Project Overview
Cloudhub is designed to provide a simple and efficient way to manage Docker containers. It includes features such as listing containers, starting and stopping containers, restarting containers, and displaying Docker version information. Additionally, Cloudhub includes a daemon that can be installed as a systemd service to monitor containers and send notifications when changes are detected.

## Core Architecture
The Cloudhub project is organized into several packages:

* `internal/cli`: This package contains the command-line interface for Cloudhub, including the main entry point and all the subcommands.
* `internal/compose`: This package provides functionality for working with Docker Compose stacks.
* `internal/config`: This package is responsible for loading and storing configuration data.
* `internal/daemon`: This package contains the implementation of the Cloudhub daemon, which monitors containers and sends notifications.
* `internal/docker`: This package provides a client for interacting with the Docker API.
* `internal/notify`: This package defines the interface for notifiers, which are used to send notifications when changes are detected.

## How the Code Works
The main flow of the code is as follows:

1. The `main` function in `cmd/main.go` calls the `Execute` function in `internal/cli/root.go`, which sets up the command-line interface and executes the selected subcommand.
2. Each subcommand is implemented in a separate file in the `internal/cli` package, and they all follow a similar pattern:
	* They define a `cobra.Command` struct to represent the subcommand.
	* They implement the `Run` function for the subcommand, which performs the necessary actions.
3. The `daemon` subcommand is special, as it sets up and runs the Cloudhub daemon. The daemon is implemented in the `internal/daemon` package, and it uses the `internal/docker` package to interact with the Docker API.
4. The daemon monitors containers and sends notifications when changes are detected. It uses the `internal/notify` package to send notifications.

## Implemented Features
The following features are implemented in Cloudhub:

* Listing containers: The `ls` subcommand lists all containers, and the `ls running`, `ls stopped`, and `ls paused` subcommands list containers in specific states.
* Starting and stopping containers: The `start` and `stop` subcommands start and stop containers, respectively.
* Restarting containers: The `restart` subcommand restarts a container.
* Displaying Docker version information: The `version` subcommand displays version information about the Docker daemon.
* Daemon: The `daemon` subcommand sets up and runs the Cloudhub daemon, which monitors containers and sends notifications when changes are detected.
* Notifications: The daemon sends notifications when changes are detected, using the `internal/notify` package.
* Stacks: The `stack` subcommand provides functionality for working with Docker Compose stacks, including listing stacks, starting and stopping stacks, and restarting stacks.

## Installation
To install Cloudhub, you can use the following steps:

1. Download the latest release from the GitHub repository.
2. Extract the archive to a directory of your choice.
3. Run the `install.sh` script to install Cloudhub to `/usr/local/bin/cloudhub`.
4. Initialize the configuration by running `cloudhub config init`.
5. Install the daemon by running `cloudhub daemon install`.

## Configuration
Cloudhub uses a configuration file to store settings. The configuration file is located at `~/.config/cloudhub/config.yaml`. You can edit this file to customize the behavior of Cloudhub.

The following environment variables are used by Cloudhub:

* `CLOUDHUB_CONFIG_DIR`: The directory where the configuration file is located.
* `CLOUDHUB_STACK_DIR`: The directory where Docker Compose stacks are located.

## API Endpoints
Cloudhub does not provide any API endpoints. It is a command-line tool that interacts with the Docker API directly.

## Development Guide
To contribute to Cloudhub, you can follow these steps:

1. Clone the repository from GitHub.
2. Install the dependencies by running `go get`.
3. Build the project by running `go build`.
4. Run the tests by running `go test`.
5. Make changes to the code and submit a pull request.

## License
Cloudhub is licensed under the MIT License.