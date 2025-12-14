# Cloudhub

**Cloudhub** is a lightweight CLI tool for monitoring and managing Docker containers in a homelab environment. It provides commands to list, start, stop, and restart containers, manage Docker‑Compose stacks, and run a daemon that watches container state changes and sends optional notifications via **ntfy**.

---

## Table of Contents
1. [Project Overview](#project-overview)  
2. [Core Architecture](#core-architecture)  
3. [How the Code Works](#how-the-code-works)  
4. [Implemented Features](#implemented-features)  
5. [Installation & Build](#installation--build)  
6. [Configuration](#configuration)  
7. [CLI Commands & Flags](#cli-commands--flags)  
8. [Development Guide](#development-guide)  
9. [Roadmap](#roadmap)  
10. [License](#license)  

---

## Project Overview
Cloudhub is a **command‑line interface (CLI)** that interacts with the local Docker daemon to:

* Inspect containers (list, filter by state).  
* Control container lifecycle (start, stop, restart).  
* Manage Docker‑Compose *stacks* (discover, `up`, `down`, `restart`).  
* Run a background **daemon** that periodically checks container states, detects changes (new, stopped, started, disappeared, state transitions) and optionally pushes notifications to an **ntfy** server.

All functionality is implemented in pure Go, using the official Docker SDK and the Cobra library for CLI parsing.

---

## Core Architecture

| Package | Responsibility |
|---------|----------------|
| `cmd/main.go` | Entry point – calls `cli.Execute()`. |
| `internal/cli` | Cobra command definitions (`root`, `ls`, `start`, `stop`, `restart`, `version`, `stack`, `daemon`). Handles flag parsing and orchestrates calls to lower‑level packages. |
| `internal/config` | Loads `config.yaml` (if present), provides defaults, and exposes configuration structs (`StackPath`, daemon interval, ntfy settings). |
| `internal/docker` | Thin wrapper around the Docker SDK: creates a client, implements `Ping`, `GetVersion`, `ListContainers`, `StartConatiner`, `StopContainer`, `RestartContainer`, `GetContainer`, and client cleanup. |
| `internal/compose` | Discovers stacks under a configurable directory, validates presence of `docker-compose.yml`, and runs `docker compose` commands (`up`, `down`, `restart`) in the stack’s directory. |
| `internal/daemon` | Implements `Monitor` that runs a ticker, polls container list, detects state changes, and forwards messages to registered `notify.Notifier`s. |
| `internal/notify` | Defines a `Notifier` interface and an `NtfyNotifier` implementation that POSTs plain‑text messages to an ntfy server/topic. |
| `Makefile` | Convenience targets: `build`, `run`, `clean`. Produces binary `dist/cloudhub`. |

The architecture follows a **layered** approach: CLI → Service (daemon/compose) → Docker client → Docker daemon, with configuration and notification handling injected where needed.

---

## How the Code Works

1. **Startup**  
   * `cmd/main.go` calls `cli.Execute()`.  
   * Cobra parses the first argument (e.g., `ls`, `stack`, `daemon`) and dispatches to the appropriate command handler.

2. **Configuration**  
   * `config.Load()` looks for `config.yaml` in the current working directory.  
   * If missing, defaults are applied (`/opt/stack` for stacks, `60s` daemon interval, ntfy disabled).

3. **Container Commands** (`ls`, `start`, `stop`, `restart`, `version`)  
   * Each command creates a Docker client via `docker.NewClient()`.  
   * The client calls the Docker SDK to perform the requested operation and prints human‑readable output.

4. **Stack Management** (`stack ls|up|down|restart`)  
   * `compose.FindStacks()` scans the configured stack directory for sub‑folders containing a `docker-compose.yml`.  
   * `compose.GetStack(name)` returns a `Stack` struct.  
   * Stack methods (`Up`, `Down`, `Restart`) invoke `docker compose` with the appropriate arguments inside the stack’s directory.

5. **Daemon** (`daemon run`)  
   * Flags `--interval`, `--ntfy-server`, `--ntfy-topic` can override config values.  
   * `daemon.NewMonitor(interval, notifiers)` creates a monitor with a Docker client and any enabled notifiers.  
   * `Monitor.Start()` runs a ticker at the configured interval:  
     * Calls `ListContainers(true)` to get the current state of all containers.  
     * Compares with the previous snapshot, detects **new**, **stopped**, **started**, **state changes**, and **disappeared** containers.  
     * Prints a message to stdout and forwards it to each notifier (`NtfyNotifier` if enabled).  
   * The daemon gracefully stops on `SIGINT`/`SIGTERM`.

6. **Notification** (`notify.NtfyNotifier`)  
   * Constructs a URL `server/topic`.  
   * Sends a POST request with `text/plain` payload containing the message.  
   * Errors are logged but do not abort the daemon.

---

## Implemented Features

| Feature | CLI Command(s) | Description |
|---------|----------------|-------------|
| **List containers** | `cloudhub ls` (all) <br> `cloudhub ls running` <br> `cloudhub ls stopped` <br> `cloudhub ls paused` | Shows a table with `NAME`, `STATE`, `STATUS`, `ID`. |
| **Start container** | `cloudhub start <name|id>` | Starts a stopped container. |
| **Stop container** | `cloudhub stop <name|id>` | Stops a running container (10 s timeout). |
| **Restart container** | `cloudhub restart <name|id>` | Restarts a container (10 s timeout). |
| **Docker version** | `cloudhub version` | Prints Docker daemon version details. |
| **Stack discovery** | `cloudhub stack ls` | Lists stacks (folders with `docker-compose.yml`). |
| **Stack lifecycle** | `cloudhub stack up <stack>` <br> `cloudhub stack down <stack>` <br> `cloudhub stack restart <stack>` | Executes `docker compose up -d`, `down`, or a full restart for the selected stack. |
| **Daemon monitoring** | `cloudhub daemon run` | Periodically polls containers, detects state changes, prints messages, and optionally notifies via ntfy. |
| **Configurable interval** | `--interval <seconds>` (daemon flag) | Overrides the daemon tick interval. |
| **Ntfy notifications** | `--ntfy-server <url>` <br> `--ntfy-topic <topic>` (daemon flags) | Enables ntfy notifications and overrides server/topic. |
| **Graceful shutdown** | `Ctrl‑C` or `SIGTERM` | Stops the daemon cleanly. |

---

## Installation & Build

### Prerequisites
* Go **1.18+** (module support).  
* Docker daemon reachable via the environment (`DOCKER_HOST`, etc.).  
* (Optional) `docker compose` command available in `$PATH` for stack operations.  

### Steps

```bash
# Clone the repository
git clone https://github.com/your-org/cloudhub.git
cd cloudhub

# Build the binary (produces ./dist/cloudhub)
make build

# Run the binary directly
./dist/cloudhub --help
```

### Makefile Targets
| Target | Description |
|--------|-------------|
| `make build` | Compiles the binary into `dist/cloudhub`. |
| `make run ARGS="daemon run"` | Builds then runs the binary with optional arguments. |
| `make clean` | Removes the binary and any coverage artifacts. |

---

## Configuration

Cloudhub reads a YAML file named **`config.yaml`** from the current working directory.

### Default Values (when `config.yaml` is absent)

```yaml
stack_path: /opt/stack          # Directory where Docker‑Compose stacks live
daemon:
  interval: 60                  # Seconds between daemon checks
notifications:
  ntfy:
    enabled: false
    server: ""
    topic: ""
```

### Example `config.yaml`

```yaml
stack_path: /home/user/my-stacks
daemon:
  interval: 30
notifications:
  ntfy:
    enabled: true
    server: https://ntfy.sh
    topic: cloudhub-alerts
```

### Overriding via Flags (daemon only)

| Flag | Effect |
|------|--------|
| `--interval <seconds>` | Overrides `daemon.interval`. |
| `--ntfy-server <url>` | Sets ntfy server and implicitly enables ntfy. |
| `--ntfy-topic <topic>` | Sets ntfy topic and implicitly enables ntfy. |

If a flag is supplied, it takes precedence over the configuration file.

---

## CLI Commands & Flags

```text
cloudhub
├─ ls [type]               List containers (all, running, stopped, paused)
│   ├─ running
│   ├─ stopped
│   └─ paused
├─ start <container>       Start a stopped container
├─ stop <container>        Stop a running container
├─ restart <container>     Restart a container
├─ version                 Show Docker daemon version information
├─ stack
│   ├─ ls                  List discovered stacks
│   ├─ up <stack>          Bring a stack up (docker compose up -d)
│   ├─ down <stack>        Bring a stack down
│   └─ restart <stack>     Restart a stack (down + up)
└─ daemon
    └─ run                 Run the monitoring daemon
        --interval int     Tick interval in seconds (default from config)
        --ntfy-server str  Ntfy server URL (overrides config)
        --ntfy-topic str   Ntfy topic name (overrides config)
```

All commands print clear success (`✅`) or error (`❌`) messages.

---

## Development Guide

1. **Fork & Clone**  
   ```bash
   git clone https://github.com/<your-username>/cloudhub.git
   cd cloudhub
   ```

2. **Run Tests (if added later)**  
   The repository currently contains no test files; add unit tests under `*_test.go` as needed.

3. **Make Changes**  
   * Follow the existing package layout.  
   * Use `go fmt` and `go vet` to keep code tidy.  

4. **Build & Verify**  
   ```bash
   make build
   ./dist/cloudhub --help
   ```

5. **Submit a Pull Request**  
   * Ensure the PR builds on the latest `main`.  
   * Include documentation updates (README, comments) for any new commands or flags.

---

## Roadmap

| Milestone | Description |
|-----------|-------------|
| **Extended Notification Backends** | Add support for Slack, Discord, email, etc., alongside ntfy. |
| **Resource‑Usage Monitoring** | Capture CPU / memory stats per container and expose them via the daemon. |
| **Web UI** | Provide a minimal web dashboard for real‑time container status and stack control. |
| **Improved Config Discovery** | Allow config file location via environment variable (`CLOUDHUB_CONFIG`) or `$XDG_CONFIG_HOME`. |
| **Unit & Integration Tests** | Add comprehensive test coverage for Docker client wrapper, daemon logic, and CLI commands. |
| **Cross‑Platform Packaging** | Distribute pre‑built binaries for Linux, macOS, and Windows (e.g., via GitHub Releases). |

Contributions that address any of the above items are welcome.

---

## License

This project is licensed under the **MIT License**. See the `LICENSE` file for details.