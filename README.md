# Cloudhub

**Cloudhub** is a lightweight CLI tool for monitoring and managing Docker containers and Docker‑Compose stacks in a homelab environment. It provides commands to list, start, stop, and restart containers, manage stacks, and run a daemon that continuously watches container state changes and sends notifications via **ntfy**.

---

## Table of Contents

- [Project Overview](#project-overview)  
- [Core Architecture](#core-architecture)  
- [How the Code Works](#how-the-code-works)  
- [Implemented Features](#implemented-features)  
- [Installation & Build](#installation--build)  
- [Configuration](#configuration)  
- [CLI Usage & Endpoints](#cli-usage--endpoints)  
- [Development Guide](#development-guide)  
- [Roadmap](#roadmap)  
- [License](#license)  

---

## Project Overview

Cloudhub is a **command‑line** utility that interacts with the Docker Engine API to:

1. **Inspect containers** – list all containers or filter by state (running, exited, paused).  
2. **Control containers** – start, stop, or restart a container by name or ID.  
3. **Manage Docker‑Compose stacks** – discover stacks under a predefined directory (`/opt/cloudhub`) and run `up`, `down`, or `restart` on them.  
4. **Run a daemon** – periodically poll container states, detect changes (new, stopped, started, disappeared, or any state transition) and push human‑readable notifications to an **ntfy** server.

All functionality is encapsulated in a set of internal packages (`cli`, `docker`, `compose`, `daemon`, `notify`) and exposed through a **Cobra**‑based CLI.

---

## Core Architecture

```
cloudhub/
├─ cmd/
│   └─ main.go                # entry point – invokes CLI
├─ internal/
│   ├─ cli/                   # Cobra commands & flag handling
│   │   ├─ root.go            # root command definition
│   │   ├─ ls.go              # list containers (all / running / stopped / paused)
│   │   ├─ start.go           # start a container
│   │   ├─ stop.go            # stop a container
│   │   ├─ restart.go         # restart a container
│   │   ├─ version.go         # Docker daemon version info
│   │   ├─ stack.go           # stack discovery & compose actions (ls, up, down, restart)
│   │   └─ daemon.go          # daemon sub‑command (run)
│   ├─ docker/                # thin wrapper around Docker SDK
│   │   └─ client.go          # client creation, Ping, Version, List, Start/Stop/Restart, etc.
│   ├─ compose/               # stack discovery & docker‑compose execution
│   │   └─ manager.go         # FindStacks, GetStack, Up/Down/Restart helpers
│   ├─ daemon/                # monitoring loop that polls containers and sends notifications
│   │   └─ monitor.go
│   └─ notify/                # notifier abstraction
│       └─ notifier.go        # Ntfy implementation (HTTP POST)
├─ Makefile                   # build/run/clean helpers
└─ go.mod / go.sum            # module definition
```

### Package Responsibilities

| Package | Responsibility |
|---------|-----------------|
| `cli`   | Parse user input, expose sub‑commands (`ls`, `start`, `stop`, `restart`, `stack`, `daemon`, `version`). |
| `docker`| Direct Docker Engine interaction (list containers, control lifecycle, fetch version). |
| `compose`| Locate Docker‑Compose stacks under `/opt/cloudhub` and invoke `docker compose` commands in the stack directory. |
| `daemon`| Periodic container state snapshot, diff detection, and dispatch of notifications. |
| `notify`| Define a generic `Notifier` interface; provide an `NtfyNotifier` that posts plain‑text messages to an ntfy server. |

---

## How the Code Works

1. **Program start** – `cmd/main.go` calls `cli.Execute()`.  
2. **Cobra** builds the command tree (`rootCmd` → sub‑commands). Flags for the daemon (`--interval`, `--ntfy-server`, `--ntfy-topic`) are registered in `daemon.go`.  
3. **Container commands** (`ls`, `start`, `stop`, `restart`) create a Docker client (`docker.NewClient()`), perform the requested operation, and close the client.  
4. **Stack commands** (`stack ls`, `stack up/down/restart`) use `compose.FindStacks()` to discover directories containing a `docker-compose.yml`. The selected stack runs `docker compose` with the appropriate arguments (`up -d`, `down`, etc.) via `exec.Command`.  
5. **Daemon** (`cloudhub daemon run`)  
   * Builds a slice of `notify.Notifier` based on CLI flags (currently only ntfy).  
   * Instantiates a `daemon.Monitor` with the polling interval and notifiers.  
   * Starts a ticker loop that calls `Monitor.checkContainers()` every interval.  
   * `checkContainers` fetches the current container list, builds a map of `name → state`, and invokes `detectChanges` to compare with the previous snapshot.  
   * Detected events (new, stopped, started, state change, disappeared) are printed to stdout and forwarded to each notifier via `Notifier.Send`.  
   * Graceful shutdown is handled via OS signals (`SIGINT`, `SIGTERM`).  

---

## Implemented Features

| Feature | CLI Command | Description |
|---------|-------------|-------------|
| **List containers** | `cloudhub ls` (all) <br> `cloudhub ls running` <br> `cloudhub ls stopped` <br> `cloudhub ls paused` | Shows a table with `NAME`, `STATE`, `STATUS`, `ID`. |
| **Start container** | `cloudhub start <container>` | Starts a stopped container (by name or ID). |
| **Stop container** | `cloudhub stop <container>` | Stops a running container (by name or ID). |
| **Restart container** | `cloudhub restart <container>` | Restarts a container (by name or ID). |
| **Docker version** | `cloudhub version` | Prints Docker daemon version details. |
| **Stack discovery** | `cloudhub stack ls` | Lists stacks (directories with `docker-compose.yml`) under `/opt/cloudhub`. |
| **Stack up** | `cloudhub stack up <stack>` | Executes `docker compose up -d` in the stack directory. |
| **Stack down** | `cloudhub stack down <stack>` | Executes `docker compose down`. |
| **Stack restart** | `cloudhub stack restart <stack>` | Runs `down` then `up -d`. |
| **Daemon monitoring** | `cloudhub daemon run` | Periodically polls containers, detects state changes, and sends notifications. |
| **Ntfy notifications** | `--ntfy-server`, `--ntfy-topic` flags on daemon | Sends plain‑text messages to `<server>/<topic>` via HTTP POST. |

---

## Installation & Build

### Prerequisites

* Go 1.22+ (module aware)  
* Docker Engine (client must be able to connect to the Docker socket)  
* (Optional) `docker compose` command available in `$PATH` for stack management.  

### Build

```bash
# Clone the repository
git clone https://github.com/your-org/cloudhub.git
cd cloudhub

# Build the binary (output: dist/cloudhub)
make build
```

### Run

```bash
# Execute the binary directly
./dist/cloudhub <command> [flags]

# Example: list all containers
./dist/cloudhub ls
```

You can also use `make run ARGS="ls"` to build and run in one step.

### Clean

```bash
make clean
```

---

## Configuration

| Source | Setting | Default | Description |
|--------|---------|---------|-------------|
| **Flag** | `--interval` (int) | `10` seconds | Polling interval for the daemon. |
| **Flag** | `--ntfy-server` (string) | `https://ntfy.dakhlaoui.tn` | Base URL of the ntfy server. |
| **Flag** | `--ntfy-topic` (string) | `cloudhub` | Topic name used when posting notifications. |
| **Environment** | Docker client uses standard Docker environment variables (`DOCKER_HOST`, `DOCKER_TLS_VERIFY`, etc.) as handled by `github.com/docker/docker/client`. | – | No additional env vars are required by Cloudhub itself. |

> **Note:** The daemon only creates an ntfy notifier when a non‑empty `--ntfy-topic` is supplied. If omitted, the daemon runs silently (no external notifications).

---

## CLI Usage & Endpoints

Cloudhub is a **CLI tool**, not an HTTP server, so there are no REST endpoints. Below is a quick reference of available commands.

### Root

```bash
cloudhub               # shows help with available sub‑commands
```

### Container Management

| Command | Syntax | Example |
|---------|--------|---------|
| List all containers | `cloudhub ls` | `cloudhub ls` |
| List running containers | `cloudhub ls running` | `cloudhub ls running` |
| List stopped containers | `cloudhub ls stopped` | `cloudhub ls stopped` |
| List paused containers | `cloudhub ls paused` | `cloudhub ls paused` |
| Start a container | `cloudhub start <name|id>` | `cloudhub start my-web` |
| Stop a container | `cloudhub stop <name|id>` | `cloudhub stop my-db` |
| Restart a container | `cloudhub restart <name|id>` | `cloudhub restart my-proxy` |
| Show Docker version | `cloudhub version` | `cloudhub version` |

### Stack Management

| Command | Syntax | Example |
|---------|--------|---------|
| List stacks | `cloudhub stack ls` | `cloudhub stack ls` |
| Bring a stack up | `cloudhub stack up <stack>` | `cloudhub stack up homeassistant` |
| Bring a stack down | `cloudhub stack down <stack>` | `cloudhub stack down homeassistant` |
| Restart a stack | `cloudhub stack restart <stack>` | `cloudhub stack restart homeassistant` |

> Stacks are discovered under **`/opt/cloudhub`**. Each sub‑directory that contains a `docker-compose.yml` file is considered a stack.

### Daemon

```bash
cloudhub daemon run [--interval <seconds>] [--ntfy-server <url>] [--ntfy-topic <topic>]
```

*Runs continuously, printing state changes and optionally posting them to ntfy.*

**Example with notifications:**

```bash
cloudhub daemon run --interval 30 --ntfy-server https://ntfy.sh --ntfy-topic myhomelab
```

The daemon can be stopped gracefully with `Ctrl+C` (SIGINT) or by sending `SIGTERM`.

---

## Development Guide

### Repository Layout

```
cmd/                # entry point (main)
internal/
  cli/              # Cobra commands
  docker/           # Docker SDK wrapper
  compose/          # Stack discovery & compose execution
  daemon/           # Monitoring loop
  notify/           # Notifier abstraction (ntfy)
Makefile            # build/run/clean shortcuts
go.mod / go.sum    # module definition
```

### Adding a New Command

1. Create a new file under `internal/cli/` (e.g., `mycmd.go`).  
2. Define a `*cobra.Command` and implement its `Run`/`RunE`.  
3. Register the command in `init()` with `rootCmd.AddCommand(myCmd)`.  
4. Use the existing `docker.NewClient()` or other internal packages as needed.

### Adding a New Notifier

1. Implement the `notify.Notifier` interface (`Send(message string) error`).  
2. Add a constructor (e.g., `NewSlackNotifier`).  
3. Extend `daemon.go` to instantiate the new notifier based on additional flags or env vars.

### Testing

*The repository currently does not contain test files.*  
When adding new functionality, consider writing unit tests for:

* Docker client wrappers (use the Docker SDK mock or a test daemon).  
* Stack discovery (`compose.FindStacks`).  
* Monitor change detection (`daemon.detectChanges`).  

Run tests with:

```bash
go test ./...
```

### Linting & Formatting

```bash
go fmt ./...
go vet ./...
```

---

## Roadmap

| Milestone | Description |
|-----------|-------------|
| **Improved error handling & logging** | Replace `fmt.Printf` with a structured logger (e.g., `zap` or `logrus`). |
| **Additional notifier back‑ends** | Add Slack, Discord, or email notifiers behind the `notify.Notifier` interface. |
| **Config file support** | Allow a YAML/JSON config file to specify default daemon interval, ntfy settings, and stacks directory. |
| **Test coverage** | Introduce unit and integration tests for all core packages. |
| **Cross‑platform stack discovery** | Make the default stacks directory configurable via flag or env var. |
| **Docker‑Compose v2 compatibility** | Detect and use `docker compose` vs `docker-compose` binaries automatically. |

Contributions that address any of the above items are welcome!

---

## License

This project is licensed under the **MIT License**.

---