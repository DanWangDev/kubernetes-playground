# Module 05: Ingress & HTTP Routing

## What Is Ingress?

**Ingress** provides Layer 7 (HTTP/HTTPS) routing to Services. Unlike Services (which are L4), Ingress understands HTTP — it can route by hostname, path, headers, and terminate TLS.

Ingress is two things:
1. **Ingress Controller** — a pod that does the actual routing (nginx, traefik, haproxy, etc.)
2. **Ingress Resource** — YAML that defines routing rules

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
spec:
  rules:
  - host: app.example.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: api-service
            port:
              number: 80
```

## Core Concepts

### 1. L4 vs L7

| | Service (L4) | Ingress (L7) |
|---|---|---|
| **OSI Layer** | Transport (TCP/UDP) | Application (HTTP) |
| **Routing** | By IP:port | By hostname + path |
| **TLS** | Passthrough | Termination + re-encryption |
| **Features** | Load balancing | Routing, TLS, rate limiting, rewrites, auth |

### 2. Ingress Controller Installation (kind)

Kubernetes doesn't include an Ingress controller by default. For kind:

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

This deploys the NGINX Ingress Controller as a DaemonSet/Deployment in `ingress-nginx` namespace, listening on ports 80 and 443 of the kind node.

### 3. Path-Based Routing

Route requests to different backends based on URL path:

```yaml
spec:
  rules:
  - http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: api-service
            port:
              number: 80
      - path: /
        pathType: Prefix
        backend:
          service:
            name: web-service
            port:
              number: 80
```

pathType options:
- **Prefix**: matches path prefix (`/api`, `/api/v1`, `/api/v1/users`)
- **Exact**: matches exact path (`/api` only, not `/api/`)
- **ImplementationSpecific**: controller-specific matching

### 4. Host-Based Routing

Route to different backends by Host header:

```yaml
spec:
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-service
            port:
              number: 80
  - host: web.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: web-service
            port:
              number: 80
```

### 5. TLS Termination

The Ingress controller terminates TLS and forwards plain HTTP to the backend:

```yaml
spec:
  tls:
  - hosts:
    - app.example.com
    secretName: app-tls-secret
```

The Secret must be type `kubernetes.io/tls` with `tls.crt` and `tls.key` keys:

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout tls.key -out tls.crt \
  -subj "/CN=app.example.com"
kubectl create secret tls app-tls-secret --cert=tls.crt --key=tls.key
```

## What You'll Practice

1. Installing the NGINX Ingress Controller on kind
2. Deploying two backend applications (echo API + nginx web)
3. Creating path-based Ingress rules
4. Testing host-based routing with curl -H "Host: ..."
5. Creating a self-signed TLS certificate and enabling HTTPS
6. Understanding Ingress annotations

## Key Gotchas

- **Ingress controllers are NOT built-in** — you MUST install one. Different cloud providers have different default controllers.
- **Ingress and Service work at different layers** — Ingress routes HTTP to Services; Services route TCP to Pods.
- **defaultBackend** — if no rules match, the default backend handles the request. Often returns 404.
- **TLS Secret must be in the same namespace** as the Ingress resource.
- **Host field is optional** — an Ingress without hosts matches all incoming traffic (wildcard).
- **Path order matters** — `/` catches everything. Put more specific paths first.
