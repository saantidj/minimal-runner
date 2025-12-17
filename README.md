# Simple Task Runner

A minimal "GitHub Actions" clone using NATS JetStream. Submit bash scripts as tasks, execute them on runners, and stream logs in real-time.

## Prerequisites

- Docker (for NATS server)
- Go 1.21+
- NATS CLI (`brew install nats-io/nats-tools/nats`)

## Start NATS Server

```bash
docker compose up
```

## Run the Task Runner

```bash
go run main.go
```

The runner will wait for a task (up to 1 month).

## Submit a Task

```bash
# Simple task
nats pub tasks.job-001 "echo hello world"

# Multi-line task with sleep
nats pub tasks.job-002 "echo start; sleep 2; echo middle; sleep 2; echo done"

# Task with stderr output
nats pub tasks.job-003 "echo stdout line; echo stderr line >&2; echo another stdout"
```

The task ID is the part after `tasks.` (e.g., `job-001`).

## View Task Logs

### Real-time (streaming)

```bash
# Watch all logs in real-time
nats sub 'logs.*' --raw
```

### Historical (from stream)

```bash
# View logs for a specific task
nats stream view LOGS --subject "logs.job-001"

# View all logs
nats stream view LOGS
```

### Log Format

- Regular output lines appear as-is
- Stderr lines are prefixed with `ERROR::`
- When task completes, `EXIT:<code>` is published (e.g., `EXIT:0` for success)

### Example Output

```
hello world
ERROR::some error message
EXIT:0
```

## Architecture

```
┌─────────────┐     tasks.<id>     ┌─────────────┐    logs.<id>      ┌─────────────┐
│  NATS CLI   │ ─────────────────► │   RUNNER    │ ═══════════════►  │  NATS CLI   │
│ (publish)   │                    │   (Go app)  │   (line by line)  │ (view)      │
└─────────────┘                    └─────────────┘                    └─────────────┘
                   TASKS stream                       LOGS stream
```

## Notes

- Scripts are saved to `/tmp/<task-id>.sh` (not deleted, for debugging)
- Execution timeout: 5 minutes
- Logs are streamed line-by-line in real-time as the script executes
# minimal-runner
