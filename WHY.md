Recently I was surprised how easy it is to build a minimal ephemeral task runner today. With a durable message stream and Docker restarting containers, you can get something useful in basically one page of AI-written code.

For message processing, I use **NATS** because it already has most of the tools I need. It’s small and easy.

For ephemeral runs, I use **Docker** with its ability to restart containers on exit, and to run multiple replicas for concurrent runners:

```yaml
services:
  runner:
    restart: always
    deploy:
      replicas: 3
```

In NATS I create/use two JetStream streams:

- **TASKS** (`tasks.*`) - stores bash scripts to execute
- **LOGS** (`logs.*`) - stores execution output, line by line

For creating and viewing tasks/jobs I just use the `nats` CLI.

The runner is a Docker container that:

1. Waits for the next task from the TASKS stream
2. Saves the script to `/tmp/<id>.sh` and executes it with bash
3. Pipes stdout/stderr to the LOGS stream in real time (stderr prefixed with `ERROR::`)
4. Exits, then Docker restarts it (`restart: always`)

As a user, you can execute shell scripts on the runner like:

```bash
cat ./example.sh | nats pub tasks.job-001
```

And see stdout/stderr logs either in real time or later:

```bash
# realtime
nats sub 'logs.job-001' --raw

# history
nats stream view LOGS --subject "logs.job-001"
```

The runner itself was written by AI in Go, because in Bash it would be a bit harder to read. It’s small and readable, you can see it in the repository.

Repo: https://github.com/istarkov/minimal-runner

**P.S.** This is just a minimal idea. You can add tags/metadata, retries, timeouts, scheduling, etc. You can also scale it across multiple machines (even across regions) - runners can live anywhere as long as they can connect to NATS.
