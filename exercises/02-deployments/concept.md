# Module 02: Deployments & ReplicaSets

## What Are Deployments?

A **Deployment** manages a set of identical Pods. Instead of creating Pods one-by-one, you declare *how many replicas you want* and what the Pod template looks like. Kubernetes then works continuously to make reality match your declaration.

Deployments solve the fundamental problem with raw Pods: Pods die (node failures, evictions, updates), and someone needs to recreate them. A Deployment owns a **ReplicaSet**, which in turn owns the Pods.

```
Deployment → ReplicaSet → Pod (× N replicas)
```

## Core Concepts

### 1. ReplicaSet: The Pod Counter

A **ReplicaSet** guarantees that a specified number of identical Pods are running at all times:

- If a Pod dies, the ReplicaSet creates a replacement
- If too many Pods exist, the ReplicaSet terminates the extras
- The ReplicaSet uses a **label selector** to track "its" Pods

You rarely create ReplicaSets directly — the Deployment creates and manages them for you.

### 2. Declarative vs Imperative

| | Imperative | Declarative |
|---|---|---|
| **Command** | `kubectl run nginx --image=nginx` | `kubectl apply -f deployment.yaml` |
| **Mental model** | "Do this, then that" | "Here's the desired state. Make it so." |
| **Idempotent** | No | Yes (run `apply` 100 times, same result) |
| **Git-friendly** | No (no file to commit) | Yes (YAML is committed) |

Kubernetes is a **declarative** system at its core. You declare desired state in YAML, and controllers continuously reconcile reality toward that state.

### 3. Deployment YAML Structure

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3                    # How many pods?
  selector:                      # How to find my pods?
    matchLabels:
      app: nginx
  template:                      # Pod to create
    metadata:
      labels:
        app: nginx               # MUST match selector
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
```

Key observation: the `template` is just a Pod spec. The `selector.matchLabels` must match the `template.metadata.labels`.

### 4. Rolling Update (Default Strategy)

When you change the Pod template (e.g., update the image tag), the Deployment performs a **rolling update**:

```
Step 1: Create new ReplicaSet (v2) with 0 pods
Step 2: Scale up new ReplicaSet by 1, scale down old by 1
Step 3: Repeat until old ReplicaSet has 0 pods, new has all replicas
```

The rolling update is controlled by two parameters:

| Parameter | Meaning | Default |
|-----------|---------|---------|
| `maxSurge` | Maximum extra pods above `replicas` during update | 25% |
| `maxUnavailable` | Maximum pods that can be unavailable during update | 25% |

With 3 replicas, `maxSurge=1` and `maxUnavailable=1`:
- At peak: 4 pods running (3 desired + 1 surge)
- At trough: 2 pods available (3 desired - 1 unavailable)
- Never more than 4, never fewer than 2

### 5. Rollout History

Every change to the Pod template creates a new **revision**:

```
kubectl rollout history deployment/nginx-deployment
```

The old ReplicaSet is kept (scaled to 0) to enable fast rollbacks:

```
kubectl rollout undo deployment/nginx-deployment
kubectl rollout undo deployment/nginx-deployment --to-revision=2
```

`rollout undo` simply scales the previous ReplicaSet back up and the current one down. This takes seconds, not minutes.

### 6. Recreate Strategy

For workloads where running old and new versions simultaneously would cause problems (e.g., database schema migrations), there's the **Recreate** strategy:

```yaml
spec:
  strategy:
    type: Recreate
```

With Recreate, ALL old pods are terminated before ANY new pods are created. This causes downtime but avoids version conflicts.

### 7. Pod Template Hashing

Every ReplicaSet gets a `pod-template-hash` label (e.g., `pod-template-hash: 7d5f8c9b6`). This hash changes whenever the Pod template changes, which triggers a new ReplicaSet and a rolling update.

You can see this in action:

```
kubectl get replicasets -l app=nginx
```

## What You'll Practice

1. Creating a Deployment and watching the ReplicaSet and Pods appear
2. Scaling up/down with `kubectl scale` and by editing the YAML
3. Triggering a rolling update by changing the image tag
4. Watching the rolling update in real time
5. Rolling back to a previous revision
6. Pausing and resuming a rollout mid-update
7. Understanding the Deployment → ReplicaSet → Pod ownership chain

## Key Gotchas

- **Don't delete ReplicaSets directly** — the Deployment will recreate them. Delete the Deployment to clean up.
- **Don't edit Pods owned by a Deployment** — the ReplicaSet will replace them with the template version. Edit the Deployment's template instead.
- **Selector must match template labels** — if they don't match, `kubectl apply` rejects the manifest with a validation error.
- **Rollout history has limits** — by default, only the last 10 revisions are kept (`revisionHistoryLimit`). Older revisions' ReplicaSets are deleted.
- **Scaling a Deployment doesn't create a new revision** — only changes to the Pod template trigger a new ReplicaSet. Running `kubectl scale` just changes the replica count on the current ReplicaSet.
- **`kubectl apply` vs `kubectl create`** — `apply` is declarative (create-or-update); `create` is imperative (fails if resource exists). Always prefer `apply`.
