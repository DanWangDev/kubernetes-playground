# Module 11: RBAC & Security

## What Is RBAC?

**Role-Based Access Control (RBAC)** controls who can do what in Kubernetes. It's the primary authorization mechanism:

- **ServiceAccount** — identity for pods (instead of users, which are for humans)
- **Role** — set of permissions within a namespace
- **ClusterRole** — set of permissions cluster-wide
- **RoleBinding** — grants a Role to a ServiceAccount/User/Group in a namespace
- **ClusterRoleBinding** — grants a ClusterRole cluster-wide

```
ServiceAccount → RoleBinding → Role → Permissions (get pods, create deployments, etc.)
```

## Core Concepts

### 1. ServiceAccount

Every namespace has a `default` ServiceAccount. Pods use it unless you specify a custom one:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: app-sa
  namespace: playground
```

### 2. Role

A Role defines permissions within a namespace:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: pod-reader
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["pods/log"]
  verbs: ["get"]
```

### 3. RoleBinding

Binds a Role to a ServiceAccount:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: read-pods
subjects:
- kind: ServiceAccount
  name: app-sa
roleRef:
  kind: Role
  name: pod-reader
  apiGroup: rbac.authorization.k8s.io
```

### 4. SecurityContext

Pod and container-level security settings:

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  capabilities:
    drop: ["ALL"]
  readOnlyRootFilesystem: true
```

### 5. NetworkPolicy

A firewall for pods — controls ingress/egress traffic:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
```

NetworkPolicies require a CNI that supports them (Calico, Cilium, Weave). kind's default CNI (kindnet) does NOT support NetworkPolicy.

## What You'll Practice

1. Creating a custom ServiceAccount
2. Defining a Role with specific permissions
3. Binding the Role to the ServiceAccount
4. Testing permissions with `kubectl auth can-i`
5. Applying restrictive SecurityContext
6. Understanding NetworkPolicy (conceptual on kind)

## Key Gotchas

- **Default SA has limited permissions** — pods can't list other pods by default.
- **RBAC is additive** — there are no "deny" rules. You can only grant, not revoke.
- **NetworkPolicy requires a CNI** — kind's default CNI doesn't support it. Use Calico for testing.
- **SecurityContext is NOT a security boundary** — it's a best-effort mechanism. Don't rely on it for multi-tenant isolation.
- **ClusterRoleBindings are dangerous** — they grant cluster-wide access. Be very careful.
