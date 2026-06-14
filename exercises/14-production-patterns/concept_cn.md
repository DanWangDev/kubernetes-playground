# 模块 14：生产模式

## 什么是生产模式？

这些是在生产 Kubernetes 集群中使用的调度、可用性和弹性模式。它们控制 Pod 运行的位置、如何分布以及中断期间会发生什么。

## 核心概念

### 1. Pod 亲和性 — 将相关工作负载放在一起

```yaml
affinity:
  podAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          app: cache
      topologyKey: kubernetes.io/hostname
```

### 2. Pod 反亲和性 — 分散 Pod

```yaml
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          app: web
      topologyKey: kubernetes.io/hostname
```

### 3. 节点选择器 / 节点亲和性

```yaml
nodeSelector:
  disktype: ssd
```

### 4. 污点与容忍（Taints & Tolerations）

污点排斥 Pod，容忍允许 Pod 被调度到有污点的节点：

```bash
kubectl taint nodes worker-1 gpu=true:NoSchedule
```

### 5. Pod 中断预算（PDB）

```yaml
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: critical-app
```

### 6. 拓扑分散约束

```yaml
topologySpreadConstraints:
- maxSkew: 1
  topologyKey: topology.kubernetes.io/zone
  whenUnsatisfiable: DoNotSchedule
```

### 7. PriorityClass

更高优先级的 Pod 可以抢占低优先级的 Pod。

## 关键注意事项

- **反亲和性可能阻止调度**——如果 Pod 数量超过节点数量，Pod 保持 Pending
- **PDB 仅涵盖自愿中断**——节点崩溃不包括在内
- **污点 + 容忍 ≠ 认证**——它们是调度提示，不是安全边界
- **多节点集群**是节点级亲和性和拓扑分散有意义的前提
