# 模块 06：存储（PV/PVC）

## 什么是 PersistentVolume？

Kubernetes 将存储分为两个抽象：

- **PersistentVolume（PV）** — 集群中的一块存储（由管理员配置或动态创建）
- **PersistentVolumeClaim（PVC）** — 用户对存储的请求（类似 Pod 请求 CPU/内存）

这种分离意味着开发者不需要了解底层存储基础设施。他们只需声明所需，Kubernetes 会将他们的声明绑定到可用存储。

```
管理员创建：PV（10Gi，RWO，SSD）
用户请求：PVC（5Gi，RWO）→ 绑定到 PV
Pod 使用：PVC 作为卷
```

## 核心概念

### 1. PV 和 PVC 绑定

PVC 绑定到满足其要求的**最小** PV。如果没有匹配的 PV，PVC 保持 Pending 状态。

### 2. 访问模式

| 模式 | 缩写 | 含义 |
|------|------|------|
| **ReadWriteOnce** | RWO | 一个节点可读写挂载 |
| **ReadOnlyMany** | ROM | 多个节点可只读挂载 |
| **ReadWriteMany** | RWX | 多个节点可读写挂载 |

### 3. 静态 vs 动态供应

| | 静态 | 动态 |
|---|---|---|
| **PV 创建者** | 管理员手动 | StorageClass 自动 |
| **创建时机** | PVC 之前 | PVC 创建时 |
| **回收策略** | 手动清理 | 自动（Delete）或手动（Retain） |

### 4. StorageClass 与动态供应

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ssd
provisioner: kubernetes.io/aws-ebs
reclaimPolicy: Delete
```

当 PVC 引用此 StorageClass 时，会自动创建 PV（如果供应器支持）。

### 5. 回收策略

| 策略 | 行为 |
|------|------|
| **Retain** | PVC 删除时 PV 不删除，需手动清理 |
| **Delete**（默认） | PVC 删除时 PV 和相关存储被删除 |

### 6. 卷模式

- **Filesystem**（默认）：作为目录挂载
- **Block**：作为原始块设备挂载

## 你将练习的内容

1. 使用 hostPath 创建静态 PV
2. 创建 PVC 并观察绑定
3. 在 Pod 中挂载 PVC 并验证数据持久性
4. 重建 Pod 并看到数据保留
5. 为动态供应创建 StorageClass
6. 在没有预先存在的 PV 的情况下创建 PVC（动态供应）
7. 理解回收策略

## 关键注意事项

- **PV 是集群范围的，PVC 是有命名空间的**
- **一对一绑定** — 一个 PV 只能绑定到一个 PVC
- **hostPath 不用于生产** — 它将数据绑定到特定节点
- **没有匹配 PV 时 PVC 保持 Pending** — 检查 `kubectl describe pvc`
- **Retain 不会自动清理** — 删除 Retain 策略的 PVC 后，PV 保持 "Released" 状态
