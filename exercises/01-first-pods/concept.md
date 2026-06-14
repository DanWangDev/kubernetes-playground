# Module 01: First Pods

## What Is a Pod?

In Kubernetes, a **Pod** is the smallest deployable unit. It's a group of one or more containers that share:

- **Network namespace** — same IP address and port space
- **IPC namespace** — can communicate via SystemV semaphores or POSIX shared memory
- **Storage volumes** — shared filesystem mounts

Think of a Pod as a logical host: it runs your application containers, just like a VM would,
but lighter and more focused. Most Pods run a single container, but multi-container Pods
are common when you need tight coupling (sidecars, proxies, log shippers).

## Core Concepts

### 1. Pod Lifecycle

Every Pod goes through a defined lifecycle:

```
Pending → Running → Succeeded (or Failed)
```

| Phase | Meaning |
|-------|---------|
| **Pending** | Pod accepted by the API server, but one or more containers not yet running. Includes image pulling and scheduling time. |
| **Running** | Pod bound to a node, all containers created, at least one still running. |
| **Succeeded** | All containers terminated successfully (exit code 0). Only for restartPolicy: Never or OnFailure. |
| **Failed** | All containers terminated, at least one with non-zero exit code. |
| **Unknown** | Pod state cannot be determined (usually node communication error). |

You can check a Pod's phase with:

```
kubectl get pod <name> -o jsonpath='{.status.phase}'
```

### 2. kubectl: Your Everyday Tool

`kubectl` is the CLI that talks to the Kubernetes API server. These are the commands you'll use most:

| Command | Purpose |
|---------|---------|
| `kubectl get pods` | List pods in the current namespace |
| `kubectl get pods -n <ns>` | List pods in a specific namespace |
| `kubectl get pods -o wide` | Show more detail (node, IP) |
| `kubectl describe pod <name>` | Full detail about a pod, including events |
| `kubectl logs <name>` | Stream stdout from the first container |
| `kubectl logs <name> -c <container>` | Stream stdout from a specific container |
| `kubectl exec <name> -- <command>` | Run a command inside the pod |
| `kubectl delete pod <name>` | Delete a pod |
| `kubectl apply -f <file>` | Create or update resources from a YAML file |

### 3. Namespaces

Namespaces partition cluster resources among multiple users or projects. They're like folders:

- Resources in different namespaces are isolated from each other
- Namespace-scoped resources include Pods, Deployments, Services, ConfigMaps, Secrets
- Cluster-scoped resources (Nodes, PVs, Namespaces themselves) don't belong to a namespace
- The default namespace is `default` — use namespaces to organize your work

```
kubectl get namespaces
kubectl create namespace playground
```

### 4. Pod YAML Structure

Every Pod manifest has four required top-level fields:

```yaml
apiVersion: v1          # API version of this resource type
kind: Pod               # Resource type
metadata:               # Identifying information
  name: my-pod
  labels:
    app: demo
spec:                   # Desired state
  containers:
  - name: nginx
    image: nginx:alpine
    ports:
    - containerPort: 80
```

### 5. Labels and Selectors

**Labels** are key-value pairs attached to Kubernetes objects. They're the glue that connects resources:

```yaml
labels:
  app: web
  environment: staging
  tier: frontend
```

Labels enable loose coupling — a Service finds its target Pods by label selector, not by pod name or IP. You'll use them constantly:

```
kubectl get pods -l app=web
kubectl get pods -l 'environment in (staging, production)'
```

### 6. Multi-Container Pods

A Pod can contain multiple containers that share fate — they're co-located on the same node,
share the same network namespace (reach each other at `localhost`), and can share volumes.
Common patterns:

- **Sidecar**: augments the main container (e.g., log shipper, proxy, config reloader)
- **Ambassador**: proxies traffic to/from the main container
- **Adapter**: normalizes output from the main container

```yaml
spec:
  containers:
  - name: app
    image: my-app:latest
  - name: sidecar
    image: busybox
    command: ["tail", "-f", "/var/log/app.log"]
```

## What You'll Practice

1. Creating a namespace to organize your work
2. Launching a simple nginx pod and inspecting its lifecycle
3. Adding labels to pods and querying by label selector
4. Executing commands inside running containers
5. Checking pod logs and event history
6. Deploying a multi-container pod (app + sidecar pattern)
7. Cleaning up resources

## Key Gotchas

- **Pods are immutable** — you can't change a pod's spec after creation. Edit the YAML, delete the pod, and recreate it. (Deployments solve this in Module 02.)
- **Pods are ephemeral** — when a pod dies, it's gone forever (including its IP, local storage, and logs, unless you have centralized logging). Don't depend on pod IPs.
- **Get vs Describe** — `kubectl get` gives a summary. `kubectl describe` gives the full story, including the events log (scheduling decisions, image pull progress, probe results).
- **Namespaces don't auto-create** — you must create a namespace before you can use it. Applying a YAML that references a nonexistent namespace fails.
- **Image pull takes time** — the first time you run an image on a node, Kubernetes must download it. Status shows `ErrImagePull` or `ImagePullBackOff` if the image is unavailable.
- **localhost works across containers** — in a multi-container pod, containers reach each other at `localhost:<port>`. This is because they share a network namespace.
