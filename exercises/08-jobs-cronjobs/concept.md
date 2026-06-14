# Module 08: Jobs & CronJobs

## What Are Jobs?

A **Job** creates one or more Pods and ensures a specified number successfully complete. Unlike Deployments (which keep pods running forever), Jobs are for finite tasks: batch processing, database migrations, report generation.

A **CronJob** creates Jobs on a schedule — like cron, but for Kubernetes.

## Core Concepts

### 1. Job Patterns

| Pattern | completions | parallelism | Use Case |
|---------|-------------|-------------|----------|
| **Single** | 1 | 1 | One-shot task (migration, backup) |
| **Fixed count** | N | M | Process N items with M workers |
| **Work queue** | — (unset) | M | Process items from a queue until empty |

### 2. Job YAML

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: hello-job
spec:
  completions: 1
  parallelism: 1
  backoffLimit: 4
  template:
    spec:
      containers:
      - name: hello
        image: busybox
        command: ["echo", "Hello, Kubernetes!"]
      restartPolicy: Never
```

Key fields:
- **completions**: how many successful pod completions needed (default: 1)
- **parallelism**: max pods running concurrently (default: 1)
- **backoffLimit**: retries before marking Failed (default: 6)
- **restartPolicy**: must be `Never` or `OnFailure` (NOT `Always`)

### 3. Job Lifecycle

```
Job created → Pod(s) created → Pod completes → Job checks completions
  → If done: Job = Complete
  → If pod fails: retry (backoffLimit times)
  → If pod succeeds: count toward completions
```

After completion, Job pods stick around for log inspection. Set `ttlSecondsAfterFinished` to auto-cleanup.

### 4. CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: daily-report
spec:
  schedule: "0 6 * * *"     # Every day at 6am
  concurrencyPolicy: Forbid  # Skip if previous still running
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 1
  jobTemplate:
    spec: ...
```

Schedule uses standard cron syntax with 5 fields: `minute hour day month weekday`.

### 5. Concurrency Policy

| Policy | Behavior |
|--------|----------|
| **Allow** (default) | Multiple Jobs can run concurrently |
| **Forbid** | Skip new Job if previous still running |
| **Replace** | Kill previous Job and start new one |

### 6. Indexed Jobs

Set `completionMode: Indexed` to give each pod a unique index (0..completions-1) in the `JOB_COMPLETION_INDEX` env var. This lets each pod process its own shard.

## What You'll Practice

1. Creating a simple one-shot Job
2. Running a parallel Job with multiple workers
3. Seeing a failed Job and its backoff limit
4. Creating a CronJob and seeing it spawn Jobs
5. Configuring TTL for automatic cleanup

## Key Gotchas

- **restartPolicy must be Never or OnFailure** — `Always` (default) is invalid for Jobs.
- **Job pods stick around** — completed/failed Job pods remain for log inspection. Use TTL to clean them up.
- **CronJob history accumulates** — set `successfulJobsHistoryLimit` and `failedJobsHistoryLimit`.
- **CronJob timezone** — Kubernetes 1.24+ supports `timeZone` in CronJob spec.
- **Missed schedules** — if the controller is down when a CronJob should fire, it may or may not catch up (controlled by `startingDeadlineSeconds`).
