# Module 04: ConfigMaps & Secrets

## What Are ConfigMaps and Secrets?

**ConfigMaps** store non-sensitive configuration data as key-value pairs. **Secrets** do the same for sensitive data (passwords, tokens, keys). Both decouple configuration from container images — you change config without rebuilding images.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  LOG_LEVEL: debug
  API_URL: https://api.example.com
---
apiVersion: v1
kind: Secret
metadata:
  name: app-secret
type: Opaque
data:
  API_KEY: c2VjcmV0LWtleQ==   # base64("secret-key")
```

## Core Concepts

### 1. Injection Patterns

Three ways to consume ConfigMaps/Secrets in a Pod:

| Pattern | Pod YAML | Best For |
|---------|----------|----------|
| **env var (single key)** | `valueFrom.configMapKeyRef` | Simple values |
| **envFrom (all keys)** | `envFrom.configMapRef` | Importing entire config |
| **Volume mount (files)** | `volumes[].configMap` | Config files, certs, complex data |

### 2. Environment Variable Injection

```yaml
spec:
  containers:
  - name: app
    env:
    - name: LOG_LEVEL
      valueFrom:
        configMapKeyRef:
          name: app-config
          key: LOG_LEVEL
    - name: API_KEY
      valueFrom:
        secretKeyRef:
          name: app-secret
          key: API_KEY
```

### 3. envFrom — Bulk Import

```yaml
spec:
  containers:
  - name: app
    envFrom:
    - configMapRef:
        name: app-config
    - secretRef:
        name: app-secret
```

This creates environment variables for EVERY key in the ConfigMap/Secret. Be careful about key collisions (ConfigMap keys override earlier ones; Secrets don't override ConfigMaps of the same key — the pod fails to start).

### 4. Volume Mount (Files)

```yaml
spec:
  containers:
  - name: app
    volumeMounts:
    - name: config
      mountPath: /etc/config
      readOnly: true
  volumes:
  - name: config
    configMap:
      name: app-config
```

Each ConfigMap key becomes a file in `/etc/config/`. `LOG_LEVEL` → `/etc/config/LOG_LEVEL`. This is great for config files that need to be read by the application at startup.

### 5. Secrets Are Base64, Not Encrypted

**Critical**: Kubernetes Secrets are base64-encoded, NOT encrypted. Anyone with `get secrets` RBAC can decode them:

```
kubectl get secret app-secret -o jsonpath='{.data.API_KEY}' | base64 -d
```

In production, layer additional protection:
- **Encryption at rest**: configure etcd encryption
- **External secret stores**: HashiCorp Vault, AWS Secrets Manager, Sealed Secrets
- **RBAC**: restrict Secret access to only the pods that need them

### 6. Immutable ConfigMaps/Secrets

Set `immutable: true` to prevent changes. This protects against accidental edits:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
immutable: true
data:
  LOG_LEVEL: debug
```

To update an immutable config, delete it and recreate with a new name.

### 7. Config Updates and Pod Restarts

ConfigMaps mounted as volumes auto-update in the pod (eventually, with a kubelet sync delay). But:
- ConfigMaps injected as **environment variables** do NOT auto-update — the pod must restart
- Even with volume mounts, the app must reread the file (many apps don't)
- The recommended pattern: change the ConfigMap name (e.g., `app-config-v2`) and update the Deployment to reference it, which triggers a rolling restart

### 8. Secret Types

| Type | Use |
|------|-----|
| **Opaque** (default) | Arbitrary key-value pairs |
| **kubernetes.io/tls** | TLS certificate + private key |
| **kubernetes.io/dockerconfigjson** | Docker registry credentials |
| **kubernetes.io/basic-auth** | Username + password |
| **kubernetes.io/ssh-auth** | SSH private key |

## What You'll Practice

1. Creating a ConfigMap from literal values and a file
2. Injecting config as environment variables (single and bulk)
3. Creating an Opaque Secret and mounting it
4. Observing that Secrets are base64-encoded (NOT encrypted)
5. Using immutable ConfigMaps
6. Mounting a ConfigMap as a volume (files)

## Key Gotchas

- **Secrets are NOT encrypted** — they're base64-encoded. Protect with RBAC and encryption at rest.
- **Size limit**: ConfigMaps and Secrets have a 1MB limit. For larger data, use volumes.
- **env var updates require pod restart** — changing a ConfigMap does NOT update env vars in running pods.
- **Key naming**: ConfigMap/Secret keys must be valid DNS subdomain names. Use underscores in env var names, hyphens for file names.
- **Immutable means truly immutable** — you can't edit an immutable ConfigMap, not even to add a key. Delete and recreate.
- **Secret data must be base64-encoded** in the YAML. Use `echo -n "value" | base64` to create the encoded value.
