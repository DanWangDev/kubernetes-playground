# Kubernetes Playground

Hands-on Kubernetes learning environment with guided exercises, IaC manifests,
and progressive modules that take you from first pod to production patterns.

## Quick Start

### Prerequisites

- **Go 1.22+** ([download](https://go.dev/dl/))
- **Docker** ([Docker Desktop](https://www.docker.com/products/docker-desktop/) or [Rancher Desktop](https://rancherdesktop.io/))
- **kubectl** ([install guide](https://kubernetes.io/docs/tasks/tools/))

### Setup

```bash
# Clone the repo
git clone https://github.com/danwa/kubernetes-playground.git
cd kubernetes-playground

# Create a local Kubernetes cluster (30 seconds)
make cluster-create

# Verify the cluster is ready
make cluster-status
```

### Run Your First Exercise

```bash
# Run an exercise (non-interactive)
make exercise-01

# Run with step-by-step explanations (interactive)
go run ./exercises/01-first-pods/ --step
```

## Learning Path

| Module | Topic | What You'll Learn |
|--------|-------|-------------------|
| 01 | First Pods | Pod lifecycle, kubectl basics, labels, multi-container pods |
| 02 | Deployments | Rolling updates, rollbacks, ReplicaSets, declarative management |
| 03 | Services | ClusterIP, NodePort, DNS, port-forward, headless services |
| 04 | ConfigMaps & Secrets | Env injection, volume mounts, immutable configs, secret types |
| 05 | Ingress | L7 routing, NGINX controller, path/host routing, TLS |
| 06 | Storage | PV, PVC, StorageClass, static vs dynamic provisioning |
| 07 | StatefulSets | Stable identity, ordered deploy/scale, per-pod storage |
| 08 | Jobs & CronJobs | Completions, parallelism, schedules, TTL cleanup |
| 09 | Probes & Resources | Liveness/readiness/startup probes, QoS classes, OOMKilled |
| 10 | Autoscaling (HPA) | Metrics Server, CPU/memory targets, scale policies |
| 11 | RBAC & Security | ServiceAccounts, Roles, NetworkPolicy, SecurityContext |
| 12 | Helm | Chart structure, templating, install/upgrade/rollback |
| 13 | Observability | Prometheus, Grafana, ServiceMonitors, log sidecars |
| 14 | Production Patterns | Affinity, anti-affinity, PDBs, topology spread, taints |

## Project Structure

```
kubernetes-playground/
├── cmd/                  # CLI tools (cluster create/delete/status)
├── pkg/                  # Shared Go packages
│   ├── kubectl/          # kubectl CLI wrapper
│   ├── logger/           # Structured exercise output
│   ├── prompt/           # Interactive step mode
│   ├── validate/         # Assertions and polling helpers
│   └── env/              # Configuration
├── exercises/            # One directory per module
│   ├── 01-first-pods/    # concept.md + manifests/ + main.go
│   ├── 02-deployments/
│   └── ...
├── test/                 # Go tests (mirrors exercises/ structure)
└── Makefile              # Task runner
```

## Available Commands

| Command | Description |
|---------|-------------|
| `make cluster-create` | Create a kind Kubernetes cluster |
| `make cluster-delete` | Destroy the kind cluster |
| `make cluster-reset` | Delete + recreate (fresh start) |
| `make cluster-status` | Show cluster info and nodes |
| `make exercise-NN` | Run a module non-interactively |
| `go run ./exercises/NN-xxx/ --step` | Run step-by-step |
| `make test` | Run all tests |
| `make lint` | Run go vet |
| `make build` | Build all packages |

## Design Decisions

- **Go** — native language of Kubernetes (kubectl, client-go, Helm all Go)
- **kind** — Kubernetes in Docker, no VM overhead, fast CI
- **Raw kubectl** — exercises shell out to real kubectl (teach the CLI, not an SDK)
- **Raw YAML → Helm** — modules 1-11 use raw manifests; module 12 introduces Helm
- **Sequential modules** — each builds on previous concepts
- **Bilingual** — concept.md (English) + concept_cn.md (中文) per module

## License

MIT
