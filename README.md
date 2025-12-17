# Simple Task Runner

A minimal "GitHub Actions" clone using NATS JetStream. Submit bash scripts as tasks, execute them on ephemeral runners with controlled concurrency, and stream logs in real-time or view them historically.

📖 [Why I built this](WHY.md)

## How it works

Two JetStream streams:

- **TASKS** (`tasks.*`) — stores bash scripts to execute
- **LOGS** (`logs.*`) — stores execution output, line by line

The runner (`main.go`) does one thing:

1. Waits for next task from TASKS stream
2. Saves script to `/tmp/<id>.sh`, executes with bash
3. Pipes stdout/stderr to LOGS stream in real-time (stderr prefixed with `ERROR::`)
4. Publishes `EXIT:<code>` when done, exits

Runners are **ephemeral** — each container processes one task and then exits. Docker automatically restarts it via `restart: always` policy, providing a fresh environment for the next task.

```
nats pub tasks.<id> "script"  →  TASKS stream  →  runner  →  LOGS stream  →  nats sub logs.<id>
```

## Quick Start

```bash
docker compose up -d
```

## Submit a Task

_Install nats cli tools https://github.com/nats-io/natscli?tab=readme-ov-file#installation_

```bash
# Simple task
nats pub tasks.job-001 "echo hello world"

# Multi-line task with sleep
nats pub tasks.job-002 "echo start; sleep 2; echo middle; sleep 2; echo done"

# Task with stderr output
nats pub tasks.job-003 "echo stdout line; echo stderr line >&2; echo another stdout"

# Exec file on runner
cat ./example.sh | nats pub tasks.job-002
```

The task ID is the part after `tasks.` (e.g., `job-001`).

## View Task Logs

### Real-time (streaming)

```bash
# Per task
nats sub 'logs.job-001' --raw
# Or watch all logs in real-time
nats sub 'logs.*' --raw
```

### Historical

```bash
# View logs for a specific task
nats stream view LOGS --subject "logs.job-001"
# Or view all logs
nats stream view LOGS
```

## Configuration

### Worker Count

To change the number of parallel workers, modify the `replicas` value in `docker-compose.yml`:

```yaml
runner:
  deploy:
    replicas: 3 # Change this number
```

Then restart the services:

```bash
docker compose up -d
```
