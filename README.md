# Cloudhub

**Cloudhub** is a lightweight CLI tool for monitoring and managing Docker containers in a homelab environment. It provides commands to list, start, stop, and restart containers, manage Docker‑Compose stacks, and run a daemon that continuously watches container state changes and sends optional notifications via **ntfy**.

---

## Table of Contents
1. [Project Overview](#project-overview)  
2. [Core Architecture](#core-architecture)  
3. [How the Code Works](#how-the-code-works)  
4. [Implemented Features](#implemented-features)  
5. [Installation](#installation)  
6. [Configuration](#configuration)  
7. [CLI Commands (API Endpoints)](#cli-commands)  
8. [Development Guide](#development-guide)  
9. [Roadmap](#roadmap)  
10. [License](#license)  

---

## Project Overview
Cloudhub is a **command‑line** application written in Go that interacts with the Docker daemon through the official Docker SDK. Its primary responsibilities are:

* Querying Docker for container information.
* Controlling container lifecycle (start, stop, restart).
* Managing Docker‑Compose “stacks” located under a fixed directory (`/opt/cloudhub`).
* Running a background daemon that periodically checks container states and emits notifications when containers appear, disappear, or change state.

The tool is intended for homelab operators who want a simple, scriptable interface without a full‑blown UI.

---

## Core Architecture

```
cloudhub/
├── cmd/
│   └── main.go                # Application entry point
├── internal/
│   ├── cli/                   # Cobra based CLI commands
│   │   ├── root.go
│   │   ├── ls.go
│   │   ├── start.go
│   │   ├── stop.go
│   │   ├── restart.go
│   │   ├── version.go
│   │   ├── daemon.go
│   │   └── stack.go
│   ├── config/                # YAML configuration loading
│   │   └── config.go
│   ├── daemon/                # Monitoring daemon implementation
│   │   └── monitor.go
│   ├── docker/                # Thin wrapper around Docker SDK
│   │   └── client.go
│   ├── notify/                # Notification abstraction (ntfy)
│   │   └── notifier.go
│   └── compose/               # Docker‑Compose stack management
│       └── manager.go
├── Makefile                   # Build/run helpers
└── go.mod / go.sum
```

* **CLI (`internal/cli`)** – defines the command hierarchy using Cobra, parses flags, and calls the service layer.
* **Docker client (`internal/docker`)** – encapsulates Docker SDK calls (list, start, stop, restart, version, ping).
* **Compose manager (`internal/compose`)** – discovers stacks under `/opt/cloudhub` and runs `docker compose` commands in the stack directory.
* **Configuration (`internal/config`)** – loads `config.yaml` from the current working directory, applies defaults (e.g., daemon interval = 60 s).
* **Daemon (`internal/daemon`)** – runs a ticker loop, compares current container states with the previous snapshot, and triggers notifications on changes.
* **Notifier (`internal/notify`)** – defines a `Notifier` interface; the only concrete implementation is `NtfyNotifier`, which POSTs plain‑text messages to an ntfy server.

---

## How the Code Works

1. **Program start** – `cmd/main.go` calls `cli.Execute()`.
2. **Cobra parses the command line** and dispatches to the appropriate sub‑command implementation.
3. **Configuration loading** (daemon only)  
   * `config.Load()` reads `config.yaml` (if present) and falls back to `config.DefaultConfig()`.  
   * CLI flags (`--interval`, `--ntfy-server`, `--ntfy-topic`) can override config values.
4. **Docker interactions** – each CLI command creates a `docker.NewClient()`, performs the requested operation, prints human‑readable output, and closes the client.
5. **Stack management** – `compose.FindStacks()` scans `/opt/cloudhub` for directories containing a `docker-compose.yml`. `stack up/down/restart` invoke `docker compose` with the appropriate arguments inside the stack’s directory.
6. **Daemon flow** (`cloudhub daemon run`)  
   * Builds a `daemon.Monitor` with the chosen interval and any configured notifiers.  
   * Starts a ticker; on each tick it calls `monitor.checkContainers()`.  
   * `checkContainers` fetches the full container list, builds a map of `name → state`, and calls `detectChanges` to compare with the previous snapshot.  
   * Detected events (new, disappeared, state changes) are printed and sent to each notifier (`NtfyNotifier` if enabled).  
   * The daemon runs until it receives an OS interrupt (`SIGINT`/`SIGTERM`), at which point `monitor.Stop()` is called.
7. **Notification** – `notify.NewNtfyNotifier(server, topic)` creates a notifier that POSTs the message to `http://<server>/<topic>` with `Content-Type: text/plain`. Errors are logged to stdout.

---

## Implemented Features

| Feature | CLI Command | Description |
|---------|-------------|-------------|
| List containers | `cloudhub ls` (and sub‑commands `running`, `stopped`, `paused`) | Shows name, state, status, and short ID in a table. |
| Start container | `cloudhub start <container>` | Starts a stopped container by name or ID. |
| Stop container | `cloudhub stop <container>` | Stops a running container gracefully (10 s timeout). |
| Restart container | `cloudhub restart <container>` | Restarts a container (stop → start). |
| Docker version | `cloudhub version` | Prints Docker daemon version information. |
| Stack discovery | `cloudhub stack ls` | Lists stacks (directories with `docker-compose.yml`) under `/opt/cloudhub`. |
| Stack lifecycle | `cloudhub stack up|down|restart <stack>` | Executes `docker compose up -d`, `down`, or a full restart for the specified stack. |
| Daemon monitoring | `cloudhub daemon run` | Periodically polls container states, detects changes, and optionally sends ntfy notifications. |
| Configurable interval & ntfy | Flags `--interval`, `--ntfy-server`, `--ntfy-topic` (daemon) | Override config values at runtime. |
| Notification via ntfy | `notify.NtfyNotifier` | Sends plain‑text messages to an ntfy server/topic. |

---

## Installation

### Prerequisites
* Go 1.22+ (or any recent version that satisfies the `go.mod` constraints)
* Docker daemon reachable from the host where Cloudhub runs
* (Optional) An ntfy server if you want notifications

### Build & Run

```bash
# Clone the repository
git clone https://github.com/your-org/cloudhub.git
cd cloudhub

# Build the binary (output: dist/cloudhub)
make build

# Run the binary (example)
./dist/cloudhub version
```

The provided `Makefile` also offers a convenient `make run` target:

```bash
make run ARGS="daemon run --interval 30"
```

---

## Configuration

Cloudhub reads a YAML file named `config.yaml` from the **current working directory**. If the file does not exist, defaults are used.

### Default configuration (`config.DefaultConfig()`)

```yaml
daemon:
  interval: 60   # seconds between daemon checks

notifications:
  ntfy:
    enabled: false
    server: ""    # e.g., "https://ntfy.sh"
    topic: ""     # e.g., "cloudhub"
```

### Overriding via CLI flags (daemon only)

| Flag | Description |
|------|-------------|
| `--interval <seconds>` | Override the daemon tick interval. |
| `--ntfy-server <url>` | Set the ntfy server URL and implicitly enable ntfy notifications. |
| `--ntfy-topic <topic>` | Set the ntfy topic name and implicitly enable ntfy notifications. |

If a flag is supplied, it takes precedence over the value from `config.yaml`.

---

## CLI Commands (API Endpoints)

Cloudhub is a **CLI**, not a network service. The public interface consists of the following commands:

| Command | Usage | Description |
|---------|-------|-------------|
| `cloudhub ls [type]` | `cloudhub ls` <br> `cloudhub ls running` <br> `cloudhub ls stopped` <br> `cloudhub ls paused` | List containers (all or filtered by state). |
| `cloudhub start <container>` | `cloudhub start my_container` | Start a stopped container. |
| `cloudhub stop <container>` | `cloudhub stop my_container` | Stop a running container. |
| `cloudhub restart <container>` | `cloudhub restart my_container` | Restart a container. |
| `cloudhub version` | – | Show Docker daemon version information. |
| `cloudhub stack ls` | – | List discovered Docker‑Compose stacks. |
| `cloudhub stack up <stack>` | – | Run `docker compose up -d` for the stack. |
| `cloudhub stack down <stack>` | – | Run `docker compose down` for the stack. |
| `cloudhub stack restart <stack>` | – | Restart the stack (`down` then `up`). |
| `cloudhub daemon run` | `cloudhub daemon run [flags]` | Start the monitoring daemon (supports `--interval`, `--ntfy-server`, `--ntfy-topic`). |

All commands output human‑readable status messages and exit with a non‑zero code on error.

---

## Development Guide

### Repository Layout
* **`cmd/`** – entry point (`main.go`).
* **`internal/`** – private packages (CLI, Docker wrapper, daemon, config, notification, compose).
* **`Makefile`** – common build/run/clean targets.

### Building locally
```bash
go test ./...          # (no tests currently, but run to verify build)
make build
```

### Adding a new CLI command
1. Create a new file under `internal/cli/` with a `cobra.Command`.
2. Register the command in `init()` by calling `rootCmd.AddCommand(yourCmd)`.
3. Use the existing Docker client (`docker.NewClient()`) or other internal services as needed.

### Extending notifications
* Implement a new type that satisfies `notify.Notifier` (method `Send(message string) error`).
* Add a constructor (e.g., `NewSlackNotifier`) and expose configuration flags if required.
* Append the new notifier to the `notifiers` slice in `runDaemon()` based on configuration.

### Contributing
1. Fork the repository.
2. Create a feature branch (`git checkout -b feat/your-feature`).
3. Write code and, where applicable, add unit tests.
4. Run `go vet ./...` and `golint` (if used) to keep code quality.
5. Submit a Pull Request with a clear description of the change.

---

## Roadmap

| Milestone | Description |
|-----------|-------------|
| **v0.2 – Additional Notifiers** | Add support for Slack, Discord, or email notifications alongside ntfy. |
| **v0.3 – Configurable Stacks Directory** | Allow the stacks root (`/opt/cloudhub`) to be overridden via config or flag. |
| **v0.4 – Unit & Integration Tests** | Introduce a test suite covering Docker client wrapper, daemon logic, and CLI commands (using mocks). |
| **v0.5 – Cross‑Platform Support** | Ensure the tool works on Windows (adjust Docker client initialization and path handling). |
| **v0.6 – Export Metrics** | Provide Prometheus metrics endpoint for the daemon (container state counts, errors, etc.). |

The above items are derived from existing `TODO`‑style gaps (e.g., only one notifier implementation, hard‑coded stacks directory, lack of tests).

---

## License

This project is licensed under the **MIT License**. See the `LICENSE` file for details.