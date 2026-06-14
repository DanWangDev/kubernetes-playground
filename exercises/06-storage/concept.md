# Module 06: Storage (PV/PVC)

## What Are PersistentVolumes?

Kubernetes separates storage into two abstractions:

- **PersistentVolume (PV)** — a piece of storage in the cluster (admin-provisioned or dynamically created)
- **PersistentVolumeClaim (PVC)** — a request for storage by a user (like a Pod requests CPU/memory)

This separation means developers don't need to know about the underlying storage infrastructure. They just claim what they need, and Kubernetes binds their claim to available storage.

```
Admin creates: PV (10Gi, RWO, SSD)
User requests: PVC (5Gi, RWO) → binds to PV
Pod uses: PVC as a volume
```

## Core Concepts

### 1. PV and PVC Binding

```yaml
# PersistentVolume (cluster-scoped)
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv-hostpath
spec:
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: /data/pv
---
# PersistentVolumeClaim (namespaced)
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pvc-claim
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 500Mi
```

A PVC binds to the **smallest** PV that satisfies its requirements. If no PV matches, the PVC stays Pending.

### 2. Access Modes

| Mode | Abbreviation | Meaning |
|------|-------------|---------|
| **ReadWriteOnce** | RWO | One node can mount as read-write |
| **ReadOnlyMany** | ROM | Many nodes can mount as read-only |
| **ReadWriteMany** | RWX | Many nodes can mount as read-write |

Note: RWO doesn't mean "one pod" — multiple pods on the same node can use the same RWO volume.

### 3. Static vs Dynamic Provisioning

| | Static | Dynamic |
|---|---|---|
| **PV created by** | Admin manually | StorageClass automatically |
| **When** | Before PVC | When PVC is created |
| **Use case** | Pre-existing storage, NFS shares | Most workloads (automatic) |
| **Reclaim policy** | Manual cleanup | Automatic (Delete) or manual (Retain) |

### 4. StorageClass and Dynamic Provisioning

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ssd
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp3
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
```

When a PVC references this StorageClass, a PV is automatically created (if the provisioner supports it). kind's default StorageClass uses `rancher.io/local-path`.

### 5. Reclaim Policies

| Policy | Behavior |
|--------|----------|
| **Retain** | PV is NOT deleted when PVC is deleted. Must be manually cleaned up. |
| **Delete** (default) | PV and associated storage are deleted when PVC is deleted. |
| **Recycle** | Deprecated. Basic scrub (`rm -rf /`) and make available again. |

### 6. Volume Modes

- **Filesystem** (default): mounted as a directory
- **Block**: mounted as a raw block device (no filesystem)

### 7. hostPath on kind

Since kind runs Kubernetes in Docker containers, `hostPath` volumes map to paths **inside the kind container**, not your host. This is fine for learning — the data persists across Pod restarts but not across cluster teardown.

## What You'll Practice

1. Creating a static PV with hostPath
2. Creating a PVC and observing the binding
3. Mounting a PVC in a Pod and verifying data persistence
4. Recreating a Pod and seeing the data survive
5. Creating a StorageClass for dynamic provisioning
6. Creating a PVC without a pre-existing PV (dynamic provisioning)
7. Understanding reclaim policies

## Key Gotchas

- **PVs are cluster-scoped, PVCs are namespaced** — a PVC can only bind to PVs in the same... wait, PVs don't have namespaces! Any PVC can bind to any PV that matches.
- **One-to-one binding** — a PV can only bind to ONE PVC. No sharing a PV across multiple claims.
- **hostPath is NOT for production** — it ties data to a specific node. Use cloud volumes or network storage.
- **PVC stays Pending without a matching PV** — check `kubectl describe pvc` for events.
- **Retain doesn't auto-clean** — after deleting a PVC with Retain, the PV stays in "Released" state and can't be reused until manually cleaned.
- **kind's default StorageClass** — kind includes a `local-path` provisioner by default. You can use dynamic provisioning right away.
