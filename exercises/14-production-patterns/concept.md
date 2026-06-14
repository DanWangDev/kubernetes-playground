# Module 14: Production Patterns

## What Are Production Patterns?

These are scheduling, availability, and resilience patterns used in production Kubernetes clusters. They control WHERE pods run, HOW they're spread, and WHAT happens during disruptions.

## Core Concepts

### 1. Pod Affinity — Co-locate Related Workloads

```yaml
affinity:
  podAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          app: cache
      topologyKey: kubernetes.io/hostname
```

This ensures the pod runs on the same node as pods with `app: cache`.

### 2. Pod Anti-Affinity — Spread Pods Apart

```yaml
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          app: web
      topologyKey: kubernetes.io/hostname
```

This ensures pods are spread across different nodes (or zones).

**Required** = pod won't schedule if rule can't be satisfied. **Preferred** = scheduler tries but will schedule if not possible.

### 3. Node Selector / Node Affinity

Simplest form:
```yaml
nodeSelector:
  disktype: ssd
```

More expressive:
```yaml
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: topology.kubernetes.io/zone
          operator: In
          values: [us-east-1a]
```

### 4. Taints and Tolerations

Taints REPEL pods from nodes. Tolerations ALLOW pods onto tainted nodes.

```bash
# Taint a node (dedicated for GPU workloads)
kubectl taint nodes worker-1 gpu=true:NoSchedule

# Pod toleration
tolerations:
- key: gpu
  operator: Equal
  value: "true"
  effect: NoSchedule
```

Taint effects:
- **NoSchedule**: don't schedule new pods here
- **PreferNoSchedule**: try not to schedule here
- **NoExecute**: evict existing pods that don't tolerate

### 5. Pod Disruption Budget (PDB)

Limits voluntary disruptions (node drains, cluster autoscaler):

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: app-pdb
spec:
  minAvailable: 2          # OR maxUnavailable: 1
  selector:
    matchLabels:
      app: critical-app
```

With `minAvailable: 2`, at most (replicas - 2) pods can be disrupted at once.

### 6. Topology Spread Constraints

Evenly distribute pods across topology domains:

```yaml
topologySpreadConstraints:
- maxSkew: 1
  topologyKey: topology.kubernetes.io/zone
  whenUnsatisfiable: DoNotSchedule
  labelSelector:
    matchLabels:
      app: web
```

This ensures pod counts across zones differ by at most 1.

### 7. PriorityClass

Higher-priority pods can preempt lower-priority ones:

```yaml
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: high-priority
value: 1000
globalDefault: false
```

## What You'll Practice

1. Pod affinity and anti-affinity rules
2. Node selectors for hardware-specific workloads
3. Tainting a node and adding pod tolerations
4. Creating a Pod Disruption Budget
5. Applying topology spread constraints
6. Understanding PriorityClass

## Key Gotchas

- **Anti-affinity can prevent scheduling** — if requiredDuringScheduling and you have more replicas than nodes, pods stay Pending.
- **PDBs only cover VOLUNTARY disruptions** — node crashes or pod evictions are not covered.
- **Taints + tolerations ≠ authentication** — they're scheduling hints, not security boundaries.
- **Topology spread needs enough nodes** — maxSkew enforcement requires enough topology domains.
- **Multi-node clusters** are needed for node-level affinity, anti-affinity, and topology spread to be meaningful. Kind single-node demonstrates the pattern but won't show real distribution.
