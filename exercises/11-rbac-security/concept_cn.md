# 模块 11：RBAC 与安全

## 什么是 RBAC？

**基于角色的访问控制（RBAC）** 控制谁能在 Kubernetes 中做什么：

- **ServiceAccount** — Pod 的身份标识
- **Role** — 命名空间内的权限集
- **ClusterRole** — 集群范围的权限集
- **RoleBinding** — 将 Role 授予 ServiceAccount/User/Group
- **ClusterRoleBinding** — 集群范围授权

## 核心概念

### 1. ServiceAccount

每个命名空间都有 `default` ServiceAccount。Pod 默认使用它，除非指定自定义 SA。

### 2. Role

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: pod-reader
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch"]
```

### 3. RoleBinding

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
```

### 4. SecurityContext

```yaml
securityContext:
  runAsNonRoot: true
  capabilities:
    drop: ["ALL"]
  readOnlyRootFilesystem: true
```

### 5. NetworkPolicy

Pod 防火墙——控制入口/出口流量。需要支持 NetworkPolicy 的 CNI（kind 默认不支持）。

## 你将练习的内容

1. 创建自定义 ServiceAccount
2. 定义具有特定权限的 Role
3. 将 Role 绑定到 ServiceAccount
4. 使用 `kubectl auth can-i` 测试权限
5. 应用限制性 SecurityContext

## 关键注意事项

- **默认 SA 权限有限**——Pod 默认不能列出其他 Pod
- **RBAC 是累加的**——没有"拒绝"规则。只能授予，不能撤销
- **NetworkPolicy 需要 CNI 支持**——kind 默认不支持
- **SecurityContext 不是安全边界**——不要依赖它进行多租户隔离
