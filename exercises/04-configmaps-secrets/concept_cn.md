# 模块 04：ConfigMap 与 Secret

## 什么是 ConfigMap 和 Secret？

**ConfigMap** 以键值对形式存储非敏感的配置数据。**Secret** 为敏感数据（密码、令牌、密钥）做同样的事。两者都将配置与容器镜像解耦——你无需重建镜像即可更改配置。

## 核心概念

### 1. 注入模式

三种在 Pod 中使用 ConfigMap/Secret 的方式：

| 模式 | Pod YAML | 最适合 |
|------|----------|--------|
| **环境变量（单个键）** | `valueFrom.configMapKeyRef` | 简单值 |
| **envFrom（所有键）** | `envFrom.configMapRef` | 导入整个配置 |
| **卷挂载（文件）** | `volumes[].configMap` | 配置文件、证书、复杂数据 |

### 2. 环境变量注入

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
```

### 3. envFrom — 批量导入

```yaml
spec:
  containers:
  - name: app
    envFrom:
    - configMapRef:
        name: app-config
```

这为 ConfigMap 中的每个键创建环境变量。注意键冲突——ConfigMap 键会覆盖之前的键。

### 4. 卷挂载（文件）

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

每个 ConfigMap 键成为 `/etc/config/` 中的一个文件。`LOG_LEVEL` → `/etc/config/LOG_LEVEL`。

### 5. Secret 是 Base64 编码，而非加密

**关键**：Kubernetes Secret 是 base64 编码的，而非加密的。任何拥有 `get secrets` RBAC 的人都可以解码它们：

```
kubectl get secret app-secret -o jsonpath='{.data.API_KEY}' | base64 -d
```

### 6. 不可变 ConfigMap/Secret

设置 `immutable: true` 可防止更改。要更新不可变配置，请删除并用新名称重新创建。

### 7. 配置更新与 Pod 重启

- 以卷形式挂载的 ConfigMap 会自动更新（最终，有 kubelet 同步延迟）
- 以**环境变量**注入的 ConfigMap 不会自动更新——Pod 必须重启
- 推荐模式：更改 ConfigMap 名称（如 `app-config-v2`）并更新 Deployment，从而触发滚动重启

### 8. Secret 类型

| 类型 | 用途 |
|------|------|
| **Opaque**（默认） | 任意键值对 |
| **kubernetes.io/tls** | TLS 证书 + 私钥 |
| **kubernetes.io/dockerconfigjson** | Docker 仓库凭据 |
| **kubernetes.io/basic-auth** | 用户名 + 密码 |

## 你将练习的内容

1. 从字面值和文件创建 ConfigMap
2. 将配置作为环境变量注入（单个和批量）
3. 创建 Opaque Secret 并挂载
4. 观察 Secret 是 base64 编码的（而非加密）
5. 使用不可变 ConfigMap
6. 将 ConfigMap 作为卷挂载（文件形式）

## 关键注意事项

- **Secret 不是加密的**——只是 base64 编码。通过 RBAC 和静态加密进行保护。
- **大小限制**：ConfigMap 和 Secret 有 1MB 限制。更大的数据请使用卷。
- **环境变量更新需要 Pod 重启**——更改 ConfigMap 不会更新运行中 Pod 的环境变量。
- **键命名**：ConfigMap/Secret 键必须是有效的 DNS 子域名。
- **不可变意味着真正不可变**——既不能编辑不可变 ConfigMap，也不能添加键。删除后重新创建。
