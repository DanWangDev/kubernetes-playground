# Module 09: Probes & Resource Management

## What Are Probes?

**Probes** are health checks that Kubernetes runs against your containers. They determine whether a container is alive, ready to serve traffic, and has started successfully.

## Core Concepts

### 1. Three Probe Types

| Probe | Question | Failure Action |
|-------|----------|----------------|
| **Liveness** | Is the app alive? (not deadlocked) | Restart the container |
| **Readiness** | Can the app serve requests? | Remove from Service endpoints |
| **Startup** | Has the app finished starting? | Disables liveness/readiness until success |

### 2. Probe Mechanisms

```yaml
# HTTP probe (most common)
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5

# Command probe
readinessProbe:
  exec:
    command: ["cat", "/tmp/healthy"]
  initialDelaySeconds: 5
  periodSeconds: 5

# TCP probe
startupProbe:
  tcpSocket:
    port: 8080
  failureThreshold: 30
  periodSeconds: 10
```

### 3. Probe Timing

| Parameter | Meaning | Default |
|-----------|---------|---------|
| `initialDelaySeconds` | Wait before first probe | 0 |
| `periodSeconds` | How often to probe | 10 |
| `timeoutSeconds` | Probe timeout | 1 |
| `successThreshold` | Consecutive successes to pass | 1 |
| `failureThreshold` | Consecutive failures to fail | 3 |

### 4. Resource Requests and Limits

```yaml
resources:
  requests:        # Minimum guaranteed (scheduler uses this)
    cpu: 100m      # 0.1 CPU cores
    memory: 128Mi
  limits:          # Maximum allowed (container throttled/killed above this)
    cpu: 500m
    memory: 256Mi
```

### 5. QoS Classes

| Class | Condition | Eviction Priority |
|-------|-----------|-------------------|
| **Guaranteed** | requests = limits (both CPU and memory) | Last (safest) |
| **Burstable** | requests < limits, or only one set | Middle |
| **BestEffort** | No requests or limits set | First to be evicted |

Check a pod's QoS class: `kubectl get pod <name> -o jsonpath='{.status.qosClass}'`

### 6. OOMKilled

When a container exceeds its memory limit, it's killed with OOMKilled. The pod restarts (if restartPolicy allows). CPU is throttled, not killed.

```
kubectl describe pod <name> | grep OOMKilled
```

## What You'll Practice

1. Creating a pod with a readiness probe
2. Observing how a failing liveness probe restarts a container
3. Using a startup probe for slow-starting apps
4. Setting resource requests and limits
5. Understanding QoS classes

## Key Gotchas

- **No liveness probe is often fine** — if your app crashes, the process exits and Kubernetes notices. Liveness probes are for deadlock detection.
- **Liveness probe failure = restart** — set conservative thresholds. A slow app + aggressive liveness probe = restart loop.
- **Readiness probe failure ≠ restart** — the pod stays running but receives no traffic.
- **CPU is compressible** (throttled), **memory is not** (OOMKilled).
- **Without requests**, the scheduler doesn't know how much your pod needs — it could be packed too tightly.
