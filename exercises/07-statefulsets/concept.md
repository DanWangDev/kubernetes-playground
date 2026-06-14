# Module 07: StatefulSets

## What Is a StatefulSet?

A **StatefulSet** manages stateful applications — databases, message queues, distributed caches. Unlike Deployments (where pods are interchangeable), StatefulSet pods have **sticky identities**:

- **Stable network identity**: pod-0, pod-1, pod-2 (not random suffixes)
- **Stable storage**: each pod gets its own PVC that survives rescheduling
- **Ordered deployment**: 0 starts first, then 1, then 2 (sequential)
- **Ordered teardown**: N-1 stops first, then N-2, etc. (reverse)

```
StatefulSet → pod-0 (PVC-0) | pod-1 (PVC-1) | pod-2 (PVC-2)
```

## Core Concepts

### 1. Stable Network Identity

Each StatefulSet pod gets a predictable DNS name:

```
<pod-name>.<headless-service>.<namespace>.svc.cluster.local
```

For a StatefulSet named `redis` with a headless service `redis-svc`:
- `redis-0.redis-svc.playground.svc.cluster.local`
- `redis-1.redis-svc.playground.svc.cluster.local`

This requires a **headless Service** (`clusterIP: None`).

### 2. Headless Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: sts-svc
spec:
  clusterIP: None          # Headless!
  selector:
    app: sts-demo
  ports:
  - port: 80
```

With `clusterIP: None`, DNS returns individual Pod IPs instead of a single Service IP. This is required for StatefulSets.

### 3. Ordered Deployment and Scaling

StatefulSets deploy and scale in strict order:

- **Scale up**: pod-0 → pod-1 → pod-2 (each must be Ready before next starts)
- **Scale down**: pod-2 → pod-1 → pod-0 (reverse order, each must terminate before next)
- **Rolling update**: pod-N-1 → pod-N-2 → ... → pod-0 (reverse order, one at a time)

This guarantees that your application's quorum or leader election logic works correctly during changes.

### 4. PVC Templates

```yaml
spec:
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 100Mi
```

Each pod automatically gets its own PVC named `data-sts-demo-0`, `data-sts-demo-1`, etc. These PVCs are NOT deleted when the pod is deleted or the StatefulSet is scaled down — storage outlives pods.

### 5. StatefulSet vs Deployment

| | Deployment | StatefulSet |
|---|---|---|
| **Pod identity** | Random (nginx-7d5f8c9b6-abc12) | Predictable (redis-0, redis-1) |
| **Pod naming** | `<deploy>-<rs-hash>-<pod-hash>` | `<sts>-<ordinal>` |
| **Startup order** | Parallel | Sequential (0, 1, 2...) |
| **Storage** | Shared PVC (if any) | Per-pod PVC |
| **Scaling down** | Random pod deleted | Highest ordinal first |
| **Use case** | Stateless web servers | Databases, queues, caches |

## What You'll Practice

1. Creating a headless Service for StatefulSet DNS
2. Deploying a StatefulSet with sequential pod creation
3. Observing stable pod names (nginx-sts-0, nginx-sts-1, nginx-sts-2)
4. Scaling down and watching reverse-order termination

## Key Gotchas

- **Headless Service is required** — a StatefulSet won't work without a headless Service for network identity.
- **PVCs survive pod deletion** — scaling down doesn't delete PVCs. You must clean them up manually.
- **Ordered operations are slow** — each pod must be Ready before the next starts. For large clusters, this takes time.
- **Rolling updates are reverse-order** — pod N-1 updates first, then N-2. This protects the "newest" pod.
- **`podManagementPolicy: Parallel`** bypasses ordering — all pods start/stop simultaneously. Use for fast scaling when order doesn't matter.
