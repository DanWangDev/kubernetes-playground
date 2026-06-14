# Module 13: Observability

## What Is Observability?

Observability means understanding what's happening inside your system from its external outputs. In Kubernetes, this breaks down into:

- **Logs**: what your application writes to stdout/stderr
- **Metrics**: numerical measurements over time (CPU, memory, request rates)
- **Events**: Kubernetes cluster-level happenings (pod scheduled, image pulled, probe failed)
- **Traces**: distributed request flows across services

## Core Concepts

### 1. kubectl logs — First Line of Defense

```
kubectl logs <pod>                    # Current container logs
kubectl logs <pod> -c <container>     # Specific container in multi-container pod
kubectl logs <pod> --tail=50          # Last 50 lines
kubectl logs <pod> --since=5m         # Last 5 minutes
kubectl logs <pod> --timestamps       # Show timestamps
kubectl logs <pod> --previous         # Logs from previous (crashed) container
kubectl logs -l app=nginx             # All pods matching a label
```

### 2. kubectl get events — Cluster Activity Log

```
kubectl get events -n playground --sort-by='.lastTimestamp'
```

Events show: scheduling decisions, image pulls, probe results, volume mounts, scaling actions. They're the first debugging stop when something isn't working.

### 3. kubectl describe — Deep Inspection

```
kubectl describe pod <name>
```

Shows the full pod state: conditions, container statuses, recent events, resource usage. More detailed than `kubectl get`.

### 4. Prometheus + Grafana

The standard Kubernetes monitoring stack:

- **Prometheus**: scrapes metrics from pods and stores time-series data
- **Grafana**: dashboards and alerting on top of Prometheus data
- **kube-prometheus-stack**: Helm chart that installs both + Alertmanager

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install monitoring prometheus-community/kube-prometheus-stack
```

### 5. Structured Logging

Write JSON to stdout. Kubernetes doesn't care about the format, but log aggregation tools (Elasticsearch, Loki, Datadog) parse structured logs much better:

```
{"level":"info","ts":"2026-06-14T10:00:00Z","msg":"request processed","method":"GET","path":"/api","status":200,"duration_ms":45}
```

### 6. Sidecar Log Pattern

When an application writes logs to a file (not stdout), use a sidecar container to tail the file to stdout:

```yaml
containers:
- name: app
  image: myapp
  volumeMounts:
  - name: logs
    mountPath: /var/log
- name: log-sidecar
  image: busybox
  command: ["tail", "-f", "/var/log/app.log"]
  volumeMounts:
  - name: logs
    mountPath: /var/log
volumes:
- name: logs
  emptyDir: {}
```

## What You'll Practice

1. Using kubectl logs with various flags
2. Viewing cluster events as a debugging tool
3. Deep-inspecting pods with kubectl describe
4. Understanding the Prometheus architecture
5. Deploying a log sidecar pattern

## Key Gotchas

- **Events are NOT persistent** — they're deleted after 1 hour by default. Don't rely on them for long-term audit.
- **`--previous` is invaluable** — when a container crashes and restarts, the current logs are from the new container. Use `--previous` to see why it crashed.
- **Prometheus is resource-heavy** — for learning, use a small instance. The full kube-prometheus-stack needs significant resources.
- **Sidecar logs are NOT kubectl logs** — `kubectl logs <pod>` shows stdout from the main container (or with `-c`). Logs that go through a sidecar to a file need to be read from the file.
