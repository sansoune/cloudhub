# Cloudhub

**Cloudhub** is a lightweight CLI tool for monitoring and managing Docker containers in a homelab environment. It provides commands to list, start, stop, and restart containers, manage Docker‑Compose stacks, and run a daemon that continuously watches container state changes.

---

## Table of Contents
1. [Project Overview](#project-overview)  
2. [Core Architecture](#core-architecture)  
3. [How the Code Works](#how-the-code-works)  
4. [Implemented Features](#implemented-features)  
5. [Installation](#installation)  
6. [Configuration](#configuration)  
7. [CLI Commands](#cli-commands)  
8. [Development Guide](#development-guide)  
9. [Roadmap](#roadmap)  
10. [License](#license)  

---

## Project Overview
Cloudhub is a **command‑line interface (CLI)** written in Go that interacts with the Docker daemon via the official Docker SDK. It is intended for homelab users who need a simple, scriptable way to:

* Inspect containers (all, running, stopped, paused)  
* Control container lifecycle (start, stop, restart)  
* Manage groups of services defined by Docker‑Compose stacks  
* Run a background daemon that periodically checks container states and reports changes (new, stopped, started, disappeared, etc.)

The tool does **not** expose an HTTP API; all interactions happen through the terminal.

---

## Core Architecture

```
cloudhub/
├─ cmd/
│   └─ main.go                # Entry point – invokes CLI Execute()
├─ internal/
│   ├─ cli/                   # Cobra based command definitions
│   │   ├─ root.go            # Root command (cloudhub)
│   │   ├─ daemon.go          # `cloudhub daemon run` implementation
│   │   ├─ ls.go              # `cloudhub ls` sub‑commands
│   │   ├─ start.go           # `cloudhub start`
│   │   ├─ stop.go            # `cloudhub stop`
│   │   ├─ restart.go         # `cloudhub restart`
│   │   ├─ version.go         # `cloudhub version`
│   │   └─ stack.go           # `cloudhub stack` (ls, up, down, restart)
│   ├─ daemon/                # Periodic monitor implementation
│   │   └─ monitor.go
│   ├─ compose/               # Docker‑Compose stack handling
│   │   └─ manager.go
│   └─ docker/                # Thin wrapper around Docker SDK
│       └─ client.go
└─ Makefile                   # Build/run helpers
```

* **CLI layer (`internal/cli`)** – defines all user‑facing commands using the Cobra library. Each command creates a Docker client (or stack manager) and delegates work to the lower layers.
* **Docker abstraction (`internal/docker`)** – encapsulates Docker SDK calls (list, start, stop, restart, version, ping). Returns a simplified `Container` struct for the CLI.
* **Compose layer (`internal/compose`)** – discovers stacks under a fixed directory (`/opt/cloudhub`), validates the presence of `docker-compose.yml`, and runs `docker compose` commands (`up`, `down`, `restart`) in the stack’s directory.
* **Daemon layer (`internal/daemon`)** – runs a ticker‑based loop that calls `ListContainers(true)` on each tick, compares the current state with the previous snapshot, and prints human‑readable change messages.
* **Entry point (`cmd/main.go`)** – simply calls `cli.Execute()` to start the Cobra command tree.

---

## How the Code Works

1. **Program start** – `cmd/main.go` calls `cli.Execute()`. Cobra parses the command line and routes to the appropriate sub‑command.
2. **Sub‑command execution** – Each command (`ls`, `start`, `stop`, `restart`, `stack`, `daemon`) creates a Docker client via `docker.NewClient()`.  
   * For stack commands, `compose.GetStack()` resolves the stack directory and then invokes `docker compose` through `exec.Command`.
3. **Docker interactions** – The Docker client wraps the official SDK:
   * `ListContainers(all)` returns a slice of `docker.Container` (ID, Name, State, Status).  
   * `StartConatiner`, `StopContainer`, `RestartContainer` locate the container (by name, full ID, or prefix) and invoke the corresponding SDK call.
4. **Daemon mode** – `cloudhub daemon run --interval N`:
   * Builds a `daemon.Monitor` with the requested tick interval.  
   * Starts a goroutine that runs `monitor.Start()`.  
   * The monitor performs an initial container snapshot, then on each tick:
     * Calls `ListContainers(true)` to get the current state.
     * Detects new containers, disappeared containers, and state transitions (e.g., running → exited) and prints messages.
   * Graceful shutdown is handled via OS signals (`SIGINT`, `SIGTERM`).
5. **Output** – All commands write human‑readable tables or status messages to `stdout`. Errors are printed with a leading ❌ emoji for clarity.

---

## Implemented Features

| Feature | CLI Command | Description |
|---------|-------------|-------------|
| List containers (all) | `cloudhub ls` | Shows name, state, status, and short ID for every container. |
| List running containers | `cloudhub ls running` | Filters to containers whose state is `running`. |
| List stopped containers | `cloudhub ls stopped` | Filters to containers whose state is `exited`. |
| List paused containers | `cloudhub ls paused` | Filters to containers whose state is `paused`. |
| Start a container | `cloudhub start <name|id>` | Starts a stopped container. |
| Stop a container | `cloudhub stop <name|id>` | Stops a running container (10 s timeout). |
| Restart a container | `cloudhub restart <name|id>` | Restarts a container (stop → start). |
| Show Docker version | `cloudhub version` | Prints Docker daemon version information. |
| List Docker‑Compose stacks | `cloudhub stack ls` | Scans `/opt/cloudhub` for directories containing `docker-compose.yml`. |
| Bring a stack up | `cloudhub stack up <stack>` | Executes `docker compose up -d` in the stack directory. |
| Bring a stack down | `cloudhub stack down <stack>` | Executes `docker compose down`. |
| Restart a stack | `cloudhub stack restart <stack>` | Runs `down` then `up -d`. |
| Daemon monitoring | `cloudhub daemon run --interval N` | Periodically checks container states and reports changes. |

---

## Installation

### Prerequisites
* Go 1.22+ (the repository uses Go modules)  
* Docker daemon reachable from the host (Docker CLI must be functional)  
* `docker compose` command available (used by the stack manager)

### Build & Run

```bash
# Clone the repository
git clone https://github.com/your-org/cloudhub.git
cd cloudhub

# Build the binary (Makefile target)
make build

# The binary will be placed at ./dist/cloudhub
./dist/cloudhub --help
```

You can also run directly via the Makefile:

```bash
make run ARGS="ls"
```

### Run from source (without Make)

```bash
go run ./cmd/main.go <command>
```

---

## Configuration

| Parameter | Source | Default | Description |
|-----------|--------|---------|-------------|
| `--interval` | Flag on `cloudhub daemon run` | `10` seconds | Tick interval for the daemon monitor. |
| Stacks directory | Hard‑coded constant `defaultStacksDir` in `internal/compose/manager.go` | `/opt/cloudhub` | Directory where stack sub‑folders containing `docker-compose.yml` are searched. |
| Docker client | Reads Docker environment variables (`DOCKER_HOST`, `DOCKER_TLS_VERIFY`, etc.) via `client.NewClientWithOpts(client.FromEnv)` | – | Standard Docker SDK configuration. |

No additional environment variables or configuration files are required.

---

## CLI Commands

```
cloudhub
├─ daemon
│   └─ run               # Run the monitoring daemon (default interval 10s)
├─ ls [type]             # List containers (all, running, stopped, paused)
│   ├─ running
│   ├─ stopped
│   └─ paused
├─ start <container>     # Start a stopped container
├─ stop <container>      # Stop a running container
├─ restart <container>   # Restart a container
├─ version               # Show Docker daemon version information
└─ stack
    ├─ ls                # List discovered stacks
    ├─ up <stack>        # Bring a stack up (docker compose up -d)
    ├─ down <stack>      # Bring a stack down
    └─ restart <stack>   # Restart a stack (down + up)
```

All commands provide helpful usage text (`--help`) via Cobra.

---

## Development Guide

1. **Fork & clone** the repository.  
2. **Create a feature branch**: `git checkout -b feature/your‑feature`.  
3. **Run tests** (if any) and linting: the project currently has no test suite, but you can run `go vet ./...` and `staticcheck ./...` for static analysis.  
4. **Make changes** – follow the existing package layout. New CLI commands should be added under `internal/cli` using Cobra. New Docker interactions belong in `internal/docker`.  
5. **Run the binary** with `make run ARGS="your command"` to verify behavior.  
6. **Commit & push** and open a Pull Request against the `main` branch.  

### Building locally

```bash
go build -o cloudhub ./cmd/main.go
```

### Running the daemon in development

```bash
./cloudhub daemon run --interval 5
```

---

## Roadmap

| Milestone | Description |
|-----------|-------------|
| **Improved error handling & logging** | Replace `fmt.Printf` with a structured logger (e.g., `logrus` or `zap`) and propagate errors up the command chain. |
| **Configurable stacks directory** | Expose the stacks base path via a flag or environment variable instead of the hard‑coded `/opt/cloudhub`. |
| **Extended Compose support** | Add commands for `logs`, `ps`, and custom compose arguments. |
| **Unit & integration tests** | Add test coverage for Docker client wrapper (using a mock Docker server) and for stack discovery logic. |
| **Cross‑platform support** | Ensure the daemon works on Windows (currently uses Unix signals). |
| **Packaging** | Provide Homebrew, Snap, or Debian packages for easier installation. |

Contributions that address any of the above items are welcome.

---

## License

This project is licensed under the **MIT License**. See the `LICENSE` file for details.