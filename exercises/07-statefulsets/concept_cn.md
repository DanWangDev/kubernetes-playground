# 模块 07：StatefulSet

## 什么是 StatefulSet？

**StatefulSet** 管理有状态应用——数据库、消息队列、分布式缓存。与 Deployment（Pod 可互换）不同，StatefulSet Pod 具有**粘性标识**：

- **稳定的网络标识**：pod-0、pod-1、pod-2（非随机后缀）
- **稳定的存储**：每个 Pod 拥有自己的 PVC，在重新调度后仍然存在
- **有序部署**：0 先启动，然后是 1，然后是 2（顺序）
- **有序销毁**：N-1 先停止，然后是 N-2，以此类推（逆序）

## 核心概念

### 1. 稳定的网络标识

每个 StatefulSet Pod 获得可预测的 DNS 名称：

```
<pod-name>.<headless-service>.<namespace>.svc.cluster.local
```

### 2. 无头 Service

```yaml
spec:
  clusterIP: None          # 无头！
  selector:
    app: sts-demo
```

使用 `clusterIP: None`，DNS 返回单个 Pod IP 而非单个 Service IP。StatefulSet 必须使用无头 Service。

### 3. 有序部署和扩缩

- **扩容**：pod-0 → pod-1 → pod-2（每个就绪后才启动下一个）
- **缩容**：pod-2 → pod-1 → pod-0（逆序，每个终止后才停止下一个）

### 4. PVC 模板

```yaml
spec:
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 100Mi
```

每个 Pod 自动获得自己的 PVC。缩容时 PVC 不会被删除——存储比 Pod 持久。

### 5. StatefulSet vs Deployment

| | Deployment | StatefulSet |
|---|---|---|
| **Pod 标识** | 随机（nginx-7d5f8c9b6-abc12） | 可预测（redis-0, redis-1） |
| **启动顺序** | 并行 | 顺序（0, 1, 2...） |
| **存储** | 共享 PVC（如果有） | 每个 Pod 独立 PVC |
| **缩容** | 随机 Pod 被删除 | 最高序号先删除 |

## 你将练习的内容

1. 为 StatefulSet DNS 创建无头 Service
2. 部署 StatefulSet 并观察顺序 Pod 创建
3. 观察稳定的 Pod 名称
4. 缩容并观察逆序终止

## 关键注意事项

- **必须有无头 Service**——StatefulSet 需要它来进行网络标识
- **PVC 在 Pod 删除后仍然存在**——缩容不会删除 PVC，必须手动清理
- **有序操作较慢**——每个 Pod 必须就绪后下一个才启动
- **滚动更新是逆序的**——Pod N-1 先更新，然后是 N-2
