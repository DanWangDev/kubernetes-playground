# 模块 02：Deployment 与 ReplicaSet

## 什么是 Deployment？

**Deployment** 管理一组相同的 Pod。你不是逐个创建 Pod，而是声明*需要多少个副本*以及 Pod 模板的模样。Kubernetes 随后持续工作，使实际状态匹配你的声明。

Deployment 解决了原始 Pod 的根本问题：Pod 会死亡（节点故障、驱逐、更新），需要有人重新创建它们。一个 Deployment 拥有一个 **ReplicaSet**，而 ReplicaSet 又拥有 Pod。

```
Deployment → ReplicaSet → Pod（× N 副本）
```

## 核心概念

### 1. ReplicaSet：Pod 计数器

**ReplicaSet** 保证指定数量的相同 Pod 始终运行：

- 如果 Pod 死亡，ReplicaSet 创建替代品
- 如果 Pod 过多，ReplicaSet 终止多余的
- ReplicaSet 使用**标签选择器**跟踪"它的"Pod

你很少直接创建 ReplicaSet——Deployment 为你创建和管理它们。

### 2. 声明式 vs 命令式

| | 命令式 | 声明式 |
|---|---|---|
| **命令** | `kubectl run nginx --image=nginx` | `kubectl apply -f deployment.yaml` |
| **思维模型** | "做这个，然后做那个" | "这是期望状态，让它实现" |
| **幂等** | 否 | 是（运行 `apply` 100 次，结果相同） |
| **Git 友好** | 否（没有文件可提交） | 是（YAML 已提交） |

Kubernetes 的核心是**声明式**系统。你在 YAML 中声明期望状态，控制器持续将现实调和到该状态。

### 3. Deployment YAML 结构

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3                    # 多少个 Pod？
  selector:                      # 如何找到我的 Pod？
    matchLabels:
      app: nginx
  template:                      # 要创建的 Pod
    metadata:
      labels:
        app: nginx               # 必须匹配 selector
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
```

关键观察：`template` 就是 Pod spec。`selector.matchLabels` 必须匹配 `template.metadata.labels`。

### 4. 滚动更新（默认策略）

当你更改 Pod 模板（例如更新镜像标签），Deployment 执行**滚动更新**：

```
步骤 1：创建新的 ReplicaSet（v2），0 个 Pod
步骤 2：将新 ReplicaSet 扩容 1，旧 ReplicaSet 缩容 1
步骤 3：重复直到旧 ReplicaSet 为 0，新 ReplicaSet 拥有全部副本
```

滚动更新由两个参数控制：

| 参数 | 含义 | 默认值 |
|------|------|--------|
| `maxSurge` | 更新期间超过 `replicas` 的最大额外 Pod 数 | 25% |
| `maxUnavailable` | 更新期间不可用的最大 Pod 数 | 25% |

对于 3 个副本，`maxSurge=1` 和 `maxUnavailable=1`：
- 峰值：4 个 Pod 运行（3 期望 + 1 超额）
- 谷值：2 个 Pod 可用（3 期望 - 1 不可用）
- 永不超过 4，永不低过 2

### 5. 发布历史

每次更改 Pod 模板都会创建新的**版本**：

```
kubectl rollout history deployment/nginx-deployment
```

旧的 ReplicaSet 被保留（缩容到 0）以实现快速回滚：

```
kubectl rollout undo deployment/nginx-deployment
kubectl rollout undo deployment/nginx-deployment --to-revision=2
```

`rollout undo` 只是将前一个 ReplicaSet 扩容回来，将当前 ReplicaSet 缩容。这只需几秒，而非几分钟。

### 6. Recreate 策略

对于同时运行新旧版本会导致问题的工作负载（例如数据库模式迁移），有 **Recreate** 策略：

```yaml
spec:
  strategy:
    type: Recreate
```

使用 Recreate 时，所有旧 Pod 先终止，然后任何新 Pod 才创建。这会造成停机，但避免了版本冲突。

### 7. Pod 模板哈希

每个 ReplicaSet 获得一个 `pod-template-hash` 标签（如 `pod-template-hash: 7d5f8c9b6`）。只要 Pod 模板发生变化，哈希就会改变，触发新的 ReplicaSet 和滚动更新。

你可以观察这一过程：

```
kubectl get replicasets -l app=nginx
```

## 你将练习的内容

1. 创建 Deployment 并观察 ReplicaSet 和 Pod 的出现
2. 使用 `kubectl scale` 和编辑 YAML 进行扩缩容
3. 通过更改镜像标签触发滚动更新
4. 实时观察滚动更新
5. 回滚到之前的版本
6. 在更新中途暂停和恢复发布
7. 理解 Deployment → ReplicaSet → Pod 的所有权链

## 关键注意事项

- **不要直接删除 ReplicaSet**——Deployment 会重新创建它们。删除 Deployment 来清理。
- **不要编辑 Deployment 拥有的 Pod**——ReplicaSet 会用模板版本替换它们。改为编辑 Deployment 的模板。
- **Selector 必须匹配模板标签**——如果不匹配，`kubectl apply` 会以验证错误拒绝清单。
- **发布历史有限制**——默认只保留最后 10 个版本（`revisionHistoryLimit`）。旧版本的 ReplicaSet 会被删除。
- **扩缩容 Deployment 不会创建新版本**——只有 Pod 模板的更改才会触发新的 ReplicaSet。运行 `kubectl scale` 只是更改当前 ReplicaSet 的副本数量。
- **`kubectl apply` vs `kubectl create`**——`apply` 是声明式的（创建或更新）；`create` 是命令式的（如果资源存在则失败）。始终优先使用 `apply`。
