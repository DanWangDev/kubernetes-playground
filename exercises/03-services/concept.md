# Module 03: Services & Networking

## What Is a Service?

Pod IPs are **ephemeral**. When a Pod dies and a new one replaces it, the IP changes. A **Service** provides a stable virtual IP and DNS name that load-balances traffic across a set of Pods identified by label selectors.

Think of a Service as a load balancer with a fixed address. Clients connect to the Service, and the Service forwards traffic to healthy Pods — regardless of which specific Pods exist at any moment.

```
Client → Service (ClusterIP: 10.96.0.42) → Pod A (10.244.1.5)
                                          → Pod B (10.244.1.6)
                                          → Pod C (10.244.2.3)
```

## Core Concepts

### 1. Service Types

| Type | Scope | Use Case |
|------|-------|----------|
| **ClusterIP** (default) | Internal only | Backend communication between microservices |
| **NodePort** | External via `<NodeIP>:<NodePort>` | Development, demos, simple external access |
| **LoadBalancer** | External via cloud LB | Production ingress (AWS NLB, GCP LB) |
| **ExternalName** | DNS CNAME | Point to external services |

### 2. ClusterIP — Internal Load Balancing

The default and most common type. Creates a virtual IP reachable only from within the cluster:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-svc
spec:
  type: ClusterIP
  selector:
    app: nginx           # Routes to Pods with this label
  ports:
  - port: 80             # Service port
    targetPort: 80       # Container port on the Pod
```

This creates a stable DNS name: `nginx-svc.playground.svc.cluster.local`

### 3. DNS Resolution

Every Service gets a DNS A record. From any Pod in the cluster:

```
# Full name (works from any namespace)
curl nginx-svc.playground.svc.cluster.local

# Short name (works from same namespace)
curl nginx-svc
```

The DNS format is: `<service>.<namespace>.svc.cluster.local`

### 4. Endpoints

Services don't directly send traffic to Pods. They maintain an **Endpoints** (or **EndpointSlice**) object that lists the IPs of matching Pods:

```
kubectl get endpoints nginx-svc
```

When Pods are added or removed (by a Deployment scaling up/down), the Endpoints are automatically updated. kube-proxy on each node watches Endpoints and programs iptables rules to route traffic.

### 5. NodePort — Direct Node Access

NodePort builds on ClusterIP and additionally opens a port (30000-32767) on **every node**. Traffic to `<AnyNodeIP>:<NodePort>` reaches the Service:

```yaml
spec:
  type: NodePort
  ports:
  - port: 80
    targetPort: 80
    nodePort: 30080      # Optional; auto-assigned if omitted
```

NodePort is primarily for development and demos. For production, use Ingress or LoadBalancer.

### 6. Port Forwarding

`kubectl port-forward` creates a temporary tunnel from your local machine to a Pod or Service:

```
kubectl port-forward svc/nginx-svc 8080:80 -n playground
```

This is a development tool — not production. It's perfect for testing and debugging.

### 7. Headless Services

A Service with `clusterIP: None` doesn't get a virtual IP. DNS resolves directly to Pod IPs:

```yaml
spec:
  clusterIP: None
  selector:
    app: cassandra
```

Headless Services are the building block for StatefulSets (Module 07) — each Pod gets its own DNS name.

### 8. Multi-Port Services

A Service can expose multiple ports with names:

```yaml
spec:
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
```

Named ports make intent clear and enable port remapping without changing the container.

## What You'll Practice

1. Creating a ClusterIP Service and accessing it from inside the cluster
2. Using a debug pod to test internal DNS and connectivity
3. Creating a NodePort Service and accessing it from your host
4. Using `kubectl port-forward` for local development access
5. Understanding Endpoints and how they track Pod IPs
6. Testing multi-port Services

## Key Gotchas

- **Service selectors use labels, not names** — the Service finds Pods by matching labels. If labels don't match, the Service has no Endpoints and traffic fails silently.
- **NodePort range is 30000-32767** — you can't use port 80 as a NodePort. kind maps NodePorts to your host automatically.
- **ClusterIP is NOT a real network interface** — it's a virtual IP implemented by kube-proxy via iptables/IPVS. You can ping Pod IPs but you can't ping ClusterIPs.
- **port vs targetPort** — `port` is the Service's listening port; `targetPort` is the container's port. They can differ (e.g., Service port 80 → container port 8080).
- **DNS caching in apps** — some apps cache DNS lookups and won't see updated Endpoints. Use `targetPort` and headless services for such cases.
- **EndpointSlice vs Endpoints** — Kubernetes 1.21+ uses EndpointSlices (smaller objects, better scaling). Endpoints are still there for backward compatibility.
