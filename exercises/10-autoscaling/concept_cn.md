# 模块 10：自动扩缩（HPA）

## 什么是 HPA？

**Horizontal Pod Autoscaler（HPA）** 基于观察到的指标（CPU、内存或自定义指标）自动扩缩 Deployment 或 StatefulSet 中的 Pod 数量。

```
HPA 观察指标 → 计算期望副本数 → 扩缩 Deployment
```

## 核心概念

### 1. HPA 算法

```
期望副本数 = ceil（当前副本数 × 当前指标值 / 期望指标值）
```

示例：1 个副本使用 80% CPU，目标 50%：
```
期望 = ceil(1 × 80 / 50) = ceil(1.6) = 2 个副本
```

### 2. HPA YAML

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
spec:
  scaleTargetRef:
    kind: Deployment
    name: app-deployment
  minReplicas: 1
  maxReplicas: 5
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        averageUtilization: 50
```

### 3. 要求

HPA 需要：
1. 集群中安装 **Metrics Server**
2. 目标 Pod 设置**资源请求**（HPA 需要基线）

### 4. Metrics Server

```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
kubectl top nodes
```

### 5. 稳定与缩容保护

```yaml
behavior:
  scaleDown:
    stabilizationWindowSeconds: 300  # 缩容前等待 5 分钟
```

### 6. HPA 状态

```
kubectl get hpa app-hpa
```

显示当前与目标利用率及当前副本数。默认每 15 秒更新一次。

## 关键注意事项

- **必须安装 Metrics Server**——默认不包含
- **Pod 需要资源请求**——否则 HPA 没有百分比计算的基线
- **HPA 需要时间**——指标每 15 秒收集一次
- **缩容有意较慢**——防止抖动。扩容较快
- **HPA 不适用于 DaemonSet**
