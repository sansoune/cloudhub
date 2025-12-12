# Cloudhub – Docker Homelab Monitor & Management CLI

**Cloudhub** is a lightweight, opinionated command‑line tool for monitoring and managing Docker containers and Docker‑Compose stacks in a homelab environment. It provides a set of intuitive commands for listing containers, controlling their lifecycle, handling stacks, and running a background daemon that continuously watches container state changes.

---

## Table of Contents
1. [Project Overview](#project-overview)  
2. [Core Architecture](#core-architecture)  
3. [How the Code Works (Main Flow)](#how-the-code-works-main-flow)  
4. [Implemented Features](#implemented-features)  
5. [Installation & Build](#installation--build)  
6. [Configuration](#configuration)  
7. [CLI Usage (Endpoints)](#cli-usage-endpoints)  
8. [Development Guide](#development-guide)  
9. [Roadmap & Open TODOs](#roadmap--open-todos)  
10. [License](#license)  

---

## Project Overview

Cloudhub is a **CLI‑only** application (no HTTP server) that interacts with the local Docker daemon via the official Docker Go SDK. Its primary responsibilities are:

* **Container management** – list, start, stop, restart Docker containers.
* **Stack management** – discover Docker‑Compose stacks under a configurable directory (`/opt/cloudhub` by default) and run `docker compose up/down/restart` on them.
* **Daemon monitoring** – run a background process that periodically polls Docker for container state, reporting new containers, state transitions, and disappeared containers.

The tool is intended for homelab operators who want a single binary to keep an eye on their Docker workloads without the overhead of a full‑blown monitoring stack.

---

## Core Architecture

```
cloudhub/
├─ cmd/
│   └─ main.go                → entry point, calls cli.Execute()
├─ internal/
│   ├─ cli/                   → Cobra‑based command hierarchy
│   │   ├─ root.go            → root command definition
│   │   ├─ daemon.go          → `cloudhub daemon run` implementation
│   │   ├─ ls.go              → `cloudhub ls` (all/running/stopped/paused)
│   │   ├─ start.go           → `cloudhub start <container>`
│   │   ├─ stop.go            → `cloudhub stop <container>`
│   │   ├─ restart.go         → `cloudhub restart <container>`
│   │   ├─ version.go         → `cloudhub version` (Docker daemon version)
│   │   └─ stack.go           → `cloudhub stack` sub‑commands (ls, up, down, restart)
│   ├─ compose/               → Stack discovery & Docker‑Compose wrapper
│   │   └─ manager.go         → FindStacks, GetStack, and compose command runner
│   ├─ daemon/                → Monitoring daemon implementation
│   │   └─ monitor.go         → Periodic container state checks, change detection
│   └─ docker/                → Thin wrapper around Docker SDK
│       └─ client.go          → Client struct, container CRUD helpers, version, ping
└─ go.mod                     → module definition
```

### Package Responsibilities

| Package | Responsibility |
|---------|-----------------|
| **cmd** | Minimal bootstrap – calls the CLI executor. |
| **internal/cli** | Defines the Cobra command tree, parses flags, and orchestrates calls to other packages. |
| **internal/docker** | Encapsulates Docker SDK usage (client creation, container CRUD, version retrieval). |
| **internal/compose** | Discovers stacks in `/opt/cloudhub`, validates presence of `docker-compose.yml`, and runs `docker compose` commands in the stack’s directory. |
| **internal/daemon** | Implements a ticker‑based monitor that polls Docker, keeps a snapshot of previous container states, and prints human‑readable alerts on changes. |

---

## How the Code Works (Main Flow)

1. **Program start** – `cmd/main.go` calls `cli.Execute()`.
2. **Cobra initialization** – `internal/cli/root.go` creates the root command (`cloudhub`). All sub‑commands (`ls`, `start`, `stop`, `restart`, `stack`, `daemon`, `version`) are attached in their respective `init()` functions.
3. **Command execution** – When a user runs a sub‑command:
   * The command’s `Run`/`RunE` handler creates a Docker client via `docker.NewClient()`.
   * The handler calls the appropriate method on the client (`ListContainers`, `StartConatiner`, `StopContainer`, `RestartContainer`, `GetVersion`) or on a `compose.Stack` (`Up`, `Down`, `Restart`).
4. **Daemon mode** – `cloudhub daemon run`:
   * Parses `--interval` flag (default **10 s**).
   * Instantiates a `daemon.Monitor` with the interval.
   * Starts the monitor in a goroutine; the main goroutine waits for OS signals (`SIGINT`, `SIGTERM`) or monitor errors.
   * The monitor periodically (`time.Ticker`) calls `checkContainers()`, which:
     * Retrieves the full container list (`ListContainers(true)`).
     * Builds a map of `containerName → state`.
     * Calls `detectChanges()` to compare with the previous snapshot and prints:
       * **NEW** – container appeared.
       * **DISAPPEARED** – container vanished.
       * **ALERT / RECOVERED** – transitions between `running` and `exited`.
       * Generic **CHANGE** for any other state shift.
5. **Graceful shutdown** – On signal receipt, `monitor.Stop()` closes its internal channel, causing the monitor loop to exit cleanly.

---

## Implemented Features

| Feature | CLI Command | Description |
|---------|-------------|-------------|
| **List containers** | `cloudhub ls` (default) | Shows all containers. |
| | `cloudhub ls running` | Shows only containers with state `running`. |
| | `cloudhub ls stopped` | Shows containers with state `exited`. |
| | `cloudhub ls paused` | Shows containers with state `paused`. |
| **Start container** | `cloudhub start <name|id>` | Starts a stopped container. |
| **Stop container** | `cloudhub stop <name|id>` | Stops a running container (10 s timeout). |
| **Restart container** | `cloudhub restart <name|id>` | Restarts a container (10 s timeout). |
| **Docker version** | `cloudhub version` | Prints Docker daemon version details. |
| **Stack discovery** | `cloudhub stack ls` | Lists stacks (directories containing `docker-compose.yml`) under `/opt/cloudhub`. |
| **Stack up** | `cloudhub stack up <stack>` | Executes `docker compose up -d` in the stack directory. |
| **Stack down** | `cloudhub stack down <stack>` | Executes `docker compose down`. |
| **Stack restart** | `cloudhub stack restart <stack>` | Runs `down` then `up -d`. |
| **Daemon monitoring** | `cloudhub daemon run --interval <seconds>` | Starts a background monitor that reports container state changes. |
| **Graceful signal handling** | (daemon) | Handles `Ctrl+C` / `SIGTERM` to stop monitoring cleanly. |

---

## Installation & Build

### Prerequisites
* Go **1.22** (or later) – the repository uses Go modules.
* Docker daemon reachable from the host (default Unix socket or configured via `DOCKER_HOST`).

### Steps

```bash
# 1. Clone the repository
git clone https://github.com/your-org/cloudhub.git
cd cloudhub

# 2. Tidy modules (optional, ensures dependencies are fetched)
go mod tidy

# 3. Build the binary
go build -o cloudhub ./cmd/main.go

# 4. Verify installation
./cloudhub version
```

You can also install directly with `go install`:

```bash
go install github.com/your-org/cloudhub/cmd/cloudhub@latest
```

The binary will be placed in `$GOPATH/bin`.

---

## Configuration

| Source | Variable / Flag | Default | Effect |
|--------|----------------|---------|--------|
| **Flag** | `--interval` (daemon) | `10` seconds | Tick interval for the monitoring daemon. |
| **Env** | `DOCKER_HOST` | Docker SDK default (`unix:///var/run/docker.sock`) | Docker daemon endpoint. |
| **Env** | `DOCKER_API_VERSION` | Docker SDK auto‑detect | API version used by the Docker client. |
| **Constant** | `defaultStacksDir` (internal/compose) | `/opt/cloudhub` | Directory where stack sub‑folders are searched. Change by editing `manager.go` or forking the repo. |

No additional configuration files are required.

---

## CLI Usage (Endpoints)

Below is a concise cheat‑sheet of the available commands.

```
cloudhub [command]

Root Commands:
  daemon      Daemon Management
  ls          List containers (all/running/stopped/paused)
  start       Start a stopped container
  stop        Stop a running container
  restart     Restart a container
  stack       Manage Docker‑Compose stacks
  version     Show Docker daemon version

Flags:
  -h, --help   help for cloudhub
```

### Detailed Sub‑Commands

| Command | Syntax | Example |
|---------|--------|---------|
| **Daemon** | `cloudhub daemon run [--interval N]` | `cloudhub daemon run --interval 15` |
| **List all** | `cloudhub ls` | `cloudhub ls` |
| **List running** | `cloudhub ls running` | `cloudhub ls running` |
| **List stopped** | `cloudhub ls stopped` | `cloudhub ls stopped` |
| **List paused** | `cloudhub ls paused` | `cloudhub ls paused` |
| **Start** | `cloudhub start <container>` | `cloudhub start my‑web` |
| **Stop** | `cloudhub stop <container>` | `cloudhub stop my‑web` |
| **Restart** | `cloudhub restart <container>` | `cloudhub restart my‑web` |
| **Version** | `cloudhub version` | `cloudhub version` |
| **Stack list** | `cloudhub stack ls` | `cloudhub stack ls` |
| **Stack up** | `cloudhub stack up <stack>` | `cloudhub stack up home‑media` |
| **Stack down** | `cloudhub stack down <stack>` | `cloudhub stack down home‑media` |
| **Stack restart** | `cloudhub stack restart <stack>` | `cloudhub stack restart home‑media` |

All commands output human‑readable tables or status messages. Errors are prefixed with `❌` and successes with `✅`.

---

## Development Guide

### Repository Layout

```
cmd/                → binary entry point
internal/
  cli/              → Cobra commands, flag handling
  compose/          → Stack discovery & docker‑compose wrapper
  daemon/           → Monitoring daemon implementation
  docker/           → Docker SDK wrapper (client, container helpers)
go.mod, go.sum      → module definition
```

### Contributing Steps

1. **Fork** the repository.
2. **Create a feature branch**: `git checkout -b feat/your-feature`.
3. **Write code** adhering to Go conventions (`go fmt`, `golint` optional).
4. **Add tests** (currently none – consider adding unit tests for `docker` wrapper and `daemon` logic).
5. **Run `go vet` / `staticcheck`** to catch issues.
6. **Commit** with clear messages.
7. **Open a Pull Request** against the `main` branch.

### Building & Testing Locally

```bash
# Run the binary directly
go run ./cmd/main.go version

# Run unit tests (once they exist)
go test ./...
```

### Extending the Tool

* **Add new sub‑commands** – create a new file under `internal/cli`, define a `cobra.Command`, and register it in `init()`.
* **Support additional Docker‑Compose actions** – extend `compose.Stack` with new wrapper methods.
* **Improve daemon output** – replace `fmt.Printf` with a structured logger (e.g., `logrus` or `zap`) for better machine‑readability.

---

## Roadmap & Open TODOs

| Milestone | Description | Status |
|-----------|-------------|--------|
| **Enhanced logging** | Replace ad‑hoc `fmt` prints with a configurable logger (levels, JSON output). | Planned |
| **Configurable stacks directory** | Allow the stacks base path to be set via env var or CLI flag instead of hard‑coded `/opt/cloudhub`. | Planned |
| **Unit & integration tests** | Add test coverage for Docker client wrapper, daemon change detection, and stack discovery. | Needed |
| **Additional Docker actions** | Implement `docker exec`, `logs`, `inspect` sub‑commands. | Idea |
| **Health‑check endpoint** | Expose a minimal HTTP endpoint for external monitoring tools (e.g., Prometheus). | Future |
| **Cross‑platform support** | Verify behavior on Windows (path handling, default socket). | To verify |
| **Daemon output formats** | Provide `--output json` flag for machine‑readable alerts. | Planned |

Contributions that address any of the above items are welcome.

---

## License

```
MIT License

Copyright (c) <year> <author>

Permission is hereby granted, free of charge, to any person obtaining a copy
...
```

The project is released under the **MIT License**. See the `LICENSE` file for the full text.